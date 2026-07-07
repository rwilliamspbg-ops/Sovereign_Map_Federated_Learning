import json
from pathlib import Path

dashboard_path = Path("grafana/provisioning/dashboards/operations_overview.json")
dashboard = json.loads(dashboard_path.read_text())

for panel in dashboard.get("panels", []):
    grid = panel.get("gridPos", {})
    if grid.get("y", 0) >= 48:
        for target in panel.get("targets", []):
            expr = target.get("expr", "").strip()
            if expr and not expr.endswith("or vector(0)"):
                target["expr"] = f"{expr} or vector(0)"

dashboard_path.write_text(json.dumps(dashboard, indent=2))
print("Dashboard queries updated.")
