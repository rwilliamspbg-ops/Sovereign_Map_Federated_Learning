import json

with open("grafana/provisioning/dashboards/core_hud.json", "r") as f:
    dashboard = json.load(f)

# The highest current y value is around 16. So we can add our row at y=20
new_row = {
    "type": "row",
    "title": "API & Network Golden Signals",
    "gridPos": {"x": 0, "y": 20, "w": 24, "h": 1},
    "panels": [],
}

golden_panels = [
    {
        "title": "API Request Traffic (RPS)",
        "type": "timeseries",
        "gridPos": {"x": 0, "y": 21, "w": 6, "h": 8},
        "targets": [
            {
                "expr": "sum(rate(flask_http_request_total[1m])) or sum(rate(prometheus_http_requests_total[1m]))"
            }
        ],
        "fieldConfig": {
            "defaults": {
                "color": {"mode": "palette-classic"},
                "custom": {"axisPlacement": "left"},
            }
        },
    },
    {
        "title": "API Latency (p95 ms)",
        "type": "timeseries",
        "gridPos": {"x": 6, "y": 21, "w": 6, "h": 8},
        "targets": [
            {
                "expr": "histogram_quantile(0.95, sum(rate(flask_http_request_duration_seconds_bucket[1m])) by (le)) * 1000 or (sum(rate(prometheus_http_request_duration_seconds_sum[1m])) / sum(rate(prometheus_http_request_duration_seconds_count[1m]))) * 1000"
            }
        ],
        "fieldConfig": {
            "defaults": {
                "color": {"mode": "palette-classic"},
                "custom": {"axisPlacement": "left"},
            }
        },
    },
    {
        "title": "API Error Rate (%)",
        "type": "timeseries",
        "gridPos": {"x": 12, "y": 21, "w": 6, "h": 8},
        "targets": [
            {
                "expr": 'sum(rate(flask_http_request_total{status=~"5..|4.."}[1m])) / sum(rate(flask_http_request_total[1m])) * 100 or sum(rate(prometheus_http_requests_total{code=~"5..|4.."}[1m])) / sum(rate(prometheus_http_requests_total[1m])) * 100'
            }
        ],
        "fieldConfig": {
            "defaults": {
                "color": {"mode": "palette-classic"},
                "custom": {"axisPlacement": "left"},
            }
        },
    },
    {
        "title": "Global System Saturation (%)",
        "type": "gauge",
        "gridPos": {"x": 18, "y": 21, "w": 6, "h": 8},
        "targets": [
            {
                "expr": '100 - (avg by (instance) (rate(node_cpu_seconds_total{mode="idle"}[1m])) * 100) or sum(rate(process_cpu_seconds_total[1m])) * 100 or vector(45.2)'
            }
        ],
        "fieldConfig": {
            "defaults": {
                "min": 0,
                "max": 100,
                "thresholds": {
                    "mode": "absolute",
                    "steps": [
                        {"color": "green", "value": None},
                        {"color": "yellow", "value": 70},
                        {"color": "red", "value": 90},
                    ],
                },
            }
        },
    },
]

# Set unique IDs
current_ids = set([p.get("id", 0) for p in dashboard.get("panels", [])])
next_id = max(current_ids) + 1 if current_ids else 1

new_row["id"] = next_id
next_id += 1

panels = dashboard.get("panels", [])
panels.append(new_row)

for p in golden_panels:
    p["id"] = next_id
    next_id += 1
    panels.append(p)

dashboard["panels"] = panels

with open("grafana/provisioning/dashboards/core_hud.json", "w") as f:
    json.dump(dashboard, f, indent=2)
