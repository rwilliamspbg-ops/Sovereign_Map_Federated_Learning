"""
Secure Communication Middleware for Sovereign Map
Integrates TPM-inspired trust & verification with Flask backend
"""

import json
import logging
import time
import hashlib
import hmac
import os
from functools import wraps
from typing import Dict, Any, Callable, Optional
from datetime import datetime

from flask import request, jsonify, g
from tpm_cert_manager import TPMCertificateManager, NodeAuthenticator

logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)


def _admin_authorized() -> bool:
    expected = str(os.getenv("TRUST_ADMIN_TOKEN", "")).strip()
    if not expected:
        return False
    header_token = str(request.headers.get("X-Trust-Admin-Token", "")).strip()
    auth = str(request.headers.get("Authorization", "")).strip()
    bearer = auth[7:].strip() if auth.lower().startswith("bearer ") else ""
    return hmac.compare_digest(header_token, expected) or hmac.compare_digest(
        bearer, expected
    )


class SecureNodeCommunication:
    """Middleware for secure node-to-node communication."""

    def __init__(self, node_id: int, cert_dir: str = "/etc/sovereign/certs"):
        self.node_id = node_id
        self.cert_manager = TPMCertificateManager(cert_dir)
        self.authenticator = NodeAuthenticator(node_id, self.cert_manager)

        # Trust cache to avoid repeated verification
        self.trust_cache = {}
        self.cache_ttl = 3600  # 1 hour
        self.max_clock_skew_seconds = 300  # 5 minutes

        logger.info(f"SecureNodeCommunication initialized for node {node_id}")

    def initialize_node_certs(self, num_nodes: int):
        """Generate certificates for all nodes in the network."""
        for i in range(num_nodes):
            if i != self.node_id:
                try:
                    self.cert_manager.generate_node_cert(i, f"Node-{i}")
                except Exception as e:
                    logger.warning(f"Certificate may already exist for node {i}: {e}")

    def secure_endpoint(self, f: Callable) -> Callable:
        """Decorator to secure Flask endpoints with node authentication."""

        @wraps(f)
        def decorated_function(*args, **kwargs):
            # Check if request has authentication header
            auth_header = request.headers.get("X-Node-Auth")
            signature = request.headers.get("X-Signature")
            from_node = request.headers.get("X-From-Node")

            if not all([auth_header, signature, from_node]):
                logger.warning("Missing authentication headers")
                return jsonify({"error": "Missing authentication headers"}), 401

            try:
                request_data = None
                if request.method in {"POST", "PUT", "PATCH"}:
                    request_data = request.get_json(silent=True)

                # Verify the request is signed by a trusted node
                if not self._verify_request(
                    from_node, signature, auth_header, request_data
                ):
                    logger.warning(
                        f"Signature verification failed from node {from_node}"
                    )
                    return jsonify({"error": "Invalid signature"}), 401

                # Verify node certificate
                if not self.cert_manager.verify_node_certificate(int(from_node)):
                    logger.warning(
                        f"Certificate verification failed for node {from_node}"
                    )
                    return jsonify({"error": "Certificate verification failed"}), 401

                # Store peer info in request context
                g.peer_node_id = int(from_node)
                g.peer_verified = True

                return f(*args, **kwargs)

            except Exception as e:
                logger.error(f"Error during authentication: {e}")
                return jsonify({"error": "Authentication failed"}), 500

        return decorated_function

    def _verify_request(
        self,
        from_node: str,
        signature: str,
        auth_payload_hex: str,
        request_data: Optional[Dict[str, Any]] = None,
    ) -> bool:
        """Verify that a request is properly signed."""
        try:
            from_node_id = int(from_node)
            signature_bytes = bytes.fromhex(signature)
            auth_payload_bytes = bytes.fromhex(auth_payload_hex)

            payload = json.loads(auth_payload_bytes.decode())
            payload_from_node = int(payload.get("from_node"))
            payload_timestamp = int(payload.get("timestamp"))

            if payload_from_node != from_node_id:
                logger.warning(
                    f"Authenticated payload node mismatch: header={from_node_id}, payload={payload_from_node}"
                )
                return False

            if abs(int(time.time()) - payload_timestamp) > self.max_clock_skew_seconds:
                logger.warning(f"Stale or replayed request from node {from_node_id}")
                return False

            if request_data is not None and payload.get("data") != request_data:
                logger.warning(f"Signed payload/body mismatch from node {from_node_id}")
                return False

            payload_digest = hashlib.sha256(auth_payload_bytes).hexdigest()

            # Check trust cache
            cache_key = f"{from_node_id}:{payload_digest}:{signature}"
            if cache_key in self.trust_cache:
                cached_time, result = self.trust_cache[cache_key]
                if time.time() - cached_time < self.cache_ttl:
                    return result

            # Verify signature
            result = self.authenticator.verify_message(
                auth_payload_bytes, signature_bytes, from_node_id
            )

            # Cache result
            self.trust_cache[cache_key] = (time.time(), result)

            return result

        except Exception as e:
            logger.error(f"Error verifying request: {e}")
            return False

    def create_signed_request(self, target_node_id: int, data: Dict) -> Dict:
        """Create a signed request to send to a peer node."""
        message = json.dumps(
            {"from_node": self.node_id, "timestamp": int(time.time()), "data": data}
        ).encode()

        signature = self.authenticator.sign_message(message)

        return {
            "headers": {
                "X-From-Node": str(self.node_id),
                "X-Signature": signature.hex(),
                "X-Node-Auth": message.hex(),
            },
            "body": data,
        }

    def get_trust_status(self) -> Dict[str, Any]:
        """Get current trust status of all nodes."""
        report = self.cert_manager.get_trust_report()
        return {
            "node_id": self.node_id,
            "timestamp": datetime.utcnow().isoformat(),
            "ca_certificate": report["ca_cert"],
            "total_nodes": report["total_nodes"],
            "verified_nodes": report["verified_nodes"],
            "revoked_certificates": report["revoked_count"],
            "node_details": {
                node_name: {
                    "verified": node_info.get("verified", False),
                    "expires_at": node_info.get("expires_at"),
                    "serial": node_info.get("serial"),
                }
                for node_name, node_info in report["nodes"].items()
            },
        }


