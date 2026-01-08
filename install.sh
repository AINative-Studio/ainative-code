#!/bin/bash

# AINative Code Installation Script
# This script installs the latest version of AINative Code for Linux and macOS

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
REPO="AINative-Studio/ainative-code"
INSTALL_DIR="${INSTALL_DIR:-/usr/local/bin}"
TEMP_DIR="$(mktemp -d)"

# Cleanup on exit
trap 'rm -rf "$TEMP_DIR"' EXIT

# Functions
print_info() {
    echo -e "${BLUE}[INFO]${NC} $1" >&2
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1" >&2
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1" >&2
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1" >&2
}

detect_platform() {
    local os
    local arch

    # Detect OS
    case "$(uname -s)" in
        Linux*)     os="Linux";;
        Darwin*)    os="Darwin";;
        *)
            print_error "Unsupported operating system: $(uname -s)"
            exit 1
            ;;
    esac

    # Detect architecture
    case "$(uname -m)" in
        x86_64|amd64)   arch="x86_64";;
        aarch64|arm64)  arch="arm64";;
        *)
            print_error "Unsupported architecture: $(uname -m)"
            exit 1
            ;;
    esac

    echo "${os}_${arch}"
}

get_latest_version() {
    local latest_url="https://api.github.com/repos/${REPO}/releases/latest"
    local version

    print_info "Fetching latest version..."

    if command -v curl >/dev/null 2>&1; then
        version=$(curl -fsSL "$latest_url" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')
    elif command -v wget >/dev/null 2>&1; then
        version=$(wget -qO- "$latest_url" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')
    else
        print_error "Neither curl nor wget is available. Please install one of them."
        exit 1
    fi

    if [ -z "$version" ]; then
        print_error "Failed to fetch latest version"
        exit 1
    fi

    echo "$version"
}

download_file() {
    local url="$1"
    local output="$2"

    print_info "Downloading from $url..."

    if command -v curl >/dev/null 2>&1; then
        curl -fsSL "$url" -o "$output"
    elif command -v wget >/dev/null 2>&1; then
        wget -qO "$output" "$url"
    else
        print_error "Neither curl nor wget is available"
        exit 1
    fi
}

verify_checksum() {
    local file="$1"
    local expected_checksum="$2"

    print_info "Verifying checksum..."

    if command -v sha256sum >/dev/null 2>&1; then
        local actual_checksum
        actual_checksum=$(sha256sum "$file" | awk '{print $1}')
    elif command -v shasum >/dev/null 2>&1; then
        local actual_checksum
        actual_checksum=$(shasum -a 256 "$file" | awk '{print $1}')
    else
        print_warning "sha256sum or shasum not found. Skipping checksum verification."
        return 0
    fi

    if [ "$actual_checksum" != "$expected_checksum" ]; then
        print_error "Checksum verification failed!"
        print_error "Expected: $expected_checksum"
        print_error "Got:      $actual_checksum"
        exit 1
    fi

    print_success "Checksum verified"
}

install_binary() {
    local binary_path="$1"
    local install_path="${INSTALL_DIR}/ainative-code"

    print_info "Installing to $install_path..."

    # Check if we need sudo
    if [ -w "$INSTALL_DIR" ]; then
        cp "$binary_path" "$install_path"
        chmod +x "$install_path"
    else
        print_warning "Installing to $INSTALL_DIR requires root privileges"
        sudo cp "$binary_path" "$install_path"
        sudo chmod +x "$install_path"
    fi

    print_success "Installation complete!"
}

check_dependencies() {
    local missing_deps=()

    if ! command -v tar >/dev/null 2>&1; then
        missing_deps+=("tar")
    fi

    if ! command -v gzip >/dev/null 2>&1; then
        missing_deps+=("gzip")
    fi

    if [ ${#missing_deps[@]} -ne 0 ]; then
        print_error "Missing required dependencies: ${missing_deps[*]}"
        print_error "Please install them and try again."
        exit 1
    fi
}

main() {
    echo ""
    echo "╔════════════════════════════════════════════╗"
    echo "║      AINative Code Installation Script    ║"
    echo "╚════════════════════════════════════════════╝"
    echo ""

    # Check dependencies
    check_dependencies

    # Detect platform
    local platform
    platform=$(detect_platform)
    print_info "Detected platform: $platform"

    # Get latest version
    local version
    version=$(get_latest_version)
    print_info "Latest version: $version"

    # Construct download URLs
    local version_without_v="${version#v}"

    # Special handling for macOS universal binary
    local archive_name
    if [[ "$platform" == Darwin_* ]]; then
        archive_name="ainative-code_${version_without_v}_Darwin_all.tar.gz"
    else
        archive_name="ainative-code_${version_without_v}_${platform}.tar.gz"
    fi

    local download_url="https://github.com/${REPO}/releases/download/${version}/${archive_name}"
    local checksums_url="https://github.com/${REPO}/releases/download/${version}/checksums.txt"

    # Download archive
    local archive_path="${TEMP_DIR}/${archive_name}"
    download_file "$download_url" "$archive_path"

    # Download and verify checksum
    local checksums_path="${TEMP_DIR}/checksums.txt"
    download_file "$checksums_url" "$checksums_path"

    local expected_checksum
    expected_checksum=$(grep "$archive_name" "$checksums_path" | awk '{print $1}')

    if [ -z "$expected_checksum" ]; then
        print_warning "Could not find checksum for $archive_name. Skipping verification."
    else
        verify_checksum "$archive_path" "$expected_checksum"
    fi

    # Extract archive
    print_info "Extracting archive..."
    tar -xzf "$archive_path" -C "$TEMP_DIR"

    # Find the binary
    local binary_path="${TEMP_DIR}/ainative-code"
    if [ ! -f "$binary_path" ]; then
        print_error "Binary not found in archive"
        exit 1
    fi

    # Install binary
    install_binary "$binary_path"

    # Verify installation
    print_info "Verifying installation..."
    if command -v ainative-code >/dev/null 2>&1; then
        local installed_version
        installed_version=$(ainative-code version --short 2>/dev/null || echo "unknown")
        print_success "AINative Code $installed_version installed successfully!"
    else
        print_warning "Installation complete, but ainative-code is not in PATH"
        print_warning "You may need to add $INSTALL_DIR to your PATH"
    fi

    echo ""
    echo "╔════════════════════════════════════════════╗"
    echo "║           Installation Complete!           ║"
    echo "╚════════════════════════════════════════════╝"
    echo ""
    echo "Get started with:"
    echo "  ainative-code version    # Show version"
    echo "  ainative-code chat       # Start a chat session"
    echo ""
    echo "For more information, visit:"
    echo "  https://github.com/${REPO}"
    echo ""
}

# Run main function
main "$@"
