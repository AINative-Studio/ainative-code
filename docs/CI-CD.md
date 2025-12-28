# CI/CD Pipeline Documentation

## Overview

AINative Code uses GitHub Actions for continuous integration and continuous deployment. The pipeline is designed to ensure code quality, security, and reliable releases across multiple platforms.

## Workflows

### 1. CI Workflow (`ci.yml`)

**Trigger**: Push to `main` or `develop` branches, pull requests, manual dispatch

**Jobs**:

#### Lint
- Runs `golangci-lint` with comprehensive rule set
- Configuration: `.golangci.yml`
- Timeout: 5 minutes

#### Test
- **Matrix Strategy**:
  - OS: Ubuntu, macOS, Windows
  - Go versions: 1.21, 1.22
- Runs unit tests with race detection
- Generates coverage reports
- **Coverage Threshold**: 80% minimum
- Uploads coverage to Codecov (Ubuntu + Go 1.21 only)

#### Integration Tests
- Runs after lint and test jobs pass
- Executes tests tagged with `integration`
- Location: `./tests/` directory

#### Build
- **Matrix Strategy**: Builds for multiple platforms
  - macOS: amd64, arm64
  - Linux: amd64, arm64
  - Windows: amd64
- Creates build artifacts
- Artifacts retained for 7 days

#### Security Scan
- Runs `gosec` security scanner
- Uploads SARIF report to GitHub Security
- Does not fail the build (informational)

#### Dependency Check
- Runs `govulncheck` for vulnerability scanning
- Checks against Go vulnerability database

### 2. Release Workflow (`release.yml`)

**Trigger**: Push of tags matching `v*.*.*`, manual dispatch

**Jobs**:

#### Create Release
- Generates changelog from git commits
- Creates GitHub release with installation instructions
- Marks pre-release for versions with `-` (e.g., `v1.0.0-beta`)

#### Build and Upload Assets
- **Matrix Strategy**: Same as CI build job
- Builds optimized binaries with version information
- Compresses binaries (tar.gz for Unix, zip for Windows)
- Generates SHA256 checksums
- Uploads all artifacts to GitHub Release

#### Docker Publish
- Builds multi-platform Docker images (linux/amd64, linux/arm64)
- Publishes to GitHub Container Registry (ghcr.io)
- **Tags**:
  - Semantic version: `v1.2.3`
  - Major.minor: `v1.2`
  - Major: `v1`
  - Latest: `latest` (only for default branch)

### 3. Dependency Updates Workflow (`dependency-updates.yml`)

**Trigger**: Weekly on Mondays at 9:00 AM UTC, manual dispatch

**Jobs**:

#### Update Dependencies
- Updates all Go dependencies to latest versions
- Runs `go mod tidy`
- Checks for vulnerabilities
- Creates PR automatically if changes detected

#### Check Outdated
- Reports outdated dependencies
- Uses `go-mod-outdated` tool

#### Check Licenses
- Scans dependency licenses
- Uses `go-licenses` tool
- Ensures compliance

## Configuration Files

### `.golangci.yml`

Comprehensive linting configuration with:
- 40+ enabled linters
- Custom rules for code quality
- Security checks
- Performance optimizations
- Test-specific exclusions

**Key Linters**:
- `errcheck`: Unchecked errors
- `gosec`: Security issues
- `govet`: Code correctness
- `staticcheck`: Advanced static analysis
- `revive`: Style and best practices
- `gocritic`: Comprehensive checks

### `.codecov.yml`

Code coverage configuration:
- **Project coverage**: 80% target
- **Patch coverage**: 80% target
- **Component tracking**: Auth, LLM, TUI, Config, API, Database
- **Exclusions**: Tests, generated code, examples, docs

### `Dockerfile`

Multi-stage build for optimal image size:
- **Build stage**: Go 1.21 Alpine
- **Runtime stage**: Alpine latest
- **User**: Non-root user (ainative:1000)
- **Size**: ~20MB (after optimization)
- **Security**: No root privileges, minimal dependencies

### `Makefile`

Comprehensive development commands:
- `make build`: Build for current platform
- `make build-all`: Build for all platforms
- `make test`: Run unit tests
- `make test-coverage`: Generate coverage report
- `make lint`: Run linter
- `make ci`: Run full CI locally
- `make release`: Create release artifacts

