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
    def __init__(self, source_file: str, state_file: str | None = None):
        self.source_file = self._resolve_source_file(source_file)
        self.state_file = self._resolve_state_file(state_file)
        self.persistence_file = self._select_persistence_file()
        self.registry = CollectorRegistry()
        self._lock = Lock()
        self._last_payload: Dict[str, Any] = {}

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
        self.token_supply_minted = Gauge(
            "tokenomics_token_supply_minted",
            "Circulating minted token supply",
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
        self.bridge_transfers_total = Gauge(
            "tokenomics_bridge_transfers_total",
            "Total bridge transfer executions observed by tokenomics telemetry",
            registry=self.registry,
        )
        self.bridge_routes_active = Gauge(
            "tokenomics_bridge_routes_active",
            "Active bridge policy routes",
            registry=self.registry,
        )
        self.validator_count = Gauge(
            "tokenomics_validator_count",
            "Active validator count for the token economy",
            registry=self.registry,
        )
        self.chain_height = Gauge(
            "tokenomics_chain_height",
            "Observed blockchain height exposed to tokenomics telemetry",
            registry=self.registry,
        )
        self.fl_verification_ratio = Gauge(
            "tokenomics_fl_verification_ratio",
            "Federated learning proof verification ratio as a 0-1 value",
            registry=self.registry,
        )
        self.fl_average_confidence_bps = Gauge(
            "tokenomics_fl_average_confidence_bps",
            "Average FL proof verification confidence in basis points",
            registry=self.registry,
        )
        self.stake_participation_ratio = Gauge(
            "tokenomics_stake_participation_ratio",
            "Stake participation ratio as a 0-1 value",
            registry=self.registry,
        )
        self.stake_concentration_gini = Gauge(
            "tokenomics_stake_concentration_gini",
            "Stake concentration gini coefficient as a 0-1 value",
            registry=self.registry,
        )
        self.unique_wallets_count = Gauge(
            "tokenomics_unique_wallets_count",
            "Unique wallet count",
            registry=self.registry,
        )
        self.wallet_average_balance = Gauge(
            "tokenomics_wallet_average_balance",
            "Average wallet balance",
            registry=self.registry,
        )
        self.top_10_holder_concentration = Gauge(
            "tokenomics_top_10_holder_concentration",
            "Top 10 holder concentration as a 0-1 value",
            registry=self.registry,
        )
        self.wallet_liquidity_ratio = Gauge(
            "tokenomics_wallet_liquidity_ratio",
            "Wallet liquidity ratio as a 0-1 value",
            registry=self.registry,
        )
        self.wallets_by_balance_bucket_large = Gauge(
            "tokenomics_wallets_by_balance_bucket_large",
            "Wallet count in the large balance bucket",
            registry=self.registry,
        )
        self.wallets_by_balance_bucket_medium = Gauge(
            "tokenomics_wallets_by_balance_bucket_medium",
            "Wallet count in the medium balance bucket",
            registry=self.registry,
        )
        self.wallets_by_balance_bucket_small = Gauge(
            "tokenomics_wallets_by_balance_bucket_small",
            "Wallet count in the small balance bucket",
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

    @staticmethod
    def _safe_optional_float(data: Dict[str, Any], *keys: str) -> float | None:
        for key in keys:
            if key not in data:
                continue
            try:
                return float(data[key])
            except (TypeError, ValueError):
                continue
        return None

    @staticmethod
    def _set_if_present(gauge: Gauge, value: float | None, minimum: float = 0.0):
        if value is None:
            return
        gauge.set(max(minimum, value))

    @staticmethod
    def _resolve_source_file(source_file: str) -> str:
        expanded = os.path.expanduser(source_file)
        if os.path.isdir(expanded):
            resolved = os.path.join(expanded, "tokenomics-telemetry.json")
            logger.warning(
                "TOKENOMICS_SOURCE_FILE points to a directory; using %s", resolved
            )
            return resolved
        return expanded

    @staticmethod
    def _resolve_state_file(state_file: str | None) -> str | None:
        if not state_file:
            return None
        return os.path.expanduser(state_file)

    @staticmethod
    def _is_writable_path(path: str) -> bool:
        if os.path.isdir(path):
            return os.access(path, os.W_OK)
        if os.path.exists(path):
            return os.access(path, os.W_OK)
        parent = os.path.dirname(path) or "."
        return os.path.exists(parent) and os.access(parent, os.W_OK)

    def _select_persistence_file(self) -> str | None:
        candidates = []
        if self.state_file:
            candidates.append(self.state_file)
        candidates.append(self.source_file)

        for candidate in candidates:
            if self._is_writable_path(candidate):
                if candidate != self.source_file:
                    logger.info(
                        "Persisting tokenomics payloads to writable state file %s",
                        candidate,
                    )
                return candidate

        logger.warning(
            "Tokenomics payload persistence disabled: no writable path available "
            "(source=%s, state=%s)",
            self.source_file,
            self.state_file,
        )
        return None

    def _apply_payload(self, payload: Dict[str, Any]):
        self._last_payload = dict(payload)
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
        bridge_transfers_total = self._safe_optional_float(
            payload,
            "bridge_transfers_total",
            "bridge_transfer_count",
            "bridge_total_transfers",
        )
        bridge_routes_active = self._safe_optional_float(
            payload,
            "bridge_routes_active",
            "bridge_policy_routes_active",
        )
        chain_height = self._safe_optional_float(
            payload,
            "chain_height",
            "blockchain_height",
            "ledger_height",
        )
        fl_verification_ratio = self._safe_optional_float(
            payload,
            "fl_verification_ratio",
            "verification_ratio",
        )
        fl_average_confidence_bps = self._safe_optional_float(
            payload,
            "fl_average_confidence_bps",
            "fl_verification_confidence_bps",
            "verification_confidence_bps",
        )
        minted_supply = self._safe_optional_float(
            payload,
            "token_supply_minted",
            "circulating_supply",
            "circulating_token_supply",
        )
        validator_count = self._safe_optional_float(
            payload,
            "validator_count",
            "validators_count",
            "active_validators",
        )
        stake_participation_ratio = self._safe_optional_float(
            payload,
            "stake_participation_ratio",
            "validator_participation_ratio",
        )
        stake_concentration_gini = self._safe_optional_float(
            payload,
            "stake_concentration_gini",
            "holder_concentration_gini",
        )
        unique_wallets_count = self._safe_optional_float(
            payload,
            "unique_wallets_count",
            "wallet_count",
            "unique_holders_count",
        )
        wallet_average_balance = self._safe_optional_float(
            payload,
            "wallet_average_balance",
            "average_wallet_balance",
            "avg_wallet_balance",
        )
        top_10_holder_concentration = self._safe_optional_float(
            payload,
            "top_10_holder_concentration",
            "top_10_holder_concentration_ratio",
            "top_10_holder_concentration_percent",
        )
        wallet_liquidity_ratio = self._safe_optional_float(
            payload,
            "wallet_liquidity_ratio",
            "liquidity_ratio",
            "wallet_liquidity_ratio_percent",
        )
        wallets_bucket_large = self._safe_optional_float(
            payload,
            "wallets_by_balance_bucket_large",
            "wallet_balance_bucket_large",
        )
        wallets_bucket_medium = self._safe_optional_float(
            payload,
            "wallets_by_balance_bucket_medium",
            "wallet_balance_bucket_medium",
        )
        wallets_bucket_small = self._safe_optional_float(
            payload,
            "wallets_by_balance_bucket_small",
            "wallet_balance_bucket_small",
        )

        self.mint_rate.set(max(0.0, mint_rate))
        self.token_supply_total.set(max(0.0, supply))
        self.bridge_inflow.set(max(0.0, inflow))
        self.bridge_outflow.set(max(0.0, outflow))
        self.bridge_net_flow.set(inflow - outflow)
        self.bridge_escrow_total.set(max(0.0, escrow))
        self.bridge_collateral_ratio.set(max(0.0, collateral))
        self.bridge_settlement_share.set(max(0.0, settlement_share))
        self.bridge_volume_24h.set(max(0.0, volume_24h))
        self._set_if_present(self.bridge_transfers_total, bridge_transfers_total)
        self._set_if_present(self.bridge_routes_active, bridge_routes_active)
        self._set_if_present(self.chain_height, chain_height)
        self._set_if_present(self.fl_verification_ratio, fl_verification_ratio)
        self._set_if_present(self.fl_average_confidence_bps, fl_average_confidence_bps)
        self._set_if_present(self.token_supply_minted, minted_supply)
        self._set_if_present(self.validator_count, validator_count)
        self._set_if_present(self.stake_participation_ratio, stake_participation_ratio)
        self._set_if_present(self.stake_concentration_gini, stake_concentration_gini)
        self._set_if_present(self.unique_wallets_count, unique_wallets_count)
        self._set_if_present(self.wallet_average_balance, wallet_average_balance)
        self._set_if_present(
            self.top_10_holder_concentration, top_10_holder_concentration
        )
        self._set_if_present(self.wallet_liquidity_ratio, wallet_liquidity_ratio)
        self._set_if_present(self.wallets_by_balance_bucket_large, wallets_bucket_large)
        self._set_if_present(
            self.wallets_by_balance_bucket_medium, wallets_bucket_medium
        )
        self._set_if_present(self.wallets_by_balance_bucket_small, wallets_bucket_small)
        self.last_update_ts.set(time.time())

    def _persist_payload(self, payload: Dict[str, Any]):
        if not self.persistence_file:
            return
        directory = os.path.dirname(self.persistence_file)
        if directory:
            os.makedirs(directory, exist_ok=True)
        temp_file = f"{self.persistence_file}.tmp"
        with open(temp_file, "w", encoding="utf-8") as handle:
            json.dump(payload, handle, indent=2)
        os.replace(temp_file, self.persistence_file)

    def load_source_file(self):
        try:
            with open(self.source_file, "r", encoding="utf-8") as handle:
                payload = json.load(handle)
            if isinstance(payload, dict):
                with self._lock:
                    self._apply_payload(payload)
        except FileNotFoundError:
            logger.warning("Tokenomics source file missing: %s", self.source_file)
        except json.JSONDecodeError as exc:
            logger.error("Invalid tokenomics source JSON: %s", exc)
        except OSError as exc:
            logger.error(
                "Unable to read tokenomics source file %s: %s", self.source_file, exc
            )

    def ingest_event(self, payload: Dict[str, Any]):
        with self._lock:
            self._apply_payload(payload)
            try:
                self._persist_payload(self._last_payload)
            except OSError as exc:
                logger.error(
                    "Unable to persist tokenomics payload to %s: %s",
                    self.source_file,
                    exc,
                )

    def generate_metrics(self) -> bytes:
        return generate_latest(self.registry)


def run_simulation(exporter):
    import time, random

    last_escrow = 85000450
    last_wallets = 8540
    last_balance = 14630
    while True:
        last_escrow += random.randint(-50000, 100000)
        last_wallets += random.randint(0, 5)
        last_balance += random.randint(-10, 15)
        payload = {
            "mint_rate_per_min": random.uniform(140.0, 160.0),
            "token_supply_total": 1000000000,
            "token_supply_minted": 125000000 + random.randint(1000, 5000),
            "bridge_inflow_per_min": random.uniform(400.0, 500.0),
            "bridge_outflow_per_min": random.uniform(250.0, 350.0),
            "bridge_escrow_total": last_escrow,
            "bridge_transfers_total": random.uniform(10000, 14000),
            "bridge_routes_active": 2,
            "bridge_collateral_ratio_percent": random.uniform(150.0, 160.0),
            "chain_height": random.uniform(800, 1200),
            "fl_verification_ratio": random.uniform(0.96, 0.995),
            "fl_average_confidence_bps": random.uniform(9400, 9900),
            "unique_wallets_count": last_wallets,
            "wallet_average_balance": last_balance,
            # Ratios are represented as 0..1 values for downstream dashboard formulas.
            "top_10_holder_concentration": 0.124,
            "wallet_liquidity_ratio": random.uniform(0.34, 0.35),
        }
        exporter.ingest_event(payload)
        time.sleep(10)


def create_app(source_file: str):
    app = Flask(__name__)
    state_file = os.getenv("TOKENOMICS_STATE_FILE")
    exporter = TokenomicsMetricsExporter(source_file=source_file, state_file=state_file)
    exporter.load_source_file()
    import threading

    t = threading.Thread(target=run_simulation, args=(exporter,), daemon=True)
    t.start()

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
