# CodeQL Configuration

## Overview

This repository uses [CodeQL](https://codeql.github.com/) for automated security analysis. CodeQL scans the codebase for potential security vulnerabilities and code quality issues.

## Configuration

### Analyzed Languages

CodeQL is configured to analyze:
- **Go** - Backend services and core packages
- **JavaScript/TypeScript** - Frontend and Node.js packages
- **Python** - Backend scripts and federated learning components

### Excluded Languages

**Java/Kotlin** is currently excluded from analysis because:
- The mobile app in `mobile-apps/android-node-app/` is a reference implementation
- It lacks a complete Gradle build configuration
- Build dependencies are not fully specified
- Manual build steps would be required for analysis

**Future**: When the mobile app build is fully configured, add `'java-kotlin'` to the language matrix in `.github/workflows/codeql-analysis.yml`.

## Ignored Paths

The following directories are excluded from CodeQL analysis:

- Workflow-level ignore config: `.github/codeql/codeql-config.yml`
- Additional repository-wide ignore patterns: `.codeqlignore`

Key excluded directories:

- `mobile-apps/` - Incomplete Android app
- `go-mobile/` - Mobile bindings (generated code)
- `archive/` - Legacy code
- `test-data/`, `test-results/`, `results/` - Test artifacts
- `node_modules/`, `vendor/` - Dependencies
- Generated protobuf files (`*_pb.go`, `*.pb.go`)

## Running CodeQL

### Automatic Scans

CodeQL runs automatically on:
- Every push to `main` branch
- Every pull request to `main` branch
- Weekly (Mondays at midnight UTC)

### Manual Scan

To manually trigger a CodeQL scan:

1. Go to **Actions** tab in GitHub
2. Select **CodeQL Security Analysis** workflow
3. Click **Run workflow**

### Viewing Results

CodeQL results are available in:
- **Security** tab → **Code scanning alerts**
- Pull request checks (for PR scans)
- Workflow run logs

## Query Suites

CodeQL uses two query suites:
- **security-extended** - Extended security queries
- **security-and-quality** - Security + code quality queries

This provides comprehensive coverage of:
- SQL injection
- Cross-site scripting (XSS)
- Command injection
- Path traversal
- Insecure deserialization
- Cryptographic issues
- Authentication bypass
- And 100+ other vulnerability patterns

## Fixing CodeQL Alerts

When CodeQL reports an alert:

1. **Review the Alert**: Check the security explanation and code path
2. **Assess Severity**: Prioritize Critical and High severity issues
3. **Implement Fix**: Follow CodeQL's remediation guidance
4. **Verify Fix**: Ensure tests pass and functionality is preserved
5. **Close Alert**: The alert auto-closes on next scan after fix is merged

## Troubleshooting

### "No Java/Kotlin code found" Error

If this appears together with messages like:
- `Java was extracted with build-mode set to 'none'`
- `Could not process Kotlin files without a build`
- `Required Gradle version not specified`

then the repository is usually running **GitHub Default Setup** in parallel with this advanced workflow.

**Resolution**:
1. Open **Security** → **Code scanning** → **CodeQL** settings
2. Disable/remove **Default setup**
3. Keep only `.github/workflows/codeql-analysis.yml` (advanced setup)
4. Re-run the workflow from **Actions**

After that, CodeQL should analyze only the configured matrix languages and stop attempting Java/Kotlin extraction.

### Build Failures

If CodeQL autobuild fails:

1. Check the workflow run logs
2. Ensure language setup steps succeeded
3. Verify `go.mod`, `package.json`, or `requirements.txt` are valid
4. Consider adding manual build steps if needed

### False Positives

To suppress a false positive:

1. Add a comment near the code: `// CodeQL [issue-id]: Reason for suppression`
2. Or add the path to `.codeqlignore` if appropriate

## Advanced Configuration

### Custom Queries

To add custom CodeQL queries:

1. Create `.github/codeql/queries/` directory
2. Add `.ql` query files
3. Update workflow to include: `config-file: .github/codeql/codeql-config.yml`

### Language-Specific Build

For languages requiring compilation:

```yaml
- name: Manual Build
  if: matrix.language == 'go'
  run: |
    go build ./...
```

## Security Best Practices

1. **Review alerts promptly** - Don't let security issues accumulate
2. **Keep dependencies updated** - Many vulnerabilities come from dependencies
3. **Enable branch protection** - Require CodeQL checks to pass before merging
4. **Regular scans** - Weekly scans catch new vulnerability patterns
5. **Security policy** - See [SECURITY.md](../SECURITY.md) for reporting vulnerabilities

## Resources

- [CodeQL Documentation](https://codeql.github.com/docs/)
- [Supported Languages](https://codeql.github.com/docs/codeql-overview/supported-languages-and-frameworks/)
- [Query Reference](https://codeql.github.com/codeql-query-help/)
- [CodeQL CLI](https://codeql.github.com/docs/codeql-cli/)

## Contact

For questions about CodeQL configuration or security scanning:
- Open an issue in this repository
- Review [SECURITY.md](../SECURITY.md) for security-specific inquiries
- Check GitHub's [Code Scanning documentation](https://docs.github.com/en/code-security/code-scanning)

---

**Last Updated**: February 28, 2026  
**CodeQL Version**: Latest (auto-updated by GitHub)
