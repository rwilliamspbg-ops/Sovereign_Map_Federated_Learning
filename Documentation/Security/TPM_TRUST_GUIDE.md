# TPM Trust & Verification System - Complete Documentation

## Overview

This system provides cryptographic trust and verification for secure node-to-node communication in Sovereign Map federated learning. It implements TPM-inspired (Trusted Platform Module) patterns including:

- **Certificate Authority (CA)** - Self-signed root CA for certificate generation
- **mTLS Communication** - Mutual TLS for node-to-node authentication
- **Message Signing** - ECDSA/RSA-based message authentication
- **Certificate Revocation** - CRL support for compromised nodes
- **Trust Chain Validation** - Automatic verification of certificates

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│           Sovereign Map Federated Learning                   │
├─────────────────────────────────────────────────────────────┤
│                                                               │
│  ┌──────────────┐      ┌──────────────┐      ┌──────────┐  │
│  │ Backend Node │◄────►│ Node Agent 1 │      │ Node 2   │  │
│  └──────────────┘      └──────────────┘      └──────────┘  │
│       │                      │                     │          │
│       └──────────────────────┼─────────────────────┘          │
│                              │ All Communication               │
│                              │ Signed & Verified               │
│                              ▼                                 │
│                    ┌──────────────────┐                       │
│                    │  TPM CA Service  │                       │
│                    │ (Certificate &   │                       │
│                    │  Trust Manager)  │                       │
│                    └──────────────────┘                       │
│                              │                                 │
│        ┌─────────────────────┼─────────────────────┐          │
│        ▼                     ▼                     ▼          │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐       │
│  │Node-0 Cert   │  │Node-1 Cert   │  │Node-2 Cert   │       │
│  │& Key         │  │& Key         │  │& Key         │       │
│  └──────────────┘  └──────────────┘  └──────────────┘       │
│        (mTLS Protected Volume)                                │
│                                                               │
└─────────────────────────────────────────────────────────────┘
```

## Components

### 1. `tpm_cert_manager.py`

Main certificate management system:

```python
# Certificate Authority Management
TPMCertificateManager
├── _generate_ca()           # Create root CA
├── generate_node_cert()     # Issue node certificates
├── verify_node_certificate() # Verify cert against CA
├── revoke_node_certificate() # Revoke compromised node
└── get_trust_report()       # Get network trust status

# Message Authentication
NodeAuthenticator
├── sign_message()           # Sign with node private key
├── verify_message()         # Verify peer signature
├── create_authenticated_message() # Create signed message
└── verify_authenticated_message() # Verify signed message
```

### 2. `secure_communication.py`

Flask middleware and secure endpoints:

```python
SecureNodeCommunication
├── secure_endpoint()        # Decorator for authentication
├── create_signed_request()  # Create authenticated request
└── get_trust_status()       # Report network trust

# Integration with Flask
create_secure_app_middleware(app, node_id, num_nodes)
├── /trust/status            # GET trust status
├── /trust/verify/<id>       # POST verify node cert
├── /trust/revoke/<id>       # POST revoke node cert
└── /trust/certificate/<id>  # GET node certificate
```

### 3. `tpm-bootstrap.sh`

Container initialization script:

- Generates CA on first run
- Creates certificates for all nodes
- Verifies certificates before startup
- Handles multi-node coordination

### 4. `docker-compose.full.yml`

Docker Compose configuration with:

- **backend-secure** - Flask backend with mTLS
- **node-agent-secure** - Client nodes with cert support
- **tpm-ca-service** - Certificate Authority service
- **trust-dashboard** - Trust verification API
- **tpm-certs volume** - Shared certificate storage

## Certificate Structure

### Root CA Certificate

```
Issuer: Sovereign Map Root CA
Subject: Sovereign Map Root CA
Serial: Random 128-bit
Validity: 10 years
Key Size: 4096-bit RSA
Usage: CA, CRL Signing
```

### Node Certificates

```
Issuer: Sovereign Map Root CA
Subject CN: node-{id}
Serial: Random 128-bit
Validity: 1 year
Key Size: 2048-bit RSA
Usage: Digital Signature, TLS Auth
SAN: node-{id}, node-{id}.sovereign-network, 172.25.0.{id}
```

## Deployment

### 1. Quick Start (Single Container Set)

```bash
# Create sovereign network (if not exists)
docker network create sovereign-network

