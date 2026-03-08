package drone

import (
	"encoding/binary"
	"testing"
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
}
