#!/bin/bash
# Test script for GitHub Issue #126: Setup wizard offers Meta Llama provider but validation rejects it
# This script verifies that Meta Llama provider is now properly supported in setup wizard validation

set -e

echo "=========================================="
echo "Testing GitHub Issue #126 Fix"
echo "Setup wizard Meta Llama provider validation"
echo "=========================================="
echo ""

# Colors
GREEN='\033[0;32m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# Test counter
TESTS_PASSED=0
TESTS_FAILED=0

# Function to report test result
report_test() {
    if [ $1 -eq 0 ]; then
        echo -e "${GREEN}✓${NC} $2"
        TESTS_PASSED=$((TESTS_PASSED + 1))
    else
        echo -e "${RED}✗${NC} $2"
        TESTS_FAILED=$((TESTS_FAILED + 1))
    fi
}

echo "Test 1: Run ValidateMetaLlamaKey unit tests"
echo "--------------------------------------------"
go test -v ./internal/setup -run TestValidateMetaLlamaKey > /tmp/test126_1.log 2>&1
result=$?
if [ $result -eq 0 ]; then
    echo "✓ ValidateMetaLlamaKey tests passed"
    report_test 0 "ValidateMetaLlamaKey method works correctly"
else
    echo "✗ ValidateMetaLlamaKey tests failed"
    cat /tmp/test126_1.log
    report_test 1 "ValidateMetaLlamaKey method works correctly"
fi
echo ""

echo "Test 2: Run ValidateProviderConfig with meta_llama"
echo "---------------------------------------------------"
go test -v ./internal/setup -run "TestValidateProviderConfig/meta_llama" > /tmp/test126_2.log 2>&1
result=$?
if [ $result -eq 0 ]; then
    echo "✓ meta_llama provider is accepted by ValidateProviderConfig"
    report_test 0 "meta_llama provider validation works"
else
    echo "✗ meta_llama provider validation failed"
    cat /tmp/test126_2.log
    report_test 1 "meta_llama provider validation works"
fi
echo ""

echo "Test 3: Run ValidateProviderConfig with meta alias"
echo "---------------------------------------------------"
go test -v ./internal/setup -run "TestValidateProviderConfig/meta_-_alias" > /tmp/test126_3.log 2>&1
result=$?
if [ $result -eq 0 ]; then
    echo "✓ meta alias is accepted by ValidateProviderConfig"
    report_test 0 "meta alias works correctly"
else
    echo "✗ meta alias validation failed"
    cat /tmp/test126_3.log
    report_test 1 "meta alias works correctly"
fi
echo ""

echo "Test 4: Run comprehensive setup wizard provider validation"
echo "-----------------------------------------------------------"
go test -v ./tests/integration -run TestSetupWizardProviderValidation > /tmp/test126_4.log 2>&1
result=$?
if [ $result -eq 0 ]; then
    echo "✓ All wizard-offered providers are supported by validation"
    report_test 0 "All wizard providers validated successfully"
else
    echo "✗ Some wizard providers failed validation"
    cat /tmp/test126_4.log
    report_test 1 "All wizard providers validated successfully"
fi
echo ""

echo "Test 5: Verify provider selection mapping"
echo "------------------------------------------"
go test -v ./tests/integration -run TestProviderSelectionMapping > /tmp/test126_5.log 2>&1
result=$?
if [ $result -eq 0 ]; then
    echo "✓ Provider selection indices map correctly to supported providers"
    report_test 0 "Provider selection mapping is correct"
else
    echo "✗ Provider selection mapping failed"
    cat /tmp/test126_5.log
    report_test 1 "Provider selection mapping is correct"
fi
echo ""

echo "Test 6: Verify Meta Llama alias support"
echo "----------------------------------------"
go test -v ./tests/integration -run TestMetaLlamaAliasSupport > /tmp/test126_6.log 2>&1
result=$?
if [ $result -eq 0 ]; then
    echo "✓ Both meta_llama and meta aliases work"
    report_test 0 "Meta Llama alias support works"
else
    echo "✗ Meta Llama alias support failed"
    cat /tmp/test126_6.log
    report_test 1 "Meta Llama alias support works"
fi
echo ""

echo "Test 7: Verify unsupported providers are still rejected"
echo "--------------------------------------------------------"
go test -v ./tests/integration -run TestUnsupportedProvider > /tmp/test126_7.log 2>&1
result=$?
if [ $result -eq 0 ]; then
    echo "✓ Unsupported providers are properly rejected"
    report_test 0 "Unsupported provider rejection works"
else
    echo "✗ Unsupported provider test failed"
    cat /tmp/test126_7.log
    report_test 1 "Unsupported provider rejection works"
fi
echo ""

echo "Test 8: Check that meta provider implementation exists"
echo "-------------------------------------------------------"
if [ -d "./internal/provider/meta" ] && [ -f "./internal/provider/meta/client.go" ]; then
    echo "✓ Meta provider implementation found"
    report_test 0 "Meta provider implementation exists"
else
    echo "✗ Meta provider implementation not found"
    report_test 1 "Meta provider implementation exists"
fi
echo ""

echo "=========================================="
echo "Test Summary"
echo "=========================================="
echo "Tests Passed: $TESTS_PASSED"
echo "Tests Failed: $TESTS_FAILED"
echo ""

if [ $TESTS_FAILED -eq 0 ]; then
    echo -e "${GREEN}All tests passed! Issue #126 is fixed.${NC}"
    echo ""
    echo "Fix Summary:"
    echo "- Added ValidateMetaLlamaKey() method to validation.go"
    echo "- Added meta_llama and meta cases to ValidateProviderConfig()"
    echo "- Added comprehensive unit and integration tests"
    echo "- Meta Llama provider offered in wizard now works correctly"
    exit 0
else
    echo -e "${RED}Some tests failed. Issue #126 may not be fully resolved.${NC}"
    exit 1
fi
