#!/bin/bash

# Test script for Issue #119: Chat with empty message validation
# This script verifies that empty messages are rejected locally without making API calls

set -e

echo "========================================"
echo "Testing Issue #119 Fix"
echo "Empty Message Validation in Chat Command"
echo "========================================"
echo ""

# Build the binary
echo "Building ainative-code binary..."
go build -o ainative-code-test ./cmd/ainative-code
echo "✓ Build successful"
echo ""

# Test counter
TESTS_RUN=0
TESTS_PASSED=0
TESTS_FAILED=0

# Function to run a test
run_test() {
    local test_name="$1"
    local command="$2"
    local expected_error="$3"
    local should_contain_error="$4" # if "yes", error should contain expected_error

    TESTS_RUN=$((TESTS_RUN + 1))
    echo "Test $TESTS_RUN: $test_name"

    # Run command and capture output
    set +e
    output=$(eval "$command" 2>&1)
    exit_code=$?
    set -e

    # Check if command failed (should fail for empty messages)
    if [ $exit_code -eq 0 ]; then
        echo "  ✗ FAILED: Command should have failed but succeeded"
        TESTS_FAILED=$((TESTS_FAILED + 1))
        return 1
    fi

    # Check if output contains expected error
    if [ "$should_contain_error" = "yes" ]; then
        if echo "$output" | grep -q "$expected_error"; then
            echo "  ✓ PASSED: Got expected error: $expected_error"
            TESTS_PASSED=$((TESTS_PASSED + 1))
            return 0
        else
            echo "  ✗ FAILED: Expected error '$expected_error' not found in output"
            echo "  Output was: $output"
            TESTS_FAILED=$((TESTS_FAILED + 1))
            return 1
        fi
    else
        if echo "$output" | grep -q "$expected_error"; then
            echo "  ✗ FAILED: Should NOT contain '$expected_error' but it did"
            TESTS_FAILED=$((TESTS_FAILED + 1))
            return 1
        else
            echo "  ✓ PASSED: Did not contain unwanted error"
            TESTS_PASSED=$((TESTS_PASSED + 1))
            return 0
        fi
    fi
}

echo "Running validation tests..."
echo ""

# Test 1: Empty string
run_test "Empty string message" \
    "/Users/aideveloper/AINative-Code/ainative-code-test chat ''" \
    "Error: message cannot be empty" \
    "yes"

# Test 2: Single space
run_test "Single space message" \
    "/Users/aideveloper/AINative-Code/ainative-code-test chat ' '" \
    "Error: message cannot be empty" \
    "yes"

# Test 3: Multiple spaces
run_test "Multiple spaces message" \
    "/Users/aideveloper/AINative-Code/ainative-code-test chat '   '" \
    "Error: message cannot be empty" \
    "yes"

# Test 4: Tab characters
run_test "Tab characters message" \
    "/Users/aideveloper/AINative-Code/ainative-code-test chat \$'\t\t'" \
    "Error: message cannot be empty" \
    "yes"

# Test 5: Newline characters
run_test "Newline characters message" \
    "/Users/aideveloper/AINative-Code/ainative-code-test chat \$'\\n'" \
    "Error: message cannot be empty" \
    "yes"

# Test 6: Mixed whitespace
run_test "Mixed whitespace message" \
    "/Users/aideveloper/AINative-Code/ainative-code-test chat \$' \\t \\n '" \
    "Error: message cannot be empty" \
    "yes"

echo ""
echo "Testing that validation happens BEFORE API calls..."
echo ""

# Test 7: Verify no provider error for empty message
# This test ensures validation happens before provider check
run_test "Empty message caught before provider check" \
    "/Users/aideveloper/AINative-Code/ainative-code-test chat ''" \
    "AI provider not configured" \
    "no"

# Test 8: Valid message should pass validation
# This test ensures valid messages get past the empty check
# (they'll fail later at API key check, which is expected)
echo "Test $((TESTS_RUN + 1)): Valid message passes empty check"
TESTS_RUN=$((TESTS_RUN + 1))
set +e
output=$(/Users/aideveloper/AINative-Code/ainative-code-test chat "hello world" --provider openai 2>&1)
exit_code=$?
set -e

if [ $exit_code -eq 0 ]; then
    echo "  ✗ FAILED: Command succeeded when API key should be missing"
    TESTS_FAILED=$((TESTS_FAILED + 1))
elif echo "$output" | grep -q "Error: message cannot be empty"; then
    echo "  ✗ FAILED: Valid message incorrectly rejected as empty"
    TESTS_FAILED=$((TESTS_FAILED + 1))
elif echo "$output" | grep -q "no API key found"; then
    echo "  ✓ PASSED: Valid message passed empty check, failed at API key check (expected)"
    TESTS_PASSED=$((TESTS_PASSED + 1))
else
    echo "  ✓ PASSED: Valid message passed empty check"
    TESTS_PASSED=$((TESTS_PASSED + 1))
fi

echo ""
echo "Testing edge cases..."
echo ""

# Test 9: Message with leading/trailing spaces
echo "Test $((TESTS_RUN + 1)): Message with leading/trailing spaces"
TESTS_RUN=$((TESTS_RUN + 1))
set +e
output=$(/Users/aideveloper/AINative-Code/ainative-code-test chat "  hello  " --provider openai 2>&1)
exit_code=$?
set -e

if echo "$output" | grep -q "Error: message cannot be empty"; then
    echo "  ✗ FAILED: Valid message with spaces incorrectly rejected"
    TESTS_FAILED=$((TESTS_FAILED + 1))
else
    echo "  ✓ PASSED: Message with spaces accepted"
    TESTS_PASSED=$((TESTS_PASSED + 1))
fi

# Test 10: Very short valid message
echo "Test $((TESTS_RUN + 1)): Single character message"
TESTS_RUN=$((TESTS_RUN + 1))
set +e
output=$(/Users/aideveloper/AINative-Code/ainative-code-test chat "a" --provider openai 2>&1)
exit_code=$?
set -e

if echo "$output" | grep -q "Error: message cannot be empty"; then
    echo "  ✗ FAILED: Single character message incorrectly rejected"
    TESTS_FAILED=$((TESTS_FAILED + 1))
else
    echo "  ✓ PASSED: Single character message accepted"
    TESTS_PASSED=$((TESTS_PASSED + 1))
fi

# Cleanup
rm -f ainative-code-test

echo ""
echo "========================================"
echo "Test Results"
echo "========================================"
echo "Tests Run:    $TESTS_RUN"
echo "Tests Passed: $TESTS_PASSED"
echo "Tests Failed: $TESTS_FAILED"
echo ""

if [ $TESTS_FAILED -eq 0 ]; then
    echo "✓ All tests passed!"
    echo ""
    echo "Summary:"
    echo "- Empty messages are rejected with friendly error"
    echo "- Whitespace-only messages are rejected"
    echo "- Validation happens BEFORE provider/API initialization"
    echo "- Valid messages pass the empty check correctly"
    echo "- No API calls are made for invalid messages"
    exit 0
else
    echo "✗ Some tests failed"
    exit 1
fi
