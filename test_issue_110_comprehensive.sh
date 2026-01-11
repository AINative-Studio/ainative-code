#!/bin/bash
# Comprehensive test for Issue #110: --config flag validation
# Tests all edge cases and scenarios

set -e

echo "=================================================="
echo "Issue #110 Comprehensive Test Suite"
echo "Testing --config flag validation"
echo "=================================================="
echo ""

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Build fresh binary
echo "Building fresh binary..."
go build -o /tmp/ainative-test-110 ./cmd/ainative-code/ || exit 1
echo -e "${GREEN}✓${NC} Build successful"
echo ""

PASSED=0
FAILED=0

# Test function
run_test() {
    local test_name="$1"
    local command="$2"
    local expected_pattern="$3"
    local should_succeed="$4"

    echo "Test: $test_name"
    echo "Command: $command"

    if eval "$command" 2>&1 | grep -q "$expected_pattern"; then
        if [ "$should_succeed" = "true" ]; then
            echo -e "${GREEN}✓ PASS${NC}: Found expected output"
            ((PASSED++))
        else
            echo -e "${GREEN}✓ PASS${NC}: Correctly detected error"
            ((PASSED++))
        fi
    else
        echo -e "${RED}✗ FAIL${NC}: Did not find expected pattern: $expected_pattern"
        ((FAILED++))
        echo "Actual output:"
        eval "$command" 2>&1 | head -10
    fi
    echo ""
}

# Test 1: Nonexistent file
run_test \
    "Nonexistent config file" \
    "/tmp/ainative-test-110 --config /totally/nonexistent/path/config.yaml version" \
    "Error: config file not found" \
    "false"

# Test 2: Nonexistent file in current directory
run_test \
    "Nonexistent file in current directory" \
    "/tmp/ainative-test-110 --config ./nonexistent.yaml version" \
    "Error: config file not found" \
    "false"

# Test 3: Directory instead of file
mkdir -p /tmp/test-config-dir-110
run_test \
    "Directory instead of file" \
    "/tmp/ainative-test-110 --config /tmp/test-config-dir-110 version" \
    "Error: config path is a directory" \
    "false"

# Test 4: Valid config file (absolute path)
echo "provider: openai" > /tmp/valid-config-110.yaml
run_test \
    "Valid config file (absolute path)" \
    "/tmp/ainative-test-110 --config /tmp/valid-config-110.yaml version" \
    "AINative Code" \
    "true"

# Test 5: Valid config file (relative path)
cd /tmp
run_test \
    "Valid config file (relative path)" \
    "/tmp/ainative-test-110 --config ./valid-config-110.yaml version" \
    "AINative Code" \
    "true"

# Test 6: Config file with spaces in path
mkdir -p "/tmp/my config dir"
echo "provider: anthropic" > "/tmp/my config dir/config.yaml"
run_test \
    "Config file with spaces in path" \
    "/tmp/ainative-test-110 --config '/tmp/my config dir/config.yaml' version" \
    "AINative Code" \
    "true"

# Test 7: No --config flag (default behavior)
run_test \
    "No --config flag (default behavior)" \
    "/tmp/ainative-test-110 version" \
    "AINative Code" \
    "true"

# Test 8: Empty config path
run_test \
    "Empty config path" \
    "/tmp/ainative-test-110 --config '' version" \
    "AINative Code" \
    "true"

# Test 9: Config file with special characters in filename
echo "provider: gemini" > /tmp/config-file_2024-test.yaml
run_test \
    "Config file with special characters" \
    "/tmp/ainative-test-110 --config /tmp/config-file_2024-test.yaml version" \
    "AINative Code" \
    "true"

# Test 10: Verify error message includes helpful guidance
echo -e "${YELLOW}Test: Error message contains guidance${NC}"
if /tmp/ainative-test-110 --config /nonexistent/file.yaml version 2>&1 | grep -q "Please check the path"; then
    echo -e "${GREEN}✓ PASS${NC}: Error message includes helpful guidance"
    ((PASSED++))
else
    echo -e "${RED}✗ FAIL${NC}: Error message lacks guidance"
    ((FAILED++))
fi
echo ""

# Cleanup
rm -f /tmp/ainative-test-110
rm -f /tmp/valid-config-110.yaml
rm -f /tmp/config-file_2024-test.yaml
rm -rf /tmp/test-config-dir-110
rm -rf "/tmp/my config dir"

# Summary
echo "=================================================="
echo "Test Results Summary"
echo "=================================================="
echo -e "Tests Passed: ${GREEN}$PASSED${NC}"
echo -e "Tests Failed: ${RED}$FAILED${NC}"
echo "Total Tests:  $((PASSED + FAILED))"
echo ""

if [ $FAILED -eq 0 ]; then
    echo -e "${GREEN}✓ All tests passed!${NC}"
    echo ""
    echo "Issue #110 is properly fixed:"
    echo "  ✓ Nonexistent config files are detected"
    echo "  ✓ Directories are rejected"
    echo "  ✓ Valid configs work correctly"
    echo "  ✓ Default behavior is preserved"
    echo "  ✓ Error messages are helpful"
    exit 0
else
    echo -e "${RED}✗ Some tests failed${NC}"
    exit 1
fi
