import json
from pathlib import Path

dashboard_path = Path("grafana/provisioning/dashboards/operations_overview.json")
dashboard = json.loads(dashboard_path.read_text())

# Add a new panel for FL Round Duration
new_panel = {
  "title": "FL Round Duration (Seconds)",
  "type": "timeseries",
  "gridPos": {
    "h": 8,
    "w": 12,
    "x": 0,
    "y": 122
  },
  "targets": [
    {
      "expr": "histogram_quantile(0.95, sum(rate(sovereignmap_fl_round_duration_seconds_bucket[$window])) by (le)) or vector(0)",
      "legendFormat": "p95 Duration"
    },
    {
      "expr": "histogram_quantile(0.5, sum(rate(sovereignmap_fl_round_duration_seconds_bucket[$window])) by (le)) or vector(0)",
      "legendFormat": "p50 Duration"
    }
  ]
}

dashboard["panels"].append(new_panel)

# Add a new panel for FL Client Errors
error_panel = {
  "title": "FL Client Errors by Type",
  "type": "timeseries",
  "gridPos": {
    "h": 8,
    "w": 12,
    "x": 12,
    "y": 122
  },
  "targets": [
    {
      "expr": "sum(increase(sovereignmap_fl_client_error_total[$window])) by (error_type) or vector(0)",
      "legendFormat": "{{error_type}}"
    }
  ]
}

dashboard["panels"].append(error_panel)

dashboard_path.write_text(json.dumps(dashboard, indent=2))
print("Dashboard updated with new panels.")
