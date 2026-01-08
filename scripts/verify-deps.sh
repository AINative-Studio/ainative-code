#!/bin/bash
# Dependency Verification Script for AINative Code

echo "=== AINative Code Dependency Verification ==="
echo ""

# Check Go installation
echo "1. Checking Go installation..."
go version
echo ""

# Check SQLC installation
echo "2. Checking SQLC installation..."
if [ -f ~/go/bin/sqlc ]; then
    ~/go/bin/sqlc version
else
    echo "ERROR: SQLC not found at ~/go/bin/sqlc"
fi
echo ""

# Verify go.mod exists
echo "3. Checking go.mod..."
if [ -f go.mod ]; then
    echo "go.mod exists"
    grep "^module" go.mod
else
    echo "ERROR: go.mod not found"
fi
echo ""

# List core dependencies
echo "4. Core dependencies:"
go list -m all | grep -E "(bubbletea|cobra|viper|jwt|sqlite3|resty)" || echo "No core dependencies found"
echo ""

# Check for required dependencies
echo "5. Verifying required dependencies..."
REQUIRED_DEPS=(
    "github.com/charmbracelet/bubbletea"
    "github.com/spf13/cobra"
    "github.com/spf13/viper"
    "github.com/golang-jwt/jwt/v5"
    "github.com/mattn/go-sqlite3"
    "github.com/go-resty/resty/v2"
)

ALL_FOUND=true
for dep in "${REQUIRED_DEPS[@]}"; do
    if go list -m all | grep -q "$dep"; then
        echo "✓ $dep"
    else
        echo "✗ $dep - MISSING"
        ALL_FOUND=false
    fi
done
echo ""

# Summary
if [ "$ALL_FOUND" = true ]; then
    echo "=== All required dependencies are installed! ==="
else
    echo "=== WARNING: Some dependencies are missing! ==="
fi
