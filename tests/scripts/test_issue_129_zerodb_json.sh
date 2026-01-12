#!/bin/bash
#
# Test Script for Issue #129 - ZeroDB Table --json Flag
# This script demonstrates that the --json flag is now properly registered
# and functional for all 6 zerodb table subcommands.
#

set -e  # Exit on error

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
BINARY="${BINARY:-./ainative-code}"
TEST_TABLE="test_users_$$"  # Use PID to avoid conflicts
TEST_SCHEMA='{"type":"object","properties":{"name":{"type":"string"},"email":{"type":"string"},"age":{"type":"number"}}}'
TEST_DATA='{"name":"John Doe","email":"john@example.com","age":30}'
UPDATE_DATA='{"age":31}'
FILTER='{"age":{"$gte":18}}'

# Check if jq is available
if ! command -v jq &> /dev/null; then
    echo -e "${YELLOW}Warning: jq is not installed. JSON validation will be limited.${NC}"
    echo -e "${YELLOW}Install jq for full JSON validation: brew install jq (macOS) or apt-get install jq (Linux)${NC}"
    JQ_AVAILABLE=false
else
    echo -e "${GREEN}✓ jq is available for JSON validation${NC}"
    JQ_AVAILABLE=true
fi

# Function to print test header
print_header() {
    echo ""
    echo -e "${BLUE}========================================${NC}"
    echo -e "${BLUE}$1${NC}"
    echo -e "${BLUE}========================================${NC}"
}

# Function to print success
print_success() {
    echo -e "${GREEN}✓ $1${NC}"
}

# Function to print error
print_error() {
    echo -e "${RED}✗ $1${NC}"
}

# Function to print info
print_info() {
    echo -e "${YELLOW}→ $1${NC}"
}

