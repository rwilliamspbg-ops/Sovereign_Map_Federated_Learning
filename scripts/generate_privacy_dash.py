import json
import os
from generate_dashboards import create_dashboard, create_panel

privacy_panels = [
    create_panel(
        "DP Gaussian Noise Level (Sigma)",
        1,
        {"h": 8, "w": 12, "x": 0, "y": 0},
        [{"expr": "fl_dp_noise_multiplier", "refId": "A"}],
    ),
    create_panel(
        "Quantization Bit Depth",
        2,
        {"h": 8, "w": 12, "x": 12, "y": 0},
        [{"expr": "fl_quantization_bits", "refId": "A"}],
    ),
    create_panel(
        "Enclave Attestation Failures",
        3,
        {"h": 8, "w": 12, "x": 0, "y": 8},
        [{"expr": "fl_tee_attestation_failures_total", "refId": "A"}],
    ),
]

os.makedirs("grafana/dashboards", exist_ok=True)
with open("grafana/dashboards/fl_privacy.json", "w") as f:
    json.dump(
        create_dashboard("FL Privacy & Security", "priv-1", privacy_panels), f, indent=2
    )

print("Privacy Dashboard regenerated.")
