"""
Tokenomics Metrics Exporter for Prometheus
----------------------------------------
Ingests minting/bridging telemetry from a JSON source and exposes stable
Prometheus metrics for dashboards and alerting.
"""

from __future__ import annotations

import json
import logging
import os
import time
from threading import Lock
from typing import Any, Dict

from flask import Flask, Response, request
from prometheus_client import (
    CONTENT_TYPE_LATEST,
    CollectorRegistry,
    Gauge,
    generate_latest,
)

logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)


class TokenomicsMetricsExporter:
    def __init__(self, source_file: str):
        self.source_file = source_file
        self.registry = CollectorRegistry()
        self._lock = Lock()

        self.mint_rate = Gauge(
            "tokenomics_mint_rate_per_min",
            "Token minting rate per minute",
            registry=self.registry,
        )
        self.token_supply_total = Gauge(
            "tokenomics_token_supply_total",
            "Total token supply",
            registry=self.registry,
        )
        self.bridge_inflow = Gauge(
            "tokenomics_bridge_inflow_per_min",
            "Bridge inflow per minute",
            registry=self.registry,
        )
        self.bridge_outflow = Gauge(
            "tokenomics_bridge_outflow_per_min",
            "Bridge outflow per minute",
            registry=self.registry,
        )
        self.bridge_net_flow = Gauge(
            "tokenomics_bridge_net_flow_per_min",
            "Bridge net flow per minute",
            registry=self.registry,
        )
        self.bridge_escrow_total = Gauge(
            "tokenomics_bridge_escrow_total",
            "Bridge escrow total",
            registry=self.registry,
        )
        self.bridge_collateral_ratio = Gauge(
            "tokenomics_bridge_collateral_ratio_percent",
            "Bridge collateral ratio percent",
            registry=self.registry,
        )
        self.bridge_settlement_share = Gauge(
            "tokenomics_bridge_settlement_share_percent",
            "Bridge settlement share percent",
            registry=self.registry,
        )
        self.bridge_volume_24h = Gauge(
            "tokenomics_bridge_volume_24h",
            "Bridge volume over trailing 24h",
            registry=self.registry,
        )
        self.last_update_ts = Gauge(
            "tokenomics_last_update_timestamp_seconds",
            "Unix timestamp of last tokenomics update",
            registry=self.registry,
        )

    @staticmethod
    def _safe_float(data: Dict[str, Any], key: str, default: float = 0.0) -> float:
        try:
            return float(data.get(key, default))
        except (TypeError, ValueError):
            return default

    def _apply_payload(self, payload: Dict[str, Any]):
        mint_rate = self._safe_float(payload, "mint_rate_per_min")
        supply = self._safe_float(payload, "token_supply_total")
        inflow = self._safe_float(payload, "bridge_inflow_per_min")
        outflow = self._safe_float(payload, "bridge_outflow_per_min")
        escrow = self._safe_float(payload, "bridge_escrow_total")
        collateral = self._safe_float(payload, "bridge_collateral_ratio_percent")
        settlement_share = self._safe_float(
            payload,
            "bridge_settlement_share_percent",
            (inflow / mint_rate * 100.0) if mint_rate > 0 else 0.0,
        )
        volume_24h = self._safe_float(payload, "bridge_volume_24h", inflow * 1440.0)

        self.mint_rate.set(max(0.0, mint_rate))
        self.token_supply_total.set(max(0.0, supply))
        self.bridge_inflow.set(max(0.0, inflow))
        self.bridge_outflow.set(max(0.0, outflow))
        self.bridge_net_flow.set(inflow - outflow)
        self.bridge_escrow_total.set(max(0.0, escrow))
        self.bridge_collateral_ratio.set(max(0.0, collateral))
        self.bridge_settlement_share.set(max(0.0, settlement_share))
        self.bridge_volume_24h.set(max(0.0, volume_24h))
        self.last_update_ts.set(time.time())

    def load_source_file(self):
        try:
            with open(self.source_file, "r", encoding="utf-8") as handle:
                payload = json.load(handle)
            if isinstance(payload, dict):
                self._apply_payload(payload)
        except FileNotFoundError:
            logger.warning("Tokenomics source file missing: %s", self.source_file)
        except json.JSONDecodeError as exc:
            logger.error("Invalid tokenomics source JSON: %s", exc)

    def ingest_event(self, payload: Dict[str, Any]):
        with self._lock:
            self._apply_payload(payload)

    def generate_metrics(self) -> bytes:
        self.load_source_file()
        return generate_latest(self.registry)


def create_app(source_file: str):
    app = Flask(__name__)
    exporter = TokenomicsMetricsExporter(source_file=source_file)

    @app.route("/metrics", methods=["GET"])
    def metrics():
        return Response(exporter.generate_metrics(), mimetype=CONTENT_TYPE_LATEST)

    @app.route("/health", methods=["GET"])
    def health():
        return {"status": "healthy", "source_file": source_file}

    @app.route("/event/tokenomics", methods=["POST"])
    def event_tokenomics():
        payload = request.get_json(silent=True) or {}
        if not isinstance(payload, dict):
            return {"status": "error", "error": "payload must be an object"}, 400
        exporter.ingest_event(payload)
        return {"status": "ok"}

    return app


if __name__ == "__main__":
    source_file = os.getenv(
        "TOKENOMICS_SOURCE_FILE", "/app/data/tokenomics-telemetry.json"
    )
    port = int(os.getenv("TOKENOMICS_METRICS_PORT", "9105"))
    logger.info("Tokenomics Metrics Exporter running on port %s", port)
    app = create_app(source_file)
    app.run(host="0.0.0.0", port=port, debug=False)
