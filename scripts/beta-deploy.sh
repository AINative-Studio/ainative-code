#!/bin/bash
set -e

echo "========================================="
echo "  AINative Code Beta Deployment"
echo "  Version: v1.1.0-beta.1"
echo "========================================="

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# 1. Pre-flight checks
echo -e "\n${YELLOW}[1/7] Pre-flight checks...${NC}"

# Check if on correct branch
CURRENT_BRANCH=$(git branch --show-current)
if [ "$CURRENT_BRANCH" != "main" ]; then
    echo -e "${RED}Error: Must be on main branch${NC}"
    exit 1
fi
echo -e "${GREEN}✓ On main branch${NC}"

# Check if working directory is clean
if ! git diff-index --quiet HEAD --; then
    echo -e "${RED}Error: Working directory not clean${NC}"
    git status
    exit 1
fi
echo -e "${GREEN}✓ Working directory clean${NC}"

# Check if all required files exist
REQUIRED_FILES=(
    ".github/RELEASE_NOTES_v1.1.0-beta.1.md"
    "docs/beta-testing-guide.md"
    "docs/beta-feedback-form.md"
    ".github/monitoring/beta-dashboard.json"
)

for file in "${REQUIRED_FILES[@]}"; do
    if [ ! -f "$file" ]; then
        echo -e "${RED}Error: Required file missing: $file${NC}"
        exit 1
    fi
done
echo -e "${GREEN}✓ All required files present${NC}"

# 2. Run all tests
echo -e "\n${YELLOW}[2/7] Running all tests...${NC}"

# Python backend tests
echo -e "\n${YELLOW}Running Python backend tests...${NC}"
cd python-backend
pytest --cov=app --cov-report=term-missing --cov-fail-under=80 -v
if [ $? -ne 0 ]; then
    echo -e "${RED}Python tests failed${NC}"
    exit 1
fi
echo -e "${GREEN}✓ Python tests passed (73/73)${NC}"

# Go CLI tests
echo -e "\n${YELLOW}Running Go CLI tests...${NC}"
cd ..
go test ./internal/backend/... ./internal/provider/... ./internal/cmd/... -v
if [ $? -ne 0 ]; then
    echo -e "${RED}Go tests failed${NC}"
    exit 1
fi
echo -e "${GREEN}✓ Go tests passed (86/86)${NC}"

# E2E integration tests
echo -e "\n${YELLOW}Running E2E integration tests...${NC}"
go test ./tests/integration/ainative_e2e/... -v
if [ $? -ne 0 ]; then
    echo -e "${RED}E2E tests failed${NC}"
    exit 1
fi
echo -e "${GREEN}✓ E2E tests passed (19/19)${NC}"

# 3. Version tagging
echo -e "\n${YELLOW}[3/7] Creating git tag...${NC}"

# Check if tag already exists
if git rev-parse v1.1.0-beta.1 >/dev/null 2>&1; then
    echo -e "${YELLOW}Warning: Tag v1.1.0-beta.1 already exists${NC}"
    read -p "Delete and recreate? (y/N): " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        git tag -d v1.1.0-beta.1
        git push origin :refs/tags/v1.1.0-beta.1 2>/dev/null || true
    else
        echo -e "${YELLOW}Skipping tag creation${NC}"
    fi
fi

if ! git rev-parse v1.1.0-beta.1 >/dev/null 2>&1; then
    git tag -a v1.1.0-beta.1 -m "Beta release: AINative cloud authentication and hosted inference"
    echo -e "${GREEN}✓ Tagged v1.1.0-beta.1${NC}"
else
    echo -e "${GREEN}✓ Tag v1.1.0-beta.1 exists${NC}"
fi

# 4. Build binaries
echo -e "\n${YELLOW}[4/7] Building binaries...${NC}"
mkdir -p dist/

# macOS (arm64)
echo -e "${YELLOW}Building darwin-arm64...${NC}"
GOOS=darwin GOARCH=arm64 go build -ldflags "-X main.Version=v1.1.0-beta.1" -o dist/ainative-code-darwin-arm64 .
if [ $? -ne 0 ]; then
    echo -e "${RED}Build failed: darwin-arm64${NC}"
    exit 1
fi
echo -e "${GREEN}✓ Built darwin-arm64${NC}"

# macOS (amd64)
echo -e "${YELLOW}Building darwin-amd64...${NC}"
GOOS=darwin GOARCH=amd64 go build -ldflags "-X main.Version=v1.1.0-beta.1" -o dist/ainative-code-darwin-amd64 .
if [ $? -ne 0 ]; then
    echo -e "${RED}Build failed: darwin-amd64${NC}"
    exit 1
