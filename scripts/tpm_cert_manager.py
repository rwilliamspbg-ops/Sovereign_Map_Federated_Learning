"""
TPM-Inspired Trust & Verification System for Sovereign Map
Generates and manages node certificates, enables secure node-to-node communication
"""

import os
import json
import hashlib
import hmac
import time
from datetime import datetime, timedelta
from pathlib import Path
from typing import Dict, Optional, Tuple
import logging

from cryptography import x509
from cryptography.x509.oid import NameOID, ExtensionOID
from cryptography.hazmat.primitives import hashes, serialization
from cryptography.hazmat.primitives.asymmetric import rsa, padding
from cryptography.hazmat.backends import default_backend

# Configure logging
logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)


class TPMCertificateManager:
    """Manages node certificates, trust chains, and verification."""

    def __init__(self, cert_dir: str = "/etc/sovereign/certs"):
        self.cert_dir = Path(cert_dir)
        self.cert_dir.mkdir(parents=True, exist_ok=True)
        os.chmod(self.cert_dir, 0o700)

        self.ca_cert_path = self.cert_dir / "ca-cert.pem"
        self.ca_key_path = self.cert_dir / "ca-key.pem"
        self.trust_store_path = self.cert_dir / "trust-store.json"
        self.crl_path = self.cert_dir / "crl.json"  # Certificate Revocation List

        self.ca_cert = None
        self.ca_key = None
        self.trust_store = {}
        self.crl = set()

        self._load_or_create_ca()
        self._load_trust_store()
        self._enforce_permissions()

    def _enforce_permissions(self):
        """Enforce restrictive permissions on sensitive certificate assets."""
        try:
            os.chmod(self.cert_dir, 0o700)
        except OSError:
            pass

        for path, mode in (
            (self.ca_key_path, 0o600),
            (self.ca_cert_path, 0o644),
            (self.trust_store_path, 0o600),
            (self.crl_path, 0o600),
        ):
            if path.exists():
                try:
                    os.chmod(path, mode)
                except OSError:
                    logger.warning(f"Could not update permissions for {path}")

    def _load_or_create_ca(self):
        """Load existing CA or create a new one."""
        if self.ca_key_path.exists() and self.ca_cert_path.exists():
            logger.info("Loading existing CA certificate")
            with open(self.ca_key_path, "rb") as f:
                self.ca_key = serialization.load_pem_private_key(
                    f.read(), password=None, backend=default_backend()
                )
            with open(self.ca_cert_path, "rb") as f:
                self.ca_cert = x509.load_pem_x509_certificate(
                    f.read(), backend=default_backend()
                )
        else:
            logger.info("Generating new CA certificate")
            self._generate_ca()

    def _generate_ca(self):
        """Generate root CA certificate and key."""
        # Generate private key
        self.ca_key = rsa.generate_private_key(
            public_exponent=65537, key_size=4096, backend=default_backend()
        )

        # Create certificate
        subject = issuer = x509.Name(
            [
                x509.NameAttribute(NameOID.COUNTRY_NAME, "US"),
                x509.NameAttribute(NameOID.STATE_OR_PROVINCE_NAME, "CA"),
                x509.NameAttribute(NameOID.LOCALITY_NAME, "San Francisco"),
                x509.NameAttribute(NameOID.ORGANIZATION_NAME, "Sovereign Map"),
                x509.NameAttribute(NameOID.COMMON_NAME, "Sovereign Map Root CA"),
            ]
        )

        self.ca_cert = (
            x509.CertificateBuilder()
            .subject_name(subject)
            .issuer_name(issuer)
            .public_key(self.ca_key.public_key())
            .serial_number(x509.random_serial_number())
            .not_valid_before(datetime.utcnow())
            .not_valid_after(datetime.utcnow() + timedelta(days=3650))
            .add_extension(
                x509.BasicConstraints(ca=True, path_length=None),
                critical=True,
            )
            .add_extension(
                x509.KeyUsage(
                    digital_signature=True,
                    content_commitment=False,
                    key_encipherment=False,
                    data_encipherment=False,
                    key_agreement=False,
                    key_cert_sign=True,
                    crl_sign=True,
                    encipher_only=False,
                    decipher_only=False,
                ),
                critical=True,
            )
            .sign(self.ca_key, hashes.SHA256(), backend=default_backend())
        )

        # Save to disk
        with open(self.ca_key_path, "wb") as f:
            f.write(
                self.ca_key.private_bytes(
                    encoding=serialization.Encoding.PEM,
                    format=serialization.PrivateFormat.PKCS8,
                    encryption_algorithm=serialization.NoEncryption(),
                )
            )
        os.chmod(self.ca_key_path, 0o600)

        with open(self.ca_cert_path, "wb") as f:
            f.write(self.ca_cert.public_bytes(serialization.Encoding.PEM))
        os.chmod(self.ca_cert_path, 0o644)

        logger.info(f"CA certificate created and saved to {self.cert_dir}")

    def generate_node_cert(self, node_id: int, node_name: str) -> Tuple[str, str]:
        """Generate a client certificate for a node."""
        # Generate node private key
        node_key = rsa.generate_private_key(
            public_exponent=65537, key_size=2048, backend=default_backend()
        )

        # Create certificate
        subject = x509.Name(
            [
                x509.NameAttribute(NameOID.COUNTRY_NAME, "US"),
                x509.NameAttribute(NameOID.STATE_OR_PROVINCE_NAME, "CA"),
                x509.NameAttribute(NameOID.ORGANIZATION_NAME, "Sovereign Map"),
                x509.NameAttribute(NameOID.ORGANIZATIONAL_UNIT_NAME, "Node"),
                x509.NameAttribute(NameOID.COMMON_NAME, f"node-{node_id}"),
            ]
        )

        cert = (
            x509.CertificateBuilder()
            .subject_name(subject)
            .issuer_name(self.ca_cert.issuer)
            .public_key(node_key.public_key())
            .serial_number(x509.random_serial_number())
            .not_valid_before(datetime.utcnow())
            .not_valid_after(datetime.utcnow() + timedelta(days=365))
            .add_extension(
                x509.SubjectAlternativeName(
                    [
                        x509.DNSName(f"node-{node_id}"),
                        x509.DNSName(f"node-{node_id}.sovereign-network"),
                        x509.IPAddress(
                            __import__("ipaddress").IPv4Address(
                                f"172.25.0.{node_id % 254 + 2}"
                            )
                        ),
                    ]
                ),
                critical=False,
            )
            .add_extension(
                x509.KeyUsage(
                    digital_signature=True,
                    content_commitment=True,
                    key_encipherment=True,
                    data_encipherment=False,
                    key_agreement=False,
                    key_cert_sign=False,
                    crl_sign=False,
                    encipher_only=False,
                    decipher_only=False,
                ),
                critical=True,
            )
            .sign(self.ca_key, hashes.SHA256(), backend=default_backend())
        )

        # Save node certificate and key
        cert_path = self.cert_dir / f"node-{node_id}-cert.pem"
        key_path = self.cert_dir / f"node-{node_id}-key.pem"

        with open(cert_path, "wb") as f:
            f.write(cert.public_bytes(serialization.Encoding.PEM))
        os.chmod(cert_path, 0o644)

        with open(key_path, "wb") as f:
            f.write(
                node_key.private_bytes(
                    encoding=serialization.Encoding.PEM,
                    format=serialization.PrivateFormat.PKCS8,
                    encryption_algorithm=serialization.NoEncryption(),
                )
            )
        os.chmod(key_path, 0o600)

        # Store in trust store
        self.trust_store[f"node-{node_id}"] = {
            "node_id": node_id,
            "name": node_name,
            "cert_path": str(cert_path),
            "public_key": cert.public_key()
            .public_bytes(
                encoding=serialization.Encoding.PEM,
                format=serialization.PublicFormat.SubjectPublicKeyInfo,
            )
            .decode(),
            "serial": hex(cert.serial_number),
            "created_at": datetime.utcnow().isoformat(),
            "expires_at": cert.not_valid_after.isoformat(),
            "verified": False,
        }

        self._save_trust_store()
        logger.info(f"Generated certificate for node-{node_id} ({node_name})")

        return str(cert_path), str(key_path)

    def verify_node_certificate(self, node_id: int) -> bool:
        """Verify a node certificate against the CA."""
        cert_path = self.cert_dir / f"node-{node_id}-cert.pem"

        if not cert_path.exists():
            logger.warning(f"Certificate not found for node-{node_id}")
            return False

        try:
            with open(cert_path, "rb") as f:
                node_cert = x509.load_pem_x509_certificate(
                    f.read(), backend=default_backend()
                )

            # Verify certificate signature
            self.ca_key.public_key().verify(
                node_cert.signature,
                node_cert.tbs_certificate_bytes,
                padding.PKCS1v15(),
                node_cert.signature_hash_algorithm,
            )

            # Check if certificate is in CRL
            if hex(node_cert.serial_number) in self.crl:
                logger.warning(f"Certificate for node-{node_id} is revoked")
                return False

            # Check expiration
            if datetime.utcnow() > node_cert.not_valid_after:
                logger.warning(f"Certificate for node-{node_id} has expired")
                return False

            self.trust_store[f"node-{node_id}"]["verified"] = True
            self._save_trust_store()
            logger.info(f"Certificate verified for node-{node_id}")
            return True

        except Exception as e:
            logger.error(f"Failed to verify certificate for node-{node_id}: {e}")
            return False

    def revoke_node_certificate(self, node_id: int) -> bool:
        """Revoke a node certificate."""
        cert_path = self.cert_dir / f"node-{node_id}-cert.pem"

        if not cert_path.exists():
            logger.warning(f"Certificate not found for node-{node_id}")
            return False

        try:
            with open(cert_path, "rb") as f:
                node_cert = x509.load_pem_x509_certificate(
                    f.read(), backend=default_backend()
                )

            self.crl.add(hex(node_cert.serial_number))
            self._save_crl()
            logger.warning(f"Certificate revoked for node-{node_id}")
            return True

        except Exception as e:
            logger.error(f"Failed to revoke certificate for node-{node_id}: {e}")
            return False

    def _load_trust_store(self):
        """Load trust store from disk."""
        if self.trust_store_path.exists():
            with open(self.trust_store_path, "r") as f:
                self.trust_store = json.load(f)
            logger.info(f"Loaded trust store with {len(self.trust_store)} entries")

        if self.crl_path.exists():
            with open(self.crl_path, "r") as f:
                self.crl = set(json.load(f))

    def _save_trust_store(self):
        """Save trust store to disk."""
        with open(self.trust_store_path, "w") as f:
            json.dump(self.trust_store, f, indent=2)
        os.chmod(self.trust_store_path, 0o600)

    def _save_crl(self):
        """Save CRL to disk."""
        with open(self.crl_path, "w") as f:
            json.dump(list(self.crl), f, indent=2)
        os.chmod(self.crl_path, 0o600)

    def get_trust_report(self) -> Dict:
        """Get a report of all nodes and their trust status."""
        return {
            "ca_cert": str(self.ca_cert_path),
            "total_nodes": len(self.trust_store),
            "verified_nodes": sum(
                1 for n in self.trust_store.values() if n.get("verified")
            ),
            "revoked_count": len(self.crl),
            "nodes": self.trust_store,
        }


