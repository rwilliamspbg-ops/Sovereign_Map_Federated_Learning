# Pull Request

## Summary

Describe what changed and why.

## Validation

List exact commands you ran locally.

```bash
# Example
make smoke
```

## Evidence

Attach concrete proof for runtime-facing changes.

- HUD and/or Grafana screenshot links (or uploaded images)
- `/health` and `/ops/health` command output snippets
- If FL behavior changed: one `trigger_fl` round log snippet or metric delta

## Checklist

- [ ] CI workflows pass on this branch
- [ ] Security checks pass (including CodeQL)
- [ ] `make smoke` passes locally
- [ ] `make screenshots-check` passes locally (or N/A for non-runtime changes)
- [ ] Documentation updated for behavior/config changes
- [ ] No secrets or credentials committed
- [ ] Any `.github/workflows/*.yml` `uses:` refs are pinned to 40-char commit SHAs
- [ ] If deployment logic changed: `deploy.yml` staging and production paths validated
- [ ] If frontend assets changed: CDN publish/invalidation plan verified
- [ ] If Phase 3D changed: dataset/device/multi-GPU config compatibility validated

## Risk and Rollback

- Risk level: low / medium / high
- Rollback plan (if needed):
- Operational blast radius:
