#!/bin/bash
# Demonstration script for Issue #128 fix
# Shows how the fix enables clean JSON piping

set -e

# Colors
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m'

BINARY="./build/ainative-code"

echo -e "${BLUE}=========================================${NC}"
echo -e "${BLUE}Issue #128 Fix Demonstration${NC}"
echo -e "${BLUE}=========================================${NC}"
echo ""

echo -e "${YELLOW}This demonstrates that JSON commands now work cleanly with jq${NC}"
echo ""

# Demo 1: Basic jq parsing
echo -e "${BLUE}Demo 1: Parse JSON with jq${NC}"
echo -e "Command: ${GREEN}session list --limit 1 --json | jq '.'${NC}"
echo ""
$BINARY session list --limit 1 --json 2>/dev/null | jq '.' | head -10
echo "..."
echo ""

# Demo 2: Extract specific field
echo -e "${BLUE}Demo 2: Extract session ID${NC}"
echo -e "Command: ${GREEN}session list --limit 1 --json | jq '.[0].id'${NC}"
echo ""
SESSION_ID=$($BINARY session list --limit 1 --json 2>/dev/null | jq -r '.[0].id')
echo "Session ID: $SESSION_ID"
echo ""

# Demo 3: Extract multiple fields
echo -e "${BLUE}Demo 3: Extract multiple fields${NC}"
echo -e "Command: ${GREEN}session list --limit 1 --json | jq '.[0] | {id, name}'${NC}"
echo ""
$BINARY session list --limit 1 --json 2>/dev/null | jq '.[0] | {id, name}'
echo ""

# Demo 4: Search and extract
echo -e "${BLUE}Demo 4: Search and extract query${NC}"
echo -e "Command: ${GREEN}session search 'test' --json | jq '.query'${NC}"
echo ""
QUERY=$($BINARY session search "test" --json 2>/dev/null | jq -r '.query')
echo "Query: $QUERY"
echo ""

# Demo 5: Show that stderr is separate
echo -e "${BLUE}Demo 5: Verify stdout/stderr separation${NC}"
echo ""
echo "Capturing stdout only (should be valid JSON):"
STDOUT=$($BINARY session list --limit 1 --json 2>/dev/null)
echo "$STDOUT" | jq '.[0].id'
echo ""

echo -e "${GREEN}=========================================${NC}"
echo -e "${GREEN}All demos completed successfully!${NC}"
echo -e "${GREEN}=========================================${NC}"
echo ""
echo "Key Takeaway:"
echo "  ✓ JSON output is clean (no log pollution)"
echo "  ✓ jq can parse output directly"
echo "  ✓ Complex jq queries work perfectly"
echo "  ✓ Logs go to stderr (won't interfere with pipes)"
echo ""
