package drone

import (
	"encoding/binary"
	"encoding/json"
	"testing"
	"time"
)

func TestProcessMAVLinkGlobalPositionInt(t *testing.T) {
	ti, err := NewTelemetryIngest(TelemetryIngestConfig{Protocol: ProtocolMAVLink})
	if err != nil {
		t.Fatalf("new ingest: %v", err)
	}

	payload := make([]byte, 28)
	lon := int32(-1220845678)
	binary.LittleEndian.PutUint32(payload[4:8], uint32(int32(374221234))) // lat 37.4221234
	binary.LittleEndian.PutUint32(payload[8:12], uint32(lon))             // lon -122.0845678
	binary.LittleEndian.PutUint32(payload[12:16], uint32(int32(25000)))   // alt mm = 25m
	binary.LittleEndian.PutUint16(payload[26:28], uint16(9000))           // heading centideg = 90deg
	binary.LittleEndian.PutUint16(payload[20:22], uint16(int16(300)))     // vx cm/s = 3m/s
	binary.LittleEndian.PutUint16(payload[22:24], uint16(int16(400)))     // vy cm/s = 4m/s

	pkt := make([]byte, 6+len(payload)+2)
	pkt[0] = 0xFE
	pkt[1] = byte(len(payload))
	pkt[2] = 1  // seq
	pkt[3] = 42 // sys id
	pkt[4] = 1  // comp id
	pkt[5] = 33 // GLOBAL_POSITION_INT
	copy(pkt[6:], payload)
	// checksum bytes left zero for parser len scan

	if err := ti.processMAVLink(pkt); err != nil {
		t.Fatalf("process mavlink: %v", err)
	}

	telem, ok := <-ti.telemetry
	if !ok {
		t.Fatalf("expected telemetry sample")
	}

	if telem.DroneID != "sys-42" {
		t.Fatalf("unexpected drone id: %s", telem.DroneID)
	}
	if telem.AltitudeMeters != 25 {
		t.Fatalf("expected altitude 25m, got %f", telem.AltitudeMeters)
	}
	if telem.HeadingDegrees != 90 {
		t.Fatalf("expected heading 90, got %f", telem.HeadingDegrees)
	}
	if telem.GroundSpeedMS < 4.9 || telem.GroundSpeedMS > 5.1 {
		t.Fatalf("expected groundspeed around 5.0 m/s, got %f", telem.GroundSpeedMS)
	}
	if telem.Source != string(ProtocolMAVLink) {
		t.Fatalf("unexpected source: %s", telem.Source)
	}
	if telem.SensorHealth != "healthy" {
		t.Fatalf("unexpected sensor health: %s", telem.SensorHealth)
	}
	if telem.Confidence < 0.5 || telem.Confidence > 1 {
		t.Fatalf("expected normalized confidence in [0.5,1], got %f", telem.Confidence)
	}
}

func TestProcessJSONAcceptsValidContractPayload(t *testing.T) {
	ti, err := NewTelemetryIngest(TelemetryIngestConfig{Protocol: ProtocolJSON})
	if err != nil {
		t.Fatalf("new ingest: %v", err)
	}

	payload := DroneTelemetry{
		ContractVersion: droneTelemetryContractVersionV1,
		DroneID:         "drone-01",
		Timestamp:       time.Now().UTC(),
		Latitude:        37.4219,
		Longitude:       -122.0840,
		AltitudeMeters:  25,
	}
	raw, _ := json.Marshal(payload)

	if err := ti.processJSON(raw); err != nil {
		t.Fatalf("process json: %v", err)
	}
	telem, ok := <-ti.telemetry
	if !ok {
		t.Fatal("expected telemetry event")
	}
	if telem.Source != string(ProtocolJSON) {
		t.Fatalf("unexpected source: %s", telem.Source)
	}
	if telem.SensorHealth != "healthy" {
		t.Fatalf("unexpected sensor health: %s", telem.SensorHealth)
	}
	if telem.Confidence <= 0 {
		t.Fatalf("expected positive confidence, got %f", telem.Confidence)
	}
}

