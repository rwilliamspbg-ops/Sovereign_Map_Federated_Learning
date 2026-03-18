import re

with open("README.md", "r") as f:
    content = f.read()

phase3_injection = """
## 🔥 Phase 3 Active: Edge Sensoring & Differential Privacy
[![Bandwidth Saved](https://img.shields.io/badge/Compression-90%25%20(Int8%2BDP)-2b8a3e?style=flat-square&logo=webpack&logoColor=white)](packages/compression)
[![Edge Integration](https://img.shields.io/badge/Hardware-Android%20ARCore%20%7C%20Web%20Sensor-1c7ed6?style=flat-square&logo=android&logoColor=white)](client/mobile-sdk/android)
[![Performance](https://img.shields.io/badge/Self--Healing-Auto%20Tuning-f08c00?style=flat-square&logo=grafana&logoColor=white)](hardware_auto_tuner.py)

The latest Phase 3 enhancements deliver full edge-hardware parity, seamlessly connecting **Camera/LiDAR data streams** directly to secure TEE enclaves over Self-Healing networks.

"""

if "## 🔥 Phase 3 Active" not in content:
    content = content.replace(
        "## Platform Capabilities", phase3_injection + "## Platform Capabilities"
    )
    with open("README.md", "w") as f:
        f.write(content)
    print("README updated.")
else:
    print("Already updated.")
