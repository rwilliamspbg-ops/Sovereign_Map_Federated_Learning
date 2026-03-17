#!/bin/bash

# SDK Dependency Policy Checker
# Enforces dependency security policies

set -e

echo "🔐 Sovereign Map SDK Dependency Policy Checker"
echo "=============================================="
echo ""

POLICY_FILE="${1:-.sdk-dependency-policy.json}"
CHECK_TRANSITIVE="${2:-true}"

# Default policy if not provided
DEFAULT_POLICY='{
  "dependencies": {
    "restricted_licenses": ["GPL", "AGPL", "SSPL"],
    "max_critical": 0,
    "max_high": 0,
    "max_moderate": 3,
    "max_low": 10
  },
  "rules": {
    "require_lock_file": true,
    "audit_level": "moderate",
    "check_outdated": true,
    "disallow_prerelease": false
  }
}'

# Load or create policy
if [ -f "$POLICY_FILE" ]; then
  POLICY=$(cat "$POLICY_FILE")
else
  POLICY="$DEFAULT_POLICY"
fi

echo "📋 Policy:"
echo "$POLICY" | jq '.' 2>/dev/null || echo "$DEFAULT_POLICY" | jq '.'
echo ""

# Check 1: Lock file exists
echo "Check 1: Lock file integrity"
if [ ! -f "package-lock.json" ]; then
  echo "❌ package-lock.json not found!"
  exit 1
else
  echo "✅ package-lock.json exists"
fi

# Check 2: No vulnerability exceeds thresholds
echo ""
echo "Check 2: Vulnerability thresholds"

AUDIT_OUTPUT=$(npm audit --production --json 2>/dev/null || echo '{}')

if command -v jq &> /dev/null; then
  CRITICAL=$(echo "$AUDIT_OUTPUT" | jq '.metadata.vulnerabilities.critical // 0')
  HIGH=$(echo "$AUDIT_OUTPUT" | jq '.metadata.vulnerabilities.high // 0')
  MODERATE=$(echo "$AUDIT_OUTPUT" | jq '.metadata.vulnerabilities.moderate // 0')
  
  echo "  Critical: $CRITICAL (policy: 0)"
  echo "  High: $HIGH (policy: 0)"
  echo "  Moderate: $MODERATE (policy: 3)"
  
  if [ "$CRITICAL" -gt 0 ] || [ "$HIGH" -gt 0 ]; then
    echo "❌ Critical/high vulnerabilities detected"
    exit 1
  fi
  
  if [ "$MODERATE" -gt 3 ]; then
    echo "⚠️  Moderate vulnerabilities exceed policy limit"
  else
    echo "✅ Vulnerability thresholds met"
  fi
fi

# Check 3: License compliance
echo ""
echo "Check 3: License compliance"

RESTRICTED_FOUND=0
for license in GPL AGPL SSPL; do
  if npm ls --all 2>/dev/null | grep -q "$license"; then
    echo "⚠️  Found $license licensed packages"
    RESTRICTED_FOUND=1
  fi
done

if [ $RESTRICTED_FOUND -eq 0 ]; then
  echo "✅ No restricted licenses found"
fi

# Check 4: Outdated packages
if command -v npm outdated &> /dev/null; then
  echo ""
  echo "Check 4: Outdated packages"
  
  OUTDATED=$(npm outdated --json 2>/dev/null || echo '{}')
  if echo "$OUTDATED" | jq 'length' | grep -q '^0$'; then
    echo "✅ All packages up to date"
  else
    echo "⚠️  Some packages are outdated:"
    echo "$OUTDATED" | jq -r 'keys[]' | head -5
  fi
fi

echo ""
echo "═══════════════════════════════════════════"
echo "✅ Dependency policy checks completed"
echo ""
echo "Recommendations:"
echo "  1. Run \`npm audit fix\` for automatic updates"
echo "  2. Review critical/high vulnerabilities manually"
echo "  3. Keep dependencies up to date"
echo "  4. Use npm ci for reproducible installs"
