#!/bin/bash

# Test script for MCP server persistence (Issue #120)
# Tests the full workflow: add → list → restart → list → remove → list

set -e  # Exit on any error

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Test configuration
TEST_MCP_CONFIG="/tmp/test-mcp-persistence-$$.json"
export MCP_CONFIG_PATH="$TEST_MCP_CONFIG"

BINARY="./build/ainative-code"

echo -e "${YELLOW}==================================================${NC}"
echo -e "${YELLOW}Testing MCP Server Persistence (Issue #120)${NC}"
echo -e "${YELLOW}==================================================${NC}"
echo ""
echo "Config file: $TEST_MCP_CONFIG"
echo ""

# Cleanup function
cleanup() {
    echo ""
    echo -e "${YELLOW}Cleaning up...${NC}"
    rm -f "$TEST_MCP_CONFIG"
    rm -f "${TEST_MCP_CONFIG}.tmp"
}

trap cleanup EXIT

# Test 1: Add a server
echo -e "${YELLOW}Test 1: Adding MCP server 'test-server'${NC}"
$BINARY mcp add-server --name test-server --url http://localhost:9000 2>&1 | grep -q "Successfully added MCP server: test-server"
if [ $? -eq 0 ]; then
    echo -e "${GREEN}✓ Server added successfully${NC}"
else
    echo -e "${RED}✗ Failed to add server${NC}"
    exit 1
fi
echo ""

# Test 2: Verify config file was created
echo -e "${YELLOW}Test 2: Verifying config file was created${NC}"
if [ -f "$TEST_MCP_CONFIG" ]; then
    echo -e "${GREEN}✓ Config file exists${NC}"
    echo "Config file contents:"
    cat "$TEST_MCP_CONFIG"
else
    echo -e "${RED}✗ Config file was not created${NC}"
    exit 1
fi
echo ""

# Test 3: List servers (should show the added server)
echo -e "${YELLOW}Test 3: Listing servers (should show test-server)${NC}"
OUTPUT=$($BINARY mcp list-servers 2>&1)
echo "$OUTPUT"
if echo "$OUTPUT" | grep -q "test-server"; then
    echo -e "${GREEN}✓ Server found in list${NC}"
else
    echo -e "${RED}✗ Server not found in list${NC}"
    exit 1
fi
echo ""

# Test 4: Add a second server
echo -e "${YELLOW}Test 4: Adding second MCP server 'test-server-2'${NC}"
$BINARY mcp add-server --name test-server-2 --url http://localhost:9001 2>&1 | grep -q "Successfully added MCP server: test-server-2"
if [ $? -eq 0 ]; then
    echo -e "${GREEN}✓ Second server added successfully${NC}"
else
    echo -e "${RED}✗ Failed to add second server${NC}"
    exit 1
fi
echo ""

# Test 5: List servers (should show both servers)
echo -e "${YELLOW}Test 5: Listing servers (should show both servers)${NC}"
OUTPUT=$($BINARY mcp list-servers 2>&1)
echo "$OUTPUT"
if echo "$OUTPUT" | grep -q "test-server" && echo "$OUTPUT" | grep -q "test-server-2"; then
    echo -e "${GREEN}✓ Both servers found in list${NC}"
else
    echo -e "${RED}✗ Not all servers found in list${NC}"
    exit 1
fi
echo ""

# Test 6: Simulate restart - verify servers are loaded from config
echo -e "${YELLOW}Test 6: Simulating restart - listing servers again${NC}"
echo "(This simulates a restart because each command creates a new registry)"
OUTPUT=$($BINARY mcp list-servers 2>&1)
echo "$OUTPUT"
if echo "$OUTPUT" | grep -q "test-server" && echo "$OUTPUT" | grep -q "test-server-2"; then
    echo -e "${GREEN}✓ Servers persisted across 'restart'${NC}"
else
    echo -e "${RED}✗ Servers not persisted${NC}"
    exit 1
fi
echo ""

# Test 7: Remove first server
echo -e "${YELLOW}Test 7: Removing 'test-server'${NC}"
$BINARY mcp remove-server --name test-server 2>&1 | grep -q "Successfully removed MCP server: test-server"
if [ $? -eq 0 ]; then
    echo -e "${GREEN}✓ Server removed successfully${NC}"
else
    echo -e "${RED}✗ Failed to remove server${NC}"
    exit 1
fi
echo ""

