import re
from pathlib import Path

path = Path("sovereignmap_production_backend_v2.py")
content = path.read_text()

# Add metric definitions
metrics_defs = """
fl_round_duration_seconds = Histogram("sovereignmap_fl_round_duration_seconds", "Time taken for a single FL round", buckets=(10, 30, 60, 120, 300, 600, 1200))
fl_client_error_total = Counter("sovereignmap_fl_client_error_total", "Total client errors reported during fit/evaluate", ["error_type"])
model_persistence_latency_seconds = Histogram("sovereignmap_model_persistence_latency_seconds", "Time taken to persist model checkpoints")
"""

if "fl_round_duration_seconds" not in content:
    content = re.sub(r'(active_nodes_gauge = Gauge\("sovereignmap_active_nodes", "Currently connected nodes"\))', r'\1' + metrics_defs, content)

# Initialize labels
labels_init = """
fl_client_error_total.labels(error_type="timeout").inc(0)
fl_client_error_total.labels(error_type="auth_failure").inc(0)
fl_client_error_total.labels(error_type="validation_rejected").inc(0)
"""

if 'fl_client_error_total.labels(error_type="timeout")' not in content:
    content = re.sub(r'(ops_control_actions_total.labels\(action="verification_policy_update"\).inc\(0\))', r'\1' + labels_init, content)

# Inject round duration tracking in the strategy
if "self.round_start_time = 0" not in content:
    content = content.replace("self.round_num = 0", "self.round_num = 0\n        self.round_start_time = time.time()")

# Record duration in aggregate_fit
duration_record = """
        if self.round_start_time > 0:
            duration = time.time() - self.round_start_time
            fl_round_duration_seconds.observe(duration)
        self.round_start_time = time.time()
"""

if "fl_round_duration_seconds.observe(duration)" not in content:
    # Find aggregate_fit definition and inject at the beginning
    content = re.sub(r'(def aggregate_fit\s*\(.*?\)\s*->\s*Tuple\[Optional\[Parameters\],\s*Dict\[str,\s*Scalar\]\]:)', r'\1' + duration_record, content)

# Record client failures
failure_record = """
        for client_proxy, exc in failures:
            err_type = "unknown"
            if "timeout" in str(exc).lower():
                err_type = "timeout"
            elif "auth" in str(exc).lower():
                err_type = "auth_failure"
            fl_client_error_total.labels(error_type=err_type).inc()
"""

if "fl_client_error_total.labels(error_type=err_type).inc()" not in content:
    content = re.sub(r'(def aggregate_fit\s*\(.*?\)\s*->\s*Tuple\[Optional\[Parameters\],\s*Dict\[str,\s*Scalar\]\]:.*?if not results:\s*return None, \{\})', r'\1' + failure_record, content, flags=re.DOTALL)

# Record persistence latency
persistence_patch = """
    start_persist = time.time()
    persist_round_snapshot(current_round, next_acc, next_loss, active_nodes)
    model_persistence_latency_seconds.observe(time.time() - start_persist)
"""
if "model_persistence_latency_seconds.observe" not in content:
    content = content.replace("persist_round_snapshot(current_round, next_acc, next_loss, active_nodes)", persistence_patch)

path.write_text(content)
print("Backend metrics updated.")
