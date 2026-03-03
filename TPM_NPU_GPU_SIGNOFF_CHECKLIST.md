# TPM + NPU + GPU Integration Sign-off Checklist

**Purpose:** Final Definition of Done (DoD) gates for TPM/NPU/GPU integration.

**Status key:**
- [ ] Not done
- [x] Done
- [N/A] Not applicable for this release

---

## 1) Branch/CI Gate (must pass first)

- [ ] `main` is clean (`git status --short` returns no changes)
- [ ] Latest `main` commit has green workflows:
  - Build and Test
  - Lint Code Base
  - HIL Tests
  - CodeQL Security Analysis
  - Build & Deploy to Production

Quick check:

```bash
gh run list --branch main --limit 20
```

---

## 2) TPM Gate (emulated in CI + optional physical verification)

### Required (release minimum)
- [ ] `tests/hil/test_tpm_emulated.py` passes
- [ ] TPM emulator tooling validated (`swtpm`, `tpm2-tools`)
- [ ] No regressions in TPM quote/attestation code paths

Run:

```bash
pytest -q tests/hil/test_tpm_emulated.py
```

### Recommended (production hardening)
- [ ] Physical TPM run executed on self-hosted runner
- [ ] Physical TPM report attached to release artifact set

---

## 3) NPU Gate (selection + performance evidence)

### Required (release minimum)
- [ ] `tests/hil/test_npu_device_selection.py` passes
- [ ] Device fallback behavior validated: NPU -> CUDA -> CPU
- [ ] `FORCE_CPU` override validated

Run:

```bash
pytest -q tests/hil/test_npu_device_selection.py
```

### Required for “NPU performance claim”
- [ ] Real NPU hardware benchmark run completed
- [ ] Throughput + latency metrics exported to `test-results/`
- [ ] Grafana panel screenshots captured (NPU utilization + speedup)

---

## 4) GPU Gate (real hardware + scaling)

### Required for “GPU acceleration claim”
- [ ] Real CUDA/ROCm benchmark run completed
- [ ] Contention + round-latency tests executed
- [ ] Result analysis generated and checked into artifacts

Commands:

```bash
python gpu-test-suite.py --benchmark
python gpu-test-suite.py --contention --nodes 20
python gpu-test-suite.py --round-latency --nodes 20
python analyze-gpu-results.py
```

### Recommended (production hardening)
- [ ] Multi-GPU coordination test completed
- [ ] Cloud GPU validation (AWS/Azure/GCP) completed

---

## 5) Federated/Byzantine Evidence Gate

- [ ] 70-99 threshold ratio sweep artifact exists and is current
- [ ] Plots regenerated and stored in artifact bundle
- [ ] Break-point statement in docs matches current run output

Commands:

```bash
python byzantine-stress-test-suite.py --threshold-ratios 70,75,80,85,90,95,99
python generate-byzantine-test-suite-plots.py
```

---

## 6) Packaging + Documentation Gate

- [ ] Artifact directory is complete: `test-results/tpm-npu-full/`
- [ ] Tarball refreshed: `test-results/tpm-npu-full-artifacts.tar.gz`
- [ ] README/INDEX/release-notes wording uses claim-safe language
- [ ] Claim boundaries aligned with `CI_STATUS_AND_CLAIMS.md`

Package:

```bash
tar -czf test-results/tpm-npu-full-artifacts.tar.gz -C test-results tpm-npu-full
```

---

## 7) Final Release Decision

**Ship-ready when all are true:**
- [ ] Section 1 green
- [ ] Section 2 required items green
- [ ] Section 3 required items green
- [ ] Section 6 green

**Ship-ready for full hardware-performance claims when all are true:**
- [ ] All above + Section 3 performance items green
- [ ] Section 4 GPU claim items green
- [ ] Physical TPM (Section 2 recommended) green

---

## 8) Commit Review Snapshot (2026-03-03)

Target commit reviewed:
- `986c69bf75caa581b2db9254a6c52d108e3b064d`

Validation references:
- [x] Commit reviewed on GitHub: https://github.com/rwilliamspbg-ops/Sovereign_Map_Federated_Learning/commit/986c69bf75caa581b2db9254a6c52d108e3b064d
- [x] Test/results/artifacts in commit are visible
- [x] Summary report added: `COMMIT_986c69b_VALIDATION.md`
- [ ] All workflow actions green for this SHA (blocked by failed `Lint Code Base` run)

Blocking run:
- `Lint Code Base`: https://github.com/rwilliamspbg-ops/Sovereign_Map_Federated_Learning/actions/runs/22641864585
- Maintainer recovery steps: `CI_LINT_RECOVERY_PLAYBOOK.md`
- Workflow hardening applied in repo: `.github/workflows/lint.yml` supports `workflow_dispatch`; default lint scope remains changed-file mode (`VALIDATE_ALL_CODEBASE: false`)

Resolution status (follow-up commit):
- Follow-up SHA: `900d663b4dc82d8f560db6fd35f107be562e5ed5`
- Required workflow set on this SHA is ✅ green:
  - Build and Test: https://github.com/rwilliamspbg-ops/Sovereign_Map_Federated_Learning/actions/runs/22642514627
  - Lint Code Base: https://github.com/rwilliamspbg-ops/Sovereign_Map_Federated_Learning/actions/runs/22642514639
  - HIL Tests: https://github.com/rwilliamspbg-ops/Sovereign_Map_Federated_Learning/actions/runs/22642514633
  - CodeQL Security Analysis: https://github.com/rwilliamspbg-ops/Sovereign_Map_Federated_Learning/actions/runs/22642514667
  - 🚀 Build & Deploy to Production: https://github.com/rwilliamspbg-ops/Sovereign_Map_Federated_Learning/actions/runs/22642514650

---

## Sign-off Record

- Release date: __________
- Commit SHA: __________
- Signed by: __________
- Notes/Exceptions: ______________________________________
