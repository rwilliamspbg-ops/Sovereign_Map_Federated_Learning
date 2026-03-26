#!/usr/bin/env bash

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
BASE_REF="${BASE_REF:-origin/main}"
BENCH_CLIENTS="${BENCH_CLIENTS:-100 500 1000}"
BENCH_RUNS="${BENCH_RUNS:-3}"
BENCH_SEED="${BENCH_SEED:-42}"
REPORT_PATH="${REPORT_PATH:-results/metrics/fedavg_benchmark_compare.md}"
PYTHON_BIN="${PYTHON_BIN:-python3}"

if ! git -C "$ROOT_DIR" rev-parse --verify "$BASE_REF" >/dev/null 2>&1; then
  echo "error: base ref '$BASE_REF' not found"
  exit 1
fi

if [[ ! -f "$ROOT_DIR/scripts/benchmark_fedavg_vectorization.py" ]]; then
  echo "error: scripts/benchmark_fedavg_vectorization.py not found"
  exit 1
fi

read -r -a CLIENT_ARGS <<< "$BENCH_CLIENTS"

TMP_DIR="$(mktemp -d)"
BASE_WORKTREE="$TMP_DIR/base"
BASE_OUT="$TMP_DIR/base_bench.txt"
CURR_OUT="$TMP_DIR/current_bench.txt"
BASE_TSV="$TMP_DIR/base.tsv"
CURR_TSV="$TMP_DIR/current.tsv"
JOINED_TSV="$TMP_DIR/joined.tsv"

cleanup() {
  git -C "$ROOT_DIR" worktree remove --force "$BASE_WORKTREE" >/dev/null 2>&1 || true
  rm -rf "$TMP_DIR"
}
trap cleanup EXIT

git -C "$ROOT_DIR" worktree add --quiet --detach "$BASE_WORKTREE" "$BASE_REF"

run_bench() {
  local repo_dir="$1"
  local out_file="$2"

  if [[ ! -f "$repo_dir/scripts/benchmark_fedavg_vectorization.py" ]]; then
    echo "error: benchmark script missing in $repo_dir"
    exit 1
  fi

  (
    cd "$repo_dir"
    "$PYTHON_BIN" scripts/benchmark_fedavg_vectorization.py \
      --clients "${CLIENT_ARGS[@]}" \
      --runs "$BENCH_RUNS" \
      --seed "$BENCH_SEED"
  ) | tee "$out_file"
}

extract_rows() {
  local in_file="$1"
  local out_file="$2"

  awk -F'|' '
    /^\|[[:space:]]*[0-9]+[[:space:]]*\|/ {
      clients = $2
      loop_ms = $3
      vec_ms = $4
      speedup = $5
      gsub(/[[:space:]]/, "", clients)
      gsub(/[[:space:]]/, "", loop_ms)
      gsub(/[[:space:]]/, "", vec_ms)
      gsub(/[[:space:]x]/, "", speedup)
      printf "%s\t%s\t%s\t%s\n", clients, loop_ms, vec_ms, speedup
    }
  ' "$in_file" | sort -n > "$out_file"
}

run_bench "$BASE_WORKTREE" "$BASE_OUT"
run_bench "$ROOT_DIR" "$CURR_OUT"

extract_rows "$BASE_OUT" "$BASE_TSV"
extract_rows "$CURR_OUT" "$CURR_TSV"

join -t $'\t' -a1 -a2 -e "NA" -o 0,1.2,1.3,1.4,2.2,2.3,2.4 "$BASE_TSV" "$CURR_TSV" > "$JOINED_TSV" || true

mkdir -p "$(dirname "$ROOT_DIR/$REPORT_PATH")"

{
  echo "# FedAvg Benchmark Comparison"
  echo
  echo "- Base ref: $BASE_REF"
  echo "- Clients: $BENCH_CLIENTS"
  echo "- Runs: $BENCH_RUNS"
  echo "- Seed: $BENCH_SEED"
  echo "- Generated at: $(date -u +"%Y-%m-%dT%H:%M:%SZ")"
  echo
  echo "| Clients | Base Loop (ms) | Base Vectorized (ms) | Base Speedup (x) | Current Loop (ms) | Current Vectorized (ms) | Current Speedup (x) | Speedup Delta % |"
  echo "|---:|---:|---:|---:|---:|---:|---:|---:|"

  if [[ ! -s "$JOINED_TSV" ]]; then
    echo "| (no benchmark rows found) | - | - | - | - | - | - | - |"
  else
    while IFS=$'\t' read -r clients base_loop base_vec base_spd curr_loop curr_vec curr_spd; do
      if [[ "$base_spd" == "NA" || "$curr_spd" == "NA" ]]; then
        delta="-"
      else
        delta="$(awk -v b="$base_spd" -v c="$curr_spd" 'BEGIN { if (b == 0) { print "0.00" } else { printf "%.2f", ((c-b)/b)*100 } }')"
      fi
      echo "| $clients | $base_loop | $base_vec | $base_spd | $curr_loop | $curr_vec | $curr_spd | $delta |"
    done < "$JOINED_TSV"
  fi
} > "$ROOT_DIR/$REPORT_PATH"

echo "wrote benchmark comparison report to $REPORT_PATH"
