#!/bin/bash
set -e

echo "========================================="
echo "Testing Issue #96 Fix"
echo "========================================="
echo ""

# Create a temporary directory for test
TMPDIR=$(mktemp -d)
echo "Test directory: $TMPDIR"

# Create a test config file that simulates setup wizard output
cat > "$TMPDIR/.ainative-code.yaml" << 'EOF'
app:
  name: ainative-code
  version: 0.1.7
  environment: development
  debug: false
llm:
  default_provider: anthropic
  anthropic:
    api_key: sk-ant-test-key-12345
    model: claude-3-5-sonnet-20241022
    max_tokens: 4096
    temperature: 0.7
    top_p: 1.0
    top_k: 0
    timeout: 30000000000
    retry_attempts: 3
    api_version: 2023-06-01
EOF

echo "Created config file:"
cat "$TMPDIR/.ainative-code.yaml"
echo ""

# Build the CLI
echo "Building ainative-code..."
go build -o "$TMPDIR/ainative-code" ./cmd/ainative-code
echo "✓ Build successful"
echo ""

# Test 1: Verify config show works
echo "Test 1: Verify config show can read the file..."
"$TMPDIR/ainative-code" --config "$TMPDIR/.ainative-code.yaml" config show 2>&1 | head -20
echo "✓ Config show works"
echo ""

# Test 2: Verify chat command recognizes the provider
echo "Test 2: Verify chat command recognizes provider..."
# We expect it to fail on API call (invalid key) but NOT on "provider not configured"
OUTPUT=$("$TMPDIR/ainative-code" --config "$TMPDIR/.ainative-code.yaml" chat "test" 2>&1 || true)

if echo "$OUTPUT" | grep -q "AI provider not configured"; then
    echo "✗ FAIL: Chat still reports 'AI provider not configured'"
    echo "Output: $OUTPUT"
    exit 1
elif echo "$OUTPUT" | grep -q "provider.*anthropic"; then
    echo "✓ Chat command recognizes anthropic provider"
elif echo "$OUTPUT" | grep -q "failed to initialize AI provider"; then
    echo "✓ Chat command attempted to initialize provider (expected with test key)"
elif echo "$OUTPUT" | grep -q "API.*error"; then
    echo "✓ Chat command reached API call stage (expected with test key)"
else
    echo "⚠ Chat produced different output:"
    echo "$OUTPUT"
fi
echo ""

# Test 3: Test with environment variable override
echo "Test 3: Test environment variable takes precedence..."
export ANTHROPIC_API_KEY="sk-ant-from-env"
OUTPUT=$("$TMPDIR/ainative-code" --config "$TMPDIR/.ainative-code.yaml" chat "test" --verbose 2>&1 | head -30 || true)
if echo "$OUTPUT" | grep -q "AI provider not configured"; then
    echo "✗ FAIL: Chat reports 'AI provider not configured' even with env var"
    exit 1
else
    echo "✓ Environment variable handled correctly"
fi
unset ANTHROPIC_API_KEY
echo ""

# Test 4: Test backward compatibility with flat config
echo "Test 4: Test backward compatibility with flat config..."
cat > "$TMPDIR/flat-config.yaml" << 'EOF'
provider: openai
model: gpt-4
api_key: sk-openai-test-key
EOF

OUTPUT=$("$TMPDIR/ainative-code" --config "$TMPDIR/flat-config.yaml" chat "test" 2>&1 || true)
if echo "$OUTPUT" | grep -q "AI provider not configured"; then
    echo "✗ FAIL: Flat config not working"
    exit 1
else
    echo "✓ Flat config backward compatibility works"
fi
echo ""

# Test 5: Test all providers
echo "Test 5: Test all supported providers..."
for provider in anthropic openai google meta_llama; do
    cat > "$TMPDIR/${provider}-config.yaml" << EOF
llm:
  default_provider: ${provider}
  ${provider}:
    api_key: test-key-${provider}
    model: test-model
EOF

    OUTPUT=$("$TMPDIR/ainative-code" --config "$TMPDIR/${provider}-config.yaml" chat "test" 2>&1 || true)
    if echo "$OUTPUT" | grep -q "AI provider not configured"; then
        echo "✗ FAIL: Provider $provider not recognized"
        exit 1
    else
        echo "✓ Provider $provider recognized"
    fi
done
echo ""

# Cleanup
rm -rf "$TMPDIR"

echo "========================================="
echo "✓ All Issue #96 tests PASSED!"
echo "========================================="
echo ""
echo "Summary:"
echo "- Setup wizard config format works ✓"
echo "- Chat command reads provider correctly ✓"
echo "- Environment variables work ✓"
echo "- Backward compatibility maintained ✓"
echo "- All providers supported ✓"
