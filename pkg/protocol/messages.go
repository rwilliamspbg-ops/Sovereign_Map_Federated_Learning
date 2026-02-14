// Copyright 2026 Sovereign-Mohawk Core Team
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package protocol

import (
	"time"
)

// ModelUpdate represents a model update from a participant node
type ModelUpdate struct {
	NodeID       string    `json:"node_id"`
	Round        int       `json:"round"`
	Weights      []byte    `json:"weights"`
	Proof        []byte    `json:"proof"`
	Timestamp    time.Time `json:"timestamp"`
	Metrics      Metrics   `json:"metrics"`
}

// Metrics holds training metrics
type Metrics struct {
	Loss     float64 `json:"loss"`
	Accuracy float64 `json:"accuracy"`
	Samples  int     `json:"samples"`
}

// AggregateModel represents the aggregated global model
type AggregateModel struct {
	Round        int       `json:"round"`
	Weights      []byte    `json:"weights"`
	Participants []string  `json:"participants"`
	Timestamp    time.Time `json:"timestamp"`
}

// RegistrationRequest is sent by a node to join the federation
type RegistrationRequest struct {
	NodeID      string `json:"node_id"`
	Capacity    int    `json:"capacity"`
	TPMAttestat []byte `json:"tpm_attestation,omitempty"`
}

// RegistrationResponse confirms node registration
type RegistrationResponse struct {
	NodeID   string `json:"node_id"`
	Approved bool   `json:"approved"`
	Round    int    `json:"round"`
}

// TrainingTask is sent to nodes to start a training round
type TrainingTask struct {
	Round         int       `json:"round"`
	GlobalWeights []byte    `json:"global_weights"`
	Epochs        int       `json:"epochs"`
	LearningRate  float64   `json:"learning_rate"`
	Deadline      time.Time `json:"deadline"`
}

// StatusUpdate is sent periodically by nodes
type StatusUpdate struct {
	NodeID    string    `json:"node_id"`
	Status    string    `json:"status"` // training, idle, error
	Round     int       `json:"round"`
	Progress  float64   `json:"progress"`
	Timestamp time.Time `json:"timestamp"`
}
