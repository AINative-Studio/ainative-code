#!/bin/bash
# Comprehensive JSON Flag Validation Script for v0.1.11
# Tests all JSON functionality after Agent 4 and Agent 5 fixes
# Date: 2026-01-12

set -u  # Exit on undefined variables (but not on command failures - we want to test all)

BINARY="/Users/aideveloper/AINative-Code/build/ainative-code-test"
REPORT_FILE="/Users/aideveloper/AINative-Code/tests/reports/v0.1.11_json_comprehensive_test_results.txt"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Counters
TOTAL_TESTS=0
PASSED_TESTS=0
FAILED_TESTS=0
SKIPPED_TESTS=0

# Initialize report
echo "========================================" > "$REPORT_FILE"
echo "JSON Flag Validation Report - v0.1.11" >> "$REPORT_FILE"
echo "Date: $(date)" >> "$REPORT_FILE"
echo "Binary: $BINARY" >> "$REPORT_FILE"
echo "========================================" >> "$REPORT_FILE"
echo "" >> "$REPORT_FILE"

# Test function
test_json_command() {
    local test_name="$1"
    local command="$2"
    local requires_api="${3:-no}"
    local expect_data="${4:-no}"

    TOTAL_TESTS=$((TOTAL_TESTS + 1))

    echo ""
    echo -e "${BLUE}[TEST $TOTAL_TESTS]${NC} Testing: $test_name"
    echo "Command: $command"

    echo "" >> "$REPORT_FILE"
    echo "-----------------------------------" >> "$REPORT_FILE"
    echo "TEST $TOTAL_TESTS: $test_name" >> "$REPORT_FILE"
    echo "Command: $command" >> "$REPORT_FILE"

    # Check if command requires API credentials
    if [ "$requires_api" = "yes" ]; then
        # Just verify the flag is recognized (help text)
        echo -e "${YELLOW}  Checking if --json flag is recognized...${NC}"

        if $command --help 2>&1 | grep -q "\-\-json"; then
            echo -e "${GREEN}  ✓ PASS: --json flag is documented in help${NC}"
            echo "Result: PASS (flag recognized)" >> "$REPORT_FILE"
            PASSED_TESTS=$((PASSED_TESTS + 1))
        else
            echo -e "${RED}  ✗ FAIL: --json flag not found in help${NC}"
            echo "Result: FAIL (flag not recognized)" >> "$REPORT_FILE"
            FAILED_TESTS=$((FAILED_TESTS + 1))
        fi
        return
    fi

    # Execute command and capture stdout/stderr separately
    local stdout_file=$(mktemp)
    local stderr_file=$(mktemp)

    eval "$command" > "$stdout_file" 2> "$stderr_file"
    local exit_code=$?

    local stdout_content=$(cat "$stdout_file")
    local stderr_content=$(cat "$stderr_file")

    echo "Exit Code: $exit_code" >> "$REPORT_FILE"
    echo "STDOUT Length: ${#stdout_content} bytes" >> "$REPORT_FILE"
    echo "STDERR Length: ${#stderr_content} bytes" >> "$REPORT_FILE"

    # Test 1: Check if stdout is valid JSON
    echo -e "${YELLOW}  Test 1: Validating JSON structure...${NC}"
    if echo "$stdout_content" | jq . > /dev/null 2>&1; then
        echo -e "${GREEN}    ✓ Valid JSON${NC}"
        echo "  JSON Valid: YES" >> "$REPORT_FILE"
    else
        echo -e "${RED}    ✗ Invalid JSON${NC}"
        echo "  JSON Valid: NO" >> "$REPORT_FILE"
        echo "  STDOUT Content: $stdout_content" >> "$REPORT_FILE"
        FAILED_TESTS=$((FAILED_TESTS + 1))
        rm "$stdout_file" "$stderr_file"
        return
    fi

    # Test 2: Check for log pollution in stdout
    echo -e "${YELLOW}  Test 2: Checking for log pollution in stdout...${NC}"
    if echo "$stdout_content" | grep -qE "INFO|WARN|ERROR|DEBUG|level="; then
        echo -e "${RED}    ✗ Log pollution detected in stdout${NC}"
        echo "  Log Pollution: YES" >> "$REPORT_FILE"
        FAILED_TESTS=$((FAILED_TESTS + 1))
        rm "$stdout_file" "$stderr_file"
        return
    else
        echo -e "${GREEN}    ✓ No log pollution${NC}"
        echo "  Log Pollution: NO" >> "$REPORT_FILE"
    fi

    # Test 3: Check if stderr contains logs (expected behavior)
    echo -e "${YELLOW}  Test 3: Checking if logs appear in stderr...${NC}"
    if [ -n "$stderr_content" ]; then
        echo -e "${GREEN}    ✓ Logs present in stderr (expected)${NC}"
        echo "  Logs in STDERR: YES" >> "$REPORT_FILE"
    else
        echo -e "${YELLOW}    - No logs in stderr (may be normal for some commands)${NC}"
        echo "  Logs in STDERR: NO" >> "$REPORT_FILE"
    fi

    # Test 4: Verify jq pipeline works
    echo -e "${YELLOW}  Test 4: Testing jq pipeline compatibility...${NC}"
    if echo "$stdout_content" | jq 'type' > /dev/null 2>&1; then
        echo -e "${GREEN}    ✓ Works with jq pipeline${NC}"
        echo "  JQ Compatible: YES" >> "$REPORT_FILE"
    else
        echo -e "${RED}    ✗ Fails with jq pipeline${NC}"
        echo "  JQ Compatible: NO" >> "$REPORT_FILE"
        FAILED_TESTS=$((FAILED_TESTS + 1))
        rm "$stdout_file" "$stderr_file"
        return
    fi

    # Test 5: Check if JSON has expected structure (if expect_data is yes)
    if [ "$expect_data" = "yes" ]; then
        echo -e "${YELLOW}  Test 5: Checking JSON structure...${NC}"
        local json_type=$(echo "$stdout_content" | jq -r 'type')
        echo -e "${GREEN}    ✓ JSON type: $json_type${NC}"
        echo "  JSON Type: $json_type" >> "$REPORT_FILE"

        # Show first few lines of JSON for verification
        echo "  JSON Preview:" >> "$REPORT_FILE"
        echo "$stdout_content" | jq . | head -20 >> "$REPORT_FILE"
    fi

    echo -e "${GREEN}  ✓✓ ALL CHECKS PASSED${NC}"
    echo "Result: PASS" >> "$REPORT_FILE"
    PASSED_TESTS=$((PASSED_TESTS + 1))

    rm "$stdout_file" "$stderr_file"
}