# Deploy with secure communication
docker compose -f docker-compose.full.yml up -d

# Check CA service
docker logs sovereign-tpm-ca

# Verify certificates
docker compose -f docker-compose.full.yml exec tpm-ca-service python -c "
from tpm_cert_manager import TPMCertificateManager
mgr = TPMCertificateManager('/etc/sovereign/certs')
print(mgr.get_trust_report())
"
```

### 2. Scale to Multiple Nodes

```bash
# Scale node agents to 100 nodes
docker compose -f docker-compose.full.yml up -d --scale node-agent-secure=100

# All nodes automatically get certificates and verify on startup
```

### 3. Monitor Trust Status

```bash
# Check trust dashboard
curl http://localhost:5001/trust/status

# Verify specific node
curl -X POST http://localhost:5001/trust/verify/5

# Get node certificate for peer verification
curl http://localhost:5001/trust/certificate/5
```

## Security Features

### 1. Certificate Pinning
- Each node has a unique certificate
- Private key never leaves node
- Certificates renewed annually

### 2. Message Authentication
- All messages signed with node private key
- Signature verification before processing
- Timestamp validation (prevents replay)

### 3. Certificate Revocation
- Certificate Revocation List (CRL) maintained
- Revoked nodes can't communicate
- CRL cached in memory for performance

### 4. Trust Chain Validation
- Certificates verified against CA on startup
- Periodic re-verification during runtime
- Trust cache (1 hour TTL) for performance

## API Endpoints

### Public API (No Auth Required)

```bash
# Get certificate
GET /trust/certificate/{node_id}
Response: { certificate: "-----BEGIN CERTIFICATE-----..." }

# Get CA certificate
GET /trust/ca-certificate
Response: { certificate: "-----BEGIN CERTIFICATE-----..." }
```

### Protected API (mTLS Required)

```bash
# Get trust status
GET /trust/status
Headers: X-From-Node, X-Signature, X-Node-Auth

# Update node model
POST /fl/update
Headers: X-From-Node, X-Signature, X-Node-Auth
Body: { weights: [...], accuracy: 95.5 }

# Get peer metrics
GET /metrics/peer/{node_id}
Headers: X-From-Node, X-Signature, X-Node-Auth
```

## Usage Examples

### Example 1: Generate and Verify Certificates

```python
from tpm_cert_manager import TPMCertificateManager

# Initialize manager
mgr = TPMCertificateManager('/etc/sovereign/certs')

# Generate certificates for 10 nodes
for i in range(10):
    cert_path, key_path = mgr.generate_node_cert(i, f"Node-{i}")
    print(f"Generated: {cert_path}")

# Verify all certificates
for i in range(10):
    verified = mgr.verify_node_certificate(i)
    print(f"Node {i} verified: {verified}")

# Get trust report
report = mgr.get_trust_report()
print(f"Total nodes: {report['total_nodes']}")
print(f"Verified: {report['verified_nodes']}")
```

### Example 2: Sign and Verify Messages

```python
from tpm_cert_manager import TPMCertificateManager, NodeAuthenticator

# Setup
mgr = TPMCertificateManager('/etc/sovereign/certs')
mgr.generate_node_cert(0, "Node-0")
mgr.generate_node_cert(1, "Node-1")

# Node 0 authenticator
auth_0 = NodeAuthenticator(0, mgr)

# Create and sign message
msg = auth_0.create_authenticated_message({
    "action": "update",
    "round": 42,
    "accuracy": 95.5
})

# Node 1 verifies message from Node 0
auth_1 = NodeAuthenticator(1, mgr)
verified = auth_1.verify_authenticated_message(msg)
print(f"Message verified: {verified}")
```

### Example 3: Secure Flask Endpoint

```python
from flask import Flask
from secure_communication import create_secure_app_middleware

