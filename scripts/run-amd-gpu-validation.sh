#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
TS="$(date -u +%Y%m%dT%H%M%SZ)"
OUT_DIR="$ROOT_DIR/test-results/amd-gpu-validation/$TS"
mkdir -p "$OUT_DIR"

PYTHON_BIN="${PYTHON_BIN:-$(command -v python3 || true)}"
NODES=8
INSTALL_HIP_TORCH=0
ROCM_INDEX_URL="https://download.pytorch.org/whl/rocm6.3"

action_usage() {
  cat <<'USAGE'
Usage: bash scripts/run-amd-gpu-validation.sh [options]

Options:
  --python-bin <path>         Python interpreter to use.
  --nodes <n>                 Node count for contention test (default: 8).
  --install-hip-torch         Install ROCm-enabled PyTorch from index URL.
  --rocm-index-url <url>      ROCm wheel index URL (default: rocm6.3).
  --help                      Show this help.
USAGE
}

while [[ $# -gt 0 ]]; do
  case "$1" in
    --python-bin)
      PYTHON_BIN="$2"
      shift 2
      ;;
    --nodes)
      NODES="$2"
      shift 2
      ;;
    --install-hip-torch)
      INSTALL_HIP_TORCH=1
      shift
      ;;
    --rocm-index-url)
      ROCM_INDEX_URL="$2"
      shift 2
      ;;
    --help)
      action_usage
      exit 0
      ;;
    *)
      echo "Unknown option: $1" >&2
      action_usage
      exit 1
      ;;
  esac
done

if [[ -z "$PYTHON_BIN" ]]; then
  echo "No python interpreter found. Set --python-bin or PYTHON_BIN." >&2
  exit 1
fi

if [[ ! -x "$PYTHON_BIN" ]]; then
  echo "Python interpreter is not executable: $PYTHON_BIN" >&2
  exit 1
fi

log() {
  echo "[$(date -u +%Y-%m-%dT%H:%M:%SZ)] $*"
}

log "AMD validation output directory: $OUT_DIR"
log "Using python: $PYTHON_BIN"

ROCM_READY=0
{
  echo "timestamp_utc=$TS"
  echo "python_bin=$PYTHON_BIN"
  echo "uname=$(uname -a)"
  echo "rocminfo=$(command -v rocminfo || true)"
  echo "rocm_smi=$(command -v rocm-smi || true)"
  echo "nvidia_smi=$(command -v nvidia-smi || true)"
  echo "dri_dir_exists=$([[ -d /dev/dri ]] && echo yes || echo no)"
  if [[ -d /dev/dri ]]; then
    ls -la /dev/dri || true
  fi
} >"$OUT_DIR/rocm-readiness.txt"

if command -v rocminfo >/dev/null 2>&1; then
  ROCM_READY=1
  rocminfo >"$OUT_DIR/rocminfo.txt" 2>&1 || true
fi
if command -v rocm-smi >/dev/null 2>&1; then
  rocm-smi >"$OUT_DIR/rocm-smi.txt" 2>&1 || true
fi

if [[ "$INSTALL_HIP_TORCH" -eq 1 ]]; then
  log "Installing ROCm-enabled PyTorch from: $ROCM_INDEX_URL"
  "$PYTHON_BIN" -m pip install --upgrade pip >"$OUT_DIR/pip-upgrade.log" 2>&1 || true
  "$PYTHON_BIN" -m pip install --upgrade --index-url "$ROCM_INDEX_URL" torch torchvision torchaudio >"$OUT_DIR/pip-install-hip.log" 2>&1 || true
fi

log "Collecting torch backend details"
"$PYTHON_BIN" - <<'PY' >"$OUT_DIR/torch-backend.json"
import json
import os

result = {
    "torch_importable": False,
    "torch_version": None,
    "cuda_available": None,
    "cuda_version": None,
    "hip_version": None,
    "device_count": None,
    "errors": [],
}

try:
    import torch
    result["torch_importable"] = True
    result["torch_version"] = torch.__version__
    result["cuda_available"] = torch.cuda.is_available()
    result["cuda_version"] = torch.version.cuda
    result["hip_version"] = getattr(torch.version, "hip", None)
    result["device_count"] = torch.cuda.device_count()
except Exception as e:
    result["errors"].append(str(e))

print(json.dumps(result, indent=2))
PY

HIP_VERSION="$($PYTHON_BIN - <<'PY'
try:
    import torch
    print(getattr(torch.version, 'hip', None) or '')
except Exception:
    print('')
PY
)"

if [[ -n "$HIP_VERSION" ]]; then
  log "ROCm HIP backend detected: $HIP_VERSION"
