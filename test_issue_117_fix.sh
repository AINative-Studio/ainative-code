#!/bin/bash

# Test script for Issue #117: Setup wizard offers outdated model that chat command rejects
# This script verifies that the fix ensures setup wizard and chat command use compatible models

set -e

echo "=========================================="
echo "Testing Issue #117 Fix"
echo "=========================================="
echo ""

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Test 1: Verify unit tests pass
echo -e "${BLUE}Test 1: Running unit tests for model compatibility${NC}"
go test -v ./tests/integration/issue_117_model_sync_test.go -run TestIssue117
if [ $? -eq 0 ]; then
    echo -e "${GREEN}✓ Unit tests passed${NC}"
else
    echo -e "${RED}✗ Unit tests failed${NC}"
    exit 1
fi
echo ""

# Test 2: Check that prompts.go has updated models
echo -e "${BLUE}Test 2: Verifying prompts.go uses Claude 4.5 models${NC}"
if grep -q "claude-sonnet-4-5-20250929" internal/setup/prompts.go; then
    echo -e "${GREEN}✓ Found claude-sonnet-4-5-20250929 in prompts.go${NC}"
else
    echo -e "${RED}✗ claude-sonnet-4-5-20250929 not found in prompts.go${NC}"
    exit 1
fi

if grep -q "claude-3-5-sonnet-20241022" internal/setup/prompts.go; then
    echo -e "${RED}✗ Old model claude-3-5-sonnet-20241022 still present in prompts.go${NC}"
    exit 1
else
    echo -e "${GREEN}✓ Old model claude-3-5-sonnet-20241022 not in prompts.go${NC}"
fi
echo ""

# Test 3: Check that wizard.go has updated default model
echo -e "${BLUE}Test 3: Verifying wizard.go uses Claude 4.5 default model${NC}"
if grep -q 'model := "claude-sonnet-4-5-20250929"' internal/setup/wizard.go; then
    echo -e "${GREEN}✓ Default model in wizard.go is claude-sonnet-4-5-20250929${NC}"
else
    echo -e "${RED}✗ Default model in wizard.go is not claude-sonnet-4-5-20250929${NC}"
    exit 1
fi

if grep -q 'model := "claude-3-5-sonnet-20241022"' internal/setup/wizard.go; then
    echo -e "${RED}✗ Old default model claude-3-5-sonnet-20241022 still in wizard.go${NC}"
    exit 1
else
    echo -e "${GREEN}✓ Old default model not in wizard.go${NC}"
fi
echo ""

# Test 4: Verify chat.go has updated default model
echo -e "${BLUE}Test 4: Verifying chat.go uses Claude 4.5 default model${NC}"
if grep -q 'return "claude-sonnet-4-5-20250929"' internal/cmd/chat.go; then
    echo -e "${GREEN}✓ Default model in chat.go is claude-sonnet-4-5-20250929${NC}"
else
    echo -e "${RED}✗ Default model in chat.go is not claude-sonnet-4-5-20250929${NC}"
    exit 1
fi
echo ""

# Test 5: Verify Anthropic provider has the models listed
echo -e "${BLUE}Test 5: Verifying Anthropic provider supports all wizard models${NC}"
wizard_models=("claude-sonnet-4-5-20250929" "claude-haiku-4-5-20251001" "claude-opus-4-1" "claude-sonnet-4-5" "claude-haiku-4-5")
for model in "${wizard_models[@]}"; do
    if grep -q "\"$model\"" internal/provider/anthropic/anthropic.go; then
        echo -e "${GREEN}✓ Model $model found in Anthropic provider${NC}"
    else
        echo -e "${RED}✗ Model $model NOT found in Anthropic provider${NC}"
        exit 1
    fi
done
echo ""

# Test 6: Build the binary to ensure no compilation errors
echo -e "${BLUE}Test 6: Building binary to ensure no compilation errors${NC}"
go build -o /tmp/ainative-code-test ./cmd/ainative-code
if [ $? -eq 0 ]; then
    echo -e "${GREEN}✓ Binary built successfully${NC}"
    rm -f /tmp/ainative-code-test
else
    echo -e "${RED}✗ Binary build failed${NC}"
    exit 1
fi
echo ""

# Test 7: Run existing setup tests to ensure no regressions
echo -e "${BLUE}Test 7: Running existing setup wizard tests for regressions${NC}"
go test -v ./internal/setup -run TestPromptModel
if [ $? -eq 0 ]; then
    echo -e "${GREEN}✓ Setup wizard tests passed${NC}"
else
    echo -e "${RED}✗ Setup wizard tests failed${NC}"
    exit 1
fi
echo ""

echo "=========================================="
echo -e "${GREEN}All tests passed! Issue #117 fix verified.${NC}"
echo "=========================================="
echo ""
echo "Summary:"
echo "- Setup wizard now offers only Claude 4.5 models"
echo "- Default model updated to claude-sonnet-4-5-20250929"
echo "- All wizard models are validated by chat command"
echo "- Old claude-3-5-sonnet-20241022 model removed from wizard"
echo ""
