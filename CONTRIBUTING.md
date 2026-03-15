# Contributing to Sovereign Map

Thanks for contributing.

## Pull Request Checklist

Before opening or merging a PR, confirm all items below:

Use the repository PR template at `.github/pull_request_template.md` when opening PRs.

- [ ] CI workflows pass on your branch
- [ ] Security checks pass (including CodeQL)
- [ ] Documentation is updated for any behavior/config changes
- [ ] No secrets or credentials are committed

## Documentation Sync Requirements

For changes that affect runtime APIs, auth, trust, or deployment defaults, documentation updates are mandatory in the same PR.

Minimum required files to update:

- `README.md` for user-facing behavior, endpoints, headers, and env vars
- `Documentation/Project/ROADMAP.md` for milestone status and next steps
- Any affected deployment/testing guide under `Documentation/`

Examples of changes that require doc sync:

- New/changed HTTP endpoints or request/response fields
- New/changed auth modes, token locations, or role checks
- New/changed environment variables or compose overlays
- New test harnesses or validation flows used for release confidence

## Definition of Done for API/Security PRs

A PR touching proof verification, hybrid verification, ledger, TPM trust, or transport security should include:

- Unit tests for happy-path and at least one negative path
- Integration coverage where route registration or middleware behavior changed
- README endpoint and auth examples aligned with the implementation
- Roadmap status update indicating whether the change is completed or follow-up work remains

## Contributor Rankings & Reward Points

The repository includes a Git-based reward scoring system to rank contributions.

- Run `make contributors-rankings`
- Outputs:
	- `test-results/contributor-rankings/CONTRIBUTOR_RANKINGS.md`
	- `test-results/contributor-rankings/contributor_rankings.json`

Current points formula:

- `points = commits*5 + additions*0.05 + deletions*0.02 + files_touched*0.5`

Default scoring window is the last 30 days. You can customize window and output paths with:

- `python3 scripts/contributor_rankings.py --since "90 days ago" --until "now" --top 50`

## CodeQL Guardrail (Temporary)

This repository uses an **advanced CodeQL workflow** at `.github/workflows/codeql-analysis.yml`.

- Do **not** enable GitHub CodeQL **Default Setup** for this repository.
- Enabling Default Setup can trigger Java/Kotlin analysis on `mobile-apps/` and fail with `build-mode: none`.
- If this happens, disable Default Setup in **Security → Code scanning** and re-run the advanced workflow.

For details, see `.github/CODEQL.md`.

## Recommended Branch Protection

For `main`, enable branch protection with:

- Require pull requests before merge
- Require status checks to pass before merge
- Require branches to be up to date before merge

At minimum, require these checks:

- `CodeQL Security Analysis / Analyze Code (go)`
- `CodeQL Security Analysis / Analyze Code (javascript-typescript)`
- `CodeQL Security Analysis / Analyze Code (python)`
- `Reproducibility Check / Clone Reproducibility`
- `Workflow Action Pin Check / Enforce SHA-Pinned Actions`
- `Governance Check / Governance Integrity`
- Your primary build/test workflow checks
