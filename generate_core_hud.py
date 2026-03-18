import json


def get_gauge_panel(
    title, result_metric, gridPos, max_val=100.0, thresholds=None, decimals=2, suffix=""
):
    if thresholds is None:
        thresholds = [
            {"color": "red", "value": None},
            {"color": "yellow", "value": max_val * 0.4},
            {"color": "green", "value": max_val * 0.8},
        ]
    return {
        "title": title,
        "type": "gauge",
        "gridPos": gridPos,
        "datasource": {"type": "prometheus", "uid": "prometheus"},
        "targets": [{"expr": result_metric, "refId": "A"}],
        "fieldConfig": {
            "defaults": {
                "min": 0,
                "max": max_val,
                "decimals": decimals,
                "unit": suffix,
                "color": {"mode": "thresholds"},
                "thresholds": {"mode": "absolute", "steps": thresholds},
            }
        },
        "options": {"text": {}, "orientation": "auto"},
    }


def get_stat_panel(title, result_metric, gridPos, color="yellow"):
    return {
        "title": title,
        "type": "stat",
        "gridPos": gridPos,
        "datasource": {"type": "prometheus", "uid": "prometheus"},
        "targets": [{"expr": result_metric, "refId": "A"}],
        "fieldConfig": {"defaults": {"color": {"mode": "fixed", "fixedColor": color}}},
        "options": {
            "graphMode": "area",
            "justifyMode": "auto",
            "orientation": "auto",
            "textMode": "auto",
        },
    }


def get_timeseries_panel(title, result_metric, gridPos):
    return {
        "title": title,
        "type": "timeseries",
        "gridPos": gridPos,
        "datasource": {"type": "prometheus", "uid": "prometheus"},
        "targets": [{"expr": result_metric, "refId": "A"}],
        "fieldConfig": {
            "defaults": {
                "color": {"mode": "palette-classic"},
                "custom": {
                    "drawStyle": "line",
                    "lineInterpolation": "smooth",
                    "lineWidth": 2,
                    "fillOpacity": 15,
                },
            }
        },
    }


dashboard = {
    "title": "Sovereign Map | Core Operations HUD",
    "uid": "sovereign-core-hud-001",
    "refresh": "5s",
    "time": {"from": "now-15m", "to": "now"},
    "tags": ["core", "hud", "sovereign"],
    "schemaVersion": 38,
    "panels": [
        # ROW 1: AI & Federated Learning
        {
            "type": "row",
            "title": "Neural Intelligence & Distributed Sync",
            "gridPos": {"h": 1, "w": 24, "x": 0, "y": 0},
        },
        get_gauge_panel(
            "FL Global Accuracy",
            "sovereignmap_fl_accuracy",
            {"h": 6, "w": 6, "x": 0, "y": 1},
            max_val=100.0,
            suffix="percent",
        ),
        get_stat_panel(
            "Active Network Nodes",
            "sovereignmap_active_nodes",
            {"h": 6, "w": 4, "x": 6, "y": 1},
            color="blue",
        ),
        get_stat_panel(
            "Total Completed FL Rounds",
            "sovereignmap_fl_rounds_total",
            {"h": 6, "w": 4, "x": 10, "y": 1},
            color="green",
        ),
        get_timeseries_panel(
            "Global Model Loss",
            "sovereignmap_fl_loss",
            {"h": 6, "w": 10, "x": 14, "y": 1},
        ),
        # ROW 2: Hardware & Edge Performance
        {
            "type": "row",
            "title": "Edge Hardware Telemetry",
            "gridPos": {"h": 1, "w": 24, "x": 0, "y": 7},
        },
        get_gauge_panel(
            "Data Compression Ratio",
            "fl_compression_ratio",
            {"h": 6, "w": 6, "x": 0, "y": 8},
            max_val=20.0,
            thresholds=[
                {"color": "red", "value": None},
                {"color": "green", "value": 5.0},
            ],
            suffix="x",
        ),
        get_gauge_panel(
            "CXL Memory Utilization",
            "process_resident_memory_bytes / 1024 / 1024",
            {"h": 6, "w": 6, "x": 6, "y": 8},
            max_val=4000.0,
            thresholds=[
                {"color": "green", "value": None},
                {"color": "red", "value": 3000},
            ],
            suffix="megabytes",
        ),
        get_timeseries_panel(
            "Attestation Latency (ms)",
            "tpm_node_attestation_latency_ms",
            {"h": 6, "w": 12, "x": 12, "y": 8},
        ),
        # ROW 3: Tokenomics & Security
        {
            "type": "row",
            "title": "Trust, Security & Protocol Economics",
            "gridPos": {"h": 1, "w": 24, "x": 0, "y": 14},
        },
        get_gauge_panel(
            "Average TEE Confidence (bps)",
            "sovereign_fl_avg_confidence_bps",
            {"h": 6, "w": 6, "x": 0, "y": 15},
            max_val=10000.0,
            thresholds=[
                {"color": "red", "value": None},
                {"color": "green", "value": 7500},
            ],
            suffix="bps",
        ),
        get_stat_panel(
            "Network Attestation Failures",
            "fl_tee_attestation_failures_total",
            {"h": 6, "w": 4, "x": 6, "y": 15},
            color="red",
        ),
        get_stat_panel(
            "Average Wallet Balance",
            "tokenomics_wallet_average_balance",
            {"h": 6, "w": 4, "x": 10, "y": 15},
            color="orange",
        ),
        get_timeseries_panel(
            "Network Stake Collateral",
            "tokenomics_staked_collateral",
            {"h": 6, "w": 10, "x": 14, "y": 15},
        ),
    ],
}

with open("grafana/provisioning/dashboards/core_hud.json", "w") as f:
    json.dump(dashboard, f, indent=2)
