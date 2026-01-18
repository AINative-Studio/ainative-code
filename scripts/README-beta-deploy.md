# Beta Deployment Script

This directory contains the automated beta deployment script for AINative Code v1.1.0-beta.1.

## Quick Start

```bash
# Navigate to project root
cd /Users/aideveloper/AINative-Code

# Run the deployment script
./scripts/beta-deploy.sh
```

## What It Does

The `beta-deploy.sh` script automates the entire beta release process:

1. **Pre-flight Checks**
   - Verifies you're on the main branch
   - Ensures working directory is clean
   - Checks all required files exist

2. **Test Execution**
   - Runs Python backend tests (73 tests, 80% coverage target)
   - Runs Go CLI tests (86 tests)
   - Runs E2E integration tests (19 tests)

3. **Version Tagging**
   - Creates git tag: `v1.1.0-beta.1`
   - Includes release message

4. **Binary Building**
   - macOS ARM64 (Apple Silicon)
   - macOS AMD64 (Intel)
   - Linux AMD64

5. **Packaging**
   - Creates tarball of Python backend
   - Generates SHA256 checksums

6. **GitHub Release**
   - Creates draft release on GitHub
   - Uploads all binaries and packages
   - Uses release notes from `.github/RELEASE_NOTES_v1.1.0-beta.1.md`

7. **Summary**
   - Displays deployment summary
   - Lists next steps

## Requirements

### Tools
- `git` - Version control
- `go` - Go compiler (1.21+)
- `pytest` - Python testing framework
- `gh` - GitHub CLI tool

### GitHub CLI Setup
```bash
# Install GitHub CLI
brew install gh

# Authenticate
gh auth login
```

### Python Environment
```bash
# Navigate to python-backend
cd python-backend

# Install dependencies
pip install -r requirements.txt

# Install dev dependencies
pip install -r requirements-dev.txt
```

## Pre-Deployment Checklist

Before running the deployment script, ensure:

- [ ] All code changes committed
- [ ] Working directory is clean (`git status`)
- [ ] On main branch (`git branch --show-current`)
- [ ] All tests passing locally
- [ ] Release notes reviewed and finalized
- [ ] Beta testing guide reviewed
- [ ] Monitoring dashboard configured

## Running the Deployment

### Dry Run (Recommended First)
```bash
# Review the script without executing
cat ./scripts/beta-deploy.sh

# Validate syntax
bash -n ./scripts/beta-deploy.sh
```

### Full Deployment
```bash
# Execute the deployment
./scripts/beta-deploy.sh
```

### Expected Output
```
=========================================
  AINative Code Beta Deployment
  Version: v1.1.0-beta.1
=========================================

[1/7] Pre-flight checks...
✓ On main branch
✓ Working directory clean
✓ All required files present

[2/7] Running all tests...
Running Python backend tests...
✓ Python tests passed (73/73)

Running Go CLI tests...
✓ Go tests passed (86/86)

Running E2E integration tests...
✓ E2E tests passed (19/19)

[3/7] Creating git tag...
✓ Tagged v1.1.0-beta.1

[4/7] Building binaries...
✓ Built darwin-arm64
✓ Built darwin-amd64
✓ Built linux-amd64

[5/7] Packaging Python backend...
✓ Packaged Python backend (2.5M)

[6/7] Creating checksums...
✓ Checksums created

[7/7] Creating GitHub release draft...
✓ GitHub draft release created

=========================================
  Beta Deployment Preparation Complete!
=========================================

Release: https://github.com/AINative-Studio/ainative-code/releases/tag/v1.1.0-beta.1

Binaries available for:
  - macOS (ARM64, Intel)
  - Linux (AMD64)

Test Results:
  - Python: 73/73 tests passed
  - Go CLI: 86/86 tests passed
  - E2E: 19/19 tests passed
  - Total: 178/178 tests passed

Next steps:
1. Review the draft release on GitHub
2. Test binaries on each platform
3. Publish release (remove draft status)
4. Push git tag: git push origin v1.1.0-beta.1
5. Notify beta testers
6. Monitor error rates and metrics
7. Collect feedback
```

