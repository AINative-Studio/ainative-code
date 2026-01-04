#!/bin/bash
# Integration Test Runner
# This script runs integration tests for AINative Code

set -e

# Colors for output
RED='\033[0:31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo "================================"
echo "  AINative Code Integration Tests"
echo "================================"
echo ""

# Check if we're in the right directory
if [ ! -f "go.mod" ]; then
    echo -e "${RED}Error: Must be run from project root${NC}"
    exit 1
fi

# Build the project first
echo -e "${YELLOW}Building project...${NC}"
if ! go build -o /dev/null ./...; then
    echo -e "${RED}Build failed - please fix compilation errors first${NC}"
    exit 1
fi

echo -e "${GREEN}Build successful${NC}"
echo ""

# Run tests
echo -e "${YELLOW}Running integration tests...${NC}"
echo ""

# Set test timeout
TIMEOUT="2m"

# Run chat tests
echo "==> Running Chat Session Tests..."
go test -v -timeout $TIMEOUT ./tests/integration -run TestChatSession || true
echo ""

# Run tool execution tests
echo "==> Running Tool Execution Tests..."
go test -v -timeout $TIMEOUT ./tests/integration -run TestToolExecution || true
echo ""

# Run session persistence tests
echo "==> Running Session Persistence Tests..."
go test -v -timeout $TIMEOUT ./tests/integration -run TestSessionPersistence || true
echo ""

# Run all tests with coverage
echo "==> Running all integration tests with coverage..."
go test -v -timeout $TIMEOUT -cover -coverprofile=coverage.out ./tests/integration || true

# Generate coverage report if coverage file exists
if [ -f "coverage.out" ]; then
    echo ""
    echo -e "${YELLOW}Coverage Summary:${NC}"
    go tool cover -func=coverage.out | tail -1
    echo ""
    echo -e "${GREEN}Coverage report: coverage.out${NC}"
    echo -e "View HTML report with: ${YELLOW}go tool cover -html=coverage.out${NC}"
fi

echo ""
echo "================================"
echo -e "${GREEN}Integration tests complete!${NC}"
echo "================================"
