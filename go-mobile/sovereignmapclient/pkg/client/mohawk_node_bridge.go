package client

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"
)

// AcceleratorBackend identifies the preferred on-device execution backend.
type AcceleratorBackend string

const (
	AcceleratorCPU    AcceleratorBackend = "cpu"
	AcceleratorNNAPI  AcceleratorBackend = "nnapi"
	AcceleratorCoreML AcceleratorBackend = "coreml"
)

// MobileCapabilities describes hardware and runtime characteristics reported by a mobile node.
type MobileCapabilities struct {
	Platform        string             `json:"platform"`
	PlatformVersion string             `json:"platform_version"`
	DeviceModel     string             `json:"device_model"`
	SecureHardware  string             `json:"secure_hardware"`
	Accelerator     AcceleratorBackend `json:"accelerator"`
	NPUAvailable    bool               `json:"npu_available"`
	MaxMemoryMB     int32              `json:"max_memory_mb"`
	CPUCores        int32              `json:"cpu_cores"`
	BatterySaver    bool               `json:"battery_saver"`
}

// SignedGradientUpdate is the transport object sent upstream by mobile clients.
type SignedGradientUpdate struct {
	NodeID             string `json:"node_id"`
	Round              int32  `json:"round"`
	GradientPayload    []byte `json:"gradient_payload"`
	GradientSignature  []byte `json:"gradient_signature"`
	SignerAlias        string `json:"signer_alias"`
	AttestationPayload []byte `json:"attestation_payload,omitempty"`
	CreatedAtUnix      int64  `json:"created_at_unix"`
}

// HardwareBackedSigner abstracts Secure Enclave / StrongBox implementations.
type HardwareBackedSigner interface {
	Sign(alias string, payload []byte) ([]byte, error)
	Attestation(alias string) ([]byte, error)
}

// hardwareSigner is registered by platform-specific glue code.
var hardwareSigner HardwareBackedSigner

// RegisterHardwareBackedSigner allows platform wrapper code to inject Secure Enclave / StrongBox signing.
func RegisterHardwareBackedSigner(signer HardwareBackedSigner) {
	hardwareSigner = signer
}

// MohawkNode is a gomobile-friendly bridge for Swift/Kotlin callers.
type MohawkNode struct {
	client       *SovereignMapClient
	nodeID       string
	signerAlias  string
	capabilities MobileCapabilities
	callTimeout  time.Duration
}

// NewMohawkNode creates a mobile bridge instance around the FL client.
func NewMohawkNode(serverAddr string, nodeID string) *MohawkNode {
	return &MohawkNode{
		client:      NewSovereignMapClient(serverAddr),
		nodeID:      nodeID,
		signerAlias: "mohawk.mobile.identity",
		capabilities: MobileCapabilities{
			Platform:       "unknown",
			SecureHardware: "none",
			Accelerator:    AcceleratorCPU,
			MaxMemoryMB:    2048,
			CPUCores:       4,
		},
		callTimeout: 15 * time.Second,
	}
}

// Connect opens a network session for FL participation.
func (n *MohawkNode) Connect() error {
	ctx, cancel := context.WithTimeout(context.Background(), n.callTimeout)
	defer cancel()
	return n.client.Connect(ctx)
}

// Disconnect terminates network connectivity.
func (n *MohawkNode) Disconnect() error {
	return n.client.Disconnect()
}

// ConfigureTraining updates local training parameters.
func (n *MohawkNode) ConfigureTraining(epochs int, batchSize int) {
	n.client.SetEpochs(epochs)
	n.client.SetBatchSize(batchSize)
}

// SetSignerAlias sets the hardware key alias used for gradient signatures.
func (n *MohawkNode) SetSignerAlias(alias string) {
	if alias != "" {
		n.signerAlias = alias
	}
}

// SetCapabilitiesJSON accepts a serialized MobileCapabilities payload from platform code.
func (n *MohawkNode) SetCapabilitiesJSON(raw string) error {
	if raw == "" {
		return errors.New("empty capabilities payload")
	}
	var caps MobileCapabilities
	if err := json.Unmarshal([]byte(raw), &caps); err != nil {
		return fmt.Errorf("invalid capabilities json: %w", err)
	}
	n.capabilities = caps
	return nil
}

// CapabilitiesJSON returns the current capability report.
func (n *MohawkNode) CapabilitiesJSON() string {
	blob, err := json.Marshal(n.capabilities)
	if err != nil {
		return "{}"
	}
	return string(blob)
}

// LoadTrainingDataFlat ingests flattened tensors from Swift/Kotlin and reshapes them for training.
func (n *MohawkNode) LoadTrainingDataFlat(features []float32, sampleCount int32, featureWidth int32, labels []int32) error {
	if sampleCount <= 0 || featureWidth <= 0 {
		return errors.New("sampleCount and featureWidth must be > 0")
	}
	expected := int(sampleCount * featureWidth)
	if len(features) != expected {
		return fmt.Errorf("invalid feature length: got=%d expected=%d", len(features), expected)
	}
	if len(labels) != int(sampleCount) {
		return fmt.Errorf("invalid label length: got=%d expected=%d", len(labels), sampleCount)
	}

	matrix := make([][]float32, sampleCount)
	for i := 0; i < int(sampleCount); i++ {
		start := i * int(featureWidth)
		end := start + int(featureWidth)
		matrix[i] = features[start:end]
	}
	return n.client.LoadData(matrix, labels)
}

// TrainLocalRound runs a local FL round and returns a serialized metrics payload.
func (n *MohawkNode) TrainLocalRound() ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), n.callTimeout)
	defer cancel()
	_, metrics, err := n.client.TrainLocal(ctx)
	if err != nil {
		return nil, err
	}
	metrics.NodeID = n.nodeID
	blob, err := json.Marshal(metrics)
	if err != nil {
		return nil, err
	}
	return blob, nil
}

// SignGradient signs an opaque gradient payload using registered hardware-backed keys.
func (n *MohawkNode) SignGradient(payload []byte) (SignedGradientUpdate, error) {
	if len(payload) == 0 {
		return SignedGradientUpdate{}, errors.New("empty gradient payload")
	}
	if hardwareSigner == nil {
		return SignedGradientUpdate{}, errors.New("no hardware signer registered")
	}

	sig, err := hardwareSigner.Sign(n.signerAlias, payload)
	if err != nil {
		return SignedGradientUpdate{}, fmt.Errorf("hardware sign failed: %w", err)
	}

	attestation, err := hardwareSigner.Attestation(n.signerAlias)
	if err != nil {
		return SignedGradientUpdate{}, fmt.Errorf("attestation fetch failed: %w", err)
	}

	status := n.client.GetMetrics()
	return SignedGradientUpdate{
		NodeID:             n.nodeID,
		Round:              status.Round,
		GradientPayload:    payload,
		GradientSignature:  sig,
		SignerAlias:        n.signerAlias,
		AttestationPayload: attestation,
		CreatedAtUnix:      time.Now().Unix(),
	}, nil
}

// StatusJSON returns bridge status for UI dashboards.
func (n *MohawkNode) StatusJSON() string {
	status := n.client.Status()
	status["node_id"] = n.nodeID
	status["capabilities"] = n.capabilities
	blob, err := json.Marshal(status)
	if err != nil {
		return "{}"
	}
	return string(blob)
}
