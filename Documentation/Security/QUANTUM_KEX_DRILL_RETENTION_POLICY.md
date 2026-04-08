<!-- markdownlint-disable MD022 MD032 -->

# Quantum KEX Drill Retention Policy

## Purpose
Define incident-grade retention controls for Quantum KEX drill evidence bundles.

## Policy Controls
- Classification: security-compliance-evidence
- Minimum retention window: 2555 days (7 years)
- Immutability requirement: required
- Integrity requirement: required checksum catalog (`checksums.sha256`)

## Required Storage Controls
- Use immutable object storage with object lock/WORM semantics.
- Enable cross-region replication for disaster resilience.
- Restrict deletion permissions to break-glass roles only.

## Transfer and Verification Procedure
1. Generate artifact bundle with `make quantum-kex-rotation-drill`.
1. Validate local checksums:

```bash
cd artifacts/quantum-kex-rotation/<drill-id>
sha256sum -c checksums.sha256
```

1. Upload to immutable storage path:

```text
security-evidence/quantum-kex-rotation/<drill-id>/
```

1. Re-validate checksums after download from immutable storage.
1. Record storage URI and verification timestamp in your governance ledger/report.

## Audit Expectations
- Every drill must include:
  - `drill-summary.json`
  - `checksums.sha256`
  - `retention-policy.json`
  - `immutability-notice.txt`
- Keep the rolling public index (`public-drill-index.md`) aligned with the latest 12 drills.

## Exception Handling
- If immutable storage is unavailable, mark the drill as provisional and complete immutable archival within 24 hours.
- If checksum verification fails at any stage, open incident triage and invalidate public claims for that drill until remediated.
