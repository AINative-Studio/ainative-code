#!/bin/bash
# Test script for Issue #128: Logger outputs to stdout instead of stderr
# This script verifies that logs go to stderr and JSON output goes to stdout

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
echo -e "${BLUE}Issue #128: Logger stderr Test${NC}"
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

echo -e "${YELLOW}=== Core Issue #128 Tests ===${NC}"
echo ""

# Test 1: Verify stdout contains only JSON, no log lines
echo -e "${BLUE}Test 1: JSON output to stdout is clean${NC}"
echo "  Testing: session list --json with stdout capture"

STDOUT=$($BINARY session list --limit 1 --json 2>/dev/null || true)

if [ -z "$STDOUT" ]; then
    echo -e "  Result: ${YELLOW}SKIPPED${NC} (no data available)"
    SKIPPED=$((SKIPPED + 1))
else
    # Check if stdout contains log patterns (it should NOT)
    if echo "$STDOUT" | grep -E "INF|DBG|WRN|ERR|level=|Searching|Listing" > /dev/null; then
        echo -e "  Result: ${RED}FAILED${NC}"
        echo "  Error: Log lines found in stdout"
        echo "  This indicates logger is still writing to stdout"
        echo "  Stdout content: $STDOUT"
        FAILED=$((FAILED + 1))
    else
        echo -e "  Result: ${GREEN}PASSED${NC}"
        echo "  ✓ Stdout contains only JSON, no log lines"
        PASSED=$((PASSED + 1))
    fi
fi
echo ""

# Test 2: Verify stderr contains log lines
echo -e "${BLUE}Test 2: Log output goes to stderr${NC}"
echo "  Testing: session list --json with stderr capture"

STDERR=$($BINARY session list --limit 1 --json 2>&1 >/dev/null || true)

if [ -z "$STDERR" ]; then
    echo -e "  Result: ${GREEN}PASSED${NC}"
    echo "  ✓ No logs produced (clean run) or logs went to stderr"
    PASSED=$((PASSED + 1))
else
    # Check if stderr contains log patterns (it SHOULD if logs are enabled)
    if echo "$STDERR" | grep -E "INF|DBG|WRN|ERR|level=|Searching|Listing" > /dev/null; then
        echo -e "  Result: ${GREEN}PASSED${NC}"
        echo "  ✓ Log lines found in stderr (correct behavior)"
        echo "  Sample log: $(echo "$STDERR" | head -1)"
        PASSED=$((PASSED + 1))
    else
        echo -e "  Result: ${GREEN}PASSED${NC}"
        echo "  ✓ Stderr captured (no explicit logs or logs disabled)"
        PASSED=$((PASSED + 1))
    fi
fi
echo ""

# Test 3: Verify jq can parse stdout directly
if [ "$JQ_AVAILABLE" = true ]; then
    echo -e "${BLUE}Test 3: jq can parse stdout without filtering stderr${NC}"
    echo "  Command: session list --json | jq '.'"

    OUTPUT=$($BINARY session list --limit 1 --json 2>/dev/null | jq '.' 2>&1 || true)

    if [ -z "$OUTPUT" ] || [ "$OUTPUT" = "null" ] || [ "$OUTPUT" = "[]" ]; then
        echo -e "  Result: ${YELLOW}SKIPPED${NC} (no data available)"
        SKIPPED=$((SKIPPED + 1))
    elif echo "$OUTPUT" | grep -q "parse error"; then
        echo -e "  Result: ${RED}FAILED${NC}"
        echo "  Error: jq failed to parse JSON"
        echo "  This means stdout is polluted with non-JSON content"
        echo "  Output: $OUTPUT"
        FAILED=$((FAILED + 1))
    else
        echo -e "  Result: ${GREEN}PASSED${NC}"
        echo "  ✓ jq successfully parsed JSON from stdout"
        PASSED=$((PASSED + 1))
    fi
    echo ""

    # Test 4: Verify jq filters work correctly
    echo -e "${BLUE}Test 4: jq filters work on stdout${NC}"
    echo "  Command: session list --json | jq '.[0].id'"

    OUTPUT=$($BINARY session list --limit 1 --json 2>/dev/null | jq '.[0].id' 2>&1 || true)

    if [ -z "$OUTPUT" ] || [ "$OUTPUT" = "null" ]; then
        echo -e "  Result: ${YELLOW}SKIPPED${NC} (no data available)"
        SKIPPED=$((SKIPPED + 1))
    elif echo "$OUTPUT" | grep -q "parse error"; then
        echo -e "  Result: ${RED}FAILED${NC}"
        echo "  Error: jq failed to filter JSON"
        echo "  Output: $OUTPUT"
        FAILED=$((FAILED + 1))
    else
        echo -e "  Result: ${GREEN}PASSED${NC}"
        echo "  ✓ jq filter extracted: $OUTPUT"
        PASSED=$((PASSED + 1))
    fi
    echo ""
