# Documentation Maintenance Runbook

This runbook defines how to keep repository documentation correct, factual, and current with engineering changes.

## Scope

Apply this process to canonical docs first:

- README.md
- docs/README.md
- Documentation/README.md
- tests/docs/TEST_ENV_SETUP.md
- Documentation/Security/CI_STATUS_AND_CLAIMS.md

Historical archives are preserved as snapshots and should not be rewritten to represent current branch behavior.

## Required Update Triggers

Update canonical docs whenever any of the following change:

- New CI workflows or major workflow behavior changes
- New required test suites or validation profiles
- Security control defaults or auth expectations
- New operator commands that replace previous recommended commands
- Any claim that no longer matches the current main branch behavior

## Documentation Correctness Checklist

For each update:

1. Verify commands from docs run successfully in a clean workspace.
2. Prefer branch-scoped workflow badges over manual status claims.
3. Keep performance and scale statements labeled as observed results, not guarantees.
4. Link to source workflows, scripts, or tests that prove the claim.
5. Remove or qualify stale absolute language like "always", "guaranteed", or "fully proven".
6. Keep one canonical location per operational topic and link to it from indexes.

## Validation Commands

Run these before finalizing documentation edits:

```bash
npm run test:setup
npm run test:full:fast
npm run test:trends
npm run test:summary:diff
```

Optional deep checks:

```bash
npm run test:full:deep
```

## Claim Policy

Use:

- CI-verified
- Observed in artifacts
- Design target
- Emulated validation

Avoid unless externally certified:

- Guaranteed
- Fully production-proved
- Formally compliant

## Suggested Ongoing Cadence

- Per PR touching CI, testing, security, or runtime behavior:
  - Update README and docs indexes in the same PR.
- Weekly:
  - Review workflow badges and stale links.
- Monthly:
  - Review claim wording and demote stale claims to historical references.

## Sources of Truth

- Workflow files in .github/workflows
- Validation scripts in tests/scripts/python and tests/scripts/ci
- Artifacts in test-results/full-validation
- Security claim boundaries in Documentation/Security/CI_STATUS_AND_CLAIMS.md
