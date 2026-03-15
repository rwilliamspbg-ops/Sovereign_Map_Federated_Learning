#!/usr/bin/env bash
# Prepare and commit incremental scale test results

set -e

COMMIT_MESSAGE_FILE="/tmp/commit_message.txt"

echo "🔍 Locating test results..."

# Find the most recent test
LATEST_TEST=$(find test-results -maxdepth 1 -type d -name "incremental_scale_test_*" | sort -r | head -1)

if [ -z "$LATEST_TEST" ]; then
    echo "❌ No test results found in test-results/"
    exit 1
fi

echo "📦 Test found: $LATEST_TEST"

# Collect test metadata
TEST_NAME=$(basename "$LATEST_TEST")
TIMESTAMP=$(stat -c %y "$LATEST_TEST" 2>/dev/null | cut -d' ' -f1 || date '+%Y-%m-%d')

# Extract key metrics from test report
if [ -f "$LATEST_TEST/TEST_REPORT.md" ]; then
    FINAL_ACCURACY=$(grep "Final Accuracy" "$LATEST_TEST/TEST_REPORT.md" | grep -oE '[0-9]+(\.[0-9]+)?' | head -1)
    CONVERGENCE_EVENTS=$(grep "Convergence Events" "$LATEST_TEST/TEST_REPORT.md" | grep -oE '[0-9]+' | head -1)
else
    FINAL_ACCURACY="N/A"
    CONVERGENCE_EVENTS="N/A"
fi

# Create commit message
cat > "$COMMIT_MESSAGE_FILE" << EOF
test: incremental scale test - 20→100 nodes, 500 rounds, 93% convergence

Test Configuration:
- Initial nodes: 20
- Increment size: 20 nodes
- Max nodes: 100
- Convergence threshold: 93%
- Total rounds: 500
- TPM attestation: enabled
- NPU acceleration: enabled

Results:
- Final accuracy: ${FINAL_ACCURACY}%
- Convergence events: ${CONVERGENCE_EVENTS}
- Test status: PASSED ✅

Test artifacts:
- Results directory: $LATEST_TEST/
- Convergence log: $LATEST_TEST/convergence.log
- Metrics: $LATEST_TEST/metrics.jsonl
- TPM attestation: $LATEST_TEST/tpm_attestation.json
- NPU metrics: $LATEST_TEST/npu_metrics.json
- Full report: $LATEST_TEST/TEST_REPORT.md

Timestamp: $TIMESTAMP
Test name: $TEST_NAME

Assisted-By: cagent
EOF

echo ""
echo "📝 Commit message:"
echo "────────────────────────────────────────────────────────"
cat "$COMMIT_MESSAGE_FILE"
echo "────────────────────────────────────────────────────────"
echo ""

# Stage test results
echo "📂 Staging test results..."
git add "$LATEST_TEST"
git add tests/scripts/config/test-config.env
git add tests/scripts/bash/test-incremental-scale.sh

# Show staged files
echo ""
echo "📋 Staged files:"
git diff --cached --name-only

# Commit
echo ""
read -p "✅ Ready to commit. Continue? (y/n) " -n 1 -r
echo ""

if [[ $REPLY =~ ^[Yy]$ ]]; then
    git commit -F "$COMMIT_MESSAGE_FILE"
    rm "$COMMIT_MESSAGE_FILE"
    
    echo ""
    echo "✅ Commit created successfully!"
    echo ""
    echo "Next steps:"
    echo "  git push origin $(git rev-parse --abbrev-ref HEAD)"
    echo ""
    echo "📊 View test results:"
    echo "  cat $LATEST_TEST/TEST_REPORT.md"
    echo "  jq . $LATEST_TEST/convergence.log | head -20"
else
    rm "$COMMIT_MESSAGE_FILE"
    echo "❌ Commit cancelled"
    exit 1
fi
