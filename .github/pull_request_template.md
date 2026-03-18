## Summary

Describe what changed and why.

## Validation

List exact commands you ran locally.

```bash
# Example
make smoke
```

## Checklist

- [ ] CI workflows pass on this branch
- [ ] Security checks pass (including CodeQL)
- [ ] `make smoke` passes locally
- [ ] Documentation updated for behavior/config changes
- [ ] No secrets or credentials committed
- [ ] Any `.github/workflows/*.yml` `uses:` refs are pinned to 40-char commit SHAs
- [ ] If deployment logic changed: `deploy.yml` staging and production paths validated
- [ ] If frontend assets changed: CDN publish/invalidation plan verified
- [ ] If Phase 3D changed: dataset/device/multi-GPU config compatibility validated

## Risk and Rollback

- Risk level: low / medium / high
- Rollback plan (if needed):
