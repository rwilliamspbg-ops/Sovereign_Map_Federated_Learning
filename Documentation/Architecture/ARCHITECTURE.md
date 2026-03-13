# Sovereign Map - System Architecture

Comprehensive technical documentation of the Sovereign Map Byzantine-tolerant federated learning system.

> Claim scope: This architecture describes intended behavior and implemented components. Validation strength varies by test type (unit/integration/emulation/benchmark). See [CI_STATUS_AND_CLAIMS.md](/Documentation/Security/CI_STATUS_AND_CLAIMS.md).

**Table of Contents**
- [High-Level Architecture](#high-level-architecture)
- [Component Details](#component-details)
- [Data Flow](#data-flow)
- [Security Architecture](#security-architecture)
- [Monitoring Architecture](#monitoring-architecture)
- [Deployment Architecture](#deployment-architecture)
- [Performance Characteristics](#performance-characteristics)
- [Design Decisions](#design-decisions)

## High-Level Architecture

```
┌─────────────────────────────────────────────────────────────────────┐
│                     Sovereign Map System                             │
├─────────────────────────────────────────────────────────────────────┤
│                                                                      │
│  ┌───────────────────────────────────────────────────────────────┐  │
│  │  FEDERATED LEARNING LAYER                                    │  │
│  │                                                               │  │
│  │  ┌──────────────────┐      ┌──────────────────────────────┐  │  │
│  │  │ Backend          │      │ Node Agents (Learners)       │  │  │
│  │  │ (Aggregator)     │◄────►│ - Train local models         │  │  │
│  │  │                  │      │ - Send weight updates        │  │  │
│  │  │ - Aggregate      │      │ - Receive aggregated weights │  │  │
│  │  │ - Convergence    │      │ - Stake management           │  │  │
│  │  │ - Governance     │      └──────────────────────────────┘  │  │
│  │  └──────────────────┘              (Scalable: 1-100K+)      │  │
│  │                                                               │  │
│  └───────────────────────────────────────────────────────────────┘  │
│                                   △                                  │
│                                   │ Signed & Encrypted              │
│                                   ▽                                  │
│  ┌───────────────────────────────────────────────────────────────┐  │
│  │  TPM TRUST & SECURITY LAYER                                   │  │
│  │                                                               │  │
│  │  ┌──────────────────────────────────────────────────────┐   │  │
│  │  │ Certificate Authority                                │   │  │
│  │  │ - Root CA (4096-bit RSA, 10-year validity)          │   │  │
│  │  │ - Node Certificates (2048-bit RSA, 1-year)          │   │  │
│  │  │ - Trust Store (JSON-based)                           │   │  │
│  │  │ - Certificate Revocation List (CRL)                  │   │  │
│  │  └──────────────────────────────────────────────────────┘   │  │
│  │                                                               │  │
│  │  ┌──────────────────────────────────────────────────────┐   │  │
│  │  │ Node Authentication                                  │   │  │
│  │  │ - Message Signing (RSA-PSS)                          │   │  │
│  │  │ - Signature Verification                             │   │  │
│  │  │ - Timestamp Validation (Replay protection)           │   │  │
│  │  │ - Trust Cache (1-hour TTL)                           │   │  │
│  │  └──────────────────────────────────────────────────────┘   │  │
│  │                                                               │  │
│  │  ┌──────────────────────────────────────────────────────┐   │  │
│  │  │ Secure Communication                                 │   │  │
│  │  │ - mTLS Endpoints                                     │   │  │
│  │  │ - Encrypted Messages                                 │   │  │
│  │  │ - Certificate Pinning                                │   │  │
│  │  └──────────────────────────────────────────────────────┘   │  │
│  │                                                               │  │
│  └───────────────────────────────────────────────────────────────┘  │
│                                   △                                  │
│                                   │                                  │
│                                   ▼                                  │
│  ┌───────────────────────────────────────────────────────────────┐  │
│  │  OBSERVABILITY & MONITORING LAYER                             │  │
│  │                                                               │  │
│  │  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐        │  │
│  │  │ Prometheus   │  │ Grafana      │  │ Alertmanager │        │  │
│  │  │              │  │              │  │              │        │  │
│  │  │ - 20+ Metrics│  │ - 3 Dashbds  │  │ - 14 Rules   │        │  │
│  │  │ - 30d Retain │  │ - Real-time  │  │ - Email/Slack│        │  │
│  │  │ - Alert Eval │  │ - 30s Refresh│  │ - Routing    │        │  │
│  │  └──────────────┘  └──────────────┘  └──────────────┘        │  │
│  │                                                               │  │
│  │  ┌──────────────────────────────────────────────────────┐   │  │
│  │  │ Loki (Optional Log Aggregation)                       │   │  │
│  │  │ - Searchable logs from all services                   │   │  │
│  │  │ - Integrated with Grafana                             │   │  │
│  │  └──────────────────────────────────────────────────────┘   │  │
│  │                                                               │  │
│  └───────────────────────────────────────────────────────────────┘  │
│                                                                      │
└─────────────────────────────────────────────────────────────────────┘
```

## Component Details

### 1. Backend Aggregator (`sovereignmap_production_backend_v2.py`)

**Responsibilities:**
- Coordinate federated learning rounds
- Collect and aggregate model updates
- Compute stake-weighted trimmed mean
- Track convergence metrics
- Manage DAO governance
- Export metrics

**Key Functions:**

```python
def fl_round_endpoint():
    """Execute single federated learning round"""
    # 1. Collect updates from all nodes
    # 2. Apply stake weights
    # 3. Compute trimmed mean aggregate
    # 4. Calculate convergence metrics
    # 5. Distribute updated model
    # 6. Record metrics & history

def convergence_tracking():
    """Maintain convergence history"""
    # Track: rounds, accuracies, losses, timestamps
    # Used for visualization & analysis
    
def byzantine_tolerance():
    """Built-in Byzantine resilience"""
    # Stake weighting reduces Byzantine impact
    # Trimmed mean removes outliers
    # Continuous trust monitoring
```

**API Endpoints:**

| Endpoint | Method | Purpose |
|----------|--------|---------|
| `/health` | GET | Health check |
| `/fl_round` | POST | Execute FL round |
| `/convergence` | GET | Get convergence history |
| `/metrics_summary` | GET | System metrics |
| `/metrics` | GET | Prometheus format |

### 2. Node Agents (`src/client.py`)

**Responsibilities:**
- Train models on local data
- Send updates to aggregator
- Receive and apply aggregated models
- Manage local state
- Report metrics
- Support Byzantine mode (testing)

**Local Training Pipeline:**

```python
class SovereignClient:
    def fit(self, parameters, config):
        # 1. Receive aggregated parameters from server
        # 2. Load local training data (MNIST subset)
        # 3. Apply differential privacy (Opacus)
        # 4. Train model for N epochs
        # 5. Return updated parameters + metadata
        
    def on_byzantine(self):
        # For testing: invert gradients
        # Simulates Byzantine attack
        # Used to validate Byzantine tolerance
```

**Data Distribution:**
- Horizontally split: Each node gets unique subset
- Non-IID data: Mimics real federated scenario
- Differential privacy: ε-privacy per node

### 3. TPM Trust System

#### 3a. Certificate Manager (`tpm_cert_manager.py`)

```python
class TPMCertificateManager:
    def __init__(self, cert_dir):
        # Initialize or load existing CA
        # Load trust store from disk
        # Load certificate revocation list
        
    def _generate_ca(self):
        # Create root CA with 4096-bit RSA
        # 10-year validity
        # Sign with itself (self-signed)
        
    def generate_node_cert(self, node_id, name):
        # Create 2048-bit RSA key pair
        # Generate certificate signed by CA
        # 1-year validity
        # Save to disk
        # Update trust store
        
    def verify_node_certificate(self, node_id):
        # Load node certificate
        # Verify CA signature
        # Check expiration
        # Check CRL
        # Update trust status
        
    def revoke_node_certificate(self, node_id):
        # Add serial number to CRL
        # Save CRL to disk
        # Immediate effect (no grace period)
```

#### 3b. Node Authenticator (`tpm_cert_manager.py`)

```python
class NodeAuthenticator:
    def sign_message(self, message):
        # Use node's private key
        # RSA-PSS signature scheme
        # SHA-256 hash
        # Return hex-encoded signature
        
    def verify_message(self, message, signature, peer_node_id):
        # Load peer's certificate
        # Extract public key
        # RSA-PSS verification
        # Return True/False
        
    def create_authenticated_message(self, data):
        # Add timestamp
        # Sign with private key
        # Return signed message object
        # Include signature + hash
```

#### 3c. Secure Communication (`secure_communication.py`)

```python
class SecureNodeCommunication:
    @app.route('/secure_endpoint')
    @secure_endpoint  # Decorator
    def protected_route():
        # Check X-From-Node header
        # Verify X-Signature
        # Check certificate validity
        # Extract peer_node_id from context
        # Execute protected logic
        
    def create_signed_request(self, target, data):
        # Serialize data
        # Sign with node key
        # Add headers
        # Return ready-to-send request
```

### 4. Metrics Exporter (`tpm_metrics_exporter.py`)

**Metrics Categories:**

```
Certificate Metrics (5):
├── tpm_certificates_total          (Gauge)
├── tpm_certificates_verified_total (Gauge)
├── tpm_certificates_revoked_total  (Gauge)
├── tpm_certificate_expiry_seconds  (Gauge + node_id label)
└── tpm_certificate_age_seconds     (Gauge + node_id label)

Trust Chain Metrics (4):
├── tpm_trust_chain_valid           (Gauge + node_id)
├── tpm_trust_verification_duration_seconds (Histogram)
├── tpm_trust_verification_failures_total   (Counter)
└── tpm_ca_certificate_valid        (Gauge)

Message Auth Metrics (4):
├── tpm_messages_signed_total       (Counter + node_id)
├── tpm_messages_verified_total     (Counter + node_id, from_node_id)
├── tpm_signature_verification_failures_total (Counter + from_node_id)
└── tpm_message_verification_duration_seconds (Histogram)

Node Trust Metrics (5):
├── tpm_node_trust_score            (Gauge + node_id, 0-100)
├── tpm_node_certificate_valid      (Gauge + node_id)
├── tpm_node_certificate_revoked    (Gauge + node_id)
├── tpm_crl_size                    (Gauge)
├── tpm_trust_cache_hits_total      (Counter)
└── tpm_trust_cache_misses_total    (Counter)
```

## Data Flow

### Federated Learning Round

```
1. INITIALIZATION
   Backend: Initialize model weights
   Nodes:   Receive initial weights
   
2. LOCAL TRAINING
   Node 1: Train on shard 1 → weights_1 + metadata
   Node 2: Train on shard 2 → weights_2 + metadata
   Node N: Train on shard N → weights_N + metadata
   (Parallel, independent training)
   
3. AGGREGATION
   All Nodes: Send (signed) updates to Backend
   Backend: Collect all updates
   Backend: Apply stake weighting
   Backend: Compute trimmed mean
   → aggregated_weights
   
4. DISTRIBUTION
   Backend: Sign aggregated weights
   Backend: Broadcast to all nodes
   All Nodes: Receive, verify signature, apply weights
   
5. METRICS UPDATE
   Backend: Calculate accuracy, loss, convergence_rate
   Backend: Export to Prometheus
   Prometheus: Scrape metrics
   Grafana: Display on dashboard
```

### Trust Verification Flow

```
1. NODE STARTUP
   Node: Request CA certificate
   Node: Request own certificate
   Node: Verify certificate against CA
   Node: Load private key
   Node: Ready for communication
   
2. MESSAGE SEND
   Node A: Create message + timestamp
   Node A: Sign with private key (RSA-PSS)
   Node A: Send with signature
   
3. MESSAGE RECEIVE
   Node B: Receive message + signature
   Node B: Extract sender (from message)
   Node B: Load sender's certificate
   Node B: Verify signature with sender's public key
   Node B: Check timestamp (prevent replay)
   Node B: Check certificate not revoked (CRL)
   Node B: Cache verification result (1 hour)
   Node B: Process message if valid
   
4. REVOCATION
   Admin: Issue revoke command for Node X
   Backend: Remove Node X certificate serial from trust store
   Backend: Add to CRL
   Backend: Push CRL to all nodes
   All Nodes: Update local CRL
   All Nodes: Reject any messages from Node X (CRL match)
```

## Security Architecture

### Threat Model

**Protected Against:**
- Byzantine node attacks: resilience validated in simulation/integration workflows
- Message tampering: RSA-PSS signatures
- Man-in-the-middle: mTLS + certificate pinning
- Replay attacks: Timestamp validation
- Compromised nodes: Certificate revocation

**Limitations:**
- Network partition: Byzantine assumptions may not hold
- 51% coordinated attack: Use Proof-of-Stake
- Cryptographic breaks: Use hardware TPM for keys
- Side-channel attacks: Not in scope (use HSM)

### Defense Layers

```
Layer 1: Cryptographic Signatures
├── Every message signed with RSA-PSS
├── Sender identity cryptographically verifiable
└── Tampering detected immediately

Layer 2: Trust Chain Validation
├── Every node certificate validated against CA
├── Expiration checked
├── Revocation list checked
└── Trust score calculated

Layer 3: Byzantine Tolerance
├── Stake-weighted aggregation
├── Outliers removed (trimmed mean)
├── Reputation tracking per node
└── Automatic node isolation

Layer 4: Monitoring & Alerting
├── Real-time metric collection
├── Anomaly detection (high failure rates)
├── Alert on certificate issues
├── Alert on Byzantine suspicion

Layer 5: Access Control
├── TLS for transport encryption
├── Private network deployment
├── Rate limiting (optional)
└── IP whitelisting (optional)
```

## Monitoring Architecture

### Metrics Collection Path

```
TPM Nodes → Exporter → Prometheus → Grafana
            (Flask)     (TSDB)     (UI)
            :9091       :9090      :3000

TSP Backend → Flask → Prometheus
Metrics         Exporter
(Built-in)      :8000/:metrics

                ↓
            Alertmanager
            (Rules Engine)
            :9093
            
                ↓
            Email/Slack
            Notifications
```

### Alert Evaluation Pipeline

```
Every 30 seconds:

1. Prometheus evaluates alert rules
   - tpm_certificate_expiry_seconds < 2592000
   - tpm_node_trust_score == 0
   - rate(tpm_signature_verification_failures[5m]) > 0.1
   
2. If condition true for specified duration:
   - Mark alert as FIRING
   - Evaluate annotation templates
   
3. Send to Alertmanager
   - Group alerts by label
   - Apply routing rules
   
4. Alertmanager routes to receivers
   - Email: security@example.com
   - Slack: #security-alerts
   - PagerDuty: security-oncall
```

## Deployment Architecture

### Single-Machine Deployment (Recommended for <1K nodes)

```
Docker Host (16GB RAM)
├── Docker Network (sovereign-network)
├── Containers:
│   ├── backend (port 8000)
│   ├── node-agent-1
│   ├── node-agent-2
│   ├── ...
│   ├── node-agent-N (scalable)
│   ├── prometheus (port 9090)
│   ├── grafana (port 3000)
│   ├── alertmanager (port 9093)
│   ├── loki (port 3100)
│   └── tpm-metrics (port 9091)
└── Volumes:
    ├── tpm-certs (certificates)
    ├── prometheus-data (metrics)
    ├── grafana-data (dashboards)
    └── alertmanager-data (config)
```

### Multi-Machine Deployment (For >1K nodes)

```
Frontend Tier:
├── Reverse Proxy (nginx/ALB)
└── Load Balancer

Application Tier:
├── Backend Node (Machine 1)
├── Backend Node (Machine 2, HA)
└── Backend Node (Machine 3, HA)

Compute Tier:
├── Node Agent Pool 1 (100 nodes)
├── Node Agent Pool 2 (100 nodes)
├── Node Agent Pool 3 (100 nodes)
└── Node Agent Pool N (100 nodes)

Observability Tier:
├── Prometheus (central metrics)
├── Grafana (central UI)
├── Alertmanager (central alerts)
└── Loki (central logs)

Storage Tier:
├── Persistent volumes (certificates, data)
├── Backup storage (S3/NFS)
└── Archive storage (long-term retention)
```

## Performance Characteristics

### Scaling Analysis

**Model: O(n log n) Aggregation**

```
Nodes | Time/Round | Accuracy | Notes
------|----------|----------|------------------
10    | 100ms    | 92%      | Base case
100   | 150ms    | 90%      | Good scaling
1K    | 200ms    | 85%      | Benchmark-observed
10K   | 300ms    | 82%      | Theoretical
100K  | 400ms    | 80%      | Extrapolated
```

### Memory Usage

```
Base: Backend + Prometheus + Grafana
      ↓
      ~2GB

+ 100 Nodes: ~100MB per node
      ↓
      ~12GB total

+ Monitoring
      ↓
      ~14GB total

Per-Node Overhead: ~140MB (including monitoring)
```

### Network Throughput

```
Per Round (1000 nodes):
├── Upload: 1000 × (model_size + metadata) = ~5MB
├── Download: 1 × aggregated_model = ~5MB
└── Total: ~10MB per round (30-second interval)

Network Requirement: ~2.7 Mbps per 1000 nodes
(Very reasonable for production)
```

## Design Decisions

### 1. Why Trimmed Mean for Aggregation?

- **Removes outliers** naturally without Byzantine detection
- **Stake weighting** aligns incentives
- **Computationally efficient** O(n log n)
- **Byzantine resilience target** up to 50% in modeled scenarios
- **Maintains convergence** even under attack

### 2. Why RSA-PSS for Message Signing?

- **Standardized & audited** cryptography
- **Good hardware support** (HSM-ready)
- **Mature security properties** when correctly implemented
- **Probabilistic signatures** with strong verification properties
- **Fast verification** <1ms per message

### 3. Why Certificate-Based Trust?

- **No PKI infrastructure** needed
- **Self-signed CA** simple deployment
- **Per-node identity** clear accountability
- **Revocation support** for compromised nodes
- **Trust scoring** mathematical basis

### 4. Why Prometheus + Grafana?

- **Industry standard** widely used
- **Operational simplicity** easy to deploy
- **Real-time** immediate visibility
- **Alerting** built-in
- **Extensible** custom metrics

### 5. Why Docker Compose (not Kubernetes)?

- **Simplicity** for development
- **Single-machine** deployment easy
- **Scaling** sufficient for <10K nodes
- **Learning curve** low
- **Production-capable** with careful configuration, hardening, and environment validation

For >10K nodes, Kubernetes recommended:
```bash
# Kubernetes deployment would use:
# - StatefulSet for backend (with PVC)
# - Deployment for node agents (HPA)
# - Service mesh (Istio) for mTLS
# - Prometheus Operator for metrics
# - Loki operator for logs
```

---

**Last Updated**: March 2026  
**Architecture Version**: v1.0.0  
**Status**: Active architecture reference (validate with CI + environment tests)