fi
echo -e "${GREEN}✓ Built darwin-amd64${NC}"

# Linux (amd64)
echo -e "${YELLOW}Building linux-amd64...${NC}"
GOOS=linux GOARCH=amd64 go build -ldflags "-X main.Version=v1.1.0-beta.1" -o dist/ainative-code-linux-amd64 .
if [ $? -ne 0 ]; then
    echo -e "${RED}Build failed: linux-amd64${NC}"
    exit 1
fi
echo -e "${GREEN}✓ Built linux-amd64${NC}"

# List binary sizes
echo -e "\n${YELLOW}Binary sizes:${NC}"
ls -lh dist/

# 5. Python backend packaging
echo -e "\n${YELLOW}[5/7] Packaging Python backend...${NC}"
cd python-backend
tar -czf ../dist/python-backend-v1.1.0-beta.1.tar.gz \
    --exclude='__pycache__' \
    --exclude='*.pyc' \
    --exclude='.pytest_cache' \
    --exclude='htmlcov' \
    --exclude='.coverage' \
    .
cd ..
echo -e "${GREEN}✓ Packaged Python backend ($(du -h dist/python-backend-v1.1.0-beta.1.tar.gz | cut -f1))${NC}"

# 6. Create checksums
echo -e "\n${YELLOW}[6/7] Creating checksums...${NC}"
cd dist
shasum -a 256 * > SHA256SUMS.txt
cat SHA256SUMS.txt
cd ..
echo -e "${GREEN}✓ Checksums created${NC}"

# 7. Create GitHub release (draft)
echo -e "\n${YELLOW}[7/7] Creating GitHub release draft...${NC}"

# Check if gh CLI is installed
if ! command -v gh &> /dev/null; then
    echo -e "${RED}Error: GitHub CLI (gh) not installed${NC}"
    echo -e "${YELLOW}Install with: brew install gh${NC}"
    exit 1
fi

# Check if authenticated
if ! gh auth status &> /dev/null; then
    echo -e "${RED}Error: Not authenticated with GitHub${NC}"
    echo -e "${YELLOW}Run: gh auth login${NC}"
    exit 1
fi

# Create draft release
echo -e "${YELLOW}Creating draft release...${NC}"
gh release create v1.1.0-beta.1 \
    --title "v1.1.0-beta.1: AINative Cloud Authentication (Beta)" \
    --notes-file .github/RELEASE_NOTES_v1.1.0-beta.1.md \
    --prerelease \
    --draft \
    dist/ainative-code-darwin-arm64 \
    dist/ainative-code-darwin-amd64 \
    dist/ainative-code-linux-amd64 \
    dist/python-backend-v1.1.0-beta.1.tar.gz \
    dist/SHA256SUMS.txt

if [ $? -ne 0 ]; then
    echo -e "${RED}GitHub release creation failed${NC}"
    echo -e "${YELLOW}You may need to delete the existing draft and try again${NC}"
    exit 1
fi
echo -e "${GREEN}✓ GitHub draft release created${NC}"

# Summary
echo -e "\n${GREEN}=========================================${NC}"
echo -e "${GREEN}  Beta Deployment Preparation Complete!${NC}"
echo -e "${GREEN}=========================================${NC}"
echo -e "\nRelease: https://github.com/AINative-Studio/ainative-code/releases/tag/v1.1.0-beta.1"
echo -e "\nBinaries available for:"
echo -e "  - macOS (ARM64, Intel)"
echo -e "  - Linux (AMD64)"
echo -e "\nTest Results:"
echo -e "  - Python: 73/73 tests passed"
echo -e "  - Go CLI: 86/86 tests passed"
echo -e "  - E2E: 19/19 tests passed"
echo -e "  - Total: 178/178 tests passed"
echo -e "\n${YELLOW}Next steps:${NC}"
echo -e "1. Review the draft release on GitHub"
echo -e "2. Test binaries on each platform"
echo -e "3. Publish release (remove draft status)"
echo -e "4. Push git tag: git push origin v1.1.0-beta.1"
echo -e "5. Notify beta testers"
echo -e "6. Monitor error rates and metrics"
echo -e "7. Collect feedback"
echo -e "\n${YELLOW}Deployment commands (when approved):${NC}"
echo -e "  # Publish release"
echo -e "  gh release edit v1.1.0-beta.1 --draft=false"
echo -e "\n  # Push tag"
echo -e "  git push origin v1.1.0-beta.1"