else
  log "ROCm HIP backend not detected (hip_version empty)"
fi

log "Running npu-gpu-cpu benchmark suite"
"$PYTHON_BIN" "$ROOT_DIR/tests/scripts/python/npu-gpu-cpu-benchmark.py" \
  --compare-devices --contention --nodes "$NODES" --json "$OUT_DIR/gpu-benchmark-results.json" \
  >"$OUT_DIR/npu-gpu-cpu-benchmark.log" 2>&1 || true

log "Running gpu test suite benchmark"
"$PYTHON_BIN" "$ROOT_DIR/tests/scripts/python/gpu-test-suite.py" \
  --benchmark --json "$OUT_DIR/gpu-test-suite-results.json" \
  >"$OUT_DIR/gpu-test-suite.log" 2>&1 || true

log "Building comparison report against CPU baselines"
"$PYTHON_BIN" - "$ROOT_DIR" "$OUT_DIR" <<'PY' >"$OUT_DIR/compare-summary.md"
import json
import os
import sys
from pathlib import Path

root = Path(sys.argv[1])
out = Path(sys.argv[2])

baseline_b1 = root / "gpu-benchmark-results.json"
baseline_b2 = root / "gpu-test-suite-results.json"
new_b1 = out / "gpu-benchmark-results.json"
new_b2 = out / "gpu-test-suite-results.json"

def load(path):
    if not path.exists():
        return None
    try:
        return json.loads(path.read_text())
    except Exception:
        return None

def get_cpu_epoch_from_b1(data):
    try:
        return float(data["comparison"]["cpu"]["avg_epoch_time_sec"])
    except Exception:
        return None

def get_cpu_epoch_from_b2(data):
    try:
        return float(data["cpu_vs_gpu"]["cpu"]["avg_epoch_time_sec"])
    except Exception:
        return None

def get_gpu_present_from_b2(data):
    try:
        return data["cpu_vs_gpu"]["gpu"] is not None
    except Exception:
        return False

base1 = load(baseline_b1)
base2 = load(baseline_b2)
cur1 = load(new_b1)
cur2 = load(new_b2)

print("# AMD Validation Comparison")
print("")
print(f"- Baseline file 1 present: {'yes' if base1 else 'no'}")
print(f"- Baseline file 2 present: {'yes' if base2 else 'no'}")
print(f"- New file 1 present: {'yes' if cur1 else 'no'}")
print(f"- New file 2 present: {'yes' if cur2 else 'no'}")
print("")

base_cpu_1 = get_cpu_epoch_from_b1(base1) if base1 else None
cur_cpu_1 = get_cpu_epoch_from_b1(cur1) if cur1 else None
base_cpu_2 = get_cpu_epoch_from_b2(base2) if base2 else None
cur_cpu_2 = get_cpu_epoch_from_b2(cur2) if cur2 else None

if base_cpu_1 is not None and cur_cpu_1 is not None:
    delta = ((cur_cpu_1 - base_cpu_1) / base_cpu_1) * 100.0
    print(f"- npu-gpu-cpu benchmark CPU epoch: baseline={base_cpu_1:.6f}s current={cur_cpu_1:.6f}s delta={delta:+.2f}%")
else:
    print("- npu-gpu-cpu benchmark CPU epoch: unavailable")

if base_cpu_2 is not None and cur_cpu_2 is not None:
    delta = ((cur_cpu_2 - base_cpu_2) / base_cpu_2) * 100.0
    print(f"- gpu-test-suite CPU epoch: baseline={base_cpu_2:.6f}s current={cur_cpu_2:.6f}s delta={delta:+.2f}%")
else:
    print("- gpu-test-suite CPU epoch: unavailable")

gpu_present = get_gpu_present_from_b2(cur2) if cur2 else False
print(f"- GPU benchmark entries present in gpu-test-suite: {'yes' if gpu_present else 'no'}")
PY

cat >"$OUT_DIR/RUN_SUMMARY.txt" <<EOF
AMD validation completed.

output_dir=$OUT_DIR
python_bin=$PYTHON_BIN
hip_version=${HIP_VERSION:-<empty>}
rocm_ready=$ROCM_READY

Artifacts:
- rocm-readiness.txt
- torch-backend.json
- npu-gpu-cpu-benchmark.log
- gpu-test-suite.log
- gpu-benchmark-results.json
- gpu-test-suite-results.json
- compare-summary.md
EOF

log "Completed. Summary file: $OUT_DIR/RUN_SUMMARY.txt"
log "Compare report: $OUT_DIR/compare-summary.md"
