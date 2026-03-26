# Release Package 2026-03-26

## Scope

This package finalizes the open ecosystem and marketplace upgrade track with:

- Marketplace, governance, and network expansion API/documentation updates.
- Frontend browser flow controls and default-view wiring.
- Observability hardening updates for dashboard query compatibility.
- Prometheus marketplace alert additions and test coverage.
- Documentation cleanup (link integrity and organization index).

## Included Documentation Updates

- docs/README.md
- docs/OPEN_ECOSYSTEM_FIRST_10_MINUTES.md
- docs/OPEN_ECOSYSTEM_SPRINT1_ROADMAP.md
- docs/OPEN_ECOSYSTEM_SPRINT2_ROADMAP.md
- docs/RELEASE_CANDIDATE_PHASE3D_SECRETS_RUNBOOK.md
- docs/PR_READY_PHASE3D_OVERHAUL.md
- docs/FEATURE_ARTIFACTS_2026-03-21.md
- docs/COMPRESSION_DEPLOYMENT.md
- docs/DATA_COMPRESSION_GUIDE.md
- docs/GPU_NPU_ACCELERATION.md
- docs/GPU_NPU_QUICK_START.md
- README.md
- SDK_API_STABILITY.md
- SDK_RELEASE_GUIDE.md

## Verification Checklist

Completed checks during package validation:

1. make lint
2. make alerts-test
3. make observability-smoke
4. npm --prefix frontend run build
5. python tests/scripts/python/test_marketplace_local_contracts.py
6. python tests/scripts/python/test_marketplace_negative_paths.py
7. make check
8. active docs link audit (0 active-doc broken links)

## Notes

- Canonical docs entrypoint is now docs/README.md.
- Root-level SDK policy and release guide files are retained as stable reference targets.
- Archive documentation was intentionally left as historical and not rewritten.
