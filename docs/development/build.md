# Build Instructions

This document provides comprehensive build instructions for the AINative Code project.

## Table of Contents

- [Quick Build](#quick-build)
- [Build Targets](#build-targets)
- [Build Configuration](#build-configuration)
- [Cross-Platform Building](#cross-platform-building)
- [Docker Builds](#docker-builds)
- [Release Builds](#release-builds)
- [Build Optimization](#build-optimization)
- [Troubleshooting](#troubleshooting)

## Quick Build

### Using Makefile (Recommended)

```bash
# Build for current platform
make build

# The binary will be at: ./build/ainative-code
```

### Manual Build

```bash
# Create build directory
mkdir -p build

# Build the application
go build -o build/ainative-code ./cmd/ainative-code

# Run the built binary
./build/ainative-code version
```

## Build Targets

### Standard Builds

#### Development Build

For local development with debug symbols:

```bash
make build
```

This creates a debug build with:
- Debug symbols included
- No optimizations
- Faster build time

#### Production Build

For production deployment with optimizations:

```bash
# Using Makefile
make release

# Or manually with optimizations
go build \
  -ldflags="-s -w -X main.version=$(git describe --tags) -X main.buildDate=$(date -u +%Y-%m-%dT%H:%M:%SZ)" \
  -o build/ainative-code \
  ./cmd/ainative-code
```

Production build flags:
- `-s`: Omit symbol table and debug info
- `-w`: Omit DWARF symbol table
- Reduces binary size by ~30%

### Platform-Specific Builds

#### macOS

```bash
# Intel (amd64)
GOOS=darwin GOARCH=amd64 go build -o build/ainative-code-darwin-amd64 ./cmd/ainative-code

# Apple Silicon (arm64)
GOOS=darwin GOARCH=arm64 go build -o build/ainative-code-darwin-arm64 ./cmd/ainative-code

# Universal Binary (both architectures)
lipo -create \
  build/ainative-code-darwin-amd64 \
  build/ainative-code-darwin-arm64 \
  -output build/ainative-code-darwin-universal
```

#### Linux

```bash
# AMD64
GOOS=linux GOARCH=amd64 go build -o build/ainative-code-linux-amd64 ./cmd/ainative-code

# ARM64 (for ARM servers/Raspberry Pi)
GOOS=linux GOARCH=arm64 go build -o build/ainative-code-linux-arm64 ./cmd/ainative-code

# ARM v7 (for older ARM devices)
GOOS=linux GOARCH=arm GOARM=7 go build -o build/ainative-code-linux-armv7 ./cmd/ainative-code
```

#### Windows

```bash
# AMD64
GOOS=windows GOARCH=amd64 go build -o build/ainative-code-windows-amd64.exe ./cmd/ainative-code

# 32-bit (if needed)
GOOS=windows GOARCH=386 go build -o build/ainative-code-windows-386.exe ./cmd/ainative-code
```

### Build All Platforms

```bash
# Build for all supported platforms at once
make build-all
```

This creates binaries for:
- macOS (Intel and Apple Silicon)
- Linux (AMD64 and ARM64)
- Windows (AMD64)

Output:
```
build/
├── ainative-code-darwin-amd64
├── ainative-code-darwin-arm64
├── ainative-code-linux-amd64
├── ainative-code-linux-arm64
└── ainative-code-windows-amd64.exe
```

## Build Configuration

### Environment Variables

```bash
# Required for SQLite support
export CGO_ENABLED=1

# For cross-compilation (may require cross-compiler)
export CGO_ENABLED=0  # Disable CGO for pure Go builds

# Set Go compiler
export GOCMD=go

# Custom build directory
export BUILD_DIR=./dist
```

### Build Variables

The Makefile supports these variables:

```bash
# Set version
make build VERSION=v1.2.3

# Set custom binary name
make build BINARY_NAME=my-ainative

# Combine multiple variables
make build VERSION=v1.2.3 BUILD_DIR=./dist
```

### LDFLAGS Configuration

Customize build-time variables:

```bash
go build -ldflags="\
  -X main.version=$(git describe --tags --always --dirty) \
  -X main.buildDate=$(date -u +%Y-%m-%dT%H:%M:%SZ) \
  -X main.gitCommit=$(git rev-parse HEAD) \
  -X main.gitBranch=$(git rev-parse --abbrev-ref HEAD) \
  -s -w" \
  -o build/ainative-code ./cmd/ainative-code
```

These variables can be displayed with:
```bash
./build/ainative-code version
```

## Cross-Platform Building

### Prerequisites for Cross-Compilation

#### Building Linux Binaries from macOS

For pure Go (CGO_ENABLED=0):
```bash
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o build/ainative-code-linux-amd64 ./cmd/ainative-code
```

For CGO-enabled builds (requires cross-compiler):
```bash
# Install cross-compilation tools
brew install FiloSottile/musl-cross/musl-cross

# Build with CGO
CC=x86_64-linux-musl-gcc \
  CGO_ENABLED=1 \
  GOOS=linux \
  GOARCH=amd64 \
  go build -o build/ainative-code-linux-amd64 ./cmd/ainative-code
```

#### Building Windows Binaries from Linux/macOS

```bash
# Install MinGW (macOS)
brew install mingw-w64

# Build Windows binary
CC=x86_64-w64-mingw32-gcc \
  CGO_ENABLED=1 \
  GOOS=windows \
  GOARCH=amd64 \
  go build -o build/ainative-code-windows-amd64.exe ./cmd/ainative-code
```

### Using Docker for Cross-Platform Builds

Build using Docker for consistent cross-platform builds:

```bash
# Build Linux binary in Docker
docker run --rm -v "$PWD":/app -w /app golang:1.21-alpine \
  sh -c "apk add --no-cache git make gcc musl-dev && make build"

# Multi-architecture build
docker buildx build \
  --platform linux/amd64,linux/arm64 \
  -t ainative-code:latest \
  .
```

## Docker Builds

### Standard Docker Build

```bash
# Build Docker image
make docker-build

# This creates:
# - ainative-code:latest
# - ainative-code:<VERSION>
```

### Multi-Stage Docker Build

The Dockerfile uses multi-stage builds for optimization:

```dockerfile
# Stage 1: Build
FROM golang:1.21-alpine AS builder
WORKDIR /build
COPY . .
RUN go build -ldflags="-s -w" -o ainative-code ./cmd/ainative-code

# Stage 2: Runtime
FROM alpine:latest
COPY --from=builder /build/ainative-code /usr/local/bin/
```

### Docker Build with Custom Arguments

```bash
# Build with version
docker build \
  --build-arg VERSION=v1.2.3 \
  --build-arg BUILD_DATE=$(date -u +%Y-%m-%dT%H:%M:%SZ) \
  -t ainative-code:v1.2.3 \
  .

# Build for specific platform
docker build --platform linux/amd64 -t ainative-code:amd64 .
docker build --platform linux/arm64 -t ainative-code:arm64 .
```

### Run Docker Build

```bash
# Run the Docker container
make docker-run

# Or manually
docker run -it --rm ainative-code:latest

# With volume mounts
docker run -it --rm \
  -v $(pwd)/configs:/home/ainative/.config/ainative-code \
  ainative-code:latest
```

## Release Builds

### Create Release Build

```bash
# Create complete release package
make release

# This creates:
# 1. Binaries for all platforms
# 2. Compressed archives (.tar.gz for Unix, .zip for Windows)
# 3. SHA256 checksums
```

Output structure:
```
build/release/
├── ainative-code-darwin-amd64.tar.gz
├── ainative-code-darwin-arm64.tar.gz
├── ainative-code-linux-amd64.tar.gz
├── ainative-code-linux-arm64.tar.gz
├── ainative-code-windows-amd64.zip
└── checksums.txt
```

### Verify Release Checksums

```bash
# Generate checksums
cd build
shasum -a 256 ainative-code-* > checksums.txt

# Verify checksums
shasum -a 256 -c checksums.txt
```

### Create GitHub Release

```bash
# Tag the release
git tag -a v1.2.3 -m "Release version 1.2.3"
git push origin v1.2.3

# The GitHub Actions workflow will automatically:
# 1. Build all platform binaries
# 2. Create release archives
# 3. Generate checksums
# 4. Create GitHub release
# 5. Upload artifacts
```

## Build Optimization

### Reducing Binary Size

#### 1. Strip Debug Information

```bash
go build -ldflags="-s -w" -o build/ainative-code ./cmd/ainative-code
```

Size reduction: ~30%

#### 2. Use UPX Compression (Optional)

```bash
# Install UPX
brew install upx  # macOS
sudo apt-get install upx  # Linux

# Compress binary
upx --best --lzma build/ainative-code

# Size reduction: ~60-70% (but slower startup)
```

#### 3. Trim Dependencies

```bash
# Remove unused dependencies
go mod tidy

# Check dependency tree
go mod graph | grep ainative-code
```

### Build Speed Optimization

#### 1. Use Build Cache

```bash
# Enable build cache (default in Go 1.10+)
export GOCACHE=$HOME/.cache/go-build

# Check cache location
go env GOCACHE

# Clean cache if needed
go clean -cache
```

#### 2. Parallel Builds

```bash
# Set parallel build count
go build -p 8 -o build/ainative-code ./cmd/ainative-code

# Or use all CPUs
go build -p $(nproc) -o build/ainative-code ./cmd/ainative-code
```

#### 3. Incremental Builds

```bash
# Build only changed packages
go install ./cmd/ainative-code

# The binary is installed to $GOPATH/bin/ainative-code
```

### Compiler Optimizations

```bash
# Enable all optimizations
go build -gcflags="-N -l" -o build/ainative-code ./cmd/ainative-code

# Disable inlining for debugging
go build -gcflags="-l" -o build/ainative-code ./cmd/ainative-code

# Enable race detector (development only)
go build -race -o build/ainative-code ./cmd/ainative-code
```

## Troubleshooting

### Common Build Issues

#### 1. CGO Errors

**Problem**: `gcc: command not found`

**Solution**:
```bash
# macOS
xcode-select --install

# Linux
sudo apt-get install build-essential

# Verify
gcc --version
```

#### 2. Module Errors

**Problem**: `cannot find module`

**Solution**:
```bash
go mod download
go mod tidy
go clean -modcache
```

#### 3. Version Information Not Embedded

**Problem**: `ainative-code version` shows "dev"

**Solution**:
```bash
# Ensure git tags exist
git describe --tags

# Or build with explicit version
make build VERSION=v1.2.3
```

#### 4. Large Binary Size

**Problem**: Binary is unexpectedly large

**Solution**:
```bash
# Use release flags
go build -ldflags="-s -w" ./cmd/ainative-code

# Check binary size
ls -lh build/ainative-code

# Analyze binary
go tool nm build/ainative-code | wc -l
```

#### 5. Cross-Compilation Fails

**Problem**: Cannot cross-compile with CGO

**Solution**:
```bash
# Option 1: Disable CGO for cross-compilation
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build ./cmd/ainative-code

# Option 2: Use Docker for cross-platform builds
docker buildx build --platform linux/amd64,linux/arm64 .

# Option 3: Install cross-compiler toolchain
```

### Build Performance Issues

```bash
# Check what's taking time
go build -x -o build/ainative-code ./cmd/ainative-code 2>&1 | head -20

# Use compiler traces
go build -a -v -o build/ainative-code ./cmd/ainative-code

# Profile the build
time make build
```

## Build Verification

### Post-Build Checks

```bash
# 1. Verify binary exists and is executable
test -x build/ainative-code && echo "Binary is executable"

# 2. Check version information
./build/ainative-code version

# 3. Verify size is reasonable
ls -lh build/ainative-code

# 4. Check dependencies
ldd build/ainative-code  # Linux
otool -L build/ainative-code  # macOS

# 5. Run smoke test
./build/ainative-code --help
```

### Automated Build Testing

```bash
# Run CI build checks locally
make ci

# This runs:
# 1. Format check
# 2. Vet
# 3. Lint
# 4. Tests with coverage
# 5. Build verification
```

## Advanced Build Techniques

### Static Linking

For fully static binaries (no external dependencies):

```bash
# Linux
CGO_ENABLED=1 \
  go build \
  -tags netgo \
  -ldflags='-extldflags "-static" -s -w' \
  -o build/ainative-code-static \
  ./cmd/ainative-code

# Verify static linking
ldd build/ainative-code-static  # Should say "not a dynamic executable"
```

### Plugin Builds

If using Go plugins:

```bash
# Build plugin
go build -buildmode=plugin -o plugin.so ./plugins/example

# Build main application
go build -o build/ainative-code ./cmd/ainative-code
```

### Custom Build Tags

```bash
# Build with specific tags
go build -tags "integration,debug" -o build/ainative-code ./cmd/ainative-code

# Example: Disable certain features
go build -tags "nologging" -o build/ainative-code ./cmd/ainative-code
```

## Continuous Integration Builds

The project uses GitHub Actions for automated builds. See `.github/workflows/ci.yml` and `.github/workflows/release.yml`.

### Local CI Simulation

```bash
# Run all CI checks
make ci

# Simulate release build
make release
```

### Build Matrix

CI builds are tested on:
- **OS**: Ubuntu, macOS, Windows
- **Go versions**: 1.21, 1.22
- **Architectures**: amd64, arm64

## Quick Reference

```bash
# Development
make build                    # Build for current platform
make run                      # Build and run
make clean                    # Clean build artifacts

# Production
make release                  # Create release build
make build-all               # Build all platforms

# Docker
make docker-build            # Build Docker image
make docker-run              # Run in container

# Verification
./build/ainative-code version    # Check version
./build/ainative-code --help     # Verify functionality

# Information
make info                    # Show build configuration
make version                 # Show version info
```

---

**Next**: [Testing Guide](testing.md) | [Debugging Guide](debugging.md) | [Code Style](code-style.md)
