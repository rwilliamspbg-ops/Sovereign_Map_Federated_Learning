#!/usr/bin/env bash

set -euo pipefail

ROOT_DIR="${ROOT_DIR:-$PWD}"
OUT_ROOT="${OUT_ROOT:-test-results/port-audit/all-compose}"

mkdir -p "$OUT_ROOT"

summary_file="$OUT_ROOT/summary.md"
json_summary="$OUT_ROOT/summary.jsonl"

echo "# All Compose Files Port Audit" > "$summary_file"
echo >> "$summary_file"
echo "Generated: $(date -u +'%Y-%m-%dT%H:%M:%SZ')" >> "$summary_file"
echo >> "$summary_file"

: > "$json_summary"

mapfile -t compose_files < <(find "$ROOT_DIR" -type f \( -name 'docker-compose*.yml' -o -name 'docker-compose*.yaml' \) | sort)

if [ "${#compose_files[@]}" -eq 0 ]; then
  echo "No docker-compose files found under $ROOT_DIR"
  exit 1
fi

echo "Found ${#compose_files[@]} compose files"

total_static=0
total_runtime=0
total_parse_issues=0

for compose_path in "${compose_files[@]}"; do
  rel_path="${compose_path#$ROOT_DIR/}"
  safe_name="$(echo "$rel_path" | tr '/.' '__')"
  out_dir="$OUT_ROOT/$safe_name"
  mkdir -p "$out_dir"

  echo "Auditing $rel_path"

  config_ok=true
  if docker compose -f "$compose_path" config >/dev/null 2>"$out_dir/config-error.log"; then
    echo "config_ok=true" > "$out_dir/config-status.env"
  else
    config_ok=false
    echo "config_ok=false" > "$out_dir/config-status.env"
  fi

  if ! node scripts/audit-compose-ports.mjs --out-dir "$out_dir" --target-file "$rel_path" >"$out_dir/audit.log" 2>&1; then
    true
  fi

  static_count="0"
  runtime_count="0"
  parse_count="0"

  if [ -f "$out_dir/compose-port-audit.json" ]; then
    static_count="$(node -e "const fs=require('fs');const p=JSON.parse(fs.readFileSync(process.argv[1],'utf8'));console.log(p.staticConflictCount||0)" "$out_dir/compose-port-audit.json")"
    runtime_count="$(node -e "const fs=require('fs');const p=JSON.parse(fs.readFileSync(process.argv[1],'utf8'));console.log(p.runtimeConflictCount||0)" "$out_dir/compose-port-audit.json")"
    parse_count="$(node -e "const fs=require('fs');const p=JSON.parse(fs.readFileSync(process.argv[1],'utf8'));console.log((p.parseIssues||[]).length)" "$out_dir/compose-port-audit.json")"
  fi

  total_static=$((total_static + static_count))
  total_runtime=$((total_runtime + runtime_count))
  total_parse_issues=$((total_parse_issues + parse_count))

  printf '{"compose_file":"%s","config_ok":%s,"static_conflicts":%s,"runtime_conflicts":%s,"parse_issues":%s}\n' \
    "$rel_path" "$config_ok" "$static_count" "$runtime_count" "$parse_count" >> "$json_summary"

  echo "- $rel_path" >> "$summary_file"
  echo "  - config_valid: $config_ok" >> "$summary_file"
  echo "  - static_conflicts: $static_count" >> "$summary_file"
  echo "  - runtime_conflicts: $runtime_count" >> "$summary_file"
  echo "  - parse_issues: $parse_count" >> "$summary_file"
done

echo >> "$summary_file"
echo "## Totals" >> "$summary_file"
echo >> "$summary_file"
echo "- static_conflicts_total: $total_static" >> "$summary_file"
echo "- runtime_conflicts_total: $total_runtime" >> "$summary_file"
echo "- parse_issues_total: $total_parse_issues" >> "$summary_file"

echo
echo "Audit complete"
echo "Summary: $summary_file"
echo "Machine-readable: $json_summary"
