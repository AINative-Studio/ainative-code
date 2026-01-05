# Installation Guide

This guide provides detailed instructions for installing AINative Code on different platforms.

## Table of Contents

1. [System Requirements](#system-requirements)
2. [Installation Methods](#installation-methods)
   - [macOS](#macos)
   - [Linux](#linux)
   - [Windows](#windows)
   - [Docker](#docker)
   - [From Source](#from-source)
3. [Verification](#verification)
4. [Initial Setup](#initial-setup)
5. [Upgrading](#upgrading)
6. [Uninstallation](#uninstallation)
7. [Troubleshooting](#troubleshooting)

## System Requirements

### Minimum Requirements

- **Operating System**: macOS 10.15+, Linux (Ubuntu 20.04+, Debian 10+, CentOS 8+), Windows 10+
- **Memory**: 2 GB RAM (4 GB recommended)
- **Disk Space**: 100 MB for installation
- **Network**: Internet connection for LLM provider access

### Recommended Requirements

- **Memory**: 4+ GB RAM
- **Processor**: Multi-core processor (for optimal performance)
- **Network**: Stable broadband connection

### Dependencies

No additional dependencies are required. AINative Code is distributed as a standalone binary.

## Installation Methods

### macOS

#### Method 1: Homebrew (Recommended)

The easiest way to install on macOS is using Homebrew:

```bash
# Add the AINative tap
brew tap ainative-studio/tap

# Install AINative Code
brew install ainative-code
```

To update later:
```bash
brew upgrade ainative-code
```

#### Method 2: Direct Download

**For Intel Macs:**
```bash
# Download the binary
curl -LO https://github.com/AINative-studio/ainative-code/releases/latest/download/ainative-code-darwin-amd64

# Make it executable
chmod +x ainative-code-darwin-amd64

# Move to a directory in your PATH
sudo mv ainative-code-darwin-amd64 /usr/local/bin/ainative-code
```

**For Apple Silicon (M1/M2/M3):**
```bash
# Download the binary
curl -LO https://github.com/AINative-studio/ainative-code/releases/latest/download/ainative-code-darwin-arm64

# Make it executable
chmod +x ainative-code-darwin-arm64

# Move to a directory in your PATH
sudo mv ainative-code-darwin-arm64 /usr/local/bin/ainative-code
```

#### Method 3: Using Install Script

```bash
curl -fsSL https://raw.githubusercontent.com/AINative-studio/ainative-code/main/install.sh | bash
```

### Linux

#### Method 1: Package Manager

**Ubuntu/Debian:**
```bash
# Add the AINative repository
curl -fsSL https://packages.ainative.studio/apt/gpg.key | sudo apt-key add -
echo "deb https://packages.ainative.studio/apt stable main" | sudo tee /etc/apt/sources.list.d/ainative.list

# Update and install
sudo apt update
sudo apt install ainative-code
```

**Fedora/RHEL/CentOS:**
```bash
# Add the AINative repository
sudo tee /etc/yum.repos.d/ainative.repo <<EOF
[ainative]
name=AINative Repository
baseurl=https://packages.ainative.studio/rpm/stable
enabled=1
gpgcheck=1
gpgkey=https://packages.ainative.studio/rpm/gpg.key
EOF

# Install
sudo dnf install ainative-code
# or for older systems
sudo yum install ainative-code
```

**Arch Linux:**
```bash
# Install from AUR
yay -S ainative-code
# or
paru -S ainative-code
```

#### Method 2: Direct Download

**For AMD64 systems:**
```bash
# Download the binary
curl -LO https://github.com/AINative-studio/ainative-code/releases/latest/download/ainative-code-linux-amd64

# Make it executable
chmod +x ainative-code-linux-amd64

# Move to a directory in your PATH
sudo mv ainative-code-linux-amd64 /usr/local/bin/ainative-code
```

**For ARM64 systems:**
```bash
# Download the binary
curl -LO https://github.com/AINative-studio/ainative-code/releases/latest/download/ainative-code-linux-arm64

# Make it executable
chmod +x ainative-code-linux-arm64

# Move to a directory in your PATH
sudo mv ainative-code-linux-arm64 /usr/local/bin/ainative-code
```

#### Method 3: Using Install Script

```bash
curl -fsSL https://raw.githubusercontent.com/AINative-studio/ainative-code/main/install.sh | bash
```

### Windows

#### Method 1: Installer (Recommended)

1. Download the installer from the [releases page](https://github.com/AINative-studio/ainative-code/releases/latest)
2. Run `ainative-code-setup.exe`
3. Follow the installation wizard
4. The installer will add AINative Code to your PATH automatically

#### Method 2: Scoop

```powershell
# Add the AINative bucket
scoop bucket add ainative https://github.com/AINative-studio/scoop-bucket

# Install
scoop install ainative-code
```

#### Method 3: Chocolatey

```powershell
# Install using Chocolatey
choco install ainative-code
```

#### Method 4: Direct Download

```powershell
# Download using PowerShell
Invoke-WebRequest -Uri "https://github.com/AINative-studio/ainative-code/releases/latest/download/ainative-code-windows-amd64.exe" -OutFile "ainative-code.exe"

# Move to a directory in your PATH (requires admin privileges)
Move-Item ainative-code.exe C:\Windows\System32\

# Or add current directory to PATH (user-level, no admin required)
$env:Path += ";$PWD"
[Environment]::SetEnvironmentVariable("Path", $env:Path, [EnvironmentVariableTarget]::User)
```

### Docker

#### Using Docker Hub

```bash
# Pull the latest image
docker pull ainativestudio/ainative-code:latest

# Run interactively
docker run -it --rm \
  -v $(pwd):/workspace \
  -e ANTHROPIC_API_KEY=$ANTHROPIC_API_KEY \
  ainativestudio/ainative-code:latest
```

#### Using GitHub Container Registry

```bash
# Pull from GitHub Container Registry
docker pull ghcr.io/ainative-studio/ainative-code:latest

# Run with mounted configuration
docker run -it --rm \
  -v $(pwd):/workspace \
  -v ~/.ainative:/root/.ainative \
  ghcr.io/ainative-studio/ainative-code:latest
```

#### Docker Compose

Create a `docker-compose.yml` file:

```yaml
version: '3.8'

services:
  ainative-code:
    image: ainativestudio/ainative-code:latest
    volumes:
      - ./workspace:/workspace
      - ~/.ainative:/root/.ainative
    environment:
      - ANTHROPIC_API_KEY=${ANTHROPIC_API_KEY}
      - OPENAI_API_KEY=${OPENAI_API_KEY}
    stdin_open: true
    tty: true
```

Run with:
```bash
docker-compose run --rm ainative-code
```

### From Source

#### Prerequisites

- Go 1.21 or higher
- Git
- Make (optional, but recommended)

#### Build Steps

```bash
# Clone the repository
git clone https://github.com/AINative-studio/ainative-code.git
cd ainative-code

# Build using Make (recommended)
make build

# Or build manually
go build -o ainative-code ./cmd/ainative-code

# Install to your system
sudo make install
# Or manually
sudo mv ainative-code /usr/local/bin/
```

#### Build Options

**Optimized Production Build:**
```bash
make build-release
```

**Build for Specific Platform:**
```bash
# macOS (Intel)
GOOS=darwin GOARCH=amd64 go build -o ainative-code-darwin-amd64 ./cmd/ainative-code

# macOS (Apple Silicon)
GOOS=darwin GOARCH=arm64 go build -o ainative-code-darwin-arm64 ./cmd/ainative-code

# Linux (AMD64)
GOOS=linux GOARCH=amd64 go build -o ainative-code-linux-amd64 ./cmd/ainative-code

# Windows
GOOS=windows GOARCH=amd64 go build -o ainative-code.exe ./cmd/ainative-code
```

**Build with Version Information:**
```bash
VERSION=$(git describe --tags --always)
BUILD_TIME=$(date -u '+%Y-%m-%d_%H:%M:%S')
go build -ldflags "-X main.Version=$VERSION -X main.BuildTime=$BUILD_TIME" ./cmd/ainative-code
```

## Verification

After installation, verify that AINative Code is installed correctly:

```bash
# Check version
ainative-code --version

# Expected output:
# ainative-code version 0.1.0
```

Check that the binary is in your PATH:

```bash
# Unix/Linux/macOS
which ainative-code

# Windows (PowerShell)
Get-Command ainative-code

# Windows (Command Prompt)
where ainative-code
```

Run the help command to ensure it's working:

```bash
ainative-code --help
```

## Initial Setup

After installation, perform the initial setup:

### 1. Create Configuration Directory

The configuration directory is created automatically on first run, but you can create it manually:

```bash
# Unix/Linux/macOS
mkdir -p ~/.config/ainative-code

# Windows (PowerShell)
New-Item -Path "$env:APPDATA\ainative-code" -ItemType Directory -Force
```

### 2. Initialize Configuration

Run the initialization command to create a default configuration file:

```bash
ainative-code init
```

This will create `~/.config/ainative-code/config.yaml` with default settings.

### 3. Configure Your LLM Provider

Set up your preferred LLM provider. For example, to use Anthropic Claude:

```bash
# Set provider
ainative-code config set llm.default_provider anthropic

# Set API key
ainative-code config set llm.anthropic.api_key "your-api-key-here"

# Or use environment variable
export ANTHROPIC_API_KEY="your-api-key-here"
```

See the [Configuration Guide](configuration.md) for detailed configuration options.

### 4. Test Your Setup

Test that everything is working:

```bash
# Start a chat session
ainative-code chat "Hello, can you help me with coding?"
```

## Upgrading

### Homebrew (macOS)

```bash
brew upgrade ainative-code
```

### Package Managers (Linux)

```bash
# Ubuntu/Debian
sudo apt update && sudo apt upgrade ainative-code

# Fedora/RHEL/CentOS
sudo dnf upgrade ainative-code

# Arch Linux
yay -Syu ainative-code
```

### Scoop (Windows)

```powershell
scoop update ainative-code
```

### Chocolatey (Windows)

```powershell
choco upgrade ainative-code
```

### Manual Upgrade

1. Download the latest version using the same method as installation
2. Replace the existing binary
3. Verify the new version:

```bash
ainative-code --version
```

### Docker

```bash
# Pull the latest image
docker pull ainativestudio/ainative-code:latest
```

## Uninstallation

### Homebrew (macOS)

```bash
brew uninstall ainative-code
brew untap ainative-studio/tap
```

### Package Managers (Linux)

```bash
# Ubuntu/Debian
sudo apt remove ainative-code

# Fedora/RHEL/CentOS
sudo dnf remove ainative-code

# Arch Linux
yay -R ainative-code
```

### Scoop (Windows)

```powershell
scoop uninstall ainative-code
```

### Chocolatey (Windows)

```powershell
choco uninstall ainative-code
```

### Manual Removal

```bash
# Remove binary
sudo rm /usr/local/bin/ainative-code

# Remove configuration (optional)
rm -rf ~/.config/ainative-code

# Remove cache (optional)
rm -rf ~/.cache/ainative-code
```

## Troubleshooting

### Command Not Found

If you get a "command not found" error:

1. Verify the binary is in your PATH:
   ```bash
   echo $PATH
   ```

2. Add the installation directory to your PATH:
   ```bash
   # Add to ~/.bashrc, ~/.zshrc, or equivalent
   export PATH="$PATH:/usr/local/bin"
   ```

3. Reload your shell configuration:
   ```bash
   source ~/.bashrc  # or ~/.zshrc
   ```

### Permission Denied

If you get permission errors:

```bash
# Make the binary executable
chmod +x /usr/local/bin/ainative-code

# Or use sudo for installation
sudo mv ainative-code /usr/local/bin/
```

### SSL/TLS Certificate Errors

If you encounter certificate errors:

```bash
# macOS
export SSL_CERT_FILE=/etc/ssl/cert.pem

# Linux
export SSL_CERT_FILE=/etc/ssl/certs/ca-certificates.crt

# Or update ca-certificates
sudo update-ca-certificates
```

### Port Already in Use (OAuth)

If the OAuth callback port (8080) is already in use:

```bash
# Use a different port
ainative-code auth login --redirect-url http://localhost:8081/callback
```

### Docker Issues

If Docker containers fail to start:

```bash
# Check Docker is running
docker info

# Pull the latest image
docker pull ainativestudio/ainative-code:latest

# Run with verbose logging
docker run -it --rm -e LOG_LEVEL=debug ainativestudio/ainative-code:latest
```

### Build from Source Issues

If building from source fails:

```bash
# Ensure Go is installed and up to date
go version  # Should be 1.21 or higher

# Clean and rebuild
make clean
make build

# Check for dependency issues
go mod tidy
go mod verify
```

## Next Steps

- [Getting Started Guide](getting-started.md) - Learn the basics
- [Configuration Guide](configuration.md) - Customize your setup
- [Providers Guide](providers.md) - Configure LLM providers
- [Authentication Guide](authentication.md) - Set up platform authentication

## Support

If you encounter issues not covered in this guide:

- Check the [Troubleshooting Guide](troubleshooting.md)
- Visit [GitHub Issues](https://github.com/AINative-studio/ainative-code/issues)
- Join [GitHub Discussions](https://github.com/AINative-studio/ainative-code/discussions)
- Email: support@ainative.studio
