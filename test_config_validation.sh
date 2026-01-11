#!/bin/bash
# Integration test for GitHub issue #110: --config flag validation
# Tests that nonexistent config files are properly detected and reported

set -e

echo "Testing GitHub Issue #110: --config flag validation"
echo "===================================================="
echo ""

# Build the binary
echo "Building ainative-code..."
go build -o /tmp/ainative-code-test ./cmd/ainative-code/
echo "Build complete."
echo ""

# Test 1: Nonexistent config file should fail
echo "Test 1: Nonexistent config file"
echo "Command: /tmp/ainative-code-test --config /nonexistent/config.yaml version"
if /tmp/ainative-code-test --config /nonexistent/config.yaml version 2>&1 | grep -q "Error: config file not found"; then
    echo "✓ PASS: Correctly detected nonexistent config file"
else
    echo "✗ FAIL: Did not detect nonexistent config file"
    exit 1
fi
echo ""

# Test 2: Directory instead of file should fail
echo "Test 2: Directory instead of file"
mkdir -p /tmp/test-config-dir
echo "Command: /tmp/ainative-code-test --config /tmp/test-config-dir version"
if /tmp/ainative-code-test --config /tmp/test-config-dir version 2>&1 | grep -q "Error: config path is a directory"; then
    echo "✓ PASS: Correctly detected directory instead of file"
else
    echo "✗ FAIL: Did not detect directory instead of file"
    exit 1
fi
echo ""

# Test 3: Valid config file should succeed
echo "Test 3: Valid config file"
echo "provider: openai" > /tmp/test-valid-config.yaml
echo "Command: /tmp/ainative-code-test --config /tmp/test-valid-config.yaml version"
if /tmp/ainative-code-test --config /tmp/test-valid-config.yaml version 2>&1 | grep -q "AINative Code"; then
    echo "✓ PASS: Successfully used valid config file"
else
    echo "✗ FAIL: Failed with valid config file"
    exit 1
fi
echo ""

# Test 4: No --config flag should work (use defaults)
echo "Test 4: No --config flag (default behavior)"
echo "Command: /tmp/ainative-code-test version"
if /tmp/ainative-code-test version 2>&1 | grep -q "AINative Code"; then
    echo "✓ PASS: Successfully ran without --config flag"
else
    echo "✗ FAIL: Failed without --config flag"
    exit 1
fi
echo ""

# Cleanup
rm -f /tmp/ainative-code-test
rm -f /tmp/test-valid-config.yaml
rm -rf /tmp/test-config-dir

echo "===================================================="
echo "All tests passed! Issue #110 is fixed."
echo ""
echo "Summary:"
echo "- Nonexistent config files are properly detected"
echo "- Directories are rejected as config files"
echo "- Valid config files work correctly"
echo "- Default behavior (no --config) still works"
