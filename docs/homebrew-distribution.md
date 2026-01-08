# Homebrew Distribution Guide

This document describes how AINative Code is distributed via Homebrew and how to maintain the Homebrew tap.

## Overview

AINative Code is available via Homebrew through a custom tap repository. The distribution process is fully automated via GitHub Actions when a new release is created.

## Repository Structure

The Homebrew tap requires a separate repository:
- Repository: `https://github.com/AINative-Studio/homebrew-tap`
- Formula location: `Formula/ainative-code.rb`

## Automated Distribution Process

When a new version is tagged (e.g., `v1.0.0`), the following happens automatically:

1. **GoReleaser builds binaries** for all platforms
2. **Checksums are generated** and published with the release
3. **GitHub Actions workflow** updates the Homebrew formula
4. **Formula is committed** to the homebrew-tap repository

## Setting Up the Homebrew Tap Repository

### 1. Create the Tap Repository

```bash
# Create a new repository named 'homebrew-tap'
gh repo create AINative-Studio/homebrew-tap --public --description "Homebrew tap for AINative Code"

# Clone the repository
git clone https://github.com/AINative-Studio/homebrew-tap.git
cd homebrew-tap

# Create the Formula directory
mkdir -p Formula

# Create initial formula (will be auto-updated by releases)
cp /path/to/ainative-code/scripts/homebrew/ainative-code.rb.template Formula/ainative-code.rb

# Commit and push
git add Formula/ainative-code.rb
git commit -m "chore: add initial ainative-code formula"
git push origin main
```

### 2. Configure GitHub Secrets

Add the following secret to the main `ainative-code` repository:

```bash
# Generate a Personal Access Token with 'repo' scope
# Settings → Developer settings → Personal access tokens → Generate new token

# Add the secret to ainative-code repository
gh secret set HOMEBREW_TAP_GITHUB_TOKEN --body "your_github_token"
```

This token allows the release workflow to update the homebrew-tap repository.

## Manual Formula Updates

If you need to manually update the formula:

### 1. Update Version and Checksums

```bash
# Set the version
VERSION="1.0.0"

# Download checksums from the release
curl -fsSL "https://github.com/AINative-Studio/ainative-code/releases/download/v${VERSION}/checksums.txt" -o checksums.txt

# Extract checksums for each platform
DARWIN_AMD64_SHA=$(grep "ainative-code_${VERSION}_Darwin_x86_64.tar.gz" checksums.txt | awk '{print $1}')
DARWIN_ARM64_SHA=$(grep "ainative-code_${VERSION}_Darwin_arm64.tar.gz" checksums.txt | awk '{print $1}')
LINUX_AMD64_SHA=$(grep "ainative-code_${VERSION}_Linux_x86_64.tar.gz" checksums.txt | awk '{print $1}')
LINUX_ARM64_SHA=$(grep "ainative-code_${VERSION}_Linux_arm64.tar.gz" checksums.txt | awk '{print $1}')
```

### 2. Update the Formula

Edit `Formula/ainative-code.rb` and update:
- `version` line
- All `sha256` values for each platform
- URLs if the naming convention changed

### 3. Test the Formula

```bash
# Test the formula locally
brew install --build-from-source ./Formula/ainative-code.rb

# Verify installation
ainative-code version

# Uninstall after testing
brew uninstall ainative-code
```

### 4. Audit the Formula

```bash
# Run Homebrew audit to check for issues
brew audit --new-formula ./Formula/ainative-code.rb

# Fix any issues reported
```

### 5. Commit and Push

```bash
git add Formula/ainative-code.rb
git commit -m "chore: update ainative-code to ${VERSION}"
git push origin main
```

## User Installation

Users can install AINative Code via Homebrew with:

```bash
# Add the tap
brew tap ainative-studio/tap

# Install ainative-code
brew install ainative-code

# Verify installation
ainative-code version
```

## Troubleshooting

### Formula Fails to Build

If users report build failures:

1. Check the build logs from failed installations
2. Verify checksums are correct
3. Test the formula on the affected platform
4. Check for platform-specific issues in GoReleaser config

### Checksum Mismatches

If checksums don't match:

1. Download the release archive manually
2. Verify the checksum: `sha256sum archive.tar.gz`
3. Compare with the checksum in checksums.txt
4. If different, the release may have been re-uploaded - update formula

### Formula Not Found

If `brew install ainative-code` fails with "formula not found":

1. Verify the tap repository exists and is public
2. Check that Formula/ainative-code.rb exists in the tap
3. Try updating Homebrew: `brew update`
4. Try re-adding the tap: `brew untap ainative-studio/tap && brew tap ainative-studio/tap`

## Maintenance

### Regular Tasks

1. **Monitor releases**: Ensure formula updates happen automatically
2. **Test installations**: Periodically test installation on different platforms
3. **Update dependencies**: Keep the formula in sync with project requirements
4. **Review issues**: Monitor homebrew-tap repository for user issues

### Version Updates

Formula updates are automatic, but verify after each release:

```bash
# Check the latest version in homebrew-tap
curl -fsSL https://raw.githubusercontent.com/AINative-Studio/homebrew-tap/main/Formula/ainative-code.rb | grep version

# Compare with latest release
gh release view --repo AINative-Studio/ainative-code
```

## Best Practices

1. **Test before releasing**: Always test the formula on multiple platforms
2. **Keep checksums accurate**: Never skip checksum verification
3. **Document changes**: Add meaningful commit messages
4. **Version consistency**: Ensure formula version matches release tag
5. **Monitor builds**: Watch GitHub Actions for any failures

## Resources

- [Homebrew Formula Cookbook](https://docs.brew.sh/Formula-Cookbook)
- [Homebrew Acceptable Formulae](https://docs.brew.sh/Acceptable-Formulae)
- [Creating Homebrew Taps](https://docs.brew.sh/How-to-Create-and-Maintain-a-Tap)
- [GoReleaser Homebrew Integration](https://goreleaser.com/customization/homebrew/)

## Support

For issues related to:
- **Homebrew installation**: Open an issue in homebrew-tap repository
- **Binary builds**: Open an issue in ainative-code repository
- **Formula updates**: Contact the maintainers via GitHub

## Checklist for New Releases

- [ ] Tag is pushed (e.g., `v1.0.0`)
- [ ] GitHub Actions workflow completes successfully
- [ ] Binaries are published to GitHub Releases
- [ ] Checksums are available in release assets
- [ ] Homebrew formula is auto-updated
- [ ] Test installation: `brew install ainative-studio/tap/ainative-code`
- [ ] Verify version: `ainative-code version`
- [ ] Test basic functionality: `ainative-code chat --help`