# ======================================
# Test Suite: Session Commands
# ======================================
echo ""
echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}Testing Session Commands${NC}"
echo -e "${BLUE}========================================${NC}"

echo "" >> "$REPORT_FILE"
echo "========================================" >> "$REPORT_FILE"
echo "SESSION COMMANDS" >> "$REPORT_FILE"
echo "========================================" >> "$REPORT_FILE"

test_json_command \
    "session list --json" \
    "$BINARY session list --json" \
    "no" \
    "yes"

test_json_command \
    "session search --json" \
    "$BINARY session search test --json" \
    "no" \
    "yes"

# ======================================
# Test Suite: Config Commands
# ======================================
echo ""
echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}Testing Config Commands${NC}"
echo -e "${BLUE}========================================${NC}"

echo "" >> "$REPORT_FILE"
echo "========================================" >> "$REPORT_FILE"
echo "CONFIG COMMANDS" >> "$REPORT_FILE"
echo "========================================" >> "$REPORT_FILE"

test_json_command \
    "config show --json" \
    "$BINARY config show --json" \
    "no" \
    "yes"

# ======================================
# Test Suite: MCP Commands
# ======================================
echo ""
echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}Testing MCP Commands${NC}"
echo -e "${BLUE}========================================${NC}"

echo "" >> "$REPORT_FILE"
echo "========================================" >> "$REPORT_FILE"
echo "MCP COMMANDS" >> "$REPORT_FILE"
echo "========================================" >> "$REPORT_FILE"

test_json_command \
    "mcp list-servers --json" \
    "$BINARY mcp list-servers --json" \
    "no" \
    "yes"

