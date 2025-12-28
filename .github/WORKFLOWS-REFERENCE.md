# GitHub Actions Workflows - Quick Reference

## Available Workflows

### 1. CI (Continuous Integration)
**File**: `.github/workflows/ci.yml`
**Triggers**: Push to main/develop, Pull Requests

**Jobs**:
- ✅ **Lint**: Code quality checks with golangci-lint
- ✅ **Test**: Unit tests on Ubuntu, macOS, Windows (Go 1.21, 1.22)
- ✅ **Integration Test**: Tagged integration tests
- ✅ **Build**: Multi-platform binary builds
- ✅ **Security Scan**: gosec security analysis
- ✅ **Dependency Check**: Vulnerability scanning

**Manual Trigger**:
```bash
gh workflow run ci.yml
```

### 2. Release
**File**: `.github/workflows/release.yml`
**Triggers**: Git tags matching `v*.*.*`

**Jobs**:
- ✅ **Create Release**: Generate GitHub release with changelog
- ✅ **Build and Upload**: Multi-platform binaries + checksums
- ✅ **Docker Publish**: Push images to ghcr.io

**Manual Trigger**:
```bash
# Create and push tag
git tag -a v1.0.0 -m "Release v1.0.0"
git push origin v1.0.0

# Or trigger manually
gh workflow run release.yml -f tag=v1.0.0
```

### 3. Dependency Updates
**File**: `.github/workflows/dependency-updates.yml`
**Triggers**: Weekly (Mondays 9 AM UTC), Manual

**Jobs**:
- ✅ **Update Dependencies**: Auto-update and create PR
- ✅ **Check Outdated**: Report outdated packages
- ✅ **Check Licenses**: Verify license compliance

**Manual Trigger**:
```bash
gh workflow run dependency-updates.yml
```

## Build Matrix

### Platforms
- darwin/amd64 (macOS Intel)
- darwin/arm64 (macOS Apple Silicon)
- linux/amd64 (Linux x64)
- linux/arm64 (Linux ARM)
- windows/amd64 (Windows x64)

### Go Versions
- 1.21 (minimum required)
- 1.22 (latest stable)

## Coverage Requirements

- **Minimum**: 80%
- **Measured on**: Ubuntu + Go 1.21
- **Tool**: Codecov

## Local Development Commands

```bash
# Run full CI locally
make ci

# Individual checks
make lint              # Run linter
make test              # Run tests
make test-coverage     # Generate coverage report
make build-all         # Build for all platforms

# Docker
make docker-build      # Build Docker image
make docker-run        # Run container

# Pre-commit checks
make pre-commit        # Format, vet, lint, test
```

## Secrets Required

| Secret | Purpose | Required For |
|--------|---------|--------------|
| `CODECOV_TOKEN` | Upload coverage | CI workflow |
| `GITHUB_TOKEN` | Auto-provided | All workflows |

## Typical Workflow Times

- **CI**: 5-8 minutes
- **Release**: 8-12 minutes
- **Dependency Updates**: 2-3 minutes

## Status Checks

Required for PR merge:
- ✅ lint
- ✅ test (ubuntu-latest, 1.21)
- ✅ build (ubuntu-latest, linux, amd64)

## Release Artifacts

Each release includes:
- ✅ 5 platform-specific binaries
- ✅ Compressed archives (tar.gz/zip)
- ✅ SHA256 checksums
- ✅ Docker images (linux/amd64, linux/arm64)
- ✅ Auto-generated changelog

## Docker Images

**Registry**: ghcr.io/ainative-studio/ainative-code

**Tags**:
- `latest` - Latest release
- `v1.2.3` - Specific version
- `v1.2` - Minor version
- `v1` - Major version

**Pull**:
```bash
docker pull ghcr.io/ainative-studio/ainative-code:latest
```

## Troubleshooting

### Failed CI?
```bash
# Check locally
make ci

# View workflow logs
gh run list
gh run view <run-id>
```

### Coverage Below Threshold?
```bash
# Check coverage
make test-coverage

# View report
open coverage.html
```

### Lint Failures?
```bash
# Run locally
make lint

# Auto-fix
make fmt
```

## Documentation

- [Complete CI/CD Guide](../docs/CI-CD.md)
- [Codecov Configuration](../.codecov.yml)
- [Linter Configuration](../.golangci.yml)
- [Makefile](../Makefile)

---

**Last Updated**: 2024-12-27
