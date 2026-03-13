#!/usr/bin/env bash
set -euo pipefail

cd "$(dirname "$0")/.."

mkdir -p Documentation/{Security,Testing,Deployment,Architecture,Guides,Performance,Project,CI,Reports,Misc}

cat > /tmp/doc_migration_map.tsv <<'MAP'
1000-NODE-NPU-TEST-FINAL-SUMMARY.md	Documentation/Testing/1000-NODE-NPU-TEST-FINAL-SUMMARY.md
1000-NODE-NPU-TEST-GUIDE.md	Documentation/Testing/1000-NODE-NPU-TEST-GUIDE.md
ARCHITECTURE.md	Documentation/Architecture/ARCHITECTURE.md
ARTIFACTS.md	Documentation/Reports/ARTIFACTS.md
BYZANTINE_STRESS_TEST_REPORT.md	Documentation/Testing/BYZANTINE_STRESS_TEST_REPORT.md
BYZANTINE_STRESS_TEST_SUITE_REPORT.md	Documentation/Testing/BYZANTINE_STRESS_TEST_SUITE_REPORT.md
CI_LINT_RECOVERY_PLAYBOOK.md	Documentation/CI/CI_LINT_RECOVERY_PLAYBOOK.md
CI_STATUS_AND_CLAIMS.md	Documentation/Security/CI_STATUS_AND_CLAIMS.md
CLEANUP_REPORT.md	Documentation/Reports/CLEANUP_REPORT.md
CODEQL_FIX_REPORT.md	Documentation/Security/CODEQL_FIX_REPORT.md
COMMIT_986c69b_VALIDATION.md	Documentation/Testing/COMMIT_986c69b_VALIDATION.md
COMMIT_PUSH_CERTIFICATE.md	Documentation/Security/COMMIT_PUSH_CERTIFICATE.md
COMPLETE_DEMO_DATA_VIEWABLE.md	Documentation/Reports/COMPLETE_DEMO_DATA_VIEWABLE.md
COMPLETION_CHECKLIST.md	Documentation/Testing/COMPLETION_CHECKLIST.md
COMPLETION_REPORT.md	Documentation/Reports/COMPLETION_REPORT.md
DASHBOARD_ARTIFACTS_CONFIRMATION.md	Documentation/Reports/DASHBOARD_ARTIFACTS_CONFIRMATION.md
DEMO_RESULTS_SUMMARY.txt	Documentation/Reports/DEMO_RESULTS_SUMMARY.txt
DEPLOYMENT.md	Documentation/Deployment/DEPLOYMENT.md
DISABLE_DEFAULT_CODEQL.md	Documentation/Security/DISABLE_DEFAULT_CODEQL.md
DOCKER_OPTIMIZATION.md	Documentation/Deployment/DOCKER_OPTIMIZATION.md
DOCKER_OPTIMIZATION_SUMMARY.md	Documentation/Deployment/DOCKER_OPTIMIZATION_SUMMARY.md
ENVIRONMENT_CLEANUP_SUMMARY.md	Documentation/Reports/ENVIRONMENT_CLEANUP_SUMMARY.md
FINAL_TEST_REPORT.md	Documentation/Testing/FINAL_TEST_REPORT.md
FINISHED.md	Documentation/Project/FINISHED.md
GENESIS_INTEGRATION_COMPLETE.md	Documentation/Project/GENESIS_INTEGRATION_COMPLETE.md
GENESIS_LAUNCH_CHECKLIST.md	Documentation/Deployment/GENESIS_LAUNCH_CHECKLIST.md
GENESIS_LAUNCH_GUIDE.md	Documentation/Deployment/GENESIS_LAUNCH_GUIDE.md
GENESIS_QUICK_START.md	Documentation/Guides/GENESIS_QUICK_START.md
GIT_COMMIT_CONFIRMATION.md	Documentation/Reports/GIT_COMMIT_CONFIRMATION.md
GPU_ACCELERATION_GUIDE.md	Documentation/Performance/GPU_ACCELERATION_GUIDE.md
GPU_TESTING_COMPLETE.md	Documentation/Testing/GPU_TESTING_COMPLETE.md
GPU_TESTING_RESULTS_REPORT.md	Documentation/Testing/GPU_TESTING_RESULTS_REPORT.md
GPU_VALIDATION_COMPLETE.md	Documentation/Testing/GPU_VALIDATION_COMPLETE.md
GRAFANA_DASHBOARDS_COMPLETE.md	Documentation/Deployment/GRAFANA_DASHBOARDS_COMPLETE.md
GRAFANA_NPU_LAPTOP_FINAL.md	Documentation/Deployment/GRAFANA_NPU_LAPTOP_FINAL.md
GRAFANA_READY.txt	Documentation/Deployment/GRAFANA_READY.txt
GRAFANA_SETUP_COMPLETE.md	Documentation/Deployment/GRAFANA_SETUP_COMPLETE.md
HIL_TESTING.md	Documentation/Testing/HIL_TESTING.md
IMPLEMENTATION_SUMMARY.md	Documentation/Architecture/IMPLEMENTATION_SUMMARY.md
INDEX.md	Documentation/Guides/INDEX.md
KUBERNETES_5000_NODE_REPORT.md	Documentation/Deployment/KUBERNETES_5000_NODE_REPORT.md
LAPTOP_RESOURCES_INTEGRATION.md	Documentation/Project/LAPTOP_RESOURCES_INTEGRATION.md
LINTER_FIXES_CONFIRMATION.md	Documentation/CI/LINTER_FIXES_CONFIRMATION.md
LINT_SECURITY_FIXES.md	Documentation/Security/LINT_SECURITY_FIXES.md
METRICS_JSON_COMMIT_CONFIRMATION.md	Documentation/Reports/METRICS_JSON_COMMIT_CONFIRMATION.md
MISSING_TESTS_ANALYSIS.md	Documentation/Testing/MISSING_TESTS_ANALYSIS.md
MOBILE_DEPLOYMENT.md	Documentation/Deployment/MOBILE_DEPLOYMENT.md
NETWORK_READINESS_ASSESSMENT.md	Documentation/Deployment/NETWORK_READINESS_ASSESSMENT.md
NPU_GPU_CPU_PERFORMANCE_ANALYSIS.md	Documentation/Performance/NPU_GPU_CPU_PERFORMANCE_ANALYSIS.md
NPU_PERFORMANCE_SCALING_COMPLETE.md	Documentation/Performance/NPU_PERFORMANCE_SCALING_COMPLETE.md
OPENCV_INSTALL.md	Documentation/Guides/OPENCV_INSTALL.md
PREREQUISITES_ENVIRONMENT_SETUP.md	Documentation/Guides/PREREQUISITES_ENVIRONMENT_SETUP.md
PREREQUISITES_QUICK_REFERENCE.md	Documentation/Guides/PREREQUISITES_QUICK_REFERENCE.md
PROJECT_COMPLETE.md	Documentation/Project/PROJECT_COMPLETE.md
QUICK_START_GUIDE.md	Documentation/Guides/QUICK_START_GUIDE.md
README_DEMO_RESULTS.md	Documentation/Reports/README_DEMO_RESULTS.md
README_UPDATE_COMPLETE.md	Documentation/Reports/README_UPDATE_COMPLETE.md
RELEASE_NOTES_TPM_NPU_2026-03-03.md	Documentation/Performance/RELEASE_NOTES_TPM_NPU_2026-03-03.md
REPRODUCIBLE_SETUP.md	Documentation/Guides/REPRODUCIBLE_SETUP.md
ROADMAP.md	Documentation/Project/ROADMAP.md
SCRIPTS_COMPLETE_GUIDE.md	Documentation/Guides/SCRIPTS_COMPLETE_GUIDE.md
SECURITY_AUDIT_2026-02-28.md	Documentation/Security/SECURITY_AUDIT_2026-02-28.md
SESSION_FINALIZATION_REPORT.md	Documentation/Reports/SESSION_FINALIZATION_REPORT.md
SESSION_KUBERNETES_5000_NODE_COMPLETE.md	Documentation/Reports/SESSION_KUBERNETES_5000_NODE_COMPLETE.md
SESSION_TESTNET_FINALIZATION_V1_1_0.md	Documentation/Reports/SESSION_TESTNET_FINALIZATION_V1_1_0.md
SSH_SETUP_GUIDE.md	Documentation/Guides/SSH_SETUP_GUIDE.md
TESTING_CLOSEOUT_FINAL.md	Documentation/Testing/TESTING_CLOSEOUT_FINAL.md
TESTING_COMPLETION_INDEX.md	Documentation/Testing/TESTING_COMPLETION_INDEX.md
TESTING_FINAL_CLOSURE.md	Documentation/Testing/TESTING_FINAL_CLOSURE.md
TESTING_PHASE_CLOSEOUT.md	Documentation/Testing/TESTING_PHASE_CLOSEOUT.md
TESTING_PHASE_COMPLETE.txt	Documentation/Testing/TESTING_PHASE_COMPLETE.txt
TESTNET_DEPLOYMENT.md	Documentation/Deployment/TESTNET_DEPLOYMENT.md
TESTNET_READY_SUMMARY.md	Documentation/Testing/TESTNET_READY_SUMMARY.md
TEST_GUIDE.md	Documentation/Testing/TEST_GUIDE.md
TEST_READY.md	Documentation/Testing/TEST_READY.md
TPM_MONITORING_GUIDE.md	Documentation/Security/TPM_MONITORING_GUIDE.md
TPM_NPU_GPU_SIGNOFF_CHECKLIST.md	Documentation/Testing/TPM_NPU_GPU_SIGNOFF_CHECKLIST.md
TPM_TESTNET_READY.md	Documentation/Security/TPM_TESTNET_READY.md
TPM_TRUST_GUIDE.md	Documentation/Security/TPM_TRUST_GUIDE.md
VALIDATION_REPORT.md	Documentation/Testing/VALIDATION_REPORT.md
MAP