## Post-Deployment Steps

### 1. Review Draft Release
```bash
# Open the draft release in browser
gh release view v1.1.0-beta.1 --web
```

### 2. Test Binaries
```bash
# Download and test each binary
cd dist/

# Test macOS ARM64
./ainative-code-darwin-arm64 --version

# Test macOS AMD64
./ainative-code-darwin-amd64 --version

# Test Linux AMD64 (if on Linux or using Docker)
./ainative-code-linux-amd64 --version
```

### 3. Verify Checksums
```bash
# Verify SHA256 checksums
cd dist/
shasum -a 256 -c SHA256SUMS.txt
```

### 4. Publish Release
```bash
# Remove draft status
gh release edit v1.1.0-beta.1 --draft=false
```

### 5. Push Git Tag
```bash
# Push the tag to remote
git push origin v1.1.0-beta.1
```

### 6. Notify Beta Testers
Send email to beta participants with:
- Link to release: https://github.com/AINative-Studio/ainative-code/releases/tag/v1.1.0-beta.1
- Link to beta testing guide: docs/beta-testing-guide.md
- Support contact: beta@ainative.studio

### 7. Monitor Metrics
Access the monitoring dashboard and watch:
- Error rates (target: <1%)
- Auth success (target: >95%)
- Chat completion success (target: >98%)
- P95 latency (target: <2s)

## Troubleshooting

### Script Fails at Pre-flight Checks

**Problem:** "Must be on main branch"
```bash
# Solution: Switch to main branch
git checkout main
```

**Problem:** "Working directory not clean"
```bash
# Solution: Commit or stash changes
git status
git add .
git commit -m "Prepare for beta release"
```

### Script Fails at Tests

**Problem:** Python tests failing
```bash
# Solution: Run tests manually to debug
cd python-backend
pytest -v

# Fix any failing tests
# Then re-run deployment script
```

**Problem:** Go tests failing
```bash
# Solution: Run tests manually to debug
go test ./internal/... -v

# Fix any failing tests
# Then re-run deployment script
```

### Script Fails at Binary Building

**Problem:** Build fails for specific platform
```bash
# Solution: Check Go installation and cross-compilation support
go version
go env GOOS GOARCH

# Try building manually
GOOS=darwin GOARCH=arm64 go build -o dist/test .
```

### Script Fails at GitHub Release

**Problem:** "GitHub CLI (gh) not installed"
```bash
# Solution: Install GitHub CLI
brew install gh
```

**Problem:** "Not authenticated with GitHub"
```bash
# Solution: Authenticate with GitHub
gh auth login
```

**Problem:** "Release already exists"
```bash
# Solution: Delete existing release or tag
gh release delete v1.1.0-beta.1
git tag -d v1.1.0-beta.1
git push origin :refs/tags/v1.1.0-beta.1
```

## Rollback Procedure

If you need to rollback the beta release:

### 1. Unpublish Release
```bash
# Make the release a draft again
gh release edit v1.1.0-beta.1 --draft=true

# Or delete the release entirely
gh release delete v1.1.0-beta.1
```

### 2. Delete Git Tag
```bash
# Delete local tag
git tag -d v1.1.0-beta.1

# Delete remote tag
git push origin :refs/tags/v1.1.0-beta.1
```

### 3. Notify Beta Testers
Send email explaining the rollback and timeline for re-release.

## Related Documentation

- [Beta Release Notes](.github/RELEASE_NOTES_v1.1.0-beta.1.md)
- [Beta Testing Guide](../docs/beta-testing-guide.md)
- [Beta Feedback Form](../docs/beta-feedback-form.md)
- [Monitoring Dashboard](.github/monitoring/beta-dashboard.json)
- [Completion Report](../docs/ISSUE_152_BETA_RELEASE_COMPLETION_REPORT.md)

## Support

For questions or issues with the deployment script:
- GitHub Issues: https://github.com/AINative-Studio/ainative-code/issues
- Email: beta@ainative.studio
- Slack: #beta-alerts

---

**Last Updated:** 2026-01-17
**Version:** v1.1.0-beta.1
