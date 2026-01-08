# Installation Guide

This guide covers all methods for installing AINative Code on Linux, macOS, and Windows.

## Table of Contents

- [Quick Install](#quick-install)
- [Package Managers](#package-managers)
- [Manual Installation](#manual-installation)
- [Building from Source](#building-from-source)
- [Docker](#docker)
- [Verification](#verification)
- [Troubleshooting](#troubleshooting)
- [Uninstallation](#uninstallation)

## Quick Install

### Linux and macOS

Download and run the installation script:

```bash
curl -fsSL https://raw.githubusercontent.com/AINative-Studio/ainative-code/main/install.sh | bash
```

The script will:
- Detect your platform and architecture automatically
- Download the latest release
- Verify checksums for security
- Install the binary to `/usr/local/bin`
- Verify the installation

#### Custom Installation Directory

```bash
export INSTALL_DIR="$HOME/.local/bin"
curl -fsSL https://raw.githubusercontent.com/AINative-Studio/ainative-code/main/install.sh | bash
```

### Windows

Download and run the PowerShell installation script:

```powershell
irm https://raw.githubusercontent.com/AINative-Studio/ainative-code/main/install.ps1 | iex
```

The script will:
- Detect your architecture (x86_64 or ARM64)
- Download the latest release
- Verify checksums for security
- Install to `%LOCALAPPDATA%\Programs\AINativeCode`
- Add to your PATH automatically

#### Custom Installation Directory

```powershell
$env:InstallDir = "$env:USERPROFILE\bin"
irm https://raw.githubusercontent.com/AINative-Studio/ainative-code/main/install.ps1 | iex
```

## Package Managers

### Homebrew (macOS and Linux)

```bash
# Add the AINative Studio tap
brew tap ainative-studio/tap

# Install AINative Code
brew install ainative-code

# Verify installation
ainative-code version
```

#### Update via Homebrew

```bash
brew update
brew upgrade ainative-code
```

### Scoop (Windows)

Coming soon! We're working on Scoop support.

### Chocolatey (Windows)

Coming soon! We're working on Chocolatey support.

## Manual Installation

### 1. Download the Binary

Visit the [releases page](https://github.com/AINative-Studio/ainative-code/releases) and download the appropriate archive for your platform:

**Linux:**
- `ainative-code_VERSION_Linux_x86_64.tar.gz` (64-bit Intel/AMD)
- `ainative-code_VERSION_Linux_arm64.tar.gz` (64-bit ARM)

**macOS:**
- `ainative-code_VERSION_Darwin_x86_64.tar.gz` (Intel Mac)
- `ainative-code_VERSION_Darwin_arm64.tar.gz` (Apple Silicon)

**Windows:**
- `ainative-code_VERSION_Windows_x86_64.zip` (64-bit Intel/AMD)
- `ainative-code_VERSION_Windows_arm64.zip` (64-bit ARM)

### 2. Verify Checksum (Recommended)

Download the `checksums.txt` file from the same release page.

**Linux/macOS:**
```bash
# Download archive and checksums
wget https://github.com/AINative-Studio/ainative-code/releases/download/v1.0.0/ainative-code_1.0.0_Linux_x86_64.tar.gz
wget https://github.com/AINative-Studio/ainative-code/releases/download/v1.0.0/checksums.txt

# Verify checksum
sha256sum -c --ignore-missing checksums.txt
```

**Windows (PowerShell):**
```powershell
# Download archive and checksums
Invoke-WebRequest -Uri "https://github.com/AINative-Studio/ainative-code/releases/download/v1.0.0/ainative-code_1.0.0_Windows_x86_64.zip" -OutFile "ainative-code.zip"
Invoke-WebRequest -Uri "https://github.com/AINative-Studio/ainative-code/releases/download/v1.0.0/checksums.txt" -OutFile "checksums.txt"

# Verify checksum
$expectedHash = (Get-Content checksums.txt | Select-String "ainative-code_1.0.0_Windows_x86_64.zip").ToString().Split()[0]
$actualHash = (Get-FileHash ainative-code.zip -Algorithm SHA256).Hash.ToLower()
if ($expectedHash -eq $actualHash) {
    Write-Host "Checksum verified!" -ForegroundColor Green
} else {
    Write-Host "Checksum mismatch!" -ForegroundColor Red
}
```

### 3. Extract and Install

**Linux/macOS:**
```bash
# Extract the archive
tar -xzf ainative-code_VERSION_OS_ARCH.tar.gz

# Move to a directory in your PATH
sudo mv ainative-code /usr/local/bin/

# Make executable (if not already)
sudo chmod +x /usr/local/bin/ainative-code

# Verify
ainative-code version
```

**Windows (PowerShell):**
```powershell
# Extract the archive
Expand-Archive -Path ainative-code.zip -DestinationPath $env:USERPROFILE\bin

# Add to PATH (if not already)
$userPath = [Environment]::GetEnvironmentVariable("Path", "User")
if ($userPath -notlike "*$env:USERPROFILE\bin*") {
    [Environment]::SetEnvironmentVariable("Path", "$userPath;$env:USERPROFILE\bin", "User")
}

# Restart your terminal and verify
ainative-code version
```

## Building from Source

### Prerequisites

- Go 1.25.5 or later
- Git
- GCC (for CGO support)

### Build Steps

```bash
# Clone the repository
git clone https://github.com/AINative-Studio/ainative-code.git
cd ainative-code

# Build the binary
make build

# Or use go build directly
go build -o ainative-code ./cmd/ainative-code

# Install to your PATH
sudo mv ainative-code /usr/local/bin/

# Verify
ainative-code version
```

### Build with Version Information

```bash
VERSION="1.0.0"
COMMIT=$(git rev-parse HEAD)
BUILD_DATE=$(date -u +"%Y-%m-%dT%H:%M:%SZ")

go build -ldflags "\
  -X github.com/AINative-studio/ainative-code/internal/cmd.version=${VERSION} \
  -X github.com/AINative-studio/ainative-code/internal/cmd.commit=${COMMIT} \
  -X github.com/AINative-studio/ainative-code/internal/cmd.buildDate=${BUILD_DATE} \
  -X github.com/AINative-studio/ainative-code/internal/cmd.builtBy=manual" \
  -o ainative-code ./cmd/ainative-code
```

## Docker

### Pull the Image

```bash
docker pull ainativestudio/ainative-code:latest
```

### Run in Docker

```bash
# Interactive mode
docker run -it --rm \
  -v $HOME/.ainative-code:/root/.ainative-code \
  ainativestudio/ainative-code:latest chat

# With environment variables
docker run -it --rm \
  -e AINATIVE_API_KEY=your_api_key \
  -v $PWD:/workspace \
  ainativestudio/ainative-code:latest chat
```

### Build Docker Image Locally

```bash
cd ainative-code
docker build -t ainative-code:local .
docker run -it --rm ainative-code:local version
```

## Verification

After installation, verify that AINative Code is working correctly:

### Check Version

```bash
ainative-code version
```

Expected output:
```
AINative Code v1.0.0
Commit:     abc1234
Built:      2024-01-07T12:00:00Z
Built by:   goreleaser
Go version: go1.25.5
Platform:   linux/amd64
```

### Check Help

```bash
ainative-code --help
```

### Run a Simple Command

```bash
ainative-code config list
```

## Troubleshooting

### Command Not Found

**Issue:** `ainative-code: command not found`

**Solution:**
1. Ensure the binary is in your PATH:
   ```bash
   which ainative-code
   ```
2. If not found, add the installation directory to your PATH:
   ```bash
   # For bash/zsh
   echo 'export PATH="$PATH:/usr/local/bin"' >> ~/.bashrc
   source ~/.bashrc

   # For fish
   set -Ua fish_user_paths /usr/local/bin
   ```

### Permission Denied

**Issue:** `Permission denied` when running the installation script

**Solution:**
```bash
# Make the script executable
chmod +x install.sh
./install.sh

# Or run with sudo if installing to /usr/local/bin
sudo bash install.sh
```

### Checksum Verification Failed

**Issue:** Checksum doesn't match during installation

**Solution:**
1. Re-download the archive - it may have been corrupted
2. Ensure you downloaded from the official GitHub releases page
3. Check your internet connection stability

### Windows: Script Execution Policy

**Issue:** PowerShell won't run the installation script

**Solution:**
```powershell
# Temporarily allow script execution
Set-ExecutionPolicy -ExecutionPolicy RemoteSigned -Scope Process

# Then run the install script
irm https://raw.githubusercontent.com/AINative-Studio/ainative-code/main/install.ps1 | iex
```

### CGO Dependency Issues

**Issue:** Build errors related to CGO when building from source

**Solution:**
- **Linux:** Install build-essential: `sudo apt-get install build-essential`
- **macOS:** Install Xcode Command Line Tools: `xcode-select --install`
- **Windows:** Install MinGW-w64 or use MSYS2

### ARM Architecture Support

**Issue:** No binary available for your ARM device

**Solution:**
- Check if there's an ARM64 release for your platform
- Build from source (Go supports cross-compilation)
- Use the Docker image (if Docker supports your ARM platform)

## Uninstallation

### Manual Uninstall

**Linux/macOS:**
```bash
# Remove the binary
sudo rm /usr/local/bin/ainative-code

# Remove configuration (optional)
rm -rf ~/.ainative-code
```

**Windows (PowerShell):**
```powershell
# Remove the binary
Remove-Item "$env:LOCALAPPDATA\Programs\AINativeCode\ainative-code.exe"

# Remove from PATH
$userPath = [Environment]::GetEnvironmentVariable("Path", "User")
$newPath = ($userPath.Split(';') | Where-Object { $_ -notlike "*AINativeCode*" }) -join ';'
[Environment]::SetEnvironmentVariable("Path", $newPath, "User")

# Remove configuration (optional)
Remove-Item -Recurse -Force "$env:USERPROFILE\.ainative-code"
```

### Homebrew Uninstall

```bash
brew uninstall ainative-code
brew untap ainative-studio/tap
```

## Next Steps

After installation:

1. **Configure API Access:**
   ```bash
   ainative-code config set provider anthropic
   ainative-code config set api_key YOUR_API_KEY
   ```

2. **Start Your First Chat:**
   ```bash
   ainative-code chat
   ```

3. **Read the Documentation:**
   - [Configuration Guide](configuration.md)
   - [Usage Guide](usage.md)
   - [Quick Start](../QUICK-START.md)

## Getting Help

If you encounter issues not covered here:

1. Check the [FAQ](faq.md)
2. Search [existing issues](https://github.com/AINative-Studio/ainative-code/issues)
3. Join our community discussions
4. Open a new issue with:
   - Your platform and architecture
   - Installation method used
   - Full error message
   - Output of `ainative-code version` (if installed)

## Version History

Check the [CHANGELOG](../CHANGELOG.md) for version history and release notes.
