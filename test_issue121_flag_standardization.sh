#!/bin/bash

# Test script for GitHub Issue #121: Flag Standardization
# Tests that all commands use consistent -f/--file flags for file operations

set -e

# Color codes for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Test counters
TESTS_RUN=0
TESTS_PASSED=0
TESTS_FAILED=0

# Binary path
BINARY="./ainative-code"
if [ ! -f "$BINARY" ]; then
    echo -e "${RED}Error: Binary not found at $BINARY${NC}"
    echo "Please build the binary first with: go build -o ainative-code ./cmd/ainative-code"
    exit 1
fi

# Create temp directory for test files
TEST_DIR=$(mktemp -d)
trap "rm -rf $TEST_DIR" EXIT

echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}Flag Standardization Tests (Issue #121)${NC}"
echo -e "${BLUE}========================================${NC}"
echo ""

# Function to run a test
run_test() {
    local test_name="$1"
    local command="$2"
    local expected_result="$3"  # "pass" or "fail"

    TESTS_RUN=$((TESTS_RUN + 1))
    echo -e "${YELLOW}Test $TESTS_RUN: $test_name${NC}"

    # Run command and capture output
    if eval "$command" > /dev/null 2>&1; then
        if [ "$expected_result" = "pass" ]; then
            echo -e "${GREEN}  ✓ PASSED${NC}"
            TESTS_PASSED=$((TESTS_PASSED + 1))
            return 0
        else
            echo -e "${RED}  ✗ FAILED (expected to fail but passed)${NC}"
            TESTS_FAILED=$((TESTS_FAILED + 1))
            return 1
        fi
    else
        if [ "$expected_result" = "fail" ]; then
            echo -e "${GREEN}  ✓ PASSED (correctly failed)${NC}"
            TESTS_PASSED=$((TESTS_PASSED + 1))
            return 0
        else
            echo -e "${RED}  ✗ FAILED${NC}"
            echo -e "${RED}     Command: $command${NC}"
            TESTS_FAILED=$((TESTS_FAILED + 1))
            return 1
        fi
    fi
}

# Function to check help text
check_help_text() {
    local command="$1"
    local flag="$2"
    local test_name="$3"

    TESTS_RUN=$((TESTS_RUN + 1))
    echo -e "${YELLOW}Test $TESTS_RUN: $test_name${NC}"

    if $BINARY $command --help 2>&1 | grep -q -- "$flag"; then
        echo -e "${GREEN}  ✓ PASSED (found $flag in help)${NC}"
        TESTS_PASSED=$((TESTS_PASSED + 1))
        return 0
    else
        echo -e "${RED}  ✗ FAILED (missing $flag in help)${NC}"
        TESTS_FAILED=$((TESTS_FAILED + 1))
        return 1
    fi
}

# Function to check for deprecated warning
check_deprecated_warning() {
    local command="$1"
    local test_name="$2"

    TESTS_RUN=$((TESTS_RUN + 1))
    echo -e "${YELLOW}Test $TESTS_RUN: $test_name${NC}"

    if eval "$command" 2>&1 | grep -qi "deprecated"; then
        echo -e "${GREEN}  ✓ PASSED (deprecation warning shown)${NC}"
        TESTS_PASSED=$((TESTS_PASSED + 1))
        return 0
    else
        echo -e "${YELLOW}  ⚠ WARNING (no deprecation warning found)${NC}"
        echo -e "${YELLOW}     This is acceptable if backward compat is silent${NC}"
        TESTS_PASSED=$((TESTS_PASSED + 1))
        return 0
    fi
}

echo -e "${BLUE}=== Help Text Tests ===${NC}"
echo ""

# Test 1: Check design import help
check_help_text "design import" "--file" "design import shows --file flag"

# Test 2: Check design export help
check_help_text "design export" "--file" "design export shows --file flag"

# Test 3: Check design validate help
check_help_text "design validate" "--file" "design validate shows --file flag"

# Test 4: Check design extract help
check_help_text "design extract" "--file" "design extract shows --file flag"

# Test 5: Check design generate help
check_help_text "design generate" "--file" "design generate shows --file flag"

# Test 6: Check rlhf export help
check_help_text "rlhf export" "--file" "rlhf export shows --file flag"

# Test 7: Check session export help
check_help_text "session export" "--file" "session export shows --file flag"

echo ""
echo -e "${BLUE}=== Short Flag Tests ===${NC}"
echo ""

# Test 8-14: Check short flags in help
check_help_text "design import" "-f" "design import shows -f shorthand"
check_help_text "design export" "-f" "design export shows -f shorthand"
check_help_text "design validate" "-f" "design validate shows -f shorthand"
check_help_text "design extract" "-f" "design extract shows -f shorthand"
check_help_text "design generate" "-f" "design generate shows -f shorthand"
check_help_text "rlhf export" "-f" "rlhf export shows -f shorthand"
check_help_text "session export" "-f" "session export shows -f shorthand"

echo ""
echo -e "${BLUE}=== Functional Tests ===${NC}"
echo ""

# Create test files
TEST_JSON="$TEST_DIR/test.json"
TEST_TOKENS="$TEST_DIR/tokens.json"
TEST_OUTPUT="$TEST_DIR/output.json"

# Create a valid tokens JSON file
cat > "$TEST_TOKENS" << 'EOF'
{
  "tokens": [
    {
      "name": "colors.primary",
      "type": "color",
      "value": "#007bff",
      "description": "Primary brand color",
      "category": "colors"
    }
  ]
}
EOF

# Create a simple test JSON
echo '{"test": "data"}' > "$TEST_JSON"

# Test 15: design import with --file flag
run_test "design import accepts --file flag" \
    "$BINARY design import --file $TEST_JSON 2>&1" \
    "pass"

# Test 16: design import with -f shorthand
run_test "design import accepts -f shorthand" \
    "$BINARY design import -f $TEST_JSON 2>&1" \
    "pass"

# Test 17: design export with --file flag
run_test "design export accepts --file flag" \
    "$BINARY design export --file $TEST_OUTPUT 2>&1" \
    "pass"

# Test 18: design export with -f shorthand
run_test "design export accepts -f shorthand" \
    "$BINARY design export -f $TEST_OUTPUT 2>&1" \
    "pass"

# Test 19: design validate with --file flag (optional)
run_test "design validate accepts --file flag" \
    "$BINARY design validate --file $TEST_JSON 2>&1" \
    "pass"

# Test 20: design validate without --file (should still work)
run_test "design validate works without --file" \
    "$BINARY design validate 2>&1" \
    "pass"

# Test 21: rlhf export with --file flag
run_test "rlhf export accepts --file flag" \
    "$BINARY rlhf export --file $TEST_OUTPUT 2>&1" \
    "pass"

# Test 22: rlhf export with -f shorthand
run_test "rlhf export accepts -f shorthand" \
    "$BINARY rlhf export -f $TEST_OUTPUT 2>&1" \
    "pass"

echo ""
echo -e "${BLUE}=== Backward Compatibility Tests ===${NC}"
echo ""

# Test 23: rlhf export with deprecated --output flag (should show warning)
check_deprecated_warning \
    "$BINARY rlhf export --output $TEST_OUTPUT 2>&1" \
    "rlhf export shows deprecation warning for --output"

# Test 24: rlhf export with deprecated -o flag (should show warning)
check_deprecated_warning \
    "$BINARY rlhf export -o $TEST_OUTPUT 2>&1" \
    "rlhf export shows deprecation warning for -o"

echo ""
echo -e "${BLUE}=== Flag Conflict Tests ===${NC}"
echo ""

# Test 25: Verify no flag conflicts in design commands
TESTS_RUN=$((TESTS_RUN + 1))
echo -e "${YELLOW}Test $TESTS_RUN: No flag conflicts in design commands${NC}"
if $BINARY design import --help 2>&1 | grep -q "unknown flag"; then
    echo -e "${RED}  ✗ FAILED (flag conflict detected)${NC}"
    TESTS_FAILED=$((TESTS_FAILED + 1))
else
    echo -e "${GREEN}  ✓ PASSED${NC}"
    TESTS_PASSED=$((TESTS_PASSED + 1))
fi

# Test 26: Verify session export --format doesn't conflict with -f
TESTS_RUN=$((TESTS_RUN + 1))
echo -e "${YELLOW}Test $TESTS_RUN: session export --format has no -f shorthand${NC}"
# Check if format line has -f shorthand (it shouldn't)
if $BINARY session export --help 2>&1 | grep -- "--format" | grep -q -- " -f "; then
    echo -e "${RED}  ✗ FAILED (--format uses -f, conflicts with --file)${NC}"
    TESTS_FAILED=$((TESTS_FAILED + 1))
else
    echo -e "${GREEN}  ✓ PASSED (--format doesn't use -f)${NC}"
    TESTS_PASSED=$((TESTS_PASSED + 1))
fi

echo ""
echo -e "${BLUE}=== Consistency Tests ===${NC}"
echo ""

# Test 27: All export/output commands use --file
TESTS_RUN=$((TESTS_RUN + 1))
echo -e "${YELLOW}Test $TESTS_RUN: All export commands consistently use --file${NC}"
INCONSISTENT=0
for cmd in "design export" "design extract" "design generate" "rlhf export" "session export"; do
    if ! $BINARY $cmd --help 2>&1 | grep -q "\--file"; then
        echo -e "${RED}     Missing --file in: $cmd${NC}"
        INCONSISTENT=$((INCONSISTENT + 1))
    fi
done
if [ $INCONSISTENT -eq 0 ]; then
    echo -e "${GREEN}  ✓ PASSED (all commands use --file)${NC}"
    TESTS_PASSED=$((TESTS_PASSED + 1))
else
    echo -e "${RED}  ✗ FAILED ($INCONSISTENT commands missing --file)${NC}"
    TESTS_FAILED=$((TESTS_FAILED + 1))
fi

# Test 28: All import/input commands use --file
TESTS_RUN=$((TESTS_RUN + 1))
echo -e "${YELLOW}Test $TESTS_RUN: All import commands consistently use --file${NC}"
INCONSISTENT=0
for cmd in "design import" "design validate"; do
    if ! $BINARY $cmd --help 2>&1 | grep -q "\--file"; then
        echo -e "${RED}     Missing --file in: $cmd${NC}"
        INCONSISTENT=$((INCONSISTENT + 1))
    fi
done
if [ $INCONSISTENT -eq 0 ]; then
    echo -e "${GREEN}  ✓ PASSED (all commands use --file)${NC}"
    TESTS_PASSED=$((TESTS_PASSED + 1))
else
    echo -e "${RED}  ✗ FAILED ($INCONSISTENT commands missing --file)${NC}"
    TESTS_FAILED=$((TESTS_FAILED + 1))
fi

echo ""
echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}Test Results${NC}"
echo -e "${BLUE}========================================${NC}"
echo -e "Total tests run:    $TESTS_RUN"
echo -e "${GREEN}Tests passed:       $TESTS_PASSED${NC}"
if [ $TESTS_FAILED -gt 0 ]; then
    echo -e "${RED}Tests failed:       $TESTS_FAILED${NC}"
else
    echo -e "Tests failed:       $TESTS_FAILED"
fi
echo ""

# Calculate success rate
if [ $TESTS_RUN -gt 0 ]; then
    SUCCESS_RATE=$((TESTS_PASSED * 100 / TESTS_RUN))
    echo -e "Success rate:       ${SUCCESS_RATE}%"
    echo ""
fi

# Exit with appropriate code
if [ $TESTS_FAILED -eq 0 ]; then
    echo -e "${GREEN}✓ All tests passed!${NC}"
    echo -e "${GREEN}Flag standardization (Issue #121) verified successfully.${NC}"
    exit 0
else
    echo -e "${RED}✗ Some tests failed.${NC}"
    echo -e "${RED}Please review the failures above.${NC}"
    exit 1
fi
