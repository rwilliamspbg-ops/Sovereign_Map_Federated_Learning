# 🔧 Fix CodeQL Analysis Failure - Quick Guide

## ❌ Current Error
```
Error: Code Scanning could not process the submitted SARIF file:
CodeQL analyses from advanced configurations cannot be processed when the default setup is enabled
```

## 🎯 Solution: Disable Default CodeQL Setup

Your repository has **both** default CodeQL and a custom advanced workflow running. GitHub doesn't allow both simultaneously.

---

## 📋 Step-by-Step Instructions

### Option 1: Web Browser (Recommended - 30 seconds)

1. **Open this URL in your browser:**
   ```
   https://github.com/rwilliamspbg-ops/Sovereign_Map_Federated_Learning/settings/security_analysis
   ```

2. **Find "Code scanning" section**
   - Scroll down to the "Code scanning" area
   - Look for "CodeQL analysis" with a **"Default"** badge or status

3. **Disable Default Setup**
   - Click the **⋮** (three dots) or **"Configure"** dropdown next to CodeQL
   - Select **"Disable CodeQL"** or **"Remove configuration"**
   - Confirm the action

4. **Verify**
   - The default setup should now show as "Not configured"
   - Your custom workflow at `.github/workflows/codeql-analysis.yml` will continue to run

---

### Option 2: GitHub CLI (If you have admin token)

```bash
# This requires a personal access token with admin:repo_hook scope
gh api repos/rwilliamspbg-ops/Sovereign_Map_Federated_Learning/code-scanning/default-setup \
  -X PATCH \
  -f state='not-configured'
```

**Note**: The current GITHUB_TOKEN doesn't have these permissions (got 403 error).

---

## ✅ What Happens After Disabling

1. **Custom workflow continues**: Your advanced CodeQL workflow (`.github/workflows/codeql-analysis.yml`) keeps running
2. **Better analysis**: Custom workflow has:
   - Extended security queries: `security-extended,security-and-quality`
   - Multiple languages: Go, JavaScript/TypeScript, Python
   - Weekly scheduled scans
3. **No more conflicts**: SARIF file uploads will succeed

---

## 🔍 Why This Happened

- GitHub automatically enables **default** CodeQL setup on many repositories
- Your repository has a **custom/advanced** CodeQL workflow with more features
- GitHub prevents running both to avoid duplicate/conflicting scans
- Default setup is simpler but less comprehensive than your custom workflow

---

## 🚀 Quick Verification

After disabling, trigger a new workflow run:

```bash
# Push a small change or manually trigger the workflow
git commit --allow-empty -m "test: Trigger CodeQL after disabling default setup"
git push
```

Or manually trigger via GitHub Actions:
```
https://github.com/rwilliamspbg-ops/Sovereign_Map_Federated_Learning/actions/workflows/codeql-analysis.yml
```

---

## 📊 Your Custom Workflow Features

Your current workflow (`.github/workflows/codeql-analysis.yml`) provides:

| Feature | Default Setup | Your Custom Workflow |
|---------|---------------|---------------------|
| Languages | Limited | Go, JS/TS, Python |
| Query Suites | Basic security | Extended + Quality |
| Schedule | Not configurable | Weekly on Mondays |
| Build Control | Limited | Custom setup steps |
| Customization | None | Fully customizable |

**Keep your custom workflow** - it's more powerful!

---

## 🆘 Troubleshooting

### Still seeing the error after disabling?

1. **Wait 1-2 minutes** for GitHub to process the configuration change
2. **Re-run the failed workflow** from GitHub Actions page
3. **Check status**: Visit the settings page again to confirm default is disabled

### Can't access settings page?

- You need **admin or write access** to the repository
- Contact the repository owner: `rwilliamspbg-ops`
- Ask them to disable default CodeQL setup

### Want to use default setup instead?

If you prefer the simpler default setup:

```bash
# Rename custom workflow to disable it
mv .github/workflows/codeql-analysis.yml .github/workflows/codeql-analysis.yml.disabled

# Re-enable default setup in GitHub settings
# Then commit and push the change
```

---

## 📞 Need Help?

- **GitHub Documentation**: https://docs.github.com/en/code-security/code-scanning/creating-an-advanced-setup-for-code-scanning/configuring-advanced-setup-for-code-scanning
- **Repository Settings**: https://github.com/rwilliamspbg-ops/Sovereign_Map_Federated_Learning/settings/security_analysis
- **Actions Page**: https://github.com/rwilliamspbg-ops/Sovereign_Map_Federated_Learning/actions

---

## ✅ Success Checklist

- [ ] Opened repository security settings
- [ ] Found "Code scanning" section
- [ ] Disabled "Default" CodeQL setup
- [ ] Confirmed default shows "Not configured"
- [ ] Triggered new workflow run
- [ ] Verified workflow completes successfully
- [ ] No SARIF upload errors in logs

**Estimated time to fix**: 30 seconds to 2 minutes
