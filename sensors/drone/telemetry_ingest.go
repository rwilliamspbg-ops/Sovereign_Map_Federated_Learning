package drone

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"math"
	"net"
	"sync"
	"time"
)

// DroneProtocol identifies communication protocol.
type DroneProtocol string

const (
	ProtocolMAVLink DroneProtocol = "mavlink"
	ProtocolJSON    DroneProtocol = "json"
	ProtocolUDP     DroneProtocol = "udp"
)

// TelemetryIngestConfig controls drone data ingestion.
type TelemetryIngestConfig struct {
	Protocol    DroneProtocol
	ListenAddr  string
	BufferSize  int
	HeartbeatHz int
}

// TelemetryIngestStats tracks ingested telemetry sample counts.
type TelemetryIngestStats struct {
	Samples       int
	BytesReceived int64
	LastUpdate    time.Time
	ActiveDrones  int
}

// DroneTelemetry contains sensor readings from aerial vehicle.
type DroneTelemetry struct {
	DroneID        string    `json:"drone_id"`
	Timestamp      time.Time `json:"timestamp"`
	Latitude       float64   `json:"latitude"`
	Longitude      float64   `json:"longitude"`
	AltitudeMeters float64   `json:"altitude_meters"`
	HeadingDegrees float64   `json:"heading_degrees"`
	GroundSpeedMS  float64   `json:"ground_speed_ms"`
	BatteryPercent float64   `json:"battery_percent"`
	FlightMode     string    `json:"flight_mode"`
	ImageData      []byte    `json:"image_data,omitempty"`
	LidarPoints    []float64 `json:"lidar_points,omitempty"`
}

// TelemetryIngest manages drone telemetry data streams.
type TelemetryIngest struct {
	mu        sync.RWMutex
	config    TelemetryIngestConfig
	conn      net.PacketConn
	stats     TelemetryIngestStats
	drones    map[string]time.Time
	telemetry chan DroneTelemetry
}

// NewTelemetryIngest creates drone ingestion service.
func NewTelemetryIngest(cfg TelemetryIngestConfig) (*TelemetryIngest, error) {
	if cfg.ListenAddr == "" {
		cfg.ListenAddr = ":14550" // Default MAVLink port
	}
	if cfg.BufferSize == 0 {
		cfg.BufferSize = 1000
	}
	if cfg.HeartbeatHz == 0 {
		cfg.HeartbeatHz = 1
	}
	if cfg.Protocol == "" {
		cfg.Protocol = ProtocolJSON
	}

	return &TelemetryIngest{
		config:    cfg,
		drones:    make(map[string]time.Time),
		telemetry: make(chan DroneTelemetry, cfg.BufferSize),
	}, nil
}

// Start begins listening for drone telemetry on UDP socket.
func (ti *TelemetryIngest) Start(ctx context.Context) error {
	conn, err := net.ListenPacket("udp", ti.config.ListenAddr)
	if err != nil {
		return fmt.Errorf("listen udp: %w", err)
	}

	ti.conn = conn

	go ti.receiveLoop(ctx)
	go ti.heartbeatMonitor(ctx)

	return nil
}

// receiveLoop continuously reads telemetry packets from UDP socket.
func (ti *TelemetryIngest) receiveLoop(ctx context.Context) {
	defer ti.conn.Close()

	buf := make([]byte, 65536)

	for {
		select {
		case <-ctx.Done():
			return
		default:
			ti.conn.SetReadDeadline(time.Now().Add(1 * time.Second))
			n, _, err := ti.conn.ReadFrom(buf)
			if err != nil {
				if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
					continue
				}
				return
			}

			if err := ti.processTelemetry(buf[:n]); err != nil {
				continue
			}
		}
	}
}

// processTelemetry parses telemetry packet and queues for processing.
func (ti *TelemetryIngest) processTelemetry(data []byte) error {
	switch ti.config.Protocol {
	case ProtocolJSON:
		return ti.processJSON(data)
	case ProtocolMAVLink:
		return ti.processMAVLink(data)
	default:
		return fmt.Errorf("unsupported protocol: %s", ti.config.Protocol)
	}
}

// processJSON parses JSON-encoded telemetry.
func (ti *TelemetryIngest) processJSON(data []byte) error {
	var telem DroneTelemetry
	if err := json.Unmarshal(data, &telem); err != nil {
		return fmt.Errorf("unmarshal json: %w", err)
	}

	return ti.enqueueTelemetry(telem, int64(len(data)))
}

func (ti *TelemetryIngest) enqueueTelemetry(telem DroneTelemetry, bytes int64) error {
	if telem.Timestamp.IsZero() {
		telem.Timestamp = time.Now().UTC()
	}
	if telem.DroneID == "" {
		telem.DroneID = "unknown"
	}

	ti.mu.Lock()
	ti.stats.Samples++
	ti.stats.BytesReceived += bytes
	ti.stats.LastUpdate = time.Now()
	ti.drones[telem.DroneID] = time.Now()
	ti.stats.ActiveDrones = len(ti.drones)
	ti.mu.Unlock()

	select {
	case ti.telemetry <- telem:
		return nil
	default:
		return fmt.Errorf("telemetry buffer full")
	}
}