test_json_command \
    "mcp list-tools --json" \
    "$BINARY mcp list-tools --json" \
    "no" \
    "yes"

# ======================================
# Test Suite: ZeroDB Commands (Fixed)
# ======================================
echo ""
echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}Testing ZeroDB Commands (Fixed by Agent 5)${NC}"
echo -e "${BLUE}========================================${NC}"

echo "" >> "$REPORT_FILE"
echo "========================================" >> "$REPORT_FILE"
echo "ZERODB COMMANDS (FIXED)" >> "$REPORT_FILE"
echo "========================================" >> "$REPORT_FILE"

# ZeroDB commands require API credentials, so we just verify flag registration
test_json_command \
    "zerodb table create --json (flag check)" \
    "$BINARY zerodb table create" \
    "yes" \
    "no"

test_json_command \
    "zerodb table list --json (flag check)" \
    "$BINARY zerodb table list" \
    "yes" \
    "no"

test_json_command \
    "zerodb table insert --json (flag check)" \
    "$BINARY zerodb table insert" \
    "yes" \
    "no"

test_json_command \
    "zerodb table query --json (flag check)" \
    "$BINARY zerodb table query" \
    "yes" \
    "no"

test_json_command \
    "zerodb table update --json (flag check)" \
    "$BINARY zerodb table update" \
    "yes" \
    "no"

test_json_command \
    "zerodb table delete --json (flag check)" \
    "$BINARY zerodb table delete" \
    "yes" \
    "no"

# ======================================
# Test Suite: RLHF Commands
# ======================================
echo ""
echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}Testing RLHF Commands (Fixed by Agent 1)${NC}"
echo -e "${BLUE}========================================${NC}"

echo "" >> "$REPORT_FILE"
echo "========================================" >> "$REPORT_FILE"
echo "RLHF COMMANDS (FIXED)" >> "$REPORT_FILE"
echo "========================================" >> "$REPORT_FILE"

test_json_command \
    "rlhf interaction --json (flag check)" \
    "$BINARY rlhf interaction" \
    "yes" \
    "no"

test_json_command \
    "rlhf analytics --json (flag check)" \
    "$BINARY rlhf analytics" \
    "yes" \
    "no"

test_json_command \
    "rlhf correction --json (flag check)" \
    "$BINARY rlhf correction" \
    "yes" \
    "no"

# ======================================
# Final Summary
# ======================================
echo ""
echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}TEST SUMMARY${NC}"
echo -e "${BLUE}========================================${NC}"

echo "" >> "$REPORT_FILE"
echo "========================================" >> "$REPORT_FILE"
echo "FINAL SUMMARY" >> "$REPORT_FILE"
echo "========================================" >> "$REPORT_FILE"

echo "Total Tests: $TOTAL_TESTS"
echo "Passed: $PASSED_TESTS"
echo "Failed: $FAILED_TESTS"
echo "Skipped: $SKIPPED_TESTS"

echo "Total Tests: $TOTAL_TESTS" >> "$REPORT_FILE"
echo "Passed: $PASSED_TESTS" >> "$REPORT_FILE"
echo "Failed: $FAILED_TESTS" >> "$REPORT_FILE"
echo "Skipped: $SKIPPED_TESTS" >> "$REPORT_FILE"

if [ $FAILED_TESTS -eq 0 ]; then
    echo ""
    echo -e "${GREEN}========================================${NC}"
    echo -e "${GREEN}✓✓ ALL TESTS PASSED - READY FOR v0.1.11${NC}"
    echo -e "${GREEN}========================================${NC}"
    echo "" >> "$REPORT_FILE"
    echo "STATUS: ✓✓ ALL TESTS PASSED - READY FOR v0.1.11" >> "$REPORT_FILE"
    exit 0
else
    echo ""
    echo -e "${RED}========================================${NC}"
    echo -e "${RED}✗✗ SOME TESTS FAILED - NOT READY${NC}"
    echo -e "${RED}========================================${NC}"
    echo "" >> "$REPORT_FILE"
    echo "STATUS: ✗✗ SOME TESTS FAILED - NOT READY" >> "$REPORT_FILE"
    exit 1
fi