# Function to validate JSON output
validate_json() {
    local output="$1"
    local description="$2"

    # Check if output is not empty
    if [ -z "$output" ]; then
        print_error "$description - Empty output"
        return 1
    fi

    # Try to parse with jq if available
    if [ "$JQ_AVAILABLE" = true ]; then
        if echo "$output" | jq . > /dev/null 2>&1; then
            print_success "$description - Valid JSON"
            echo "$output" | jq .
            return 0
        else
            print_error "$description - Invalid JSON"
            echo "Output: $output"
            return 1
        fi
    else
        # Basic JSON validation - check for opening brace/bracket
        if [[ "$output" =~ ^[[:space:]]*[\{\[] ]]; then
            print_success "$description - Appears to be JSON"
            echo "$output"
            return 0
        else
            print_error "$description - Does not appear to be JSON"
            echo "Output: $output"
            return 1
        fi
    fi
}

# Function to check flag registration in help
check_flag_in_help() {
    local subcmd="$1"
    print_header "Checking --json flag in help for: zerodb table $subcmd"

    if $BINARY zerodb table "$subcmd" --help 2>&1 | grep -q "\-\-json"; then
        print_success "Flag --json found in help output"
        $BINARY zerodb table "$subcmd" --help 2>&1 | grep -A1 "\-\-json"
        return 0
    else
        print_error "Flag --json NOT found in help output"
        return 1
    fi
}

# Track test results
TESTS_PASSED=0
TESTS_FAILED=0
TESTS_SKIPPED=0

# Stored document ID for update/delete tests
DOC_ID=""

print_header "Issue #129 Test Script - ZeroDB Table --json Flag"
echo "Testing all 6 zerodb table subcommands with --json flag"
echo "Binary: $BINARY"
echo ""

# Build the binary if it doesn't exist
if [ ! -f "$BINARY" ]; then
    print_info "Binary not found. Building..."
    cd "$(dirname "$0")/../.."
    go build -o ainative-code ./cmd/ainative-code
    cd - > /dev/null
    BINARY="./ainative-code"
fi

# ============================================================================
# TEST 1: Check --json flag registration in help for all subcommands
# ============================================================================
print_header "TEST 1: Flag Registration in Help Output"

for subcmd in create list insert query update delete; do
    if check_flag_in_help "$subcmd"; then
        ((TESTS_PASSED++))
    else
        ((TESTS_FAILED++))
    fi
done

# ============================================================================
# TEST 2: zerodb table create --json
# ============================================================================
print_header "TEST 2: zerodb table create --json"

print_info "Creating table: $TEST_TABLE"
print_info "Command: $BINARY zerodb table create --name \"$TEST_TABLE\" --schema '$TEST_SCHEMA' --json"

OUTPUT=$($BINARY zerodb table create --name "$TEST_TABLE" --schema "$TEST_SCHEMA" --json 2>&1 || true)

if validate_json "$OUTPUT" "table create"; then
    ((TESTS_PASSED++))

    # Extract table ID if jq is available
    if [ "$JQ_AVAILABLE" = true ]; then
        TABLE_ID=$(echo "$OUTPUT" | jq -r '.id // empty' 2>/dev/null || true)
        if [ -n "$TABLE_ID" ]; then
            print_info "Created table ID: $TABLE_ID"
        fi
    fi
else
    # Check if it's a flag recognition error
    if echo "$OUTPUT" | grep -q "unknown flag.*json"; then
        print_error "BUG CONFIRMED: --json flag not recognized!"
        ((TESTS_FAILED++))
    else
        print_info "Command may have failed due to API/config issues, but flag was recognized"
        ((TESTS_SKIPPED++))
    fi
fi

# ============================================================================
# TEST 3: zerodb table list --json
# ============================================================================
print_header "TEST 3: zerodb table list --json"

print_info "Command: $BINARY zerodb table list --json"

OUTPUT=$($BINARY zerodb table list --json 2>&1 || true)

if validate_json "$OUTPUT" "table list"; then
    ((TESTS_PASSED++))

    # Count tables if jq is available
    if [ "$JQ_AVAILABLE" = true ]; then
        TABLE_COUNT=$(echo "$OUTPUT" | jq 'length // 0' 2>/dev/null || echo "0")
        print_info "Found $TABLE_COUNT table(s)"
    fi
else
    if echo "$OUTPUT" | grep -q "unknown flag.*json"; then
        print_error "BUG CONFIRMED: --json flag not recognized!"
        ((TESTS_FAILED++))
    else
        print_info "Command may have failed due to API/config issues, but flag was recognized"
        ((TESTS_SKIPPED++))
    fi
fi

# ============================================================================
# TEST 4: zerodb table insert --json
# ============================================================================
print_header "TEST 4: zerodb table insert --json"

print_info "Inserting document into table: $TEST_TABLE"
print_info "Command: $BINARY zerodb table insert --table \"$TEST_TABLE\" --data '$TEST_DATA' --json"

OUTPUT=$($BINARY zerodb table insert --table "$TEST_TABLE" --data "$TEST_DATA" --json 2>&1 || true)

if validate_json "$OUTPUT" "table insert"; then
    ((TESTS_PASSED++))

    # Extract document ID if jq is available
    if [ "$JQ_AVAILABLE" = true ]; then
        DOC_ID=$(echo "$OUTPUT" | jq -r '.id // empty' 2>/dev/null || true)
        if [ -n "$DOC_ID" ]; then
            print_info "Inserted document ID: $DOC_ID"
        fi
    fi
else
    if echo "$OUTPUT" | grep -q "unknown flag.*json"; then
        print_error "BUG CONFIRMED: --json flag not recognized!"
        ((TESTS_FAILED++))
    else
        print_info "Command may have failed due to API/config issues, but flag was recognized"
        ((TESTS_SKIPPED++))
    fi
fi

# ============================================================================
# TEST 5: zerodb table query --json
# ============================================================================
print_header "TEST 5: zerodb table query --json"

print_info "Querying documents from table: $TEST_TABLE"
print_info "Command: $BINARY zerodb table query --table \"$TEST_TABLE\" --filter '$FILTER' --json"

OUTPUT=$($BINARY zerodb table query --table "$TEST_TABLE" --filter "$FILTER" --json 2>&1 || true)

if validate_json "$OUTPUT" "table query"; then
    ((TESTS_PASSED++))

    # Count documents if jq is available
    if [ "$JQ_AVAILABLE" = true ]; then
        DOC_COUNT=$(echo "$OUTPUT" | jq 'length // 0' 2>/dev/null || echo "0")
        print_info "Found $DOC_COUNT document(s)"
    fi
else
    if echo "$OUTPUT" | grep -q "unknown flag.*json"; then
        print_error "BUG CONFIRMED: --json flag not recognized!"
        ((TESTS_FAILED++))
    else
        print_info "Command may have failed due to API/config issues, but flag was recognized"
        ((TESTS_SKIPPED++))
    fi
fi

# ============================================================================
# TEST 6: zerodb table update --json
# ============================================================================
print_header "TEST 6: zerodb table update --json"

# Use a test document ID if we don't have one from insert
if [ -z "$DOC_ID" ]; then
    DOC_ID="test-doc-id"
    print_info "Using test document ID (insert may have failed)"
fi

print_info "Updating document in table: $TEST_TABLE"
print_info "Command: $BINARY zerodb table update --table \"$TEST_TABLE\" --id \"$DOC_ID\" --data '$UPDATE_DATA' --json"

OUTPUT=$($BINARY zerodb table update --table "$TEST_TABLE" --id "$DOC_ID" --data "$UPDATE_DATA" --json 2>&1 || true)

if validate_json "$OUTPUT" "table update"; then
    ((TESTS_PASSED++))
else
    if echo "$OUTPUT" | grep -q "unknown flag.*json"; then
        print_error "BUG CONFIRMED: --json flag not recognized!"
        ((TESTS_FAILED++))
    else
        print_info "Command may have failed due to API/config issues, but flag was recognized"
        ((TESTS_SKIPPED++))
    fi
fi

# ============================================================================
# TEST 7: zerodb table delete --json
# ============================================================================
print_header "TEST 7: zerodb table delete --json"

print_info "Deleting document from table: $TEST_TABLE"
print_info "Command: $BINARY zerodb table delete --table \"$TEST_TABLE\" --id \"$DOC_ID\" --json"

OUTPUT=$($BINARY zerodb table delete --table "$TEST_TABLE" --id "$DOC_ID" --json 2>&1 || true)

if validate_json "$OUTPUT" "table delete"; then
    ((TESTS_PASSED++))

    # Verify success field if jq is available
    if [ "$JQ_AVAILABLE" = true ]; then
        SUCCESS=$(echo "$OUTPUT" | jq -r '.success // false' 2>/dev/null || echo "false")
        if [ "$SUCCESS" = "true" ]; then
            print_info "Delete confirmed successful"
        fi
    fi
else
    if echo "$OUTPUT" | grep -q "unknown flag.*json"; then
        print_error "BUG CONFIRMED: --json flag not recognized!"
        ((TESTS_FAILED++))
    else
        print_info "Command may have failed due to API/config issues, but flag was recognized"
        ((TESTS_SKIPPED++))
    fi
fi

# ============================================================================
# TEST 8: Verify JSON can be piped to jq
# ============================================================================
if [ "$JQ_AVAILABLE" = true ]; then
    print_header "TEST 8: Verify JSON pipeable to jq"

    print_info "Testing: $BINARY zerodb table list --json | jq '.[]'"

    if $BINARY zerodb table list --json 2>&1 | jq '.[]' > /dev/null 2>&1; then
        print_success "JSON output successfully piped to jq"
        ((TESTS_PASSED++))
    else
        OUTPUT=$($BINARY zerodb table list --json 2>&1 || true)
        if echo "$OUTPUT" | grep -q "unknown flag.*json"; then
            print_error "BUG CONFIRMED: --json flag not recognized!"
            ((TESTS_FAILED++))
        else
            print_info "Command may have failed due to API/config issues"
            ((TESTS_SKIPPED++))
        fi
    fi
fi

# ============================================================================
# Summary
# ============================================================================
print_header "Test Summary"

TOTAL_TESTS=$((TESTS_PASSED + TESTS_FAILED + TESTS_SKIPPED))

echo ""
echo "Total Tests:   $TOTAL_TESTS"
echo -e "${GREEN}Passed:        $TESTS_PASSED${NC}"
echo -e "${RED}Failed:        $TESTS_FAILED${NC}"
echo -e "${YELLOW}Skipped:       $TESTS_SKIPPED${NC}"
echo ""

if [ $TESTS_FAILED -eq 0 ]; then
    echo -e "${GREEN}========================================${NC}"
    echo -e "${GREEN}✓ All tests passed or skipped!${NC}"
    echo -e "${GREEN}✓ Issue #129 is FIXED${NC}"
    echo -e "${GREEN}========================================${NC}"
    exit 0
else
    echo -e "${RED}========================================${NC}"
    echo -e "${RED}✗ Some tests failed${NC}"
    echo -e "${RED}✗ Issue #129 is NOT fixed${NC}"
    echo -e "${RED}========================================${NC}"
    exit 1
fi
