# Repository Settings Admin Runbook

Use this runbook when repository metadata updates are needed and automation cannot apply them due to permission limits.

## Required Permissions

- Repository admin rights on this repository
- A GitHub CLI token with `repo` scope for private repos (or `public_repo` for public-only repos)

Recommended token refresh command:

```bash
gh auth refresh -h github.com -s repo,read:org
```

## Metadata Values to Keep Current

- Description
- Homepage URL
- Topics

## Admin Update Commands

```bash
gh repo edit rwilliamspbg-ops/Sovereign_Map_Federated_Learning \
  --description "Formally verified federated learning runtime with hierarchical trust, TPM sovereignty, and auditable zk-style proof paths."

gh repo edit rwilliamspbg-ops/Sovereign_Map_Federated_Learning \
  --homepage "https://rwilliamspbg-ops.github.io/Sovereign_Map_Federated_Learning/"

gh repo edit rwilliamspbg-ops/Sovereign_Map_Federated_Learning \
  --add-topic federated-learning \
  --add-topic byzantine-fault-tolerance \
  --add-topic tpm \
  --add-topic zero-knowledge \
  --add-topic post-quantum-cryptography \
  --add-topic distributed-systems \
  --add-topic observability
```

## Verification

```bash
gh repo view rwilliamspbg-ops/Sovereign_Map_Federated_Learning \
  --json description,homepageUrl,repositoryTopics
```

## Branch Protection Required Checks Alignment

After any workflow rename, verify required check names still match the active workflow/job names in branch protection settings.

At minimum, keep these docs checks aligned:

- `docs-markdownlint / markdownlint`
- `Docs Quality / markdown-lint-and-links`
- `Docs Pages / deploy`
