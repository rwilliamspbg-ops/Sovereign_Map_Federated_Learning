"""
TPM Trust Metrics Exporter for Prometheus
Exports certificate, trust chain, and security metrics to Prometheus
"""

import json
import logging
import time
from datetime import datetime
from pathlib import Path
from typing import Dict, Optional
from threading import Thread

from prometheus_client import (
    Counter,
    Gauge,
    Histogram,
    CollectorRegistry,
    generate_latest,
    CONTENT_TYPE_LATEST,
)
from flask import Flask, Response
from tpm_cert_manager import TPMCertificateManager, NodeAuthenticator

logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)


class TPMMetricsExporter:
    """Exports TPM trust metrics to Prometheus."""

    def __init__(self, node_id: int, cert_dir: str = "/etc/sovereign/certs"):
        self.node_id = node_id
        self.cert_manager = TPMCertificateManager(cert_dir)
        self.authenticator = NodeAuthenticator(node_id, self.cert_manager)

        # Create registry
        self.registry = CollectorRegistry()

        # Certificate Metrics
        self.certs_total = Gauge(
            "tpm_certificates_total",
            "Total number of issued certificates",
            registry=self.registry,
        )

        self.certs_verified = Gauge(
            "tpm_certificates_verified_total",
            "Number of verified certificates",
            registry=self.registry,
        )

        self.certs_revoked = Gauge(
            "tpm_certificates_revoked_total",
            "Number of revoked certificates",
            registry=self.registry,
        )

        self.cert_expiry_seconds = Gauge(
            "tpm_certificate_expiry_seconds",
            "Seconds until certificate expiration",
            ["node_id"],
            registry=self.registry,
        )

        self.cert_age_seconds = Gauge(
            "tpm_certificate_age_seconds",
            "Age of certificate in seconds",
            ["node_id"],
            registry=self.registry,
        )

        # Trust Chain Metrics
        self.trust_chain_valid = Gauge(
            "tpm_trust_chain_valid",
            "Is the trust chain valid (1=valid, 0=invalid)",
            ["node_id"],
            registry=self.registry,
        )

        self.trust_verification_time = Histogram(
            "tpm_trust_verification_duration_seconds",
            "Time to verify trust chain",
            registry=self.registry,
        )

        self.trust_verification_failures = Counter(
            "tpm_trust_verification_failures_total",
            "Total trust verification failures",
            registry=self.registry,
        )

        # Message Authentication Metrics
        self.messages_signed_total = Counter(
            "tpm_messages_signed_total",
            "Total messages signed",
            ["node_id"],
            registry=self.registry,
        )

        self.messages_verified_total = Counter(
            "tpm_messages_verified_total",
            "Total messages verified",
            ["node_id", "from_node_id"],
            registry=self.registry,
        )

        self.signature_verification_failures = Counter(
            "tpm_signature_verification_failures_total",
            "Total signature verification failures",
            ["from_node_id"],
            registry=self.registry,
        )

        self.message_verification_time = Histogram(
            "tpm_message_verification_duration_seconds",
            "Time to verify message signature",
            registry=self.registry,
        )

        # Node Trust Status Metrics
        self.node_trust_score = Gauge(
            "tpm_node_trust_score",
            "Trust score for a node (0-100)",
            ["node_id"],
            registry=self.registry,
        )

        self.node_cert_valid = Gauge(
            "tpm_node_certificate_valid",
            "Is node certificate valid (1=valid, 0=invalid)",
            ["node_id"],
            registry=self.registry,
        )

        self.node_cert_revoked = Gauge(
            "tpm_node_certificate_revoked",
            "Is node certificate revoked (1=revoked, 0=valid)",
            ["node_id"],
            registry=self.registry,
        )

        # System Metrics
        self.ca_cert_valid = Gauge(
            "tpm_ca_certificate_valid",
            "Is CA certificate valid (1=valid, 0=invalid)",
            registry=self.registry,
        )

        self.crl_size = Gauge(
            "tpm_crl_size",
            "Number of revoked certificates in CRL",
            registry=self.registry,
        )

        self.trust_cache_hits = Counter(
            "tpm_trust_cache_hits_total",
            "Total trust cache hits",
            registry=self.registry,
        )

        self.trust_cache_misses = Counter(
            "tpm_trust_cache_misses_total",
            "Total trust cache misses",
            registry=self.registry,
        )

        logger.info(f"TPMMetricsExporter initialized for node {node_id}")

    def update_metrics(self):
        """Update all metrics from current trust state."""
        try:
            # Get trust report
            report = self.cert_manager.get_trust_report()

            # Update certificate metrics
            self.certs_total.set(report["total_nodes"])
            self.certs_verified.set(report["verified_nodes"])
            self.certs_revoked.set(report["revoked_count"])
            self.crl_size.set(report["revoked_count"])

            # Update CA metrics
            self.ca_cert_valid.set(1 if self.cert_manager.ca_cert else 0)

            # Update node-specific metrics
            for node_name, node_info in report["nodes"].items():
                try:
                    node_id = node_info["node_id"]

                    # Parse expiry time
                    expires_at = datetime.fromisoformat(
                        node_info["expires_at"].replace("Z", "+00:00")
                    )
                    seconds_until_expiry = (
                        expires_at - datetime.utcnow()
                    ).total_seconds()
                    self.cert_expiry_seconds.labels(node_id=node_id).set(
                        max(0, seconds_until_expiry)
                    )

                    # Parse creation time
                    created_at = datetime.fromisoformat(
                        node_info["created_at"].replace("Z", "+00:00")
                    )
                    age_seconds = (datetime.utcnow() - created_at).total_seconds()
                    self.cert_age_seconds.labels(node_id=node_id).set(age_seconds)

                    # Trust status
                    is_verified = 1 if node_info.get("verified") else 0
                    self.trust_chain_valid.labels(node_id=node_id).set(is_verified)

                    # Node certificate status
                    is_revoked = (
                        1 if node_info["serial"] in self.cert_manager.crl else 0
                    )
                    self.node_cert_revoked.labels(node_id=node_id).set(is_revoked)
                    self.node_cert_valid.labels(node_id=node_id).set(1 - is_revoked)

                    # Calculate trust score
                    trust_score = 100
                    if is_revoked:
                        trust_score = 0
                    elif not is_verified:
                        trust_score = 50
                    elif seconds_until_expiry < 86400 * 30:  # Less than 30 days
                        trust_score = 70
                    self.node_trust_score.labels(node_id=node_id).set(trust_score)

                except Exception as e:
                    logger.error(f"Error updating metrics for {node_name}: {e}")

        except Exception as e:
            logger.error(f"Error updating metrics: {e}")

    def record_message_signed(self, from_node_id: int):
        """Record a signed message."""
        self.messages_signed_total.labels(node_id=from_node_id).inc()

    def record_message_verified(self, from_node_id: int, to_node_id: int = None):
        """Record a verified message."""
        to_node_id = to_node_id or self.node_id
        self.messages_verified_total.labels(
            node_id=to_node_id, from_node_id=from_node_id
        ).inc()

    def record_signature_failure(self, from_node_id: int):
        """Record a signature verification failure."""
        self.signature_verification_failures.labels(from_node_id=from_node_id).inc()

    def record_verification_time(self, duration_seconds: float):
        """Record trust verification time."""
        self.trust_verification_time.observe(duration_seconds)

    def record_message_verification_time(self, duration_seconds: float):
        """Record message verification time."""
        self.message_verification_time.observe(duration_seconds)

    def record_verification_failure(self):
        """Record a verification failure."""
        self.trust_verification_failures.inc()

    def get_metrics(self) -> bytes:
        """Get metrics in Prometheus format."""
        self.update_metrics()
        return generate_latest(self.registry)


