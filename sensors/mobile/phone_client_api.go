package mobile

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"
)

// SensorDataType identifies mobile sensor stream type.
type SensorDataType string

const (
	DataCamera        SensorDataType = "camera"
	DataGPS           SensorDataType = "gps"
	DataAccelerometer SensorDataType = "accelerometer"
	DataGyroscope     SensorDataType = "gyroscope"
	DataMagnetometer  SensorDataType = "magnetometer"
	DataBarometer     SensorDataType = "barometer"
)

// PhoneClientConfig controls mobile ingestion API.
type PhoneClientConfig struct {
	ListenAddr  string
	TLSCertPath string
	TLSKeyPath  string
	MaxUploadMB int
	RequireAuth bool
	AuthToken   string
}

// PhoneClientStatus tracks registration status for mobile contributors.
type PhoneClientStatus struct {
	Registered  bool
	ClientID    string
	LastSeen    time.Time
	DataStreams []SensorDataType
}

// SensorFrame contains mobile sensor data snapshot.
type SensorFrame struct {
	ClientID  string                 `json:"client_id"`
	Type      SensorDataType         `json:"type"`
	Timestamp time.Time              `json:"timestamp"`
	Latitude  float64                `json:"latitude"`
	Longitude float64                `json:"longitude"`
	Altitude  float64                `json:"altitude"`
	ImageData []byte                 `json:"image_data,omitempty"`
	IMUData   map[string]interface{} `json:"imu_data,omitempty"`
}

// PhoneClientAPI manages HTTP API for mobile sensor ingestion.
type PhoneClientAPI struct {
	mu      sync.RWMutex
	config  PhoneClientConfig
	server  *http.Server
	clients map[string]PhoneClientStatus
	frames  chan SensorFrame
}

// NewPhoneClientAPI creates mobile ingestion API server.
func NewPhoneClientAPI(cfg PhoneClientConfig) (*PhoneClientAPI, error) {
	if cfg.ListenAddr == "" {
		cfg.ListenAddr = ":8443"
	}
	if cfg.MaxUploadMB == 0 {
		cfg.MaxUploadMB = 50
	}

	api := &PhoneClientAPI{
		config:  cfg,
		clients: make(map[string]PhoneClientStatus),
		frames:  make(chan SensorFrame, 1000),
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/register", api.handleRegister)
	mux.HandleFunc("/api/v1/upload", api.handleUpload)
	mux.HandleFunc("/api/v1/stream", api.handleStream)
	mux.HandleFunc("/api/v1/health", api.handleHealth)

	api.server = &http.Server{
		Addr:         cfg.ListenAddr,
		Handler:      mux,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	return api, nil
}

// Start begins HTTP server for mobile clients.
func (api *PhoneClientAPI) Start(ctx context.Context) error {
	errChan := make(chan error, 1)

	go func() {
		var err error
		if api.config.TLSCertPath != "" && api.config.TLSKeyPath != "" {
			api.server.TLSConfig = &tls.Config{MinVersion: tls.VersionTLS12}
			err = api.server.ListenAndServeTLS(api.config.TLSCertPath, api.config.TLSKeyPath)
		} else {
			err = api.server.ListenAndServe()
		}
		if err != nil && err != http.ErrServerClosed {
			errChan <- err
		}
	}()

	select {
	case err := <-errChan:
		return err
	case <-time.After(100 * time.Millisecond):
		return nil
	}
}

// handleRegister registers mobile client and establishes session.
func (api *PhoneClientAPI) handleRegister(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if api.config.RequireAuth && r.Header.Get("Authorization") != "Bearer "+api.config.AuthToken {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	var req struct {
		ClientID    string           `json:"client_id"`
		DeviceModel string           `json:"device_model"`
		Streams     []SensorDataType `json:"streams"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	api.mu.Lock()
	api.clients[req.ClientID] = PhoneClientStatus{
		Registered:  true,
		ClientID:    req.ClientID,
		LastSeen:    time.Now(),
		DataStreams: req.Streams,
	}
	api.mu.Unlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "registered", "client_id": req.ClientID})
}

// handleUpload receives sensor data from mobile client.
func (api *PhoneClientAPI) handleUpload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, int64(api.config.MaxUploadMB)*1024*1024)

	var frame SensorFrame
	if err := json.NewDecoder(r.Body).Decode(&frame); err != nil {
		http.Error(w, "invalid payload", http.StatusBadRequest)
		return
	}

	api.mu.Lock()
	if status, ok := api.clients[frame.ClientID]; ok {
		status.LastSeen = time.Now()
		api.clients[frame.ClientID] = status
	}
	api.mu.Unlock()

	select {
	case api.frames <- frame:
		w.WriteHeader(http.StatusAccepted)
	default:
		http.Error(w, "buffer full", http.StatusServiceUnavailable)
	}
}

// handleStream provides SSE stream of sensor data for visualization.
func (api *PhoneClientAPI) handleStream(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	ctx := r.Context()
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			api.mu.RLock()
			data, _ := json.Marshal(api.clients)
			api.mu.RUnlock()
			fmt.Fprintf(w, "data: %s\n\n", data)
			if f, ok := w.(http.Flusher); ok {
				f.Flush()
			}
		}
	}
}

// handleHealth returns API health status.
func (api *PhoneClientAPI) handleHealth(w http.ResponseWriter, r *http.Request) {
	api.mu.RLock()
	clientCount := len(api.clients)
	api.mu.RUnlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  "ok",
		"clients": clientCount,
	})
}

// NextFrame returns next sensor frame from ingestion queue.
func (api *PhoneClientAPI) NextFrame(ctx context.Context) (SensorFrame, error) {
	select {
	case <-ctx.Done():
		return SensorFrame{}, ctx.Err()
	case frame := <-api.frames:
		return frame, nil
	}
}

// GetClients returns list of registered mobile clients.
func (api *PhoneClientAPI) GetClients() []PhoneClientStatus {
	api.mu.RLock()
	defer api.mu.RUnlock()

	clients := make([]PhoneClientStatus, 0, len(api.clients))
	for _, status := range api.clients {
		clients = append(clients, status)
	}
	return clients
}

// Shutdown gracefully stops the API server.
func (api *PhoneClientAPI) Shutdown(ctx context.Context) error {
	close(api.frames)
	return api.server.Shutdown(ctx)
}