// processMAVLink parses MAVLink telemetry (v1/v2 framing, common message IDs).
func (ti *TelemetryIngest) processMAVLink(data []byte) error {
	parsed := 0
	i := 0
	for i < len(data) {
		magic := data[i]
		headerLen := 0
		payloadLen := 0
		msgID := 0
		sysID := uint8(0)

		switch magic {
		case 0xFE: // MAVLink v1
			if i+6 > len(data) {
				break
			}
			headerLen = 6
			payloadLen = int(data[i+1])
			sysID = data[i+3]
			msgID = int(data[i+5])
		case 0xFD: // MAVLink v2
			if i+10 > len(data) {
				break
			}
			headerLen = 10
			payloadLen = int(data[i+1])
			sysID = data[i+5]
			msgID = int(data[i+7]) | int(data[i+8])<<8 | int(data[i+9])<<16
		default:
			i++
			continue
		}

		packetLen := headerLen + payloadLen + 2 // checksum bytes
		if i+packetLen > len(data) {
			break
		}

		payloadStart := i + headerLen
		payload := data[payloadStart : payloadStart+payloadLen]
		telem, ok := parseMAVLinkTelemetry(msgID, sysID, payload)
		if ok {
			if err := ti.enqueueTelemetry(telem, int64(packetLen)); err != nil {
				return err
			}
			parsed++
		}

		i += packetLen
	}

	if parsed == 0 {
		return fmt.Errorf("no supported mavlink telemetry messages found")
	}

	return nil
}

func parseMAVLinkTelemetry(msgID int, sysID uint8, payload []byte) (DroneTelemetry, bool) {
	t := DroneTelemetry{
		DroneID:   fmt.Sprintf("sys-%d", sysID),
		Timestamp: time.Now().UTC(),
	}

	switch msgID {
	case 0: // HEARTBEAT
		if len(payload) < 9 {
			return DroneTelemetry{}, false
		}
		baseMode := payload[6]
		systemStatus := payload[7]
		t.FlightMode = fmt.Sprintf("base_mode=%d,status=%d", baseMode, systemStatus)
		return t, true

	case 1: // SYS_STATUS
		if len(payload) < 31 {
			return DroneTelemetry{}, false
		}
		// battery_remaining (int8) at offset 30
		t.BatteryPercent = float64(int8(payload[30]))
		return t, true

	case 33: // GLOBAL_POSITION_INT
		if len(payload) < 28 {
			return DroneTelemetry{}, false
		}
		lat := int32(binary.LittleEndian.Uint32(payload[4:8]))
		lon := int32(binary.LittleEndian.Uint32(payload[8:12]))
		altMM := int32(binary.LittleEndian.Uint32(payload[12:16]))
		hdgCd := binary.LittleEndian.Uint16(payload[26:28])
		vx := int16(binary.LittleEndian.Uint16(payload[20:22]))
		vy := int16(binary.LittleEndian.Uint16(payload[22:24]))

		t.Latitude = float64(lat) / 1e7
		t.Longitude = float64(lon) / 1e7
		t.AltitudeMeters = float64(altMM) / 1000.0
		if hdgCd != 0xFFFF {
			t.HeadingDegrees = float64(hdgCd) / 100.0
		}
		t.GroundSpeedMS = math.Sqrt(float64(vx*vx+vy*vy)) / 100.0
		return t, true

	case 74: // VFR_HUD
		if len(payload) < 20 {
			return DroneTelemetry{}, false
		}
		// groundspeed float32 at offset 4, heading int16 at offset 8, alt float32 at offset 12
		gsBits := binary.LittleEndian.Uint32(payload[4:8])
		altBits := binary.LittleEndian.Uint32(payload[12:16])
		t.GroundSpeedMS = float64(math.Float32frombits(gsBits))
		t.HeadingDegrees = float64(int16(binary.LittleEndian.Uint16(payload[8:10])))
		t.AltitudeMeters = float64(math.Float32frombits(altBits))
		return t, true
	}

	return DroneTelemetry{}, false
}

// heartbeatMonitor removes stale drones from active list.
func (ti *TelemetryIngest) heartbeatMonitor(ctx context.Context) {
	ticker := time.NewTicker(time.Second * time.Duration(ti.config.HeartbeatHz))
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			ti.pruneStale()
		}
	}
}

// pruneStale removes drones with no telemetry in last 30 seconds.
func (ti *TelemetryIngest) pruneStale() {
	cutoff := time.Now().Add(-30 * time.Second)

	ti.mu.Lock()
	for droneID, lastSeen := range ti.drones {
		if lastSeen.Before(cutoff) {
			delete(ti.drones, droneID)
		}
	}
	ti.stats.ActiveDrones = len(ti.drones)
	ti.mu.Unlock()
}

// NextTelemetry returns next telemetry sample from queue.
func (ti *TelemetryIngest) NextTelemetry(ctx context.Context) (DroneTelemetry, error) {
	select {
	case <-ctx.Done():
		return DroneTelemetry{}, ctx.Err()
	case telem := <-ti.telemetry:
		return telem, nil
	}
}

// Stats returns current ingestion statistics.
func (ti *TelemetryIngest) Stats() TelemetryIngestStats {
	ti.mu.RLock()
	defer ti.mu.RUnlock()
	return ti.stats
}

// GetActiveDrones returns list of drones that recently sent telemetry.
func (ti *TelemetryIngest) GetActiveDrones() []string {
	ti.mu.RLock()
	defer ti.mu.RUnlock()

	drones := make([]string, 0, len(ti.drones))
	for droneID := range ti.drones {
		drones = append(drones, droneID)
	}
	return drones
}

// Shutdown gracefully stops telemetry ingestion.
func (ti *TelemetryIngest) Shutdown(ctx context.Context) error {
	if ti.conn != nil {
		ti.conn.Close()
	}
	close(ti.telemetry)
	return nil
}
