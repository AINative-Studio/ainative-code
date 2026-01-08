# Versioning Policy

## Overview

AINative Code follows [Semantic Versioning 2.0.0](https://semver.org/) for all releases. This document outlines our versioning strategy, release process, and guidelines for maintainers.

## Version Format

Versions follow the format: `vMAJOR.MINOR.PATCH[-PRERELEASE][+BUILD]`

- **MAJOR**: Incremented for incompatible API changes
- **MINOR**: Incremented for backwards-compatible new features
- **PATCH**: Incremented for backwards-compatible bug fixes
- **PRERELEASE**: Optional identifier for pre-release versions (alpha, beta, rc)
- **BUILD**: Optional build metadata

### Examples

- `v0.1.0` - Initial minor release
- `v0.2.0` - Second minor release with new features
- `v1.0.0` - First stable release
- `v1.0.1` - Patch release with bug fixes
- `v1.1.0-beta.1` - Beta pre-release
- `v1.2.0-rc.1` - Release candidate

## Version Incrementation Rules

### MAJOR Version (v1.0.0 → v2.0.0)

Increment when making incompatible changes:
- Breaking changes to CLI commands or flags
- Breaking changes to configuration file format
- Removal of deprecated features
- Incompatible API changes for library consumers

**Note**: During initial development (0.x.x), MAJOR version remains at 0, and MINOR version increments may include breaking changes.

### MINOR Version (v1.0.0 → v1.1.0)

Increment when adding backwards-compatible functionality:
- New CLI commands or subcommands
- New configuration options (with backwards compatibility)
- New LLM provider integrations
- New features in TUI interface
- New platform integrations (ZeroDB, Strapi, etc.)
- Deprecation warnings (but feature still works)

### PATCH Version (v1.0.0 → v1.0.1)

Increment for backwards-compatible bug fixes:
- Bug fixes without behavior changes
- Performance improvements
- Documentation corrections
- Security patches (non-breaking)
- Dependency updates (patch-level)

## Pre-Release Versions

Pre-release identifiers indicate development status:

### Alpha (`-alpha.N`)

Early development, potentially unstable:
```bash
v0.2.0-alpha.1
v0.2.0-alpha.2
```

**Use when:**
- Feature is experimental
- Breaking changes may still occur
- Limited testing completed

### Beta (`-beta.N`)

Feature-complete but needs testing:
```bash
v0.2.0-beta.1
v0.2.0-beta.2
```

**Use when:**
- All planned features implemented
- API is frozen
- Needs broader testing
- Bug fixes only

### Release Candidate (`-rc.N`)

Final testing before stable release:
```bash
v0.2.0-rc.1
v0.2.0-rc.2
```

**Use when:**
- All known bugs fixed
- Ready for production unless critical issues found
- Final integration testing

## Release Process

### 1. Prepare Release

```bash
# Ensure working directory is clean
git status

# Update CHANGELOG.md with release notes
# Update version references in documentation

# Commit changes
git add CHANGELOG.md docs/
git commit -m "docs: Prepare release v0.2.0"
```

### 2. Create Git Tag

```bash
# Create annotated tag
git tag -a v0.2.0 -m "Release v0.2.0: Add Azure OpenAI provider support"

# Verify tag
git tag -l "v0.2.0" -n9
```

### 3. Push to Remote

```bash
# Push commits
git push origin main

# Push tag
git push origin v0.2.0
```

### 4. Create GitHub Release

Use GitHub CLI:
```bash
gh release create v0.2.0 \
  --title "v0.2.0 - Azure OpenAI Support" \
  --notes-file docs/releases/changelog/v0.2.0.md \
  --draft  # Remove for immediate publish
```

Or manually:
1. Go to https://github.com/AINative-studio/ainative-code/releases/new
2. Select the tag `v0.2.0`
3. Enter release title and notes
4. Attach compiled binaries if available
5. Publish release

### 5. Verify Release

```bash
# Verify tag
git describe --tags

# Verify GitHub release
gh release view v0.2.0

# Test version in binary
go run cmd/ainative-code/main.go version
```

## Automated Release Pipeline

GitHub Actions automatically handles:
1. **Build**: Compiles binaries for all platforms
2. **Test**: Runs full test suite
3. **Package**: Creates release archives
4. **Publish**: Uploads to GitHub Releases
5. **Distribute**: Updates Homebrew tap, Docker images

Triggered by pushing tags matching `v*.*.*`

## Changelog Management

### Format

Follow [Keep a Changelog](https://keepachangelog.com/) format:

```markdown
# Changelog

## [0.2.0] - 2024-01-15

### Added
- Azure OpenAI provider support
- Config value validation

### Changed
- Improved error handling in config loading

### Fixed
- Examples build errors
- Version embedding in binaries

### Security
- Updated jose2go to v1.7.0 (CVE fixes)

## [0.1.0] - 2024-01-10

### Added
- Initial release with core features
```

### Categories

- **Added**: New features
- **Changed**: Changes to existing functionality
- **Deprecated**: Soon-to-be removed features
- **Removed**: Removed features
- **Fixed**: Bug fixes
- **Security**: Security fixes

## Git Tag Conventions

### Tag Naming

- Always prefix with `v`: `v0.1.0` (not `0.1.0`)
- Use annotated tags: `git tag -a` (not lightweight tags)
- Include meaningful message: Short description of release

### Tag Messages

Good examples:
```bash
git tag -a v0.1.0 -m "Initial release with core features"
git tag -a v0.2.0 -m "Add Azure OpenAI provider and config validation"
git tag -a v1.0.0 -m "First stable release"
```

Bad examples:
```bash
git tag v0.1.0  # Lightweight tag (not annotated)
git tag -a v0.1.0 -m "v0.1.0"  # Non-descriptive message
```

## Version Detection

The build system embeds version information:

```bash
# During build
go build -ldflags "-X main.Version=$(git describe --tags --always --dirty)"

# Runtime version detection
$ ainative-code version
AINative Code v0.2.0
```

If no tag exists, version falls back to:
- Commit hash: `c76a39d`
- "dev" for development builds

## Special Cases

### Initial Development (0.x.x)

During 0.x.x versions:
- Public API is not stable
- Breaking changes may occur in MINOR versions
- PATCH versions are backwards-compatible
- Use 0.x.x until API is stable enough for 1.0.0

### Transitioning to v1.0.0

Release v1.0.0 when:
- Public API is stable and well-documented
- Core features are production-ready
- Comprehensive test coverage
- Used in production by multiple users
- Ready to commit to backwards compatibility

### Hotfixes

For critical production fixes:

```bash
# From release tag
git checkout v1.0.0
git checkout -b hotfix/1.0.1

# Make fixes
git commit -m "fix: Critical security patch"

# Tag hotfix
git tag -a v1.0.1 -m "Hotfix: Critical security patch"

# Merge back to main
git checkout main
git merge hotfix/1.0.1
git push origin main v1.0.1
```

## Version Support

### Active Support

- **Latest MAJOR.MINOR**: Full support (features + fixes)
- **Previous MINOR**: Bug fixes and security patches
- **Previous MAJOR**: Security patches only (6 months)

### End of Life

Versions reach EOL when:
- Two MAJOR versions behind current
- One year after MAJOR release
- Announced 3 months in advance

## Resources

- [Semantic Versioning 2.0.0](https://semver.org/)
- [Keep a Changelog](https://keepachangelog.com/)
- [GitHub Releases Documentation](https://docs.github.com/en/repositories/releasing-projects-on-github)
- [Git Tagging Documentation](https://git-scm.com/book/en/v2/Git-Basics-Tagging)

## Questions?

For questions about versioning:
- Open a GitHub Discussion
- Contact the maintainers team
- Review past release notes for examples