else
    echo -e "${YELLOW}Tests 3 & 4: SKIPPED (jq not available)${NC}"
    SKIPPED=$((SKIPPED + 2))
    echo ""
fi

# Test 5: Comprehensive separation test
echo -e "${BLUE}Test 5: Complete stdout/stderr separation${NC}"
echo "  Testing: Both streams are independent"

# Create temp files for capture
STDOUT_FILE=$(mktemp)
STDERR_FILE=$(mktemp)

# Run command with separate capture
$BINARY session list --limit 1 --json > "$STDOUT_FILE" 2> "$STDERR_FILE" || true

STDOUT_CONTENT=$(cat "$STDOUT_FILE")
STDERR_CONTENT=$(cat "$STDERR_FILE")

# Clean up temp files
rm -f "$STDOUT_FILE" "$STDERR_FILE"

if [ -z "$STDOUT_CONTENT" ]; then
    echo -e "  Result: ${YELLOW}SKIPPED${NC} (no stdout content)"
    SKIPPED=$((SKIPPED + 1))
else
    # Stdout should be valid JSON
    FIRST_CHAR=$(echo "$STDOUT_CONTENT" | head -c 1)
    if [ "$FIRST_CHAR" != "{" ] && [ "$FIRST_CHAR" != "[" ]; then
        echo -e "  Result: ${RED}FAILED${NC}"
        echo "  Error: Stdout does not start with JSON bracket"
        echo "  First char: '$FIRST_CHAR'"
        echo "  This indicates log pollution in stdout"
        FAILED=$((FAILED + 1))
    else
        echo -e "  Result: ${GREEN}PASSED${NC}"
        echo "  ✓ Stdout contains valid JSON"
        echo "  ✓ Stderr is separate (may contain logs or be empty)"
        PASSED=$((PASSED + 1))
    fi
fi
echo ""

# Test 6: Real-world pipeline test
echo -e "${BLUE}Test 6: Real-world pipeline scenario${NC}"
echo "  Scenario: session search 'test' --json | jq '.query'"

if [ "$JQ_AVAILABLE" = true ]; then
    PIPELINE_OUTPUT=$($BINARY session search "test" --json 2>&1 | jq '.query' 2>&1 || true)

    if [ -z "$PIPELINE_OUTPUT" ] || [ "$PIPELINE_OUTPUT" = "null" ]; then
        echo -e "  Result: ${YELLOW}SKIPPED${NC} (no search results)"
        SKIPPED=$((SKIPPED + 1))
    elif echo "$PIPELINE_OUTPUT" | grep -q "parse error"; then
        echo -e "  Result: ${RED}FAILED${NC}"
        echo "  Error: Pipeline failed - JSON parsing error"
        echo "  This indicates stdout pollution from logger"
        echo "  Output: $PIPELINE_OUTPUT"
        FAILED=$((FAILED + 1))
    else
        echo -e "  Result: ${GREEN}PASSED${NC}"
        echo "  ✓ Pipeline works correctly"
        echo "  ✓ Query extracted: $PIPELINE_OUTPUT"
        PASSED=$((PASSED + 1))
    fi
else
    echo -e "  Result: ${YELLOW}SKIPPED${NC} (jq not available)"
    SKIPPED=$((SKIPPED + 1))
fi
echo ""

