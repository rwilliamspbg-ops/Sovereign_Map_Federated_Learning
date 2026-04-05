# Sovereign Map SDK Security Policy

This document outlines security practices, vulnerability reporting, and supply-chain security commitments for the Sovereign Map SDK.

## Security Overview

The Sovereign Map SDK implements multiple layers of security:

1. **Build Integrity** — Reproducible builds with GitHub Actions
2. **Supply-Chain Transparency** — Software Bill of Materials (SBOM) for all releases
3. **Dependency Security** — Automated vulnerability scanning
4. **Code Security** — Static analysis with CodeQL
5. **Provenance** — SLSA-aligned provenance artifacts and reporting

## Reporting Security Vulnerabilities

### For Researchers & Security Professionals

If you discover a security vulnerability in the Sovereign Map SDK:

1. **Do NOT** open a public GitHub issue
2. Email **security@sovereignmap.dev** with:
   - Package name and affected version
   - Vulnerability description and impact
   - Steps to reproduce (if applicable)
   - Suggested fix (if available)

3. **Response Timeline**:
   - Acknowledgment: Within 24 hours
   - Initial assessment: Within 5 business days
   - Fix release: Within 30 days for critical issues
   - Public disclosure: Coordinated with reporter

### Example Email

```
Subject: [SECURITY] Vulnerability in @sovereignmap/core@0.2.0

Hello,

I discovered a potential security vulnerability in the Sovereign Map SDK:

- **Package**: @sovereignmap/core
- **Version**: 0.2.0
- **Issue**: [Brief description]
- **Severity**: [Critical/High/Medium/Low]
- **Impact**: [Description of potential impact]
- **PoC**: [Code or steps to reproduce]

Suggested fix: [If you have ideas for a fix]

Please advise on next steps.

Best regards,
[Your Name]
```

### Bounty Program

Sovereign Map may offer monetary rewards for responsibly-disclosed vulnerabilities. Details are provided after initial contact.

## PGP Key for Encrypted Communication

(To be added when security contact is established)

## Vulnerability Disclosure Process

### Phase 1: Private Fix (Day 0-30)
1. Security team assesses the vulnerability
2. Hotfix is prepared in a private branch
3. Tests are added to prevent regression
4. Code is reviewed and tested thoroughly
5. Backports are created for supported versions

### Phase 2: Staging Release (Day 30-35)
1. Security patch is released to staging npm tag
2. Partners and integrators are notified
3. 5-day coordination period for integration testing
4. Public disclosure is coordinated with reporter

### Phase 3: Public Release (Day 35+)
1. Security patch is published to `latest` npm tag
2. Security advisory is published
3. GitHub Security Advisory is filed
4. Full disclosure is posted on website

### Example: CVE Handling

If a vulnerability qualifies for a CVE:

1. CVE ID is requested from MITRE or Red Hat
2. CVSS score is calculated for severity assessment
3. Detailed advisory is published with:
   - Affected versions
   - Fixed versions
   - Mitigation steps
   - Acknowledgments

## Supply-Chain Security Practices

### Dependency Management

**Direct Dependencies**:
- All direct dependencies are explicitly listed in package.json
- Versions are pinned in package-lock.json
- Updates are tested before merging
- npm audit is run in CI/CD on scheduled scans and on matching push/PR events

**Transitive Dependencies**:
- Automatically inherited from direct dependency trees
- Scanned on scheduled daily npm audit runs
- Reported in Software Bill of Materials (SBOM)
- Updated when security patches are released upstream

**Update Strategy**:
- Security updates: Applied within 1 week
- Feature updates: Applied monthly
- Major version changes: Evaluated for breaking changes
- Deprecated dependencies: Replaced before removal upstream

**Lock File Integrity**:
- package-lock.json is committed to git
- Changes are reviewed in pull requests
- Lock files are verified in CI before builds
- Reproducible builds require lock file integrity

### Scanning & Monitoring

**Automated Scans**:
- npm audit runs on scheduled daily scans and on push/PR events that match security workflow path filters
- Vulnerabilities are checked with audit-level=moderate
- Critical/high severity failures block merge
- Schedule: Daily scans at 2 AM UTC

**Code Security**:
- CodeQL analysis on all TypeScript code
- Security and quality rules enabled
- Results reviewed before merge
- Findings tracked in GitHub Security tab

**License Compliance**:
- GPL/AGPL licenses flagged for review
- License changes require approval
- Incompatible licenses block release

### Release Artifacts

**Integrity**:
- All builds are reproducible
- Git commit SHA is recorded in provenance
- Build timestamps are logged
- Artifacts are hashed with SHA-256

**Attestation**:
- SLSA-aligned provenance records
- In-toto provenance format
- Supply-chain provenance recorded
- Verifiable build environment

