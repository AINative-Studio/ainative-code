#!/bin/bash
# AINative Code - Branding Verification Script
# © 2024 AINative Studio. All rights reserved.

echo "=================================="
echo "AINative Code Branding Verification"
echo "=================================="
echo ""

# Color codes
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Check for old branding in code files
echo "1. Checking for old branding in code files..."
OLD_BRAND_COUNT=$(grep -r "crush" . --include="*.go" --include="*.yaml" --include="*.yml" --include="*.json" --include="*.toml" -i 2>/dev/null | grep -v "TestNoCrushReferences" | grep -v "constants_test.go" | wc -l | xargs)

if [ "$OLD_BRAND_COUNT" -eq 0 ]; then
    echo -e "${GREEN}✓ No old branding references in code files${NC}"
else
    echo -e "${RED}✗ Found $OLD_BRAND_COUNT old branding references in code files${NC}"
fi

# Check for old branding in markdown files (excluding completion reports)
echo "2. Checking for old branding in markdown files..."
MD_OLD_BRAND_COUNT=$(grep -r "crush" . --include="*.md" -i 2>/dev/null | grep -v "TASK-002-COMPLETION-REPORT.md" | grep -v "TASK-002-SUMMARY.md" | wc -l | xargs)

if [ "$MD_OLD_BRAND_COUNT" -eq 0 ]; then
    echo -e "${GREEN}✓ No old branding references in markdown files${NC}"
else
    echo -e "${RED}✗ Found $MD_OLD_BRAND_COUNT old branding references in markdown files${NC}"
fi

# Verify module path
echo "3. Checking Go module path..."
MODULE_PATH=$(grep "module" go.mod | awk '{print $2}')

if [ "$MODULE_PATH" == "github.com/AINative-studio/ainative-code" ]; then
    echo -e "${GREEN}✓ Correct module path: $MODULE_PATH${NC}"
else
    echo -e "${RED}✗ Incorrect module path: $MODULE_PATH${NC}"
fi

# Verify branding constants exist
echo "4. Checking branding constants..."
if [ -f "internal/branding/constants.go" ]; then
    echo -e "${GREEN}✓ Branding constants file exists${NC}"
else
    echo -e "${RED}✗ Branding constants file not found${NC}"
fi

# Verify example config exists
echo "5. Checking example configuration..."
if [ -f "configs/example-config.yaml" ]; then
    echo -e "${GREEN}✓ Example configuration file exists${NC}"
else
    echo -e "${RED}✗ Example configuration file not found${NC}"
fi

# Run branding tests
echo "6. Running branding tests..."
if go test ./internal/branding/... -v > /dev/null 2>&1; then
    echo -e "${GREEN}✓ All branding tests passing${NC}"
else
    echo -e "${RED}✗ Some branding tests failing${NC}"
fi

# Verify binary name in main.go
echo "7. Checking binary configuration..."
if grep -q "ainative-code" cmd/ainative-code/main.go; then
    echo -e "${GREEN}✓ Binary correctly configured${NC}"
else
    echo -e "${YELLOW}⚠ Could not verify binary configuration${NC}"
fi

# Summary
echo ""
echo "=================================="
echo "Verification Complete"
echo "=================================="

if [ "$OLD_BRAND_COUNT" -eq 0 ] && [ "$MD_OLD_BRAND_COUNT" -eq 0 ] && [ "$MODULE_PATH" == "github.com/AINative-studio/ainative-code" ]; then
    echo -e "${GREEN}✓ All branding checks passed!${NC}"
    exit 0
else
    echo -e "${RED}✗ Some branding checks failed${NC}"
    exit 1
fi