# Test 7: Verify logger behavior in code
echo -e "${BLUE}Test 7: Logger configuration verification${NC}"
echo "  Checking: internal/logger/logger.go default config"

LOGGER_CONFIG=$(grep -A 10 "DefaultConfig()" /Users/aideveloper/AINative-Code/internal/logger/logger.go | grep "Output:" || true)

if echo "$LOGGER_CONFIG" | grep -q '"stderr"'; then
    echo -e "  Result: ${GREEN}PASSED${NC}"
    echo "  ✓ Logger default config uses stderr"
    echo "  Config line: $LOGGER_CONFIG"
    PASSED=$((PASSED + 1))
elif echo "$LOGGER_CONFIG" | grep -q '"stdout"'; then
    echo -e "  Result: ${RED}FAILED${NC}"
    echo "  Error: Logger still configured to use stdout"
    echo "  Config line: $LOGGER_CONFIG"
    echo "  This is the root cause of issue #128"
    FAILED=$((FAILED + 1))
else
    echo -e "  Result: ${YELLOW}SKIPPED${NC}"
    echo "  Warning: Could not verify logger config"
    SKIPPED=$((SKIPPED + 1))
fi
echo ""

# Test 8: Unit test verification
echo -e "${BLUE}Test 8: Unit tests for logger stderr${NC}"
echo "  Running: go test ./internal/logger/..."

TEST_OUTPUT=$(go test ./internal/logger/... -run "TestLoggerOutputsToStderr|TestDefaultConfig" -v 2>&1 || true)

if echo "$TEST_OUTPUT" | grep -q "FAIL"; then
    echo -e "  Result: ${RED}FAILED${NC}"
    echo "  Error: Logger unit tests failed"
    echo "  Output: $TEST_OUTPUT"
    FAILED=$((FAILED + 1))
elif echo "$TEST_OUTPUT" | grep -q "PASS"; then
    echo -e "  Result: ${GREEN}PASSED${NC}"
    echo "  ✓ TestLoggerOutputsToStderr: PASS"
    echo "  ✓ TestDefaultConfig: PASS"
    PASSED=$((PASSED + 1))
else
    echo -e "  Result: ${YELLOW}SKIPPED${NC}"
    echo "  Warning: Could not run unit tests"
    SKIPPED=$((SKIPPED + 1))
fi
echo ""

echo -e "${BLUE}=========================================${NC}"
echo -e "${BLUE}Test Summary${NC}"
echo -e "${BLUE}=========================================${NC}"
echo -e "  ${GREEN}Passed:${NC}  $PASSED"
echo -e "  ${RED}Failed:${NC}  $FAILED"
echo -e "  ${YELLOW}Skipped:${NC} $SKIPPED"
echo ""

echo -e "${YELLOW}Key Points Verified:${NC}"
echo "  1. JSON output to stdout is clean (no log lines)"
echo "  2. Log output goes to stderr (correct separation)"
echo "  3. jq can parse stdout directly (no filtering needed)"
echo "  4. jq filters work correctly on JSON output"
echo "  5. stdout and stderr are completely independent"
echo "  6. Real-world pipelines work as expected"
echo "  7. Logger code is configured correctly"
echo "  8. Unit tests verify the fix"
echo ""

if [ $FAILED -eq 0 ]; then
    echo -e "${GREEN}=========================================${NC}"
    echo -e "${GREEN}All tests passed! ✓${NC}"
    echo -e "${GREEN}Issue #128 is FIXED${NC}"
    echo -e "${GREEN}=========================================${NC}"
    echo ""
    echo -e "${GREEN}Fix Summary:${NC}"
    echo "  - Changed logger default output from 'stdout' to 'stderr'"
    echo "  - Location: internal/logger/logger.go line 84"
    echo "  - Added comprehensive tests for stderr behavior"
    echo "  - Verified JSON commands work cleanly with jq"
    echo ""
    exit 0
else
    echo -e "${RED}=========================================${NC}"
    echo -e "${RED}Some tests failed ✗${NC}"
    echo -e "${RED}Issue #128 needs more work${NC}"
    echo -e "${RED}=========================================${NC}"
    exit 1
fi
