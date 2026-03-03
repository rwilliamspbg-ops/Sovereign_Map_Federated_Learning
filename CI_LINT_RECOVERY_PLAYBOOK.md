# CI Lint Recovery Playbook

Purpose: restore `Lint Code Base` to green on `main`, verify all required workflows are green, and capture evidence.

## Applies To
- Workflow file: `.github/workflows/lint.yml`
- Job name: `Lint and Validate`
- Lint step: `super-linter/super-linter@v5`

## Known Failure Snapshot (2026-03-03)
- Commit: `986c69bf75caa581b2db9254a6c52d108e3b064d`
- Failed run: `https://github.com/rwilliamspbg-ops/Sovereign_Map_Federated_Learning/actions/runs/22641864585`
- Annotation seen: `/usr/bin/git` exit code `128`

---

## 1) Triage the failing run

Use either GitHub UI or CLI:

```bash
gh run view 22641864585
gh run view 22641864585 --log-failed
```

What to confirm in logs:
- Failure occurred in step `Lint Code Base`.
- `super-linter` is using changed-file mode (`VALIDATE_ALL_CODEBASE: false`).
- `actions/checkout` used `fetch-depth: 0`.

If the only error is git exit 128 inside linter changed-file detection and no deterministic file lint error is shown, treat as transient/runner-context issue.

---

## 2) Re-run the workflow (maintainer token)

From GitHub UI:
1. Open the failed run URL.
2. Click **Re-run jobs** (or **Re-run failed jobs**).

From CLI (requires token with workflow write permission):

```bash
gh run rerun 22641864585
gh run watch 22641864585 --exit-status
```

If CLI returns `Resource not accessible by integration`, use the GitHub UI with maintainer permissions.

---

## 3) Force a clean rerun on main (if rerun is unavailable)

Create a no-op commit to retrigger all workflows on `main`:

```bash
git commit --allow-empty -m "ci: retrigger lint workflow"
git push origin main
```

Then monitor:

```bash
gh run list --branch main --limit 20
gh run list --branch main --workflow "Lint Code Base" --limit 5
```

---

## 4) Deterministic fallback status

Status: applied in repository.

- `.github/workflows/lint.yml` now sets `VALIDATE_ALL_CODEBASE: true`
- Workflow now supports `workflow_dispatch` for manual maintainer reruns

Effect: removes changed-file git dependency in super-linter, reducing risk of git-exit-128 failures at the cost of longer lint runtime.

---

## 5) Green verification checklist

Confirm all required workflows on latest `main` commit are green:
- Build and Test
- Lint Code Base
- HIL Tests
- CodeQL Security Analysis
- 🚀 Build & Deploy to Production

CLI check:

```bash
gh run list --branch main --limit 20
```

Optional strict check for latest SHA:

```bash
LATEST_SHA=$(git rev-parse origin/main)
gh run list --commit "$LATEST_SHA" --limit 20 --json workflowName,conclusion,status,url
```

---

## 6) Documentation update after recovery

Once green is restored:
1. Update `TPM_NPU_GPU_SIGNOFF_CHECKLIST.md` section 1 and section 8 to mark workflows green.
2. Update `COMMIT_986c69b_VALIDATION.md` to note lint recovery run URL and final all-green status.
3. Keep links to run evidence in docs.
