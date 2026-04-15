package mobile

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"
)

const sensorContractVersionV1 = "av-v1"

const (
	errCodeInvalidPayload        = "INVALID_PAYLOAD"
	errCodeUnsupportedContract   = "UNSUPPORTED_CONTRACT_VERSION"
	errCodeMissingClientID       = "MISSING_CLIENT_ID"
	errCodeUnsupportedSensorType = "UNSUPPORTED_SENSOR_TYPE"
	errCodeFutureTimestamp       = "FUTURE_TIMESTAMP"
	errCodeStaleTimestamp        = "STALE_TIMESTAMP"
	errCodeInvalidLatitude       = "INVALID_LATITUDE"
	errCodeInvalidLongitude      = "INVALID_LONGITUDE"
	errCodeMissingImageData      = "MISSING_IMAGE_DATA"
	errCodeMissingIMUData        = "MISSING_IMU_DATA"
	errCodeClientNotRegistered   = "CLIENT_NOT_REGISTERED"
	errCodeDuplicateFrame        = "DUPLICATE_FRAME"
	errCodeOutOfOrderFrame       = "OUT_OF_ORDER_FRAME"
	errCodeBufferFull            = "BUFFER_FULL"
)

type ingestValidationError struct {
	Code       string `json:"code"`
	Message    string `json:"message"`
	StatusCode int    `json:"-"`
}

func (e *ingestValidationError) Error() string {
	if e == nil {
		return ""
	}
	return e.Message
}

type phoneIngestStats struct {
	AcceptedTotal    int64            `json:"accepted_total"`
	RejectedTotal    int64            `json:"rejected_total"`
	BufferDropsTotal int64            `json:"buffer_drops_total"`
	AcceptedByType   map[string]int64 `json:"accepted_by_type"`
	RejectedByReason map[string]int64 `json:"rejected_by_reason"`
	LastAcceptedUnix map[string]int64 `json:"last_accepted_unix_by_client"`
}

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
	ListenAddr   string
	TLSCertPath  string
	TLSKeyPath   string
	MaxUploadMB  int
	RequireAuth  bool
	AuthToken    string
	MaxClockSkew time.Duration
	MaxFrameAge  time.Duration
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
	ContractVersion string                 `json:"contract_version"`
	ClientID        string                 `json:"client_id"`
	Type            SensorDataType         `json:"type"`
	Timestamp       time.Time              `json:"timestamp"`
	Latitude        float64                `json:"latitude"`
	Longitude       float64                `json:"longitude"`
	Altitude        float64                `json:"altitude"`
	ImageData       []byte                 `json:"image_data,omitempty"`
	IMUData         map[string]interface{} `json:"imu_data,omitempty"`
}

// PhoneClientAPI manages HTTP API for mobile sensor ingestion.
type PhoneClientAPI struct {
	mu           sync.RWMutex
	config       PhoneClientConfig
	server       *http.Server
	clients      map[string]PhoneClientStatus
	frames       chan SensorFrame
	lastAccepted map[string]time.Time
	stats        phoneIngestStats
}

