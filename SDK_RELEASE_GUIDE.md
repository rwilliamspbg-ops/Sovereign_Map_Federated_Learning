# Sovereign Map SDK Release Guide

## Overview

This guide covers the release process for the Sovereign Map SDK, including version management, prerelease channels, and supply-chain security.

## Release Phases

### Phase 1: Prepare Changes
1. Develop features and fixes
2. Add test coverage as needed
3. Commit with meaningful messages (uses conventional commits)
4. Create changeset: `npm run changeset`

### Phase 2: Version Bump
1. Run changeset version: `npm run sdk:version`
   - Runs full test suite with coverage gates
   - Updates package.json versions
   - Generates CHANGELOG entries
   - Builds all packages
2. Review generated changes in version PR
3. Merge version PR to main

### Phase 3: Publish
1. Publish to npm: `npm run sdk:publish`
   - Publishes all 4 packages simultaneously
   - Requires NPM_TOKEN secret in GitHub Actions
2. Generate SBOMs: `npm run sdk:sbom`
   - Creates Software Bill of Materials in `.sbom/` directory
   - Lists all dependencies for security audits
3. Generate release notes: `npm run sdk:release-notes <version>`
   - Creates detailed release notes with:
     - Changelog summary
     - Installation instructions
     - Security artifacts info
     - API stability guarantees
4. Create GitHub release with attachments

### Phase 4: Manage Channels
1. Set release channel: `npm run sdk:channels <version> <channel>`
   - `latest` - Recommended for most users (production-ready)
   - `next` - Prerelease/beta versions
   - `stable` - Long-term support releases

## Release Channels

### Latest (Default)
- **Audience**: Most users (recommended)
- **Version**: Stable, semantic versioning
- **Lifecycle**: Active maintenance
- **Install**: `npm install @sovereignmap/core` (auto-installs latest)
- **npm Tag**: `latest`
- **Examples**: 0.2.0, 0.3.0, 1.0.0

### Next (Prerelease)
- **Audience**: Early adopters, testing
- **Version**: Beta/RC versions (e.g., 0.3.0-beta.1, 1.0.0-rc.1)
- **Lifecycle**: Short-lived, moves to latest when stable
- **Install**: `npm install @sovereignmap/core@next`
- **npm Tag**: `next`
- **Note**: Not recommended for production use

### Stable (LTS)
- **Audience**: Enterprise/long-term deployments
- **Version**: Major versions with extended support
- **Lifecycle**: 2+ years of bug fix support
- **Install**: `npm install @sovereignmap/core@stable`
- **npm Tag**: `stable`
- **Examples**: 1.0.0, 2.0.0
- **Note**: Requires special designation during release

## Versioning Strategy