**Distribution**:
- Published to npm registry only
- Never distributed via unofficial channels
- npm dist-tags prevent confusion (latest, next, stable)
- npm signing via NPM_TOKEN (environment-based)

## SLSA Framework Alignment

### Current Posture: L2-aligned controls (Hosted Build Service)

The SDK implements controls associated with SLSA Build L2 and records provenance artifacts in CI:

| Requirement | Status | Evidence |
|-------------|--------|----------|
| Build service is authenticated | ✅ | GitHub Actions with OAuth |
| Build service controls build environment | ✅ | Isolated Ubuntu runner |
| Build invocation is authenticated | ✅ | Webhook signed by GitHub |
| Builds cannot influence each other | ✅ | Separate runners per job |
| Build script is version controlled | ✅ | .github/workflows in git |
| Provenance is unavailable to attacker | ✅ | GitHub Actions logs |
| Build output includes provenance | ✅ | Attestation workflow |
| Provenance is ingestible | ✅ | in-toto JSON format |

### Path to Level 3 (Cryptographic Provenance)

Future enhancement to add:
- Sigstore/cosign integration
- Digital signatures on artifacts
- Threshold signing for major releases
- Transitive dependency provenance

### Verification

Users can verify SDK authenticity:

```bash
# Check SBOM for dependencies
npm view @sovereignmap/core

# Verify build provenance (when Sigstore enabled)
cosign verify-attestation \
  --certificate-identity https://github.com/rwilliamspbg-ops/Sovereign_Map_Federated_Learning \
  --certificate-oidc-issuer https://token.actions.githubusercontent.com \
  ghcr.io/sovereignmap/sdk@v0.2.0
```

## Security Best Practices for Users

### Installation

- Always use latest stable version: `npm install @sovereignmap/core`
- Pin versions in production: `"@sovereignmap/core": "0.2.0"`
- Audit your own dependencies: `npm audit`
- Use compatible npm versions (same as SDK: 18+)

### Configuration

- Never commit credentials (API keys, tokens)
- Use environment variables for secrets
- Enable hardware-backed security (TPM) when available
- Configure privacy budget conservatively

### Monitoring

- Subscribe to GitHub security updates
- Monitor npm audit results in CI
- Review dependency changes in pull requests
- Test major version upgrades in staging

### Reporting Issues

- Security issues: Email security@sovereignmap.dev
- Bug reports: GitHub Issues
- Feature requests: GitHub Discussions
- Questions: GitHub Discussions

## Security Contacts

| Role | Contact | Response Time |
|------|---------|----------------|
| Security Team | security@sovereignmap.dev | 24 hours |
| Maintainers | ryan@sovereignmap.dev | 48 hours |
| Issue Tracker | GitHub Issues | 5 business days |

## Changelog & Updates

Security updates are documented in:
- CHANGELOG.md — All releases and fixes
- packages/*/CHANGELOG.md — Package-specific details
- GitHub Security Advisories — CVE and vulnerability data
- Releases page — For each version

Subscribe to updates:
- GitHub: Watch repository (Release notifications)
- npm: `npm view @sovereignmap/core` (check latest version)
- Email: [Subscribe to mailing list - not yet available]

## FAQ

**Q: How often are security audits conducted?**  
A: Continuous via CI/CD, weekly scheduled scans, annual third-party audit (planned).

**Q: What's the minimum supported version for security patches?**  
A: Current minor version + 1 prior. E.g., if latest is 0.3.0, we backport to 0.2.z.

**Q: Are binary packages (WASM) scanned for vulnerabilities?**  
A: WASM source (Rust/C) is not currently scanned. Fallback to pure-JS implementation available.

**Q: Can I verify the integrity of releases offline?**  
A: Yes, via:
  - SBOM comparison (manual dependency audit)
  - Git commit verification (GPG signatures planned)
  - Hash verification against GitHub releases

**Q: What happens if a zero-day is discovered?**  
A: We follow responsible disclosure practices with 30-day coordination window before public release.

**Q: How do I stay informed about security updates?**  
A: Subscribe to GitHub release notifications or watch the repository.

## Related Documents

- [SDK API Stability Policy](./SDK_API_STABILITY.md) — Breaking change procedures
- [SDK Release Guide](./SDK_RELEASE_GUIDE.md) — Release process and procedures
- [SLSA Framework](https://slsa.dev/) — Supply-chain security standards
- [npm Security Best Practices](https://docs.npmjs.com/packages-and-modules/securing-your-code)
- [NIST Software Supply Chain Security](https://csrc.nist.gov/projects/supply-chain-risk-management)

## Acknowledgments

We thank security researchers who responsibly disclose vulnerabilities. Contributors will be acknowledged in relevant security advisories.

---

**Last Updated**: 2026-03-17  
**Version**: 1.0  
**SLSA Level**: L2 (Hosted Build Service)
