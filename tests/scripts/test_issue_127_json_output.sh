#!/bin/bash
# Test script for Issue #127: Clean JSON output for --json flags
# This script verifies that all commands with --json flags produce clean JSON
# without log lines interfering with the output.

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Build directory
BUILD_DIR="./build"
BINARY="$BUILD_DIR/ainative-code"

echo -e "${BLUE}=========================================${NC}"
echo -e "${BLUE}Issue #127: JSON Output Clean Test${NC}"
echo -e "${BLUE}=========================================${NC}"
echo ""

# Check if binary exists
if [ ! -f "$BINARY" ]; then
    echo -e "${YELLOW}Building binary...${NC}"
    go build -tags sqlite_fts5 -o "$BINARY" cmd/ainative-code/main.go
    echo -e "${GREEN}✓ Binary built successfully${NC}"
    echo ""
fi

# Check if jq is available
if ! command -v jq &> /dev/null; then
    echo -e "${YELLOW}Warning: jq is not installed. Some tests will be skipped.${NC}"
    echo "Install jq with: brew install jq"
    echo ""
    JQ_AVAILABLE=false
else
    JQ_AVAILABLE=true
    echo -e "${GREEN}✓ jq is available${NC}"
    echo ""
fi

# Test counter
PASSED=0
FAILED=0
SKIPPED=0

# Function to run a test
run_test() {
    local test_name="$1"
    local command="$2"
    local description="$3"

    echo -e "${BLUE}Test: $test_name${NC}"
    echo "  Command: $command"
    echo "  Expected: $description"

    # Run command and capture output
    OUTPUT=$(eval "$command" 2>&1 || true)

    # Check if output is empty
    if [ -z "$OUTPUT" ]; then
        echo -e "  Result: ${YELLOW}SKIPPED${NC} (no data available)"
        SKIPPED=$((SKIPPED + 1))
        echo ""
        return
    fi

    # Check for log patterns that should NOT be in JSON output
    if echo "$OUTPUT" | grep -q -i "INF\|DBG\|Searching\|Listing"; then
        echo -e "  Result: ${RED}FAILED${NC}"
        echo "  Error: Log lines found in JSON output"
        echo "  Output: $OUTPUT"
        FAILED=$((FAILED + 1))
        echo ""
        return
    fi

    # Check if first character is JSON
    FIRST_CHAR=$(echo "$OUTPUT" | head -c 1)
    if [ "$FIRST_CHAR" != "{" ] && [ "$FIRST_CHAR" != "[" ]; then
        echo -e "  Result: ${RED}FAILED${NC}"
        echo "  Error: Output does not start with { or ["
        echo "  First char: $FIRST_CHAR"
        echo "  Output: $OUTPUT"
        FAILED=$((FAILED + 1))
        echo ""
        return
    fi

    echo -e "  Result: ${GREEN}PASSED${NC}"
    PASSED=$((PASSED + 1))
    echo ""
}

# Function to run jq test
run_jq_test() {
    local test_name="$1"
    local command="$2"
    local jq_filter="$3"
    local description="$4"

    if [ "$JQ_AVAILABLE" = false ]; then
        echo -e "${BLUE}Test: $test_name${NC}"
        echo "  Command: $command | jq '$jq_filter'"
        echo -e "  Result: ${YELLOW}SKIPPED${NC} (jq not available)"
        SKIPPED=$((SKIPPED + 1))
        echo ""
        return
    fi

    echo -e "${BLUE}Test: $test_name${NC}"
    echo "  Command: $command | jq '$jq_filter'"
    echo "  Expected: $description"

    # Run command and pipe to jq
    OUTPUT=$(eval "$command" 2>&1 | jq "$jq_filter" 2>&1 || true)

    # Check if jq succeeded
    if echo "$OUTPUT" | grep -q "parse error"; then
        echo -e "  Result: ${RED}FAILED${NC}"
        echo "  Error: jq failed to parse JSON"
        echo "  Output: $OUTPUT"
        FAILED=$((FAILED + 1))
        echo ""
        return
    fi

    # Check if output is empty
    if [ -z "$OUTPUT" ] || [ "$OUTPUT" = "null" ]; then
        echo -e "  Result: ${YELLOW}SKIPPED${NC} (no data available)"
        SKIPPED=$((SKIPPED + 1))
        echo ""
        return
    fi

    echo -e "  Result: ${GREEN}PASSED${NC}"
    echo "  Output: $OUTPUT"
    PASSED=$((PASSED + 1))
    echo ""
}

echo -e "${YELLOW}=== Basic JSON Output Tests ===${NC}"
echo ""

run_test \
    "session list --json" \
    "$BINARY session list --limit 1 --json" \
    "Clean JSON array without log lines"

run_test \
    "session search --json" \
    "$BINARY session search 'test' --limit 5 --json" \
    "Clean JSON object without log lines"

echo -e "${YELLOW}=== jq Pipeline Tests ===${NC}"
echo ""

run_jq_test \
    "session list | jq identity" \
    "$BINARY session list --limit 1 --json" \
    "." \
    "Valid JSON that jq can parse"

run_jq_test \
    "session list | jq extract id" \
    "$BINARY session list --limit 1 --json" \
    ".[0].id" \
    "Extract session ID using jq"

run_jq_test \
    "session search | jq extract query" \
    "$BINARY session search 'test' --limit 1 --json" \
    ".query" \
    "Extract search query using jq"

echo -e "${YELLOW}=== Regression Test (Issue #127) ===${NC}"
echo ""

echo -e "${BLUE}Test: Issue #127 - session search --json${NC}"
echo "  Testing the exact scenario from the issue"

# Capture both stdout and stderr separately
STDOUT=$($BINARY session search "test" --json 2>/dev/null || true)
STDERR=$($BINARY session search "test" --json 2>&1 >/dev/null || true)

if [ -z "$STDOUT" ]; then
    echo -e "  Result: ${YELLOW}SKIPPED${NC} (no data available)"
    SKIPPED=$((SKIPPED + 1))
else
    # Check for the specific log line from the issue
    if echo "$STDOUT" | grep -q "Searching sessions"; then
        echo -e "  Result: ${RED}FAILED${NC}"
        echo "  Error: Found 'Searching sessions' log line in stdout"
        echo "  This is the exact bug from issue #127"
        FAILED=$((FAILED + 1))
    else
        echo -e "  Result: ${GREEN}PASSED${NC}"
        echo "  ✓ No 'Searching sessions' log line in stdout"
        PASSED=$((PASSED + 1))
    fi
fi
echo ""

echo -e "${BLUE}=========================================${NC}"
echo -e "${BLUE}Test Summary${NC}"
echo -e "${BLUE}=========================================${NC}"
echo -e "  ${GREEN}Passed:${NC}  $PASSED"
echo -e "  ${RED}Failed:${NC}  $FAILED"
echo -e "  ${YELLOW}Skipped:${NC} $SKIPPED"
echo ""

if [ $FAILED -eq 0 ]; then
    echo -e "${GREEN}All tests passed! ✓${NC}"
    echo -e "${GREEN}Issue #127 is FIXED${NC}"
    exit 0
else
    echo -e "${RED}Some tests failed ✗${NC}"
    echo -e "${RED}Issue #127 needs more work${NC}"
    exit 1
fi
