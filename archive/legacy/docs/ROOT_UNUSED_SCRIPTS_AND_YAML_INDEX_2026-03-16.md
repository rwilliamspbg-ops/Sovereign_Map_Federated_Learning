# Root Unused Scripts and YAML Archive Index (2026-03-16)

This index tracks root-level `.sh` and `.yml` files moved to `archive/legacy/` during root cleanup.

## Scope

- Source location: repository root
- Move criteria: no active in-repository references outside legacy/historical reports
- Goal: keep root focused on currently used runtime, deployment, and CI assets

## Moved Files

| Original Path | Archived Path | Reason |
|---|---|---|
| `start.sh` | `archive/legacy/scripts/start.sh` | Legacy convenience wrapper; not referenced by active docs, CI, Docker, or scripts |
| `phase-mega-master-run.sh` | `archive/legacy/scripts/phase-mega-master-run.sh` | Legacy orchestration wrapper; not referenced by active docs/CI/scripts |
| `phase-6-cleanup-aws.sh` | `archive/legacy/scripts/phase-6-cleanup-aws.sh` | Legacy teardown helper referenced only in archived orchestration notes |
| `prometheus-tpm.yml` | `archive/legacy/config/prometheus-tpm.yml` | Superseded by active monitoring config (`prometheus.yml`) and not referenced by active stack |

## Notes

- Active phase pipeline files (`phase-1/2/3/5-*`) were retained in root because they are still used by test orchestration scripts.
- Active monitoring and alert files (`prometheus.yml`, `tpm_alerts.yml`, `dashboard_compat_rules.yml`, `fl_slo_alerts.yml`) were retained in root.