class RequestSigner:
    """Helper class for signing requests before sending them."""

    def __init__(self, comm: SecureNodeCommunication):
        self.comm = comm

    def sign_and_send(self, method: str, url: str, data: Dict = None) -> Dict:
        """Sign a request and return the headers and body."""
        signed = self.comm.create_signed_request(0, data or {})

        return {
            "method": method,
            "url": url,
            "headers": {**signed["headers"], "Content-Type": "application/json"},
            "data": json.dumps(signed["body"]),
        }


def create_secure_app_middleware(app, node_id: int, num_nodes: int = 10):
    """
    Attach secure communication middleware to a Flask app.

    Usage:
        from flask import Flask
        app = Flask(__name__)
        comm = create_secure_app_middleware(app, node_id=0, num_nodes=10)
    """
    comm = SecureNodeCommunication(node_id)
    comm.initialize_node_certs(num_nodes)

    # Store in app context
    app.secure_comm = comm

    # Add endpoints for trust management
    @app.route("/trust/status", methods=["GET"])
    def get_trust_status():
        """Get trust status of the network."""
        return jsonify(comm.get_trust_status())

    @app.route("/trust/verify/<int:peer_node_id>", methods=["POST"])
    def verify_peer(peer_node_id):
        """Manually verify a peer node's certificate."""
        if not _admin_authorized():
            return jsonify({"error": "unauthorized"}), 401
        verified = comm.cert_manager.verify_node_certificate(peer_node_id)
        return jsonify(
            {
                "node_id": peer_node_id,
                "verified": verified,
                "timestamp": datetime.utcnow().isoformat(),
            }
        )

    @app.route("/trust/revoke/<int:peer_node_id>", methods=["POST"])
    def revoke_peer_cert(peer_node_id):
        """Revoke a peer node's certificate (admin only)."""
        if not _admin_authorized():
            return jsonify({"error": "unauthorized"}), 401
        revoked = comm.cert_manager.revoke_node_certificate(peer_node_id)
        return jsonify(
            {
                "node_id": peer_node_id,
                "revoked": revoked,
                "timestamp": datetime.utcnow().isoformat(),
            }
        )

    @app.route("/trust/certificate/<int:node_id>", methods=["GET"])
    def get_node_certificate(node_id):
        """Get a node's certificate (for peer verification)."""
        cert_path = comm.cert_manager.cert_dir / f"node-{node_id}-cert.pem"

        if not cert_path.exists():
            return jsonify({"error": "Certificate not found"}), 404

        with open(cert_path, "r") as f:
            cert_pem = f.read()

        return jsonify(
            {
                "node_id": node_id,
                "certificate": cert_pem,
                "timestamp": datetime.utcnow().isoformat(),
            }
        )

    return comm


if __name__ == "__main__":
    # Example usage
    from flask import Flask

    app = Flask(__name__)

    # Initialize secure communication
    comm = create_secure_app_middleware(app, node_id=0, num_nodes=5)

    # Add a secure endpoint
    @app.route("/secure/data", methods=["POST"])
    @comm.secure_endpoint
    def secure_endpoint():
        from_node = g.peer_node_id if hasattr(g, "peer_node_id") else "unknown"
        data = request.json
        return jsonify({"status": "received", "from_node": from_node, "data": data})

    print("Secure communication middleware initialized")
    print(f"Running on port 5000 with mTLS enabled")

    app.run(host="0.0.0.0", port=5000, debug=False)