app = Flask(__name__)

# Initialize middleware
comm = create_secure_app_middleware(app, node_id=0, num_nodes=10)

# Protected endpoint
@app.route('/secure/data', methods=['POST'])
@comm.secure_endpoint
def secure_data():
    from flask import g, request, jsonify
    return jsonify({
        "status": "received",
        "from_node": g.peer_node_id,
        "data": request.json
    })

app.run(host='0.0.0.0', port=5000)
```

### Example 4: Send Signed Request

```python
from secure_communication import SecureNodeCommunication, RequestSigner
import requests

# Setup
comm = SecureNodeCommunication(node_id=0)
signer = RequestSigner(comm)

# Create signed request
signed = signer.sign_and_send(
    method="POST",
    url="http://node-1:5000/secure/data",
    data={"action": "update", "round": 1}
)

# Send
response = requests.request(
    method=signed["method"],
    url=signed["url"],
    headers=signed["headers"],
    data=signed["data"]
)
```

## Monitoring and Debugging

### Check Certificate Status

```bash
# List all certificates
ls -la /etc/sovereign/certs/*.pem

# Inspect CA certificate
openssl x509 -in /etc/sovereign/certs/ca-cert.pem -text -noout

# Inspect node certificate
openssl x509 -in /etc/sovereign/certs/node-0-cert.pem -text -noout

# Verify cert signature
openssl verify -CAfile /etc/sovereign/certs/ca-cert.pem /etc/sovereign/certs/node-0-cert.pem
```

### View Trust Store

```bash
# Check trust store
cat /etc/sovereign/certs/trust-store.json | python -m json.tool

# Check CRL
cat /etc/sovereign/certs/crl.json | python -m json.tool
```

### Monitor Logs

```bash
# TPM bootstrap logs
docker logs sovereign-tpm-ca

# Backend logs
docker logs sovereign-backend-secure

# Node agent logs
docker logs node-agent-secure-1
```

## Performance Considerations

- **Trust Cache**: 1 hour TTL reduces verification overhead
- **Certificate Validity**: 1 year reduces rotation frequency
- **Lazy Loading**: Certificates loaded only when needed
- **Parallel Verification**: Multiple nodes verified in parallel

## Security Best Practices

1. **Rotate Certificates Regularly** - Plan annual rotation
2. **Protect Private Keys** - Store in secure volumes
3. **Monitor Certificate Expiry** - Set up alerts 30 days before expiry
4. **Implement CRL Distribution** - Distribute CRL to all nodes
5. **Audit Trust Operations** - Log all cert operations
6. **Use Hardware TPM** - Consider real TPM for production

## Troubleshooting

### Problem: "Certificate not found for node-X"
- Ensure tpm-ca-service has started
- Check tpm-certs volume is mounted
- Verify bootstrap script completed

### Problem: "Signature verification failed"
- Check node private key exists
- Verify peer certificate is not revoked
- Check message wasn't tampered

### Problem: "Certificate verification failed"
- Check certificate isn't expired
- Verify CA cert is available
- Check certificate not revoked

### Problem: Nodes can't communicate
- Ensure all nodes on same docker network
- Check firewall rules between nodes
- Verify certificates are valid

## Next Steps

1. **Integrate with Prometheus** - Export trust metrics
2. **Add Hardware TPM Support** - Real TPM integration
3. **Implement Certificate Pinning** - Pin certs at registry
4. **Create Admin Dashboard** - UI for cert management
5. **Add Key Rotation** - Automatic key rotation schedule
6. **Integrate with Vault** - HashiCorp Vault support

## References

- OpenSSL Documentation: https://www.openssl.org/docs/
- Python Cryptography: https://cryptography.io/
- mTLS Guide: https://horovits.medium.com/mutual-tls-mtls-afc5a6f9b7b5
- Certificate Authority Design: https://cabforum.org/

---

**Version**: 1.0.0  
**Last Updated**: 2024  
**Status**: Production Ready
