#!/bin/bash
# Test script for GitHub Issue #123: Session list with negative limit shows all sessions instead of error
# This script validates that the session list command properly rejects negative and zero limit values

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"
TEST_BINARY="$PROJECT_ROOT/ainative-code"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo "=========================================="
echo "Testing Issue #123: Session List Limit Validation"
echo "=========================================="
echo ""

# Build the project first
echo "Building ainative-code..."
cd "$PROJECT_ROOT"
go build -o "$TEST_BINARY" ./cmd/ainative-code
echo ""

# Function to run a test
run_test() {
    local test_name="$1"
    local command="$2"
    local expected_result="$3" # "success" or "error"
    local expected_error_msg="$4"

    echo -n "Test: $test_name... "

    # Run the command and capture output
    set +e
    output=$($command 2>&1)
    exit_code=$?
    set -e

    # Check if result matches expectation
    if [ "$expected_result" == "error" ]; then
        if [ $exit_code -ne 0 ] && [[ "$output" == *"$expected_error_msg"* ]]; then
            echo -e "${GREEN}PASS${NC}"
            return 0
        else
            echo -e "${RED}FAIL${NC}"
            echo "  Expected error with message: $expected_error_msg"
            echo "  Got exit code: $exit_code"
            echo "  Output: $output"
            return 1
        fi
    else
        if [ $exit_code -eq 0 ]; then
            echo -e "${GREEN}PASS${NC}"
            return 0
        else
            echo -e "${RED}FAIL${NC}"
            echo "  Expected success but got exit code: $exit_code"
            echo "  Output: $output"
            return 1
        fi
    fi
}

# Test 1: Negative limit should return error
run_test "Negative limit (-1)" \
    "$TEST_BINARY session list --limit -1" \
    "error" \
    "Error: limit must be a positive integer"

# Test 2: Zero limit should return error
run_test "Zero limit (0)" \
    "$TEST_BINARY session list --limit 0" \
    "error" \
    "Error: limit must be a positive integer"

# Test 3: Large negative limit should return error
run_test "Large negative limit (-999)" \
    "$TEST_BINARY session list -n -999" \
    "error" \
    "Error: limit must be a positive integer"

# Test 4: Positive limit should succeed
run_test "Positive limit (5)" \
    "$TEST_BINARY session list --limit 5" \
    "success" \
    ""

# Test 5: Default limit should succeed
run_test "Default limit (no flag)" \
    "$TEST_BINARY session list" \
    "success" \
    ""

# Test 6: Large positive limit should succeed
run_test "Large positive limit (1000)" \
    "$TEST_BINARY session list -n 1000" \
    "success" \
    ""

echo ""
echo "=========================================="
echo "All tests completed successfully!"
echo "=========================================="
echo ""
echo "Summary:"
echo "  - Negative limits are properly rejected"
echo "  - Zero limit is properly rejected"
echo "  - Positive limits work correctly"
echo "  - This prevents potential performance issues from unlimited queries"
echo ""

# Clean up
rm -f "$TEST_BINARY"

exit 0
