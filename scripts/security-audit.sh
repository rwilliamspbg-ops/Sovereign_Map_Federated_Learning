#!/bin/bash

# SDK Security Audit Script
# Comprehensive security scanning for dependencies and code

set -e

AUDIT_LEVEL="${1:-moderate}"  # moderate, high, critical
QUIET="${2:-false}"

echo "🔒 Sovereign Map SDK Security Audit"
echo "===================================="
echo "Audit Level: $AUDIT_LEVEL"
echo ""

# Color codes
RED='\033[0;31m'
YELLOW='\033[1;33m'
GREEN='\033[0;32m'
NC='\033[0m' # No Color

exit_code=0

# Function to run audit and capture results
run_audit() {
  local level=$1
  echo "🔍 Running npm audit (level: $level)..."
  
  if npm audit --omit=dev --audit-level=$level 2>&1 | tee /tmp/audit-results.txt; then
    echo -e "${GREEN}✅ No vulnerabilities at level: $level${NC}"
    return 0
  else
    echo -e "${RED}❌ Vulnerabilities found at level: $level${NC}"
    return 1
  fi
}

# Function to parse and display audit results
parse_audit_results() {
  echo ""
  echo "📊 Audit Results Summary:"
  echo ""
  
  if grep -q "found 0 vulnerabilities" /tmp/audit-results.txt; then
    echo -e "${GREEN}✅ No vulnerabilities detected${NC}"
  else
    # Try to extract vulnerability counts
    grep -E "dependencies|vulnerabilities" /tmp/audit-results.txt | head -5
    
    # Show severity breakdown
    echo ""
    echo "Severity breakdown:"
    grep -E "critical|high|moderate|low" /tmp/audit-results.txt || echo "  (See full report above)"
  fi
}

# Function to check for GPL/AGPL licenses
check_licenses() {
  echo ""
  echo "📋 Checking for restricted licenses (GPL/AGPL)..."
  
  if npm ls --all 2>/dev/null | grep -i "gpl\|agpl"; then
    echo -e "${YELLOW}⚠️  Found GPL/AGPL licenses - review required${NC}"
    return 1
  else
    echo -e "${GREEN}✅ No restricted GPL/AGPL licenses${NC}"
    return 0
  fi
}

# Function to display fixes
show_fixes() {
  echo ""
  echo "🔧 Suggested Fixes:"
  echo ""
  echo "1. Automatic fix (if available):"
  echo "   npm audit fix"
  echo ""
  echo "2. Manual review:"
  echo "   npm audit --omit=dev | grep vulnerability"
  echo ""
  echo "3. Update specific package:"
  echo "   npm update <package-name>"
  echo ""
  echo "4. Check for newer versions:"
  echo "   npm outdated"
}

# Run audits
echo "Step 1: Vulnerability Audit"
if ! run_audit "$AUDIT_LEVEL"; then
  exit_code=1
fi

parse_audit_results

echo ""
echo "Step 2: License Compliance"
if ! check_licenses; then
  exit_code=1
fi

# Summary
echo ""
echo "═══════════════════════════════════════════"
if [ $exit_code -eq 0 ]; then
  echo -e "${GREEN}✅ Security audit passed${NC}"
  echo "All vulnerabilities are below threshold ($AUDIT_LEVEL)"
else
  echo -e "${RED}❌ Security audit found issues${NC}"
  show_fixes
fi

echo ""
echo "Audit Level: $AUDIT_LEVEL"
echo "Timestamp: $(date -u +%Y-%m-%dT%H:%M:%SZ)"

exit $exit_code
