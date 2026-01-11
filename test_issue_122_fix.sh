#!/bin/bash
# Test script for Issue #122: Config validate fails due to empty root provider field
# This script tests the fix for config validation logic

set -e

echo "==================================================================="
echo "Test Script for Issue #122: Config Validate Fix"
echo "==================================================================="
echo ""

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Build the binary first
echo "Building ainative-code..."
go build -o ainative-code-test ./cmd/ainative-code
if [ $? -ne 0 ]; then
    echo -e "${RED}FAILED: Build failed${NC}"
    exit 1
fi
echo -e "${GREEN}✓ Build successful${NC}"
echo ""

# Create temporary directory for test configs
TEST_DIR=$(mktemp -d)
trap "rm -rf $TEST_DIR" EXIT

echo "Test directory: $TEST_DIR"
echo ""

# Test Case 1: Config with empty root provider and valid nested provider (Issue #122)
echo "==================================================================="
echo "Test 1: Config with empty root 'provider' and valid 'llm.default_provider'"
echo "==================================================================="
cat > "$TEST_DIR/config1.yaml" << 'EOF'
provider: ""
llm:
  default_provider: anthropic
  anthropic:
    api_key: test-key
    model: claude-3-5-sonnet-20241022
EOF

echo "Config file content:"
cat "$TEST_DIR/config1.yaml"
echo ""

echo "Running: ainative-code-test config validate --config=$TEST_DIR/config1.yaml"
if ./ainative-code-test config validate --config="$TEST_DIR/config1.yaml" 2>&1; then
    echo -e "${GREEN}✓ PASS: Validation succeeded with nested provider${NC}"
else
    echo -e "${RED}✗ FAIL: Validation should succeed with valid nested provider${NC}"
    exit 1
fi
echo ""

# Test Case 2: Config with only nested provider (no root provider)
echo "==================================================================="
echo "Test 2: Config with only 'llm.default_provider' (no root field)"
echo "==================================================================="
cat > "$TEST_DIR/config2.yaml" << 'EOF'
llm:
  default_provider: openai
  openai:
    api_key: test-key
    model: gpt-4
EOF

echo "Config file content:"
cat "$TEST_DIR/config2.yaml"
echo ""

echo "Running: ainative-code-test config validate --config=$TEST_DIR/config2.yaml"
if ./ainative-code-test config validate --config="$TEST_DIR/config2.yaml" 2>&1; then
    echo -e "${GREEN}✓ PASS: Validation succeeded with only nested provider${NC}"
else
    echo -e "${RED}✗ FAIL: Validation should succeed with nested provider${NC}"
    exit 1
fi
echo ""

# Test Case 3: Config with only root provider (legacy format)
echo "==================================================================="
echo "Test 3: Config with legacy 'provider' field only"
echo "==================================================================="
cat > "$TEST_DIR/config3.yaml" << 'EOF'
provider: anthropic
model: claude-3-5-sonnet-20241022
EOF

echo "Config file content:"
cat "$TEST_DIR/config3.yaml"
echo ""

echo "Running: ainative-code-test config validate --config=$TEST_DIR/config3.yaml"
if ./ainative-code-test config validate --config="$TEST_DIR/config3.yaml" 2>&1 | tee /tmp/test3_output.txt; then
    echo -e "${GREEN}✓ PASS: Validation succeeded with legacy provider${NC}"

    # Check for migration warning
    if grep -q "legacy 'provider' field" /tmp/test3_output.txt; then
        echo -e "${GREEN}✓ PASS: Migration warning displayed${NC}"
    else
        echo -e "${YELLOW}⚠ WARNING: Migration notice not displayed${NC}"
    fi
else
    echo -e "${RED}✗ FAIL: Validation should succeed with legacy provider${NC}"
    exit 1
fi
echo ""

# Test Case 4: Config with both providers (should prefer nested)
echo "==================================================================="
echo "Test 4: Config with both providers (should prefer nested)"
echo "==================================================================="
cat > "$TEST_DIR/config4.yaml" << 'EOF'
provider: openai
llm:
  default_provider: anthropic
  anthropic:
    api_key: test-key
EOF

echo "Config file content:"
cat "$TEST_DIR/config4.yaml"
echo ""

