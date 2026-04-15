#!/usr/bin/env bash
set -euo pipefail

OWNER_REPO="${1:-rwilliamspbg-ops/Sovereign_Map_Federated_Learning}"

cat <<EOF
Repository settings update helper (prints commands only)

Target repository: ${OWNER_REPO}

Required permissions:
- Repository admin access on ${OWNER_REPO}
- GitHub CLI auth token with:
  - repo (private repositories)
  - public_repo (public-only alternative)
  - read:org (recommended)

Run these commands:

gh auth status
gh auth refresh -h github.com -s repo,read:org

gh repo edit ${OWNER_REPO} \
  --description "Formally verified federated learning runtime with hierarchical trust, TPM sovereignty, and auditable zk-style proof paths."

gh repo edit ${OWNER_REPO} \
  --homepage "https://rwilliamspbg-ops.github.io/Sovereign_Map_Federated_Learning/"

gh repo edit ${OWNER_REPO} \
  --add-topic federated-learning \
  --add-topic byzantine-fault-tolerance \
  --add-topic tpm \
  --add-topic zero-knowledge \
  --add-topic post-quantum-cryptography \
  --add-topic distributed-systems \
  --add-topic observability

Verify:
gh repo view ${OWNER_REPO} --json description,homepageUrl,repositoryTopics
EOF
