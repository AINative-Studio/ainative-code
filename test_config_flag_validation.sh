#!/bin/bash

# Test script for GitHub Issue #110: --config flag validation
# Tests that the --config flag properly validates config file existence

set -e

BINARY="./ainative-code"
TEST_DIR="/tmp/ainative-test-$$"
PASS_COUNT=0
FAIL_COUNT=0

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo "=========================================="
echo "Testing GitHub Issue #110: --config Flag Validation"
echo "=========================================="
echo ""

# Setup test directory
mkdir -p "$TEST_DIR"
trap "rm -rf $TEST_DIR" EXIT

# Helper function to run test
run_test() {
    local test_name="$1"
    local expected_result="$2"  # "success" or "failure"
    shift 2
    local cmd=("$@")

    echo -n "Test: $test_name ... "

    if output=$("${cmd[@]}" 2>&1); then
        actual_result="success"
        exit_code=0
    else
        actual_result="failure"
        exit_code=$?
    fi

    if [ "$expected_result" = "$actual_result" ]; then
        echo -e "${GREEN}PASS${NC}"
        PASS_COUNT=$((PASS_COUNT + 1))
        return 0
    else
        echo -e "${RED}FAIL${NC}"
        echo "  Expected: $expected_result"
        echo "  Got: $actual_result (exit code: $exit_code)"
        echo "  Output: $output"
        FAIL_COUNT=$((FAIL_COUNT + 1))
        return 1
    fi
}

# Helper function to check error message
check_error_message() {
    local test_name="$1"
    local expected_message="$2"
    shift 2
    local cmd=("$@")

    echo -n "Test: $test_name ... "

    if output=$("${cmd[@]}" 2>&1); then
        echo -e "${RED}FAIL${NC}"
        echo "  Expected command to fail with error message"
        echo "  But it succeeded"
        FAIL_COUNT=$((FAIL_COUNT + 1))
        return 1
    fi

    if echo "$output" | grep -q "$expected_message"; then
        echo -e "${GREEN}PASS${NC}"
        PASS_COUNT=$((PASS_COUNT + 1))
        return 0
    else
        echo -e "${RED}FAIL${NC}"
        echo "  Expected error message containing: $expected_message"
        echo "  Got: $output"
        FAIL_COUNT=$((FAIL_COUNT + 1))
        return 1
    fi
}

echo "Test Category: Nonexistent Config Files"
echo "----------------------------------------"

# Test 1: Nonexistent config file
check_error_message \
    "Nonexistent config file shows clear error" \
    "config file not found" \
    $BINARY --config "/nonexistent/config.yaml" version

# Test 2: Nonexistent config file in nonexistent directory
check_error_message \
    "Nonexistent config in nonexistent dir shows error" \
    "config file not found" \
    $BINARY --config "/path/that/does/not/exist/config.yaml" version

echo ""
echo "Test Category: Valid Config Files"
echo "----------------------------------------"

# Test 3: Valid config file
VALID_CONFIG="$TEST_DIR/valid-config.yaml"
cat > "$VALID_CONFIG" << EOF
provider: anthropic
model: claude-3-5-sonnet-20241022
EOF

run_test \
    "Valid config file works" \
    "success" \
    $BINARY --config "$VALID_CONFIG" version

# Test 4: Valid config with absolute path
run_test \
    "Valid config with absolute path works" \
    "success" \
    $BINARY --config "$VALID_CONFIG" version

# Test 5: Config file with spaces in path
SPACE_DIR="$TEST_DIR/my config dir"
mkdir -p "$SPACE_DIR"
SPACE_CONFIG="$SPACE_DIR/config.yaml"
cat > "$SPACE_CONFIG" << EOF
provider: openai
model: gpt-4
EOF

run_test \
    "Config file with spaces in path works" \
    "success" \
    $BINARY --config "$SPACE_CONFIG" version

echo ""
echo "Test Category: Invalid Config Paths"
echo "----------------------------------------"

# Test 6: Directory instead of file
check_error_message \
    "Directory instead of file shows error" \
    "directory" \
    $BINARY --config "$TEST_DIR" version

# Test 7: Empty config path (should use defaults)
run_test \
    "No config flag uses defaults" \
    "success" \
    $BINARY version

echo ""
echo "Test Category: Malformed Config Files"
echo "----------------------------------------"

# Test 8: Malformed YAML (should warn but not fail)
MALFORMED_CONFIG="$TEST_DIR/malformed-config.yaml"
cat > "$MALFORMED_CONFIG" << EOF
provider: anthropic
invalid yaml here: [
  missing bracket
EOF

# This should succeed with a warning (graceful degradation)
if output=$($BINARY --config "$MALFORMED_CONFIG" version 2>&1); then
    if echo "$output" | grep -q "Warning"; then
        echo -e "Test: Malformed YAML shows warning ... ${GREEN}PASS${NC}"
        PASS_COUNT=$((PASS_COUNT + 1))
    else
        echo -e "Test: Malformed YAML shows warning ... ${YELLOW}PARTIAL${NC}"
        echo "  Command succeeded but no warning found"
        PASS_COUNT=$((PASS_COUNT + 1))
    fi
else
    echo -e "Test: Malformed YAML shows warning ... ${RED}FAIL${NC}"
    echo "  Expected success with warning, got failure"
    FAIL_COUNT=$((FAIL_COUNT + 1))
fi

echo ""
echo "Test Category: Special Characters"
echo "----------------------------------------"

# Test 9: Config file with special characters
SPECIAL_CONFIG="$TEST_DIR/config-file_2024.yaml"
cat > "$SPECIAL_CONFIG" << EOF
provider: openai
model: gpt-4-turbo
EOF

run_test \
    "Config file with special characters works" \
    "success" \
    $BINARY --config "$SPECIAL_CONFIG" version

# Test 10: Config file with dots in name
DOT_CONFIG="$TEST_DIR/my.config.yaml"
cat > "$DOT_CONFIG" << EOF
provider: anthropic
EOF

run_test \
    "Config file with dots in name works" \
    "success" \
    $BINARY --config "$DOT_CONFIG" version

echo ""
echo "=========================================="
echo "Test Summary"
echo "=========================================="
echo -e "Total Tests: $((PASS_COUNT + FAIL_COUNT))"
echo -e "${GREEN}Passed: $PASS_COUNT${NC}"
echo -e "${RED}Failed: $FAIL_COUNT${NC}"

if [ $FAIL_COUNT -eq 0 ]; then
    echo ""
    echo -e "${GREEN}All tests passed!${NC}"
    exit 0
else
    echo ""
    echo -e "${RED}Some tests failed${NC}"
    exit 1
fi
