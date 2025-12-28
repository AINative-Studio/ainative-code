# CI/CD Post-Setup Checklist

This checklist should be completed after TASK-001 (Go Module Init) and TASK-002 (Rebranding) are finished.

## Prerequisites Completion

- [x] TASK-001: Go module initialized (`go.mod` exists)
- [x] TASK-002: AINative Code rebranding complete
- [x] `cmd/ainative-code/main.go` exists with basic CLI structure

## GitHub Repository Setup

### 1. Initialize Git Repository

```bash
cd /Users/aideveloper/AINative-Code
git init
git add .
git commit -m "feat: initial commit with CI/CD pipeline

- Add GitHub Actions workflows (CI, Release, Dependency Updates)
- Add golangci-lint configuration
- Add Codecov configuration
- Add Dockerfile and .dockerignore
- Add Makefile with development commands
- Add comprehensive documentation
- Add build status badges to README

🤖 Generated with Claude Code
Co-Authored-By: Claude <noreply@anthropic.com>"
```

### 2. Create GitHub Repository

**Option A: Using GitHub CLI (recommended)**
```bash
gh repo create AINative-studio/ainative-code \
  --public \
  --description "AI-Native Development, Natively" \
  --homepage "https://docs.ainative.studio/code"

git remote add origin https://github.com/AINative-studio/ainative-code.git
git branch -M main
git push -u origin main
```

**Option B: Using GitHub Web UI**
1. Go to https://github.com/new
2. Owner: `AINative-studio`
3. Repository name: `ainative-code`
4. Description: `AI-Native Development, Natively`
5. Visibility: Public
6. Click "Create repository"

```bash
git remote add origin https://github.com/AINative-studio/ainative-code.git
git branch -M main
git push -u origin main
```

- [ ] GitHub repository created
- [ ] Initial commit pushed

### 3. Configure Repository Settings

#### General Settings
- [ ] Set repository description: "AI-Native Development, Natively"
- [ ] Set website: https://docs.ainative.studio/code
- [ ] Add topics: `go`, `cli`, `ai`, `tui`, `bubble-tea`, `anthropic`, `openai`

#### Features
- [ ] Enable Issues
- [ ] Enable Projects
- [ ] Enable Discussions
- [ ] Enable Wikis (optional)
- [ ] Disable Allow merge commits (use squash or rebase)
- [ ] Enable Automatically delete head branches

#### Security
- [ ] Enable Dependabot alerts
- [ ] Enable Dependabot security updates
- [ ] Enable Code scanning (CodeQL)
- [ ] Enable Secret scanning

## Branch Protection

### Configure Main Branch Protection

Go to: Settings → Branches → Add rule

**Branch name pattern**: `main`

- [ ] Require a pull request before merging
  - [ ] Require approvals: 1
  - [ ] Dismiss stale pull request approvals when new commits are pushed
  - [ ] Require review from Code Owners
- [ ] Require status checks to pass before merging
  - [ ] Require branches to be up to date before merging
  - [ ] Required status checks:
    - [ ] `lint`
    - [ ] `test (ubuntu-latest, 1.21)`
    - [ ] `build (ubuntu-latest, linux, amd64, ainative-code-linux-amd64)`
- [ ] Require conversation resolution before merging
- [ ] Require signed commits (recommended)
- [ ] Require linear history
- [ ] Include administrators
- [ ] Allow force pushes: No
- [ ] Allow deletions: No

## Codecov Integration

### 1. Sign up for Codecov

1. Go to https://codecov.io
2. Sign in with GitHub
3. Add repository: `AINative-studio/ainative-code`

- [ ] Codecov account created
- [ ] Repository added to Codecov

### 2. Configure Codecov Token

1. Copy the `CODECOV_TOKEN` from Codecov dashboard
2. Go to GitHub repository → Settings → Secrets and variables → Actions
3. Click "New repository secret"
4. Name: `CODECOV_TOKEN`
5. Value: Paste the token
6. Click "Add secret"

- [ ] `CODECOV_TOKEN` added to GitHub secrets

### 3. Verify Codecov Configuration

- [ ] `.codecov.yml` exists in repository root
- [ ] Coverage threshold set to 80%
- [ ] Component tracking configured

## GitHub Actions Verification

### 1. Check Workflows

Go to: Actions tab

- [ ] CI workflow appears in workflow list
- [ ] Release workflow appears in workflow list
- [ ] Dependency Updates workflow appears in workflow list

