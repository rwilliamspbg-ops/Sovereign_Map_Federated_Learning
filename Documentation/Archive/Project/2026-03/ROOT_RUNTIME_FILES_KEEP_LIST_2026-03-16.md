# Root Runtime Files Keep List (2026-03-16)

This list records why root-level `.sh`, `.yml`, and `.yaml` files are still kept in repository root after archival cleanup.

## Keep Criteria

- Referenced by active GitHub workflows.
- Referenced by active scripts or test orchestration.
- Referenced by active documentation and deployment runbooks.
- Used as primary runtime/monitoring/deployment configuration.

## Kept Files

### Active Runtime and Monitoring
- `prometheus.yml` - Primary active Prometheus configuration.
- `alertmanager.yml` - Active alertmanager config referenced by deployment docs.
- `tpm_alerts.yml` - Active alert rule file loaded by `prometheus.yml`.
- `dashboard_compat_rules.yml` - Active alert compatibility rules for dashboards.
- `fl_slo_alerts.yml` - Active SLO alert rules.

### Active Deployment and Compose
- `docker-compose.yml` - Base stack compose file.
- `docker-compose.full.yml` - Production compose stack.
- `docker-compose.full.yml` - Full local/integration stack.
- `docker-compose.full.yml` - Monitoring overlay compose.
- `docker-compose.full.yml` - TPM monitoring overlay.
- `docker-compose.full.yml` - TPM-secure deployment overlay.
- `docker-compose.full.yml` - Participant onboarding/local participant profile.
- `docker-compose.full.yml` - Development compose profile.
- `docker-compose.full.yml` - CI compose profile.
- `docker-compose.full.yml` - 1000-node benchmark/test profile.
- `docker-compose.full.yml` - 200-node BFT test profile.
- `docker-compose.full.yml` - Large-scale scenario profile.
- `docker-compose.full.yml` - Mohawk capability profile.
- `kubernetes-5000-node-manifests.yaml` - 5000-node k8s scenario manifests.

### Active Operational Scripts
- `genesis-launch.sh` - Active launch orchestration for genesis/testnet flows.
- `tpm-bootstrap.sh` - Active TPM bootstrap flow.
- `validate-docker.sh` - Active docker validation helper.
- `validate-genesis-launch.sh` - Active genesis launch validation helper.
- `validate-secrets.sh` - Active secret validation helper.
- `phase-1-aws-setup.sh` - Active phase pipeline script.
- `phase-2-deploy-infrastructure.sh` - Active phase pipeline script.
- `phase-3-deploy-code.sh` - Active phase pipeline script.
- `phase-5-capture-results.sh` - Active phase pipeline script.

### Context-Specific/Legacy-Compatible but Still Referenced
- `deploy.sh` - Referenced by active root docs.
- `deploy_demo.sh` - Referenced by active docs and demo flow notes.

## Archived Counterpart Index

See `archive/legacy/docs/ROOT_UNUSED_SCRIPTS_AND_YAML_INDEX_2026-03-16.md` for files moved out of root.
