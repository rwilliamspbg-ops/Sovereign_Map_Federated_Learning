import json
from pathlib import Path
dashboard = json.loads(Path("grafana/provisioning/dashboards/operations_overview.json").read_text())
for panel in dashboard.get("panels", []):
    grid = panel.get("gridPos", {})
    if grid.get("y", 0) >= 48:
        print(f"Panel: {panel.get('title')} (y={grid.get('y')})")
        for target in panel.get("targets", []):
            expr = target.get("expr", "")
            print(f"  Expr: {expr}")