// NewPhoneClientAPI creates mobile ingestion API server.
func NewPhoneClientAPI(cfg PhoneClientConfig) (*PhoneClientAPI, error) {
	if cfg.ListenAddr == "" {
		cfg.ListenAddr = ":8443"
	}
	if cfg.MaxUploadMB == 0 {
		cfg.MaxUploadMB = 50
	}
	if cfg.MaxClockSkew == 0 {
		cfg.MaxClockSkew = 5 * time.Second
	}
	if cfg.MaxFrameAge == 0 {
		cfg.MaxFrameAge = 30 * time.Second
	}

	api := &PhoneClientAPI{
		config:       cfg,
		clients:      make(map[string]PhoneClientStatus),
		frames:       make(chan SensorFrame, 1000),
		lastAccepted: make(map[string]time.Time),
		stats: phoneIngestStats{
			AcceptedByType:   make(map[string]int64),
			RejectedByReason: make(map[string]int64),
			LastAcceptedUnix: make(map[string]int64),
		},
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
	if err := json.NewEncoder(w).Encode(map[string]string{"status": "registered", "client_id": req.ClientID}); err != nil {
		log.Printf("phone api register response encode: %v", err)
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
	}
}

// handleUpload receives sensor data from mobile client.
func (api *PhoneClientAPI) handleUpload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		api.writeAPIError(w, &ingestValidationError{Code: errCodeInvalidPayload, Message: "method not allowed", StatusCode: http.StatusMethodNotAllowed})
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, int64(api.config.MaxUploadMB)*1024*1024)

	var frame SensorFrame
	if err := json.NewDecoder(r.Body).Decode(&frame); err != nil {
		api.writeAPIError(w, &ingestValidationError{Code: errCodeInvalidPayload, Message: "invalid payload", StatusCode: http.StatusBadRequest})
		return
	}

	if err := api.normalizeAndValidateFrame(&frame); err != nil {
		var vErr *ingestValidationError
		if errors.As(err, &vErr) {
			api.incrementReject(vErr.Code)
			api.writeAPIError(w, vErr)
			return
		}
		api.incrementReject(errCodeInvalidPayload)
		api.writeAPIError(w, &ingestValidationError{Code: errCodeInvalidPayload, Message: err.Error(), StatusCode: http.StatusBadRequest})
		return
	}

	api.mu.Lock()
	status, ok := api.clients[frame.ClientID]
	if !ok {
		api.mu.Unlock()
		api.incrementReject(errCodeClientNotRegistered)
		api.writeAPIError(w, &ingestValidationError{Code: errCodeClientNotRegistered, Message: "client_id not registered", StatusCode: http.StatusUnauthorized})
		return
	}

	if lastTs, hasLast := api.lastAccepted[frame.ClientID]; hasLast {
		if frame.Timestamp.Equal(lastTs) {
			api.mu.Unlock()
			api.incrementReject(errCodeDuplicateFrame)
			api.writeAPIError(w, &ingestValidationError{Code: errCodeDuplicateFrame, Message: "duplicate frame timestamp", StatusCode: http.StatusConflict})
			return
		}
		if frame.Timestamp.Before(lastTs) {
			api.mu.Unlock()
			api.incrementReject(errCodeOutOfOrderFrame)
			api.writeAPIError(w, &ingestValidationError{Code: errCodeOutOfOrderFrame, Message: "out-of-order frame timestamp", StatusCode: http.StatusConflict})
			return
		}
	}

	status.LastSeen = time.Now()
	api.clients[frame.ClientID] = status
	api.lastAccepted[frame.ClientID] = frame.Timestamp
	api.stats.AcceptedTotal++
	api.stats.AcceptedByType[string(frame.Type)]++
	api.stats.LastAcceptedUnix[frame.ClientID] = frame.Timestamp.Unix()
	api.mu.Unlock()

	select {
	case api.frames <- frame:
		w.WriteHeader(http.StatusAccepted)
	default:
		api.mu.Lock()
		api.stats.BufferDropsTotal++
		api.stats.RejectedTotal++
		api.stats.RejectedByReason[errCodeBufferFull]++
		api.mu.Unlock()
		api.writeAPIError(w, &ingestValidationError{Code: errCodeBufferFull, Message: "buffer full", StatusCode: http.StatusServiceUnavailable})
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
			data, err := json.Marshal(api.clients)
			api.mu.RUnlock()
			if err != nil {
				log.Printf("phone api stream marshal: %v", err)
				return
			}
			if _, err := fmt.Fprintf(w, "data: %s\n\n", data); err != nil {
				log.Printf("phone api stream write: %v", err)
				return
			}
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
	if err := json.NewEncoder(w).Encode(map[string]interface{}{
		"status":       "ok",
		"clients":      clientCount,
		"ingest_stats": api.snapshotStats(),
	}); err != nil {
		log.Printf("phone api health response encode: %v", err)
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
	}
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

func (api *PhoneClientAPI) normalizeAndValidateFrame(frame *SensorFrame) error {
	if frame == nil {
		return &ingestValidationError{Code: errCodeInvalidPayload, Message: "frame is required", StatusCode: http.StatusBadRequest}
	}
	if frame.ContractVersion != sensorContractVersionV1 {
		return &ingestValidationError{Code: errCodeUnsupportedContract, Message: fmt.Sprintf("unsupported contract_version: expected %s", sensorContractVersionV1), StatusCode: http.StatusBadRequest}
	}
	if frame.ClientID == "" {
		return &ingestValidationError{Code: errCodeMissingClientID, Message: "client_id is required", StatusCode: http.StatusBadRequest}
	}
	if !isValidSensorDataType(frame.Type) {
		return &ingestValidationError{Code: errCodeUnsupportedSensorType, Message: fmt.Sprintf("unsupported sensor type: %s", frame.Type), StatusCode: http.StatusBadRequest}
	}

	now := time.Now().UTC()
	if frame.Timestamp.IsZero() {
		frame.Timestamp = now
	}
	if frame.Timestamp.After(now.Add(api.config.MaxClockSkew)) {
		return &ingestValidationError{Code: errCodeFutureTimestamp, Message: "timestamp too far in future", StatusCode: http.StatusBadRequest}
	}
	if frame.Timestamp.Before(now.Add(-api.config.MaxFrameAge)) {
		return &ingestValidationError{Code: errCodeStaleTimestamp, Message: "timestamp too old", StatusCode: http.StatusBadRequest}
	}

	if frame.Latitude < -90 || frame.Latitude > 90 {
		return &ingestValidationError{Code: errCodeInvalidLatitude, Message: "latitude out of range", StatusCode: http.StatusBadRequest}
	}
	if frame.Longitude < -180 || frame.Longitude > 180 {
		return &ingestValidationError{Code: errCodeInvalidLongitude, Message: "longitude out of range", StatusCode: http.StatusBadRequest}
	}

	switch frame.Type {
	case DataCamera:
		if len(frame.ImageData) == 0 {
			return &ingestValidationError{Code: errCodeMissingImageData, Message: "camera frame missing image_data", StatusCode: http.StatusBadRequest}
		}
	case DataAccelerometer, DataGyroscope, DataMagnetometer, DataBarometer:
		if len(frame.IMUData) == 0 {
			return &ingestValidationError{Code: errCodeMissingIMUData, Message: "imu frame missing imu_data", StatusCode: http.StatusBadRequest}
		}
	case DataGPS:
		// No additional payload requirements beyond coordinates.
	}

	return nil
}

func (api *PhoneClientAPI) writeAPIError(w http.ResponseWriter, err *ingestValidationError) {
	if err == nil {
		err = &ingestValidationError{Code: errCodeInvalidPayload, Message: "unknown error", StatusCode: http.StatusBadRequest}
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(err.StatusCode)
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"error": map[string]string{
			"code":    err.Code,
			"message": err.Message,
		},
	})
}