class NodeAuthenticator:
    """Handles node-to-node authentication and message signing."""

    def __init__(self, node_id: int, cert_manager: TPMCertificateManager):
        self.node_id = node_id
        self.cert_manager = cert_manager
        self.node_key_path = self.cert_manager.cert_dir / f"node-{node_id}-key.pem"
        self.node_cert_path = self.cert_manager.cert_dir / f"node-{node_id}-cert.pem"

        # Load node's private key
        if not self.node_key_path.exists():
            logger.warning(f"Private key not found for node {node_id}")
            self.node_key = None
        else:
            with open(self.node_key_path, "rb") as f:
                self.node_key = serialization.load_pem_private_key(
                    f.read(), password=None, backend=default_backend()
                )

    def sign_message(self, message: bytes) -> bytes:
        """Sign a message with the node's private key."""
        if not self.node_key:
            raise ValueError(f"Node {self.node_id} key not available")

        signature = self.node_key.sign(
            message,
            padding.PSS(
                mgf=padding.MGF1(hashes.SHA256()), salt_length=padding.PSS.MAX_LENGTH
            ),
            hashes.SHA256(),
        )
        return signature

    def verify_message(
        self, message: bytes, signature: bytes, peer_node_id: int
    ) -> bool:
        """Verify a message signature from a peer node."""
        peer_cert_path = self.cert_manager.cert_dir / f"node-{peer_node_id}-cert.pem"

        if not peer_cert_path.exists():
            logger.warning(f"Certificate not found for peer node {peer_node_id}")
            return False

        try:
            with open(peer_cert_path, "rb") as f:
                peer_cert = x509.load_pem_x509_certificate(
                    f.read(), backend=default_backend()
                )

            peer_public_key = peer_cert.public_key()
            peer_public_key.verify(
                signature,
                message,
                padding.PSS(
                    mgf=padding.MGF1(hashes.SHA256()),
                    salt_length=padding.PSS.MAX_LENGTH,
                ),
                hashes.SHA256(),
            )
            logger.info(f"Message signature verified from node {peer_node_id}")
            return True

        except Exception as e:
            logger.error(
                f"Failed to verify message signature from node {peer_node_id}: {e}"
            )
            return False

    def create_authenticated_message(self, data: Dict) -> Dict:
        """Create a message with signature and timestamp."""
        timestamp = int(time.time())
        message_data = json.dumps(
            {"from_node": self.node_id, "timestamp": timestamp, "data": data}
        ).encode()

        signature = self.sign_message(message_data)

        return {
            "from_node": self.node_id,
            "timestamp": timestamp,
            "data": data,
            "signature": signature.hex(),
            "message_hash": hashlib.sha256(message_data).hexdigest(),
        }

    def verify_authenticated_message(self, message: Dict) -> bool:
        """Verify an authenticated message."""
        try:
            # Reconstruct original message
            message_data = json.dumps(
                {
                    "from_node": message["from_node"],
                    "timestamp": message["timestamp"],
                    "data": message["data"],
                }
            ).encode()

            # Verify hash
            expected_hash = hashlib.sha256(message_data).hexdigest()
            if expected_hash != message["message_hash"]:
                logger.warning(
                    f"Message hash mismatch from node {message['from_node']}"
                )
                return False

            # Verify signature
            signature = bytes.fromhex(message["signature"])
            return self.verify_message(message_data, signature, message["from_node"])

        except Exception as e:
            logger.error(f"Failed to verify authenticated message: {e}")
            return False


if __name__ == "__main__":
    # Example usage
    mgr = TPMCertificateManager()

    # Generate certificates for 5 nodes
    for i in range(5):
        cert_path, key_path = mgr.generate_node_cert(i, f"Node-{i}")
        print(f"Generated cert for node {i}: {cert_path}")

    # Verify certificates
    for i in range(5):
        verified = mgr.verify_node_certificate(i)
        print(f"Node {i} verified: {verified}")

    # Get trust report
    report = mgr.get_trust_report()
    print("\n=== Trust Report ===")
    print(json.dumps(report, indent=2))

    # Demonstrate message authentication
    print("\n=== Message Authentication ===")
    auth_0 = NodeAuthenticator(0, mgr)
    auth_1 = NodeAuthenticator(1, mgr)

    msg = auth_0.create_authenticated_message({"action": "update", "round": 1})
    verified = auth_1.verify_authenticated_message(msg)
    print(f"Message from node 0 to node 1 verified: {verified}")