echo "Running: ainative-code-test config validate --config=$TEST_DIR/config4.yaml"
if ./ainative-code-test config validate --config="$TEST_DIR/config4.yaml" 2>&1 | tee /tmp/test4_output.txt; then
    echo -e "${GREEN}✓ PASS: Validation succeeded${NC}"

    # Check that it uses anthropic (nested), not openai (root)
    if grep -q "Provider: anthropic" /tmp/test4_output.txt; then
        echo -e "${GREEN}✓ PASS: Correctly prioritized nested provider${NC}"
    else
        echo -e "${RED}✗ FAIL: Should prioritize nested provider over root${NC}"
        exit 1
    fi
else
    echo -e "${RED}✗ FAIL: Validation should succeed${NC}"
    exit 1
fi
echo ""

# Test Case 5: Config with no provider (should fail)
echo "==================================================================="
echo "Test 5: Config with no provider (should fail)"
echo "==================================================================="
cat > "$TEST_DIR/config5.yaml" << 'EOF'
model: claude-3-5-sonnet-20241022
EOF

echo "Config file content:"
cat "$TEST_DIR/config5.yaml"
echo ""

echo "Running: ainative-code-test config validate --config=$TEST_DIR/config5.yaml"
if ./ainative-code-test config validate --config="$TEST_DIR/config5.yaml" 2>&1; then
    echo -e "${RED}✗ FAIL: Validation should fail with no provider${NC}"
    exit 1
else
    echo -e "${GREEN}✓ PASS: Validation correctly failed with no provider${NC}"
fi
echo ""

# Test Case 6: Config with empty root and empty nested provider (should fail)
echo "==================================================================="
echo "Test 6: Config with both providers empty (should fail)"
echo "==================================================================="
cat > "$TEST_DIR/config6.yaml" << 'EOF'
provider: ""
llm:
  default_provider: ""
EOF

echo "Config file content:"
cat "$TEST_DIR/config6.yaml"
echo ""

echo "Running: ainative-code-test config validate --config=$TEST_DIR/config6.yaml"
if ./ainative-code-test config validate --config="$TEST_DIR/config6.yaml" 2>&1; then
    echo -e "${RED}✗ FAIL: Validation should fail with empty providers${NC}"
    exit 1
else
    echo -e "${GREEN}✓ PASS: Validation correctly failed with empty providers${NC}"
fi
echo ""

# Test Case 7: Config with invalid provider value
echo "==================================================================="
echo "Test 7: Config with invalid provider value (should fail)"
echo "==================================================================="
cat > "$TEST_DIR/config7.yaml" << 'EOF'
llm:
  default_provider: invalid-provider
EOF

echo "Config file content:"
cat "$TEST_DIR/config7.yaml"
echo ""

echo "Running: ainative-code-test config validate --config=$TEST_DIR/config7.yaml"
if ./ainative-code-test config validate --config="$TEST_DIR/config7.yaml" 2>&1; then
    echo -e "${RED}✗ FAIL: Validation should fail with invalid provider${NC}"
    exit 1
else
    echo -e "${GREEN}✓ PASS: Validation correctly failed with invalid provider${NC}"
fi
echo ""

# Test Case 8: Test all valid providers
echo "==================================================================="
echo "Test 8: Test all valid providers"
echo "==================================================================="
for provider in anthropic openai ollama google bedrock azure meta_llama meta; do
    echo "Testing provider: $provider"
    cat > "$TEST_DIR/config_$provider.yaml" << EOF
llm:
  default_provider: $provider
EOF

    if ./ainative-code-test config validate --config="$TEST_DIR/config_$provider.yaml" 2>&1 | grep -q "Configuration is valid"; then
        echo -e "${GREEN}✓ PASS: Provider '$provider' validated${NC}"
    else
        echo -e "${RED}✗ FAIL: Provider '$provider' should be valid${NC}"
        exit 1
    fi
done
echo ""

# Clean up test binary
rm -f ainative-code-test

echo "==================================================================="
echo -e "${GREEN}ALL TESTS PASSED!${NC}"
echo "==================================================================="
echo ""
echo "Summary:"
echo "- Empty root provider with valid nested provider: ✓"
echo "- Only nested provider (no root): ✓"
echo "- Legacy root provider only: ✓"
echo "- Both providers (prioritization): ✓"
echo "- No provider (error handling): ✓"
echo "- Empty providers (error handling): ✓"
echo "- Invalid provider (error handling): ✓"
echo "- All valid providers: ✓"
echo ""
echo "Issue #122 has been successfully fixed!"