## Secrets Required

Configure these secrets in GitHub repository settings:

| Secret | Description | Required For |
|--------|-------------|--------------|
| `CODECOV_TOKEN` | Codecov upload token | Coverage reporting |
| `GITHUB_TOKEN` | Automatic (provided by GitHub) | All workflows |

## Branch Protection Rules

Recommended settings for `main` branch:

- ✅ Require pull request before merging
- ✅ Require approvals: 1
- ✅ Require status checks to pass:
  - `lint`
  - `test (ubuntu-latest, 1.21)`
  - `build (ubuntu-latest, linux, amd64, ainative-code-linux-amd64)`
- ✅ Require branches to be up to date
- ✅ Include administrators
- ✅ Allow force pushes: No
- ✅ Allow deletions: No

## Release Process

### Automated Release (Recommended)

1. Create and push a tag:
   ```bash
   git tag -a v1.0.0 -m "Release v1.0.0"
   git push origin v1.0.0
   ```

2. GitHub Actions automatically:
   - Creates GitHub release
   - Builds binaries for all platforms
   - Uploads artifacts
   - Publishes Docker images
   - Generates changelog

### Manual Release

Use workflow dispatch in GitHub Actions UI:
1. Go to Actions → Release
2. Click "Run workflow"
3. Enter tag name (e.g., `v1.0.0`)
4. Click "Run workflow"

## Local Development

### Run CI Checks Locally

```bash
# Format code
make fmt

# Run linter
make lint

# Run tests with coverage
make test-coverage

# Run full CI suite
make ci
```

### Build for All Platforms

```bash
make build-all
```

### Test Docker Build

```bash
make docker-build
make docker-run
```

## Performance Optimizations

### Build Caching
- Go module cache enabled in all workflows
- Docker layer caching using GitHub Actions cache

### Matrix Builds
- Parallel execution across multiple OS and architectures
- Typical build time: 5-8 minutes

### Artifact Management
- Build artifacts retained for 7 days
- Release artifacts permanent
- Docker images use multi-stage builds

## Monitoring and Debugging

### View Workflow Runs
1. Navigate to Actions tab
2. Select workflow
3. View run details and logs

### Download Build Artifacts
1. Go to completed workflow run
2. Scroll to "Artifacts" section
3. Download desired artifact

### Debug Failed Builds

Enable debug logging:
```bash
# Set repository secret
ACTIONS_STEP_DEBUG=true
ACTIONS_RUNNER_DEBUG=true
```

## Security Best Practices

### Code Scanning
- **gosec**: Security-focused Go linter
- **govulncheck**: Vulnerability scanning
- SARIF reports uploaded to GitHub Security

### Container Security
- Non-root user in Docker images
- Minimal base images (Alpine)
- No secrets in images
- Regular base image updates

### Dependency Management
- Weekly automated dependency updates
- License compliance checking
- Vulnerability scanning on every PR

## Troubleshooting

### Coverage Below Threshold

```bash
# Check coverage locally
make test-coverage

# View HTML report
open coverage.html
```

### Linter Failures

```bash
# Run linter locally
make lint

# Auto-fix issues
make fmt
```

### Build Failures

```bash
# Clean and rebuild
make clean
make build
```

### Docker Build Issues

```bash
# Build locally with detailed output
docker build --progress=plain -t ainative-code:test .
```

## Continuous Improvement

### Metrics to Track
- Build duration
- Test coverage percentage
- Number of failing builds
- Release frequency
- Dependency update frequency

### Regular Reviews
- Review and update linter rules quarterly
- Update Go version when new stable releases available
- Review and optimize build times
- Update dependencies regularly

## References

- [GitHub Actions Documentation](https://docs.github.com/en/actions)
- [golangci-lint Configuration](https://golangci-lint.run/usage/configuration/)
- [Codecov Documentation](https://docs.codecov.com/)
- [Docker Best Practices](https://docs.docker.com/develop/dev-best-practices/)
- [Go Build Tags](https://pkg.go.dev/cmd/go#hdr-Build_constraints)

---

**Last Updated**: 2024-12-27