while IFS=$'\t' read -r old new; do
  if [[ -f "$old" ]]; then
    mkdir -p "$(dirname "$new")"
    mv "$old" "$new"
  fi
done < /tmp/doc_migration_map.tsv

find . -type f -iname '*.md' -print > /tmp/all_md_files.txt
while IFS=$'\t' read -r old new; do
  esc_old=$(printf '%s' "$old" | sed 's/[.[\\*^$()+?{|]/\\&/g')
  esc_new=$(printf '%s' "$new" | sed 's/[&]/\\&/g')
  while IFS= read -r mf; do
    perl -0777 -i -pe "s#\]\((?:\./)?${esc_old}\)#](/${esc_new})#g; s#\]\((?:\.\./)+${esc_old}\)#](/${esc_new})#g" "$mf"
  done < /tmp/all_md_files.txt
done < /tmp/doc_migration_map.tsv

find . -type d \( -name .git -o -name node_modules -o -name .venv -o -name __pycache__ \) -prune -o -type f \( -iname '*.md' -o -iname '*.txt' -o -iname '*.rst' -o -iname '*.adoc' \) -print | sed 's#^./##' | sort > /tmp/all_docs_paths.txt

make_topic_index() {
  topic="$1"
  pattern="$2"
  out="Documentation/${topic}_INDEX.md"
  {
    echo "# ${topic} Documentation Index"
    echo
    echo "## Files"
    echo
    grep -Ei "$pattern" /tmp/all_docs_paths.txt | sort | while IFS= read -r p; do
      if [[ "$p" == Documentation/* ]]; then
        link="./${p#Documentation/}"
      else
        link="../$p"
      fi
      echo "- [$p]($link)"
    done
  } > "$out"
}

make_topic_index "SECURITY" 'SECURITY|TPM|TRUST|CODEQL|AUDIT|CLAIMS|CERT'
make_topic_index "TESTING" 'TEST|VALIDATION|VERIFY|CHECKLIST|HIL|BYZANTINE'
make_topic_index "DEPLOYMENT" 'DEPLOY|LAUNCH|DOCKER|KUBERNETES|K8S|MONITOR|GRAFANA|PROMETHEUS|ALERT'
make_topic_index "ARCHITECTURE" 'ARCHITECTURE|TOPOLOGY|IMPLEMENTATION|BLUEPRINT|DESIGN|SERVICE|COORDINATOR'

cat > Documentation/README.md <<'EOF'
# Documentation Hub

This folder centralizes repository documentation discovery.

## Start Here

- Master categorized index: [MASTER_DOCUMENTATION_INDEX.md](MASTER_DOCUMENTATION_INDEX.md)
- Security index: [SECURITY_INDEX.md](SECURITY_INDEX.md)
- Testing index: [TESTING_INDEX.md](TESTING_INDEX.md)
- Deployment index: [DEPLOYMENT_INDEX.md](DEPLOYMENT_INDEX.md)
- Architecture index: [ARCHITECTURE_INDEX.md](ARCHITECTURE_INDEX.md)

## Topic Folders

- [Security](Security)
- [Testing](Testing)
- [Deployment](Deployment)
- [Architecture](Architecture)
- [Guides](Guides)
- [Performance](Performance)
- [CI](CI)
- [Reports](Reports)
- [Project](Project)
- [Misc](Misc)
EOF

echo "done"