package drone

import (
	"context"
	"encoding/json"
	"fmt"
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

	ti.mu.Lock()
	ti.stats.Samples++
	ti.stats.BytesReceived += int64(len(data))
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

// processMAVLink parses MAVLink telemetry (stub for future implementation).
func (ti *TelemetryIngest) processMAVLink(data []byte) error {
	// TODO: Implement MAVLink protocol parsing
	// For now, return stub
	return fmt.Errorf("mavlink parsing not implemented")
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
