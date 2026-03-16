#!/bin/bash
# TPM Certificate Bootstrap Script
# Generates and manages node certificates for secure communication

set -e

CERT_DIR="${CERT_DIR:-/etc/sovereign/certs}"
NODE_ID="${NODE_ID:-0}"
NUM_NODES="${NUM_NODES:-10}"
LOG_FILE="/var/log/sovereign/tpm-bootstrap.log"

# Create directories
mkdir -p "$CERT_DIR"
mkdir -p "$(dirname "$LOG_FILE")"

# Logging function
log() {
    echo "[$(date +'%Y-%m-%d %H:%M:%S')] $*" | tee -a "$LOG_FILE"
}

log "========================================="
log "TPM Certificate Bootstrap for Node $NODE_ID"
log "========================================="
log "Cert directory: $CERT_DIR"
log "Total nodes: $NUM_NODES"

# Check if CA exists
if [ ! -f "$CERT_DIR/ca-cert.pem" ]; then
    log "CA certificate not found. Generating new CA..."
    python3 << EOF
import sys
sys.path.insert(0, '/app')
from tpm_cert_manager import TPMCertificateManager

mgr = TPMCertificateManager('$CERT_DIR')
print("CA generated successfully")
EOF
    log "CA certificate generated"
else
    log "CA certificate already exists"
fi

# Generate certificates for all nodes (if this is node 0)
if [ "$NODE_ID" -eq 0 ]; then
    log "Node 0 - Generating certificates for all nodes..."
    python3 << EOF
import sys
sys.path.insert(0, '/app')
from tpm_cert_manager import TPMCertificateManager

mgr = TPMCertificateManager('$CERT_DIR')

# Generate certs for all nodes
for i in range($NUM_NODES):
    try:
        cert_path, key_path = mgr.generate_node_cert(i, f"Node-{i}")
        print(f"Generated certificate for node {i}")
    except Exception as e:
        print(f"Certificate may already exist for node {i}: {e}")

# Verify all certificates
for i in range($NUM_NODES):
    verified = mgr.verify_node_certificate(i)
    print(f"Node {i} verified: {verified}")

# Print trust report
import json
report = mgr.get_trust_report()
print(json.dumps(report, indent=2))
EOF
    log "All node certificates generated and verified"
else
    log "Node $NODE_ID - Waiting for CA and certificates..."
    
    # Wait for CA to be available (up to 30 seconds)
    max_attempts=30
    attempts=0
    while [ ! -f "$CERT_DIR/ca-cert.pem" ] && [ $attempts -lt $max_attempts ]; do
        log "Waiting for CA certificate... (attempt $((attempts+1))/$max_attempts)"
        sleep 1
        attempts=$((attempts+1))
    done
    
    if [ ! -f "$CERT_DIR/ca-cert.pem" ]; then
        log "ERROR: CA certificate not available after 30 seconds"
        exit 1
    fi
    
    log "CA certificate found"
    
    # Verify this node's certificate
    python3 << EOF
import sys
sys.path.insert(0, '/app')
from tpm_cert_manager import TPMCertificateManager

mgr = TPMCertificateManager('$CERT_DIR')
verified = mgr.verify_node_certificate($NODE_ID)
print(f"Node $NODE_ID certificate verified: {verified}")

if verified:
    print("Node is ready for secure communication")
else:
    print("WARNING: Node certificate verification failed")
EOF
fi

log "Bootstrap complete for Node $NODE_ID"
log "Certificates available at: $CERT_DIR"

# If a command is provided, run it
if [ $# -gt 0 ]; then
    log "Running command: $*"
    exec "$@"
fi