func TestProcessJSONRejectsInvalidContractVersion(t *testing.T) {
	ti, err := NewTelemetryIngest(TelemetryIngestConfig{Protocol: ProtocolJSON})
	if err != nil {
		t.Fatalf("new ingest: %v", err)
	}

	payload := DroneTelemetry{
		ContractVersion: "legacy-v0",
		DroneID:         "drone-01",
		Timestamp:       time.Now().UTC(),
		Latitude:        37.4219,
		Longitude:       -122.0840,
	}
	raw, _ := json.Marshal(payload)

	if err := ti.processJSON(raw); err == nil {
		t.Fatalf("expected invalid contract version error")
	}
}

func TestProcessJSONRejectsStaleTimestamp(t *testing.T) {
	ti, err := NewTelemetryIngest(TelemetryIngestConfig{Protocol: ProtocolJSON, MaxSampleAge: 5 * time.Second})
	if err != nil {
		t.Fatalf("new ingest: %v", err)
	}

	payload := DroneTelemetry{
		ContractVersion: droneTelemetryContractVersionV1,
		DroneID:         "drone-01",
		Timestamp:       time.Now().UTC().Add(-1 * time.Minute),
		Latitude:        37.4219,
		Longitude:       -122.0840,
	}
	raw, _ := json.Marshal(payload)

	if err := ti.processJSON(raw); err == nil {
		t.Fatalf("expected stale timestamp error")
	}
}

func TestProcessJSONRejectsFutureTimestamp(t *testing.T) {
	ti, err := NewTelemetryIngest(TelemetryIngestConfig{Protocol: ProtocolJSON, MaxClockSkew: 1 * time.Second})
	if err != nil {
		t.Fatalf("new ingest: %v", err)
	}

	payload := DroneTelemetry{
		ContractVersion: droneTelemetryContractVersionV1,
		DroneID:         "drone-01",
		Timestamp:       time.Now().UTC().Add(10 * time.Second),
		Latitude:        37.4219,
		Longitude:       -122.0840,
	}
	raw, _ := json.Marshal(payload)

	if err := ti.processJSON(raw); err == nil {
		t.Fatalf("expected future timestamp error")
	}
}

func TestProcessJSONRejectsInvalidCoordinates(t *testing.T) {
	ti, err := NewTelemetryIngest(TelemetryIngestConfig{Protocol: ProtocolJSON})
	if err != nil {
		t.Fatalf("new ingest: %v", err)
	}

	payload := DroneTelemetry{
		ContractVersion: droneTelemetryContractVersionV1,
		DroneID:         "drone-01",
		Timestamp:       time.Now().UTC(),
		Latitude:        99.99,
		Longitude:       -222,
	}
	raw, _ := json.Marshal(payload)

	if err := ti.processJSON(raw); err == nil {
		t.Fatalf("expected invalid coordinate error")
	}
}

func TestProcessJSONRejectsDuplicateAndOutOfOrderSamples(t *testing.T) {
	ti, err := NewTelemetryIngest(TelemetryIngestConfig{Protocol: ProtocolJSON})
	if err != nil {
		t.Fatalf("new ingest: %v", err)
	}

	ts := time.Now().UTC()
	payload := DroneTelemetry{
		ContractVersion: droneTelemetryContractVersionV1,
		DroneID:         "drone-01",
		Timestamp:       ts,
		Latitude:        37.4219,
		Longitude:       -122.0840,
	}
	raw, _ := json.Marshal(payload)

	if err := ti.processJSON(raw); err != nil {
		t.Fatalf("expected first sample accepted: %v", err)
	}
	if err := ti.processJSON(raw); err == nil {
		t.Fatalf("expected duplicate sample error")
	}

	payload.Timestamp = ts.Add(-1 * time.Second)
	rawOlder, _ := json.Marshal(payload)
	if err := ti.processJSON(rawOlder); err == nil {
		t.Fatalf("expected out-of-order sample error")
	}

	stats := ti.Stats()
	if stats.RejectedTotal < 2 {
		t.Fatalf("expected at least 2 rejects, got %d", stats.RejectedTotal)
	}
}