# Test 8: List servers (should only show test-server-2)
echo -e "${YELLOW}Test 8: Listing servers (should only show test-server-2)${NC}"
OUTPUT=$($BINARY mcp list-servers 2>&1)
echo "$OUTPUT"
# Check that test-server-2 is present but test-server is not (must avoid matching test-server-2)
if echo "$OUTPUT" | grep -q "test-server-2" && ! echo "$OUTPUT" | grep -E "^test-server[^-]|test-server   "; then
    echo -e "${GREEN}✓ Removed server is not in list${NC}"
else
    echo -e "${RED}✗ Server still appears in list or test-server-2 missing${NC}"
    exit 1
fi
echo ""

# Test 9: Verify removal persisted
echo -e "${YELLOW}Test 9: Verifying removal persisted (list again)${NC}"
OUTPUT=$($BINARY mcp list-servers 2>&1)
echo "$OUTPUT"
# Check that test-server-2 is present but test-server is not (must avoid matching test-server-2)
if echo "$OUTPUT" | grep -q "test-server-2" && ! echo "$OUTPUT" | grep -E "^test-server[^-]|test-server   "; then
    echo -e "${GREEN}✓ Removal persisted${NC}"
else
    echo -e "${RED}✗ Removal did not persist${NC}"
    exit 1
fi
echo ""

# Test 10: Verify config file reflects the removal
echo -e "${YELLOW}Test 10: Verifying config file reflects removal${NC}"
if grep -q "test-server-2" "$TEST_MCP_CONFIG" && ! grep -q "\"test-server\":" "$TEST_MCP_CONFIG"; then
    echo -e "${GREEN}✓ Config file correctly updated${NC}"
    echo "Final config file contents:"
    cat "$TEST_MCP_CONFIG"
else
    echo -e "${RED}✗ Config file not correctly updated${NC}"
    cat "$TEST_MCP_CONFIG"
    exit 1
fi
echo ""

# Test 11: Try to remove non-existent server
echo -e "${YELLOW}Test 11: Attempting to remove non-existent server (should fail)${NC}"
if $BINARY mcp remove-server --name nonexistent 2>&1 | grep -q "not found"; then
    echo -e "${GREEN}✓ Correctly reported server not found${NC}"
else
    echo -e "${RED}✗ Did not properly handle non-existent server${NC}"
    exit 1
fi
echo ""

# Test 12: Remove last server
echo -e "${YELLOW}Test 12: Removing last server 'test-server-2'${NC}"
$BINARY mcp remove-server --name test-server-2 2>&1 | grep -q "Successfully removed MCP server: test-server-2"
if [ $? -eq 0 ]; then
    echo -e "${GREEN}✓ Last server removed successfully${NC}"
else
    echo -e "${RED}✗ Failed to remove last server${NC}"
    exit 1
fi
echo ""

# Test 13: List servers (should be empty)
echo -e "${YELLOW}Test 13: Listing servers (should be empty)${NC}"
OUTPUT=$($BINARY mcp list-servers 2>&1)
echo "$OUTPUT"
if echo "$OUTPUT" | grep -q "No MCP servers registered"; then
    echo -e "${GREEN}✓ Server list is empty${NC}"
else
    echo -e "${RED}✗ Server list is not empty${NC}"
    exit 1
fi
echo ""

# Test 14: Try to add duplicate server
echo -e "${YELLOW}Test 14: Adding server for duplicate test${NC}"
$BINARY mcp add-server --name duplicate-test --url http://localhost:9000 2>&1 | grep -q "Successfully added"
echo -e "${YELLOW}Test 14: Attempting to add duplicate server (should fail)${NC}"
if $BINARY mcp add-server --name duplicate-test --url http://localhost:9000 2>&1 | grep -q "already registered"; then
    echo -e "${GREEN}✓ Correctly prevented duplicate server${NC}"
else
    echo -e "${RED}✗ Did not prevent duplicate server${NC}"
    exit 1
fi
echo ""

echo -e "${GREEN}==================================================${NC}"
echo -e "${GREEN}All tests passed! Issue #120 is FIXED ✓${NC}"
echo -e "${GREEN}==================================================${NC}"
echo ""
echo "Summary:"
echo "  ✓ Servers are persisted to $TEST_MCP_CONFIG"
echo "  ✓ Servers are loaded on startup"
echo "  ✓ Server removal is persisted"
echo "  ✓ Duplicate servers are prevented"
echo "  ✓ Non-existent servers return proper errors"