### 2. Test CI Pipeline

**Create a test PR to verify workflows run correctly:**

```bash
# Create test branch
git checkout -b test/ci-verification

# Make a small change
echo "# CI Test" >> .github/CI-TEST.md
git add .github/CI-TEST.md
git commit -m "test: verify CI pipeline"
git push origin test/ci-verification

# Create PR
gh pr create \
  --title "test: Verify CI/CD Pipeline" \
  --body "Testing GitHub Actions workflows

- [ ] Lint job completes successfully
- [ ] Test job completes on all platforms
- [ ] Build job creates artifacts
- [ ] Security scan completes
- [ ] Coverage report uploads to Codecov"
```

**Verify**:
- [ ] All CI workflow jobs complete successfully
- [ ] Status checks appear on PR
- [ ] Coverage report posted by Codecov bot
- [ ] No errors in workflow logs

**After verification**:
```bash
# Merge or close the test PR
gh pr merge --squash  # or: gh pr close
git checkout main
git pull
git branch -d test/ci-verification
```

## Docker Registry Setup

### GitHub Container Registry (GHCR)

GHCR is automatically configured for public repositories.

**Verify**:
- [ ] Repository has `packages: write` permission in workflows
- [ ] `GITHUB_TOKEN` has package write access

**First release will automatically publish Docker images to:**
`ghcr.io/ainative-studio/ainative-code`

## Release Testing

### 1. Create Pre-release for Testing

```bash
# Create a pre-release tag
git tag -a v0.1.0-beta.1 -m "Pre-release for testing

Testing:
- GitHub Actions release workflow
- Multi-platform builds
- Docker image publishing
- Artifact uploads"

git push origin v0.1.0-beta.1
```

**Verify**:
- [ ] Release workflow triggered automatically
- [ ] All platform builds complete successfully
- [ ] Release created on GitHub with:
  - [ ] Changelog generated
  - [ ] 5 binary files uploaded
  - [ ] Compressed archives (tar.gz/zip)
  - [ ] SHA256 checksum files
  - [ ] Marked as pre-release
- [ ] Docker images published to ghcr.io with tags:
  - [ ] `v0.1.0-beta.1`
  - [ ] `v0.1`
  - [ ] `v0`

### 2. Test Release Artifacts

**Download and test a binary:**
```bash
# macOS Apple Silicon example
curl -LO https://github.com/AINative-studio/ainative-code/releases/download/v0.1.0-beta.1/ainative-code-darwin-arm64
chmod +x ainative-code-darwin-arm64
./ainative-code-darwin-arm64 version

# Verify checksum
curl -LO https://github.com/AINative-studio/ainative-code/releases/download/v0.1.0-beta.1/ainative-code-darwin-arm64.sha256
shasum -a 256 -c ainative-code-darwin-arm64.sha256
```

- [ ] Binary downloads successfully
- [ ] Binary executes correctly
- [ ] Checksum verification passes

**Test Docker image:**
```bash
docker pull ghcr.io/ainative-studio/ainative-code:v0.1.0-beta.1
docker run --rm ghcr.io/ainative-studio/ainative-code:v0.1.0-beta.1 version
```

- [ ] Docker image pulls successfully
- [ ] Container runs correctly

## Documentation Updates

### 1. Update README Badges

The badges are already in README.md. Verify they work after first push:

- [ ] CI badge shows status
- [ ] Release badge shows latest version
- [ ] Codecov badge shows coverage percentage
- [ ] Go Report Card badge shows grade

### 2. Add CODEOWNERS File

Create `.github/CODEOWNERS`:
```bash
cat > .github/CODEOWNERS << 'EOF'
# CODEOWNERS file
# Users mentioned here will be automatically requested for review

# Default owners for everything
* @your-github-username

# CI/CD workflows
/.github/workflows/ @devops-team
/Dockerfile @devops-team
/.dockerignore @devops-team

# Documentation
/docs/ @documentation-team
/*.md @documentation-team
EOF

git add .github/CODEOWNERS
git commit -m "chore: add CODEOWNERS file"
git push
```

- [ ] CODEOWNERS file created and pushed

### 3. Add Issue Templates

Create issue templates in `.github/ISSUE_TEMPLATE/`:

**Bug Report** (`.github/ISSUE_TEMPLATE/bug_report.yml`)
**Feature Request** (`.github/ISSUE_TEMPLATE/feature_request.yml`)
**Documentation** (`.github/ISSUE_TEMPLATE/documentation.yml`)

