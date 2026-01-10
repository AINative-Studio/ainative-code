#!/bin/bash

# Test script for ZeroDB Production Integration Tests
# This script loads production credentials from .env and runs integration tests

set -e

echo "==================================================================="
echo "ZeroDB Production Integration Test Suite"
echo "GitHub Issue #116: Setup wizard ZeroDB/Strapi configuration"
echo "==================================================================="
echo ""

# Load environment variables
if [ -f ".env" ]; then
    echo "✓ Loading production credentials from .env..."
    export $(cat .env | grep -E "^(ZERODB_|TEST_STRAPI_)" | xargs)
    echo "✓ Environment variables loaded"
    echo ""
else
    echo "❌ ERROR: .env file not found!"
    exit 1
fi

# Verify credentials are loaded
echo "=== Production Configuration ==="
echo "ZERODB_API_BASE_URL: ${ZERODB_API_BASE_URL:-NOT SET}"
echo "ZERODB_EMAIL: ${ZERODB_EMAIL:-NOT SET}"
echo "ZERODB_API_KEY: ${ZERODB_API_KEY:0:20}... (truncated)"
echo ""

# Run production integration tests
echo "=== Running ZeroDB Production API Tests ==="
echo "Target: ${ZERODB_API_BASE_URL}"
echo ""

go test -v ./tests/integration/zerodb_production_api_test.go -timeout 5m

echo ""
echo "==================================================================="
echo "Production Integration Tests Complete"
echo "==================================================================="