func (api *PhoneClientAPI) incrementReject(code string) {
	api.mu.Lock()
	defer api.mu.Unlock()
	api.stats.RejectedTotal++
	api.stats.RejectedByReason[code]++
}

func (api *PhoneClientAPI) snapshotStats() phoneIngestStats {
	api.mu.RLock()
	defer api.mu.RUnlock()
	out := phoneIngestStats{
		AcceptedTotal:    api.stats.AcceptedTotal,
		RejectedTotal:    api.stats.RejectedTotal,
		BufferDropsTotal: api.stats.BufferDropsTotal,
		AcceptedByType:   make(map[string]int64, len(api.stats.AcceptedByType)),
		RejectedByReason: make(map[string]int64, len(api.stats.RejectedByReason)),
		LastAcceptedUnix: make(map[string]int64, len(api.stats.LastAcceptedUnix)),
	}
	for k, v := range api.stats.AcceptedByType {
		out.AcceptedByType[k] = v
	}
	for k, v := range api.stats.RejectedByReason {
		out.RejectedByReason[k] = v
	}
	for k, v := range api.stats.LastAcceptedUnix {
		out.LastAcceptedUnix[k] = v
	}
	return out
}

func isValidSensorDataType(dataType SensorDataType) bool {
	switch dataType {
	case DataCamera, DataGPS, DataAccelerometer, DataGyroscope, DataMagnetometer, DataBarometer:
		return true
	default:
		return false
	}
}