Follows [Semantic Versioning 2.0.0](https://semver.org/) with phases:

### 0.x.y Phase (Alpha, Current)
- **0.1.0-alpha.N**: Pre-release builds with unstable APIs
- **0.2.0-0.9.z**: Stable releases with public APIs
- **Breaking Changes**: Allowed with 2-minor-version notice (e.g., 0.5 → 0.7)
- **Support**: 3-month bug fix window per release

### 1.0.0+ Phase (Stable)
- **1.x.y**: Stable with strong API guarantees
- **Breaking Changes**: Only on major version (requires 6+ month notice)
- **Support**: See LTS policy

## Release Checklist

### Before Release
- [ ] All PRs merged to main
- [ ] CI passing (tests, coverage gates)
- [ ] Changelog reviewed and complete
- [ ] API changes documented
- [ ] Security review completed (if needed)

### Release Steps
```bash
# 1. Create changeset
npm run changeset
git add .changeset/*.md
git commit -m "chore(changesets): prepare v0.2.0 release"
git push

# 2. Wait for version PR and merge
# GitHub actions will create a PR automatically

# 3. When version PR is merged, automated publish begins
# (or manually trigger with):
npm run sdk:publish

# 4. Generate SBOMs and attestation
npm run sdk:sbom

# 5. Generate release notes
npm run sdk:release-notes 0.2.0

# 6. Set release channel
npm run sdk:channels 0.2.0 latest
```

### After Release
- [ ] GitHub release created with notes
- [ ] SBOMs attached to release
- [ ] Provenance attestation verified
- [ ] Announce on social channels
- [ ] Update documentation if needed
- [ ] File follow-up issues for next release

## Supply Chain Security

### Software Bill of Materials (SBOM)

Each release includes CycloneDX format SBOMs:
- Location: `.sbom/*-sbom.json`
- Format: CycloneDX 1.4
- Contents: Direct and transitive dependencies
- Usage:
  - Vulnerability scanning (e.g., with OWASP Dependency-Check)
  - License compliance audits
  - Supply-chain security reviews

### Provenance Attestation

Planned for Phase 4 final milestone:
- **Framework**: Sigstore (cosign)
- **Format**: SLSA provenance format
- **Verification**: `cosign verify-attestation` commands
- **Artifacts**: Attached to GitHub release

To verify provenance when available:
```bash
cosign verify-attestation \
  --certificate-identity https://github.com/rwilliamspbg-ops/Sovereign_Map_Federated_Learning/.github/workflows/sdk-publish.yml@refs/tags/v0.2.0 \
  --certificate-oidc-issuer https://token.actions.githubusercontent.com \
  ghcr.io/sovereignmap/sdk@sha256:...
```

### Dependency Policy

- **Pinned Versions**: All direct dependencies pinned in package-lock.json
- **Lock Files**: Committed to git for reproducible builds
- **Updates**: Automated via Dependabot, quarterly security audits
- **Vulnerability Scanning**: npm audit in CI/CD pipeline

## API Stability & Deprecation

### Stability Tiers
- **Stable**: Public APIs guaranteed through next major version
- **Subject to Change**: May change with 2-minor-version notice
- **Internal**: No compatibility guarantees

### Deprecation Process
1. Add `@deprecated` JSDoc tag with version and replacement
2. Emit runtime warning in next minor release
3. Document in CHANGELOG.md
4. Remove in release 2 minor versions later

Example:
```typescript
/**
 * @deprecated Use `newMethod()` instead. Will be removed in 0.5.0.
 */
async oldMethod(): Promise<void> { ... }
```

## Performance Baselines

Track performance across releases:

```bash
# Run benchmarks
npm run bench --prefix packages/core

# Compare to previous release
npm run bench --prefix packages/core -- --compare v0.1.0
```

## Hotfix Releases

For critical security or bug fixes:

```bash
# 1. Check out latest release tag
git checkout v0.2.0

# 2. Create hotfix branch
git checkout -b hotfix/security-fix

# 3. Apply fix and test
git commit -m "fix(core): security patch for XYZ"

# 4. Increment patch version manually
# Edit packages/*/package.json to 0.2.1

# 5. Create changeset
npm run changeset

# 6. Publish
npm run sdk:publish

# 7. Tag release
git tag v0.2.1
git push origin v0.2.1
```

## Release Notes Template

```markdown
# Sovereign Map SDK v0.2.0

## What's New
- Feature 1
- Feature 2  
- Bug fix 1

## Breaking Changes
- If any, document migration path

## Installation
\`\`\`bash
npm install @sovereignmap/{core,privacy,consensus,island}@0.2.0
\`\`\`

## Security
- SBOM: `.sbom/*.json`
- Provenance: See attestation in release assets

## Thank You
Contributors and community members who made this release possible.
```

## FAQ

**Q: How often are releases scheduled?**  
A: Every 2 weeks for feature/fix releases, with hotfixes as needed for security issues.

**Q: What's the difference between `latest` and `stable`?**  
A: `latest` gets latest features/fixes, `stable` receives only critical patches for 2+ years.

**Q: Can I use prerelease versions in production?**  
A: Not recommended. Prerelease versions are for testing; use `latest` or `stable` for production.

**Q: How do I report a security vulnerability?**  
A: Email security@sovereignmap.dev instead of using GitHub issues.

**Q: What's your Node.js version support policy?**  
A: Support current and one prior LTS version. Currently: 18+, 20+, 22+.

---

## Related Documentation

- [SDK API Stability Policy](./SDK_API_STABILITY.md)
- [CHANGELOG.md](./CHANGELOG.md)
- [Package-specific CHANGELOGs](./packages/)
- [Contributing Guide](./CONTRIBUTING.md)
