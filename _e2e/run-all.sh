#!/bin/bash

set -euo pipefail

# Get script directory
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"

# Get current branch name
CURRENT_BRANCH=$(git branch --show-current)
export GHA_BRANCH_NAME="${GHA_BRANCH_NAME:-$CURRENT_BRANCH}"

echo "===== Running CIUX E2E Tests ====="
echo "Project directory: $PROJECT_DIR"
echo "Branch: $GHA_BRANCH_NAME"
echo "======================================"

# Run ciux ignite
echo "🚀 Running ciux ignite --selector ci --branch=\"$GHA_BRANCH_NAME\" $PROJECT_DIR"
ciux ignite --selector ci --branch="$GHA_BRANCH_NAME" "$PROJECT_DIR"

echo ""
echo "✅ ciux ignite completed successfully"
echo ""

# Find and run all test scripts
echo "🧪 Running E2E tests..."
echo ""

# Counter for tests
total_tests=0
passed_tests=0
failed_tests=0

# Find all test scripts in the _e2e directory
for test_script in "$SCRIPT_DIR"/test_*.sh; do
    if [[ -f "$test_script" && -x "$test_script" ]]; then
        test_name=$(basename "$test_script")
        echo "Running $test_name..."

        total_tests=$((total_tests + 1))

        if "$test_script"; then
            echo "✅ $test_name PASSED"
            passed_tests=$((passed_tests + 1))
        else
            echo "❌ $test_name FAILED"
            failed_tests=$((failed_tests + 1))
        fi
        echo ""
    fi
done

# Summary
echo "======================================"
echo "E2E Test Summary:"
echo "Total tests: $total_tests"
echo "Passed: $passed_tests"
echo "Failed: $failed_tests"
echo "======================================"

if [[ $failed_tests -gt 0 ]]; then
    echo "❌ Some tests failed!"
    exit 1
else
    echo "✅ All tests passed!"
    exit 0
fi