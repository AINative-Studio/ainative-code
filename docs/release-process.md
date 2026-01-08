# Release Process

This document describes the automated release process for AINative Code.

## Overview

AINative Code uses an automated release pipeline powered by:
- **GoReleaser**: Multi-platform binary builds
- **GitHub Actions**: Automated CI/CD
- **Semantic Versioning**: Version management

## Release Workflow

### 1. Automated Releases (Recommended)

The release process is fully automated when you push a version tag:

```bash
# Ensure you're on main branch
git checkout main
git pull origin main

# Create a version tag (follows semantic versioning)
git tag -a v1.0.0 -m "Release version 1.0.0"

# Push the tag to trigger the release
git push origin v1.0.0
```

This will automatically:
1. Build binaries for all platforms (Linux, macOS, Windows × AMD64/ARM64)
2. Generate checksums for security verification
3. Create a GitHub Release with changelog
4. Upload all artifacts to the release
5. Update Homebrew formula (when configured)
6. Build and push Docker images (when configured)

### 2. Manual Releases

If you need to create a release manually:

```bash
# Ensure GoReleaser is installed
brew install goreleaser

# Create a snapshot build (local testing)
goreleaser build --snapshot --clean --single-target

# Create a full release (requires GITHUB_TOKEN)
export GITHUB_TOKEN="your_github_token"
goreleaser release --clean
```

## Version Numbering

We follow [Semantic Versioning 2.0.0](https://semver.org/):

- **MAJOR**: Incompatible API changes (v1.0.0 → v2.0.0)
- **MINOR**: Backwards-compatible functionality (v1.0.0 → v1.1.0)
- **PATCH**: Backwards-compatible bug fixes (v1.0.0 → v1.0.1)

### Pre-release Versions

For pre-release versions, append a suffix:

- **Alpha**: `v1.0.0-alpha.1`
- **Beta**: `v1.0.0-beta.1`
- **Release Candidate**: `v1.0.0-rc.1`

Example:
```bash
git tag -a v1.0.0-beta.1 -m "Beta release 1.0.0-beta.1"
git push origin v1.0.0-beta.1
```

## Release Checklist

Before creating a release:

- [ ] All tests pass locally: `make test`
- [ ] Code coverage meets requirements: `make coverage`
- [ ] Linter passes: `make lint`
- [ ] Update CHANGELOG.md with release notes
- [ ] Version is bumped in appropriate places
- [ ] All dependencies are up to date: `go mod tidy`
- [ ] Documentation is updated
- [ ] Integration tests pass: `make test-integration`
- [ ] Security audit complete: `make security-audit`

## GitHub Actions Workflow

The release workflow (`.github/workflows/release.yml`) runs when a tag matching `v*` is pushed:

### Jobs

1. **goreleaser**: Builds binaries and creates GitHub Release
   - Sets up Go 1.25.5
   - Installs cross-compilation tools
   - Runs GoReleaser
   - Uploads artifacts

2. **docker** (optional): Builds and pushes Docker images
   - Multi-platform builds (AMD64, ARM64)
   - Pushes to Docker Hub
   - Creates manifest for multi-arch support

3. **update-homebrew** (optional): Updates Homebrew formula
   - Checks out homebrew-tap repository
   - Updates formula with new version
   - Commits and pushes changes

### Required Secrets

Configure these secrets in GitHub repository settings:

1. **GITHUB_TOKEN**: Automatically provided by GitHub Actions
2. **HOMEBREW_TAP_GITHUB_TOKEN**: Personal access token for homebrew-tap repo
3. **DOCKER_USERNAME**: Docker Hub username (optional)
4. **DOCKER_PASSWORD**: Docker Hub password/token (optional)
5. **GPG_FINGERPRINT**: GPG key fingerprint for signing (optional)

## Build Targets

GoReleaser builds for the following platforms:

| Platform | Architecture | Notes |
|----------|-------------|-------|
| Linux    | AMD64       | Standard 64-bit Intel/AMD |
| Linux    | ARM64       | 64-bit ARM (Raspberry Pi, etc.) |
| macOS    | AMD64       | Intel Macs |
| macOS    | ARM64       | Apple Silicon (M1, M2, M3) |
| Windows  | AMD64       | Standard 64-bit Intel/AMD |
| Windows  | ARM64       | ARM64 Windows devices |

### Universal Binaries

macOS builds are combined into universal binaries that work on both Intel and Apple Silicon.

## Release Artifacts

Each release includes:

1. **Binaries**: Platform-specific executables
   - `ainative-code_VERSION_Linux_x86_64.tar.gz`
   - `ainative-code_VERSION_Linux_arm64.tar.gz`
   - `ainative-code_VERSION_Darwin_x86_64.tar.gz`
   - `ainative-code_VERSION_Darwin_arm64.tar.gz`
   - `ainative-code_VERSION_Windows_x86_64.zip`
   - `ainative-code_VERSION_Windows_arm64.zip`

2. **Checksums**: `checksums.txt` with SHA256 hashes

3. **Signatures**: GPG signatures (if configured)

4. **SBOM**: Software Bill of Materials

5. **Release Notes**: Auto-generated changelog

## Changelog Generation

The changelog is automatically generated from commit messages using conventional commits:

- `feat:` → Features section
- `fix:` → Bug Fixes section
- `security:` → Security section
- `perf:` → Performance section
- `docs:` → Documentation section

Example commit message:
```bash
git commit -m "feat: add support for Gemini 2.0 Flash model"
```

## Testing Releases

### Local Testing

Test the release process locally without publishing:

```bash
# Build snapshot (no upload)
goreleaser build --snapshot --clean

# Test the binaries
./dist/ainative-code_darwin_arm64_v8.0/ainative-code version

# Test on specific platform only
goreleaser build --snapshot --clean --single-target
```

### Pre-release Testing

Create a pre-release for testing:

```bash
git tag -a v1.0.0-rc.1 -m "Release candidate 1"
git push origin v1.0.0-rc.1
```

Pre-releases are automatically marked in GitHub and don't trigger Homebrew updates.

## Rollback Procedure

If a release has critical issues:

1. **Mark release as pre-release** in GitHub
2. **Create hotfix**:
   ```bash
   git checkout v1.0.0
   git checkout -b hotfix/1.0.1
   # Make fixes
   git commit -m "fix: critical bug in release 1.0.0"
   git tag -a v1.0.1 -m "Hotfix release 1.0.1"
   git push origin v1.0.1
   ```

3. **Update documentation** to point to new version

## Homebrew Distribution

After a successful release:

1. **Automatic Update**: GitHub Actions updates the formula
2. **Manual Update**: See [Homebrew Distribution Guide](homebrew-distribution.md)
3. **Testing**: Test installation: `brew install ainative-studio/tap/ainative-code`

### Creating the Homebrew Tap

First-time setup (one-time only):

```bash
# Create the tap repository
gh repo create AINative-Studio/homebrew-tap --public

# Clone and initialize
git clone https://github.com/AINative-Studio/homebrew-tap.git
cd homebrew-tap
mkdir -p Formula

# Create initial formula from template
cp /path/to/ainative-code/scripts/homebrew/ainative-code.rb.template Formula/ainative-code.rb

# Commit and push
git add Formula/ainative-code.rb
git commit -m "chore: add initial ainative-code formula"
git push origin main
```

## Docker Distribution

After release, Docker images are available:

```bash
# Latest release
docker pull ainativestudio/ainative-code:latest

# Specific version
docker pull ainativestudio/ainative-code:v1.0.0

# Platform-specific
docker pull ainativestudio/ainative-code:v1.0.0-amd64
docker pull ainativestudio/ainative-code:v1.0.0-arm64
```

## Monitoring Releases

### Release Status

Monitor release status in GitHub Actions:
- https://github.com/AINative-Studio/ainative-code/actions/workflows/release.yml

### Release Notifications

Set up notifications for:
- Release workflow failures
- User-reported installation issues
- Download statistics

### Metrics to Track

- Download counts per platform
- Installation success rate
- Time to release (tag → published)
- User-reported issues

## Troubleshooting

### Build Failures

**Cross-compilation errors:**
- Ensure all cross-compilation tools are installed
- Check GoReleaser configuration for correct compiler paths
- Verify CGO settings for each platform

**GoReleaser validation errors:**
```bash
# Validate configuration
goreleaser check

# Run with debug output
goreleaser release --clean --debug
```

### GitHub Actions Failures

**Token issues:**
- Verify GITHUB_TOKEN has correct permissions
- Check HOMEBREW_TAP_GITHUB_TOKEN is valid

**Build timeouts:**
- Check for slow network connections
- Verify GitHub Actions quota

### Homebrew Issues

**Formula update failures:**
- Check homebrew-tap repository exists
- Verify token permissions
- Manual update as fallback

**Checksum mismatches:**
- Re-download release assets
- Verify checksums.txt is correct
- Update formula manually if needed

## Best Practices

1. **Always test locally** before releasing
2. **Use pre-releases** for major changes
3. **Keep CHANGELOG.md updated** with meaningful notes
4. **Follow semantic versioning** strictly
5. **Test installation scripts** on all platforms
6. **Monitor release metrics** for issues
7. **Document breaking changes** clearly
8. **Maintain backwards compatibility** when possible

## Emergency Procedures

### Critical Bug in Release

1. Create hotfix branch
2. Fix the bug
3. Create patch version
4. Release immediately
5. Notify users via GitHub Release notes

### Security Vulnerability

1. **DO NOT** push public commits with details
2. Create private security advisory
3. Fix in private branch
4. Coordinate disclosure
5. Release patch version
6. Publish security advisory

### Revoke Release

```bash
# Delete tag locally
git tag -d v1.0.0

# Delete tag remotely
git push origin :refs/tags/v1.0.0

# Delete GitHub Release
gh release delete v1.0.0
```

## Resources

- [GoReleaser Documentation](https://goreleaser.com)
- [Semantic Versioning](https://semver.org)
- [GitHub Actions Documentation](https://docs.github.com/actions)
- [Homebrew Formula Cookbook](https://docs.brew.sh/Formula-Cookbook)
- [Conventional Commits](https://www.conventionalcommits.org)

## Support

For release-related issues:
- Open an issue: https://github.com/AINative-Studio/ainative-code/issues
- Contact: devops@ainative.studio