- [ ] Issue templates created

### 4. Add Pull Request Template

Create `.github/PULL_REQUEST_TEMPLATE.md`

- [ ] PR template created

## Local Development Setup

### 1. Install Development Tools

```bash
# golangci-lint
curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin

# gosec
go install github.com/securego/gosec/v2/cmd/gosec@latest

# govulncheck
go install golang.org/x/vuln/cmd/govulncheck@latest

# go-licenses (optional)
go install github.com/google/go-licenses@latest

# go-mod-outdated (optional)
go install github.com/psampaz/go-mod-outdated@latest
```

- [ ] golangci-lint installed
- [ ] gosec installed
- [ ] govulncheck installed

### 2. Test Local Development Workflow

```bash
# Test Makefile commands
make help           # Show all commands
make deps           # Download dependencies
make build          # Build binary
make test           # Run tests
make lint           # Run linter
make ci             # Run full CI locally
```

- [ ] All Makefile commands work correctly
- [ ] Local CI passes

## Dependency Updates Configuration

The dependency update workflow runs weekly on Mondays at 9 AM UTC.

**Manual trigger for testing:**
```bash
gh workflow run dependency-updates.yml
```

**Verify**:
- [ ] Workflow completes successfully
- [ ] Creates PR if updates available
- [ ] PR includes changelog
- [ ] Automated tests run on PR

## Monitoring Setup

### 1. Enable Notifications

- [ ] Watch repository for releases
- [ ] Enable notifications for failed workflow runs
- [ ] Subscribe to Dependabot alerts

### 2. Set Up Status Dashboard (Optional)

Tools to consider:
- GitHub Actions dashboard
- Codecov dashboard
- Dependabot dashboard

## Security Hardening

### 1. Review Security Settings

- [ ] Two-factor authentication enabled for all team members
- [ ] SSH keys or GPG signing configured
- [ ] Review team access levels

### 2. Configure Secret Scanning

- [ ] Custom patterns added (if needed)
- [ ] Webhooks configured for alerts

### 3. Review Security Advisories

- [ ] Subscribe to GitHub Security Advisories
- [ ] Configure notification preferences

## Final Verification

### Pre-Production Checklist

- [ ] All CI workflows passing
- [ ] Code coverage above 80%
- [ ] No security vulnerabilities
- [ ] No linter warnings
- [ ] Documentation complete and accurate
- [ ] README badges working
- [ ] Docker images building and running
- [ ] Release workflow tested
- [ ] Team has appropriate access
- [ ] Branch protection enforced

### Production Release

When ready for v1.0.0:

```bash
git checkout main
git pull

# Create release tag
git tag -a v1.0.0 -m "Release v1.0.0

Initial production release of AINative Code.

Features:
- Multi-LLM provider support
- AINative platform integration
- Beautiful TUI interface
- Cross-platform support

🤖 Generated with Claude Code
Co-Authored-By: Claude <noreply@anthropic.com>"

git push origin v1.0.0
```

- [ ] v1.0.0 release created successfully
- [ ] All artifacts uploaded
- [ ] Docker images published
- [ ] Release announcement prepared

## Maintenance Schedule

### Weekly
- [ ] Review Dependabot PRs
- [ ] Check for security advisories
- [ ] Review failed workflow runs

### Monthly
- [ ] Review and update dependencies
- [ ] Review CI/CD performance metrics
- [ ] Update documentation as needed

### Quarterly
- [ ] Review and update linter rules
- [ ] Update Go version if new stable release
- [ ] Review and optimize workflow performance
- [ ] Security audit

## Support Resources

- **GitHub Actions Docs**: https://docs.github.com/en/actions
- **Codecov Docs**: https://docs.codecov.com/
- **golangci-lint Docs**: https://golangci-lint.run/
- **Docker Docs**: https://docs.docker.com/
- **Internal Docs**: `/Users/aideveloper/AINative-Code/docs/CI-CD.md`

## Troubleshooting

Common issues and solutions documented in:
- `docs/CI-CD.md` - Troubleshooting section
- `.github/WORKFLOWS-REFERENCE.md` - Quick reference

## Completion

- [ ] All checklist items completed
- [ ] Team trained on CI/CD workflow
- [ ] Documentation reviewed and approved
- [ ] Production release successful

---

**Checklist Last Updated**: 2024-12-27
**Prepared By**: Claude (DevOps Specialist)
**Review Required**: Yes
