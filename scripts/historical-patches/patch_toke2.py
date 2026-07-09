import re

with open("tokenomics_metrics_exporter.py", "r") as f:
    text = f.read()

# Fix the duplicate threading injections
text = re.sub(
    r"(\s*import threading\s*t = threading\.Thread\(target=run_simulation, args=\(exporter,\), daemon=True\)\s*t\.start\(\)\s*)+",
    r"\n    import threading\n    t = threading.Thread(target=run_simulation, args=(exporter,), daemon=True)\n    t.start()\n",
    text,
)

# Prevent generate_metrics from always loading the source file and overriding our simulation
text = re.sub(
    r"def generate_metrics\(self\) -> bytes:\s*self\.load_source_file\(\)\s*return generate_latest\(self\.registry\)",
    r"def generate_metrics(self) -> bytes:\n        return generate_latest(self.registry)",
    text,
)

with open("tokenomics_metrics_exporter.py", "w") as f:
    f.write(text)
