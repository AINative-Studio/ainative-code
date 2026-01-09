#!/bin/bash

# Test script for install.sh fallback logic
# This script tests various scenarios for the installation fallback mechanism

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

print_test() {
    echo -e "${BLUE}[TEST]${NC} $1"
}

print_pass() {
    echo -e "${GREEN}[PASS]${NC} $1"
}

print_fail() {
    echo -e "${RED}[FAIL]${NC} $1"
}

print_info() {
    echo -e "${YELLOW}[INFO]${NC} $1"
}

# Test 1: Check that install.sh exists
test_script_exists() {
    print_test "Checking if install.sh exists..."
    if [ -f "install.sh" ]; then
        print_pass "install.sh found"
        return 0
    else
        print_fail "install.sh not found"
        return 1
    fi
}

# Test 2: Check for required functions
test_required_functions() {
    print_test "Checking for required functions in install.sh..."

    local required_functions=(
        "install_binary"
        "install_to_user_directory"
        "setup_path"
        "check_path"
        "detect_platform"
    )

    local missing_functions=()

    for func in "${required_functions[@]}"; do
        if ! grep -q "^${func}()" install.sh; then
            missing_functions+=("$func")
        fi
    done

    if [ ${#missing_functions[@]} -eq 0 ]; then
        print_pass "All required functions found"
        return 0
    else
        print_fail "Missing functions: ${missing_functions[*]}"
        return 1
    fi
}

# Test 3: Check for fallback logic
test_fallback_logic() {
    print_test "Checking for fallback installation logic..."

    if grep -q "install_to_user_directory" install.sh; then
        print_pass "Fallback function call found"
    else
        print_fail "Fallback function call not found"
        return 1
    fi

    if grep -q 'sudo -n true' install.sh; then
        print_pass "Sudo password check found"
    else
        print_fail "Sudo password check not found"
        return 1
    fi

    if grep -q '~/.local/bin\|$HOME/.local/bin' install.sh; then
        print_pass "User directory fallback path found"
    else
        print_fail "User directory fallback path not found"
        return 1
    fi

    return 0
}

# Test 4: Check for PATH setup instructions
test_path_instructions() {
    print_test "Checking for PATH setup instructions..."

    if grep -q "export PATH=" install.sh; then
        print_pass "PATH export instructions found"
        return 0
    else
        print_fail "PATH export instructions not found"
        return 1
    fi
}

# Test 5: Check that set -e is removed
test_error_handling() {
    print_test "Checking error handling..."

    if grep -q "^set -e" install.sh; then
        print_fail "Script still uses 'set -e' which prevents graceful error handling"
        return 1
    else
        print_pass "Script allows graceful error handling"
    fi

    if grep -q "return 1" install.sh; then
        print_pass "Functions use return codes"
        return 0
    else
        print_fail "Functions don't use return codes properly"
        return 1
    fi
}

# Test 6: Check for user feedback messages
test_user_feedback() {
    print_test "Checking for user feedback messages..."

    local required_messages=(
        "Falling back to user directory"
        "Sudo installation failed or was cancelled"
        "Successfully installed to"
    )

    local missing_messages=()

    for msg in "${required_messages[@]}"; do
        if ! grep -qi "$msg" install.sh; then
            missing_messages+=("$msg")
        fi
    done

    if [ ${#missing_messages[@]} -eq 0 ]; then
        print_pass "All required user feedback messages found"
        return 0
    else
        print_fail "Missing messages: ${missing_messages[*]}"
        return 1
    fi
}

# Test 7: Syntax check
test_syntax() {
    print_test "Running bash syntax check..."

    if bash -n install.sh 2>/dev/null; then
        print_pass "Script syntax is valid"
        return 0
    else
        print_fail "Script has syntax errors"
        bash -n install.sh
        return 1
    fi
}

# Main test runner
main() {
    echo ""
    echo "╔════════════════════════════════════════════╗"
    echo "║   Install Script Fallback Logic Tests     ║"
    echo "╚════════════════════════════════════════════╝"
    echo ""

    local tests=(
        "test_script_exists"
        "test_syntax"
        "test_required_functions"
        "test_fallback_logic"
        "test_path_instructions"
        "test_error_handling"
        "test_user_feedback"
    )

    local passed=0
    local failed=0

    for test in "${tests[@]}"; do
        if $test; then
            ((passed++))
        else
            ((failed++))
        fi
        echo ""
    done

    echo "╔════════════════════════════════════════════╗"
    echo "║              Test Results                  ║"
    echo "╚════════════════════════════════════════════╝"
    echo ""
    echo -e "${GREEN}Passed: $passed${NC}"
    echo -e "${RED}Failed: $failed${NC}"
    echo ""

    if [ $failed -eq 0 ]; then
        print_pass "All tests passed!"
        return 0
    else
        print_fail "Some tests failed"
        return 1
    fi
}

# Run tests
main "$@"
