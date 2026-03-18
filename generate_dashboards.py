import json
import os


def create_dashboard(title, uid, panels):
    return {
        "annotations": {"list": []},
        "editable": True,
        "fiscalYearStartMonth": 0,
        "graphTooltip": 0,
        "id": None,
        "links": [],
        "liveNow": False,
        "panels": panels,
        "refresh": "5s",
        "schemaVersion": 38,
        "style": "dark",
        "tags": [title.lower().replace(" ", "-")],
        "templating": {"list": []},
        "time": {"from": "now-6h", "to": "now"},
        "timepicker": {},
        "timezone": "",
        "title": title,
        "uid": uid,
        "version": 1,
    }


def create_panel(title, pid, gridPos, targets):
    return {
        "datasource": {"type": "prometheus", "uid": "prometheus"},
        "fieldConfig": {
            "defaults": {
                "color": {"mode": "palette-classic"},
                "custom": {
                    "axisCenteredZero": False,
                    "axisColorMode": "text",
                    "axisLabel": "",
                    "axisPlacement": "auto",
                    "barAlignment": 0,
                    "drawStyle": "line",
                    "fillOpacity": 10,
                    "gradientMode": "none",
                    "hideFrom": {"legend": False, "tooltip": False, "viz": False},
                    "lineInterpolation": "linear",
                    "lineWidth": 1,
                    "pointSize": 5,
                    "scaleDistribution": {"type": "linear"},
                    "showPoints": "auto",
                    "spanNulls": False,
                    "stacking": {"group": "A", "mode": "none"},
                    "thresholdsStyle": {"mode": "off"},
                },
                "mappings": [],
                "thresholds": {
                    "mode": "absolute",
                    "steps": [{"color": "green", "value": None}],
                },
            },
            "overrides": [],
        },
        "gridPos": gridPos,
        "id": pid,
        "options": {
            "legend": {
                "calcs": [],
                "displayMode": "list",
                "placement": "bottom",
                "showLegend": True,
            },
            "tooltip": {"mode": "single", "sort": "none"},
        },
        "targets": targets,
        "title": title,
        "type": "timeseries",
    }


tokenomics_panels = [
    create_panel(
        "Staked Collateral Over Time",
        1,
        {"h": 8, "w": 12, "x": 0, "y": 0},
        [{"expr": "tokenomics_staked_collateral", "refId": "A"}],
    ),
    create_panel(
        "Wallet Token Balances",
        2,
        {"h": 8, "w": 12, "x": 12, "y": 0},
        [{"expr": "tokenomics_wallet_balance", "refId": "A"}],
    ),
    create_panel(
        "Model Contributions (Int8)",
        3,
        {"h": 8, "w": 12, "x": 0, "y": 8},
        [{"expr": "tokenomics_contributions_total", "refId": "A"}],
    ),
]

performance_panels = [
    create_panel(
        "Edge Compression Ratio",
        1,
        {"h": 8, "w": 12, "x": 0, "y": 0},
        [{"expr": "fl_compression_ratio", "refId": "A"}],
    ),
    create_panel(
        "Self-Healing Backoff Events",
        2,
        {"h": 8, "w": 12, "x": 12, "y": 0},
        [{"expr": "fl_self_healing_events_total", "refId": "A"}],
    ),
    create_panel(
        "Aggregator Bottlenecks (429/503)",
        3,
        {"h": 8, "w": 12, "x": 0, "y": 8},
        [{"expr": "fl_aggregator_throttles_total", "refId": "A"}],
    ),
]

os.makedirs("grafana/dashboards", exist_ok=True)
with open("grafana/dashboards/fl_tokenomics.json", "w") as f:
    json.dump(
        create_dashboard("Tokenomics Overview", "tok-1", tokenomics_panels), f, indent=2
    )

with open("grafana/dashboards/fl_performance.json", "w") as f:
    json.dump(
        create_dashboard("FL Performance", "perf-1", performance_panels), f, indent=2
    )

print("Dashboards regenerated.")