def create_metrics_app(node_id: int, cert_dir: str = "/etc/sovereign/certs") -> tuple:
    """
    Create a Flask app that exports TPM metrics.

    Returns: (app, exporter)
    """
    app = Flask(__name__)
    exporter = TPMMetricsExporter(node_id, cert_dir)

    @app.route("/metrics", methods=["GET"])
    def metrics():
        """Prometheus metrics endpoint."""
        return Response(exporter.get_metrics(), mimetype=CONTENT_TYPE_LATEST)

    @app.route("/metrics/summary", methods=["GET"])
    def metrics_summary():
        """Get trust metrics summary as JSON."""
        exporter.update_metrics()
        report = exporter.cert_manager.get_trust_report()

        return {
            "node_id": node_id,
            "timestamp": datetime.utcnow().isoformat(),
            "certificates": {
                "total": report["total_nodes"],
                "verified": report["verified_nodes"],
                "revoked": report["revoked_count"],
            },
            "ca": {
                "valid": exporter.ca_cert_valid._value.get() == 1,
                "certificate_path": str(exporter.cert_manager.ca_cert_path),
            },
        }

    @app.route("/health", methods=["GET"])
    def health():
        """Health check."""
        return {
            "status": "healthy",
            "node_id": node_id,
            "certificates_loaded": len(exporter.cert_manager.trust_store),
        }

    return app, exporter


if __name__ == "__main__":
    app, exporter = create_metrics_app(node_id=0)
    logger.info("TPM Metrics Exporter running on port 9091")
    app.run(host="0.0.0.0", port=9091, debug=False)
