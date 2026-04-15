package mobile

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

type apiErrorEnvelope struct {
	Error struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

func TestHandleUploadAcceptsValidCameraFrame(t *testing.T) {
	api, err := NewPhoneClientAPI(PhoneClientConfig{})
	if err != nil {
		t.Fatalf("new api: %v", err)
	}

	api.clients["client-1"] = PhoneClientStatus{Registered: true, ClientID: "client-1"}

	payload := SensorFrame{
		ContractVersion: sensorContractVersionV1,
		ClientID:        "client-1",
		Type:            DataCamera,
		Timestamp:       time.Now().UTC(),
		Latitude:        37.4219,
		Longitude:       -122.0840,
		ImageData:       []byte{1, 2, 3},
	}

	body, _ := json.Marshal(payload)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/upload", bytes.NewReader(body))
	w := httptest.NewRecorder()

	api.handleUpload(w, req)

	if w.Code != http.StatusAccepted {
		t.Fatalf("expected status 202, got %d body=%s", w.Code, w.Body.String())
	}
}

func TestHandleUploadRejectsBadContractVersion(t *testing.T) {
	api, err := NewPhoneClientAPI(PhoneClientConfig{})
	if err != nil {
		t.Fatalf("new api: %v", err)
	}

	api.clients["client-1"] = PhoneClientStatus{Registered: true, ClientID: "client-1"}

	payload := SensorFrame{
		ContractVersion: "legacy-v0",
		ClientID:        "client-1",
		Type:            DataGPS,
		Timestamp:       time.Now().UTC(),
		Latitude:        37.4219,
		Longitude:       -122.0840,
	}

	body, _ := json.Marshal(payload)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/upload", bytes.NewReader(body))
	w := httptest.NewRecorder()

	api.handleUpload(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}

	var env apiErrorEnvelope
	if err := json.Unmarshal(w.Body.Bytes(), &env); err != nil {
		t.Fatalf("decode error payload: %v", err)
	}
	if env.Error.Code != errCodeUnsupportedContract {
		t.Fatalf("expected code %s, got %s", errCodeUnsupportedContract, env.Error.Code)
	}
}

func TestHandleUploadRejectsStaleTimestamp(t *testing.T) {
	api, err := NewPhoneClientAPI(PhoneClientConfig{MaxFrameAge: 5 * time.Second})
	if err != nil {
		t.Fatalf("new api: %v", err)
	}

	api.clients["client-1"] = PhoneClientStatus{Registered: true, ClientID: "client-1"}

	payload := SensorFrame{
		ContractVersion: sensorContractVersionV1,
		ClientID:        "client-1",
		Type:            DataGPS,
		Timestamp:       time.Now().UTC().Add(-1 * time.Minute),
		Latitude:        37.4219,
		Longitude:       -122.0840,
	}

	body, _ := json.Marshal(payload)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/upload", bytes.NewReader(body))
	w := httptest.NewRecorder()

	api.handleUpload(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}
}

func TestHandleUploadRejectsCameraWithoutImageData(t *testing.T) {
	api, err := NewPhoneClientAPI(PhoneClientConfig{})
	if err != nil {
		t.Fatalf("new api: %v", err)
	}

	api.clients["client-1"] = PhoneClientStatus{Registered: true, ClientID: "client-1"}

	payload := SensorFrame{
		ContractVersion: sensorContractVersionV1,
		ClientID:        "client-1",
		Type:            DataCamera,
		Timestamp:       time.Now().UTC(),
		Latitude:        37.4219,
		Longitude:       -122.0840,
	}

	body, _ := json.Marshal(payload)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/upload", bytes.NewReader(body))
	w := httptest.NewRecorder()

	api.handleUpload(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}
}

func TestHandleUploadRejectsUnregisteredClient(t *testing.T) {
	api, err := NewPhoneClientAPI(PhoneClientConfig{})
	if err != nil {
		t.Fatalf("new api: %v", err)
	}

	payload := SensorFrame{
		ContractVersion: sensorContractVersionV1,
		ClientID:        "unknown-client",
		Type:            DataGPS,
		Timestamp:       time.Now().UTC(),
		Latitude:        37.4219,
		Longitude:       -122.0840,
	}
	body, _ := json.Marshal(payload)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/upload", bytes.NewReader(body))
	w := httptest.NewRecorder()

	api.handleUpload(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d", w.Code)
	}
}

func TestHandleUploadRejectsFutureTimestamp(t *testing.T) {
	api, err := NewPhoneClientAPI(PhoneClientConfig{MaxClockSkew: 1 * time.Second})
	if err != nil {
		t.Fatalf("new api: %v", err)
	}
	api.clients["client-1"] = PhoneClientStatus{Registered: true, ClientID: "client-1"}

	payload := SensorFrame{
		ContractVersion: sensorContractVersionV1,
		ClientID:        "client-1",
		Type:            DataGPS,
		Timestamp:       time.Now().UTC().Add(10 * time.Second),
		Latitude:        37.4219,
		Longitude:       -122.0840,
	}
	body, _ := json.Marshal(payload)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/upload", bytes.NewReader(body))
	w := httptest.NewRecorder()
	api.handleUpload(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}
}

func TestHandleUploadRejectsOutOfRangeCoordinates(t *testing.T) {
	api, err := NewPhoneClientAPI(PhoneClientConfig{})
	if err != nil {
		t.Fatalf("new api: %v", err)
	}
	api.clients["client-1"] = PhoneClientStatus{Registered: true, ClientID: "client-1"}

	payload := SensorFrame{
		ContractVersion: sensorContractVersionV1,
		ClientID:        "client-1",
		Type:            DataGPS,
		Timestamp:       time.Now().UTC(),
		Latitude:        120,
		Longitude:       -200,
	}
	body, _ := json.Marshal(payload)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/upload", bytes.NewReader(body))
	w := httptest.NewRecorder()
	api.handleUpload(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}
}

func TestHandleUploadRejectsIMUWithoutPayload(t *testing.T) {
	api, err := NewPhoneClientAPI(PhoneClientConfig{})
	if err != nil {
		t.Fatalf("new api: %v", err)
	}
	api.clients["client-1"] = PhoneClientStatus{Registered: true, ClientID: "client-1"}

	payload := SensorFrame{
		ContractVersion: sensorContractVersionV1,
		ClientID:        "client-1",
		Type:            DataAccelerometer,
		Timestamp:       time.Now().UTC(),
		Latitude:        37.4219,
		Longitude:       -122.0840,
	}
	body, _ := json.Marshal(payload)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/upload", bytes.NewReader(body))
	w := httptest.NewRecorder()
	api.handleUpload(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}
}

func TestHandleUploadRejectsDuplicateAndOutOfOrderFrames(t *testing.T) {
	api, err := NewPhoneClientAPI(PhoneClientConfig{})
	if err != nil {
		t.Fatalf("new api: %v", err)
	}
	api.clients["client-1"] = PhoneClientStatus{Registered: true, ClientID: "client-1"}

	ts := time.Now().UTC()
	first := SensorFrame{
		ContractVersion: sensorContractVersionV1,
		ClientID:        "client-1",
		Type:            DataGPS,
		Timestamp:       ts,
		Latitude:        37.4219,
		Longitude:       -122.0840,
	}
	body1, _ := json.Marshal(first)
	req1 := httptest.NewRequest(http.MethodPost, "/api/v1/upload", bytes.NewReader(body1))
	w1 := httptest.NewRecorder()
	api.handleUpload(w1, req1)
	if w1.Code != http.StatusAccepted {
		t.Fatalf("expected first frame accepted, got %d", w1.Code)
	}

	body2, _ := json.Marshal(first)
	req2 := httptest.NewRequest(http.MethodPost, "/api/v1/upload", bytes.NewReader(body2))
	w2 := httptest.NewRecorder()
	api.handleUpload(w2, req2)
	if w2.Code != http.StatusConflict {
		t.Fatalf("expected duplicate frame rejected with 409, got %d", w2.Code)
	}

	older := first
	older.Timestamp = ts.Add(-1 * time.Second)
	body3, _ := json.Marshal(older)
	req3 := httptest.NewRequest(http.MethodPost, "/api/v1/upload", bytes.NewReader(body3))
	w3 := httptest.NewRecorder()
	api.handleUpload(w3, req3)
	if w3.Code != http.StatusConflict {
		t.Fatalf("expected out-of-order frame rejected with 409, got %d", w3.Code)
	}
}

func TestHealthIncludesIngestStatsSnapshot(t *testing.T) {
	api, err := NewPhoneClientAPI(PhoneClientConfig{})
	if err != nil {
		t.Fatalf("new api: %v", err)
	}
	api.clients["client-1"] = PhoneClientStatus{Registered: true, ClientID: "client-1"}

	payload := SensorFrame{
		ContractVersion: sensorContractVersionV1,
		ClientID:        "client-1",
		Type:            DataGPS,
		Timestamp:       time.Now().UTC(),
		Latitude:        37.4219,
		Longitude:       -122.0840,
	}
	body, _ := json.Marshal(payload)
	reqUpload := httptest.NewRequest(http.MethodPost, "/api/v1/upload", bytes.NewReader(body))
	wUpload := httptest.NewRecorder()
	api.handleUpload(wUpload, reqUpload)

	reqHealth := httptest.NewRequest(http.MethodGet, "/api/v1/health", nil)
	wHealth := httptest.NewRecorder()
	api.handleHealth(wHealth, reqHealth)

	if wHealth.Code != http.StatusOK {
		t.Fatalf("expected health 200, got %d", wHealth.Code)
	}

	var resp map[string]any
	if err := json.Unmarshal(wHealth.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode health response: %v", err)
	}
	if _, ok := resp["ingest_stats"]; !ok {
		t.Fatalf("expected ingest_stats in health response")
	}
}
