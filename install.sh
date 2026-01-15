#!/usr/bin/env bash
# AINative Code Installation Script for Linux and macOS
# This script installs the latest version of AINative Code

set -e

# Configuration
REPO="AINative-Studio/ainative-code"
BINARY_NAME="ainative-code"
INSTALL_DIR="${INSTALL_DIR:-/usr/local/bin}"
VERSION="${VERSION:-latest}"

# Colors
COLOR_RESET="\033[0m"
COLOR_INFO="\033[36m"
COLOR_SUCCESS="\033[32m"
COLOR_WARNING="\033[33m"
COLOR_ERROR="\033[31m"

# Functions
log_info() {
    echo -e "${COLOR_INFO}[Info]${COLOR_RESET} $1"
}

log_success() {
    echo -e "${COLOR_SUCCESS}[Success]${COLOR_RESET} $1"
}

log_warning() {
    echo -e "${COLOR_WARNING}[Warning]${COLOR_RESET} $1"
}

log_error() {
    echo -e "${COLOR_ERROR}[Error]${COLOR_RESET} $1"
}

check_dependencies() {
    local missing_deps=()

    for cmd in curl tar shasum; do
        if ! command -v "$cmd" &> /dev/null; then
            missing_deps+=("$cmd")
        fi
    done

    if [ ${#missing_deps[@]} -gt 0 ]; then
        log_error "Missing required dependencies: ${missing_deps[*]}"
        log_error "Please install them and try again"
        exit 1
    fi
}

get_platform() {
    local os
    local arch

    # Detect OS
    case "$(uname -s)" in
        Darwin*)
            os="Darwin"
            # macOS releases use universal binary
            arch="all"
            ;;
        Linux*)
            os="Linux"
            # Detect architecture for Linux
            case "$(uname -m)" in
                x86_64|amd64)
                    arch="x86_64"
                    ;;
                aarch64|arm64)
                    arch="arm64"
                    ;;
                *)
                    log_error "Unsupported architecture: $(uname -m)"
                    exit 1
                    ;;
            esac
            ;;
        *)
            log_error "Unsupported operating system: $(uname -s)"
            exit 1
            ;;
    esac

    echo "${os}_${arch}"
}

get_latest_version() {
    log_info "Fetching latest version..."

    local api_url="https://api.github.com/repos/${REPO}/releases/latest"
    local version

    version=$(curl -fsSL "$api_url" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')

    if [ -z "$version" ]; then
        log_error "Failed to fetch latest version"
        exit 1
    fi

    echo "$version"
}

download_file() {
    local url="$1"
    local output="$2"

    log_info "Downloading from $url..."

    if ! curl -fsSL "$url" -o "$output"; then
        log_error "Download failed"
        exit 1
    fi

    if [ ! -f "$output" ]; then
        log_error "Download failed: file not found at $output"
        exit 1
    fi
}

verify_checksum() {
    local file="$1"
    local expected="$2"

    log_info "Verifying checksum..."

    local actual
    actual=$(shasum -a 256 "$file" | awk '{print $1}')

    if [ "$actual" != "$expected" ]; then
        log_error "Checksum verification failed!"
        log_error "Expected: $expected"
        log_error "Got:      $actual"
        exit 1
    fi

    log_success "Checksum verified"
}

extract_archive() {
    local archive="$1"
    local dest="$2"

    log_info "Extracting archive..."

    if ! tar -xzf "$archive" -C "$dest"; then
        log_error "Failed to extract archive"
        exit 1
    fi
}

check_write_permission() {
    local dir="$1"

    if [ ! -w "$dir" ]; then
        log_warning "No write permission to $dir"
        log_info "Attempting to use sudo for installation..."
        USE_SUDO=true
    else
        USE_SUDO=false
    fi
}

install_binary() {
    local binary_path="$1"
    local install_dir="$2"

    log_info "Installing to $install_dir..."

    # Check if we need sudo
    check_write_permission "$install_dir"

    # Create install directory if it doesn't exist
    if [ ! -d "$install_dir" ]; then
        if [ "$USE_SUDO" = true ]; then
            sudo mkdir -p "$install_dir"
        else
            mkdir -p "$install_dir"
        fi
    fi

    # Copy binary
    if [ "$USE_SUDO" = true ]; then
        sudo cp "$binary_path" "$install_dir/$BINARY_NAME"
        sudo chmod +x "$install_dir/$BINARY_NAME"
    else
        cp "$binary_path" "$install_dir/$BINARY_NAME"
        chmod +x "$install_dir/$BINARY_NAME"
    fi
}

test_installation() {
    log_info "Verifying installation..."

    if command -v "$BINARY_NAME" &> /dev/null; then
        local version
        version=$($BINARY_NAME version --short 2>&1 || echo "unknown")
        log_success "AINative Code $version installed successfully!"
        return 0
    else
        log_warning "Installation verification failed"
        log_warning "You may need to add $INSTALL_DIR to your PATH"
        log_warning "Add this to your shell profile (~/.bashrc, ~/.zshrc, etc.):"
        log_warning "  export PATH=\"\$PATH:$INSTALL_DIR\""
        return 1
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

    # Get platform
    local platform
    platform=$(get_platform)
    log_info "Detected platform: $platform"

    # Get version
    if [ "$VERSION" = "latest" ]; then
        VERSION=$(get_latest_version)
    fi
    log_info "Installing version: $VERSION"

    # Construct download URLs
    local version_without_v="${VERSION#v}"
    local archive_name="ainative-code_${version_without_v}_${platform}.tar.gz"
    local download_url="https://github.com/${REPO}/releases/download/${VERSION}/${archive_name}"
    local checksums_url="https://github.com/${REPO}/releases/download/${VERSION}/checksums.txt"

    # Create temp directory
    local temp_dir
    temp_dir=$(mktemp -d -t ainative-code-install.XXXXXX)

    # Ensure cleanup on exit
    trap "rm -rf '$temp_dir'" EXIT

    # Download archive
    local archive_path="$temp_dir/$archive_name"
    download_file "$download_url" "$archive_path"

    # Download checksums
    local checksums_path="$temp_dir/checksums.txt"
    download_file "$checksums_url" "$checksums_path"

    # Verify checksum
    local expected_checksum
    expected_checksum=$(grep "$archive_name" "$checksums_path" | awk '{print $1}')

    if [ -z "$expected_checksum" ]; then
        log_warning "Could not find checksum for $archive_name. Skipping verification."
    else
        verify_checksum "$archive_path" "$expected_checksum"
    fi

    # Extract archive
    extract_archive "$archive_path" "$temp_dir"

    # Find binary (might be in a subdirectory)
    local binary_path
    if [ -f "$temp_dir/$BINARY_NAME" ]; then
        binary_path="$temp_dir/$BINARY_NAME"
    elif [ -f "$temp_dir/ainative-code/$BINARY_NAME" ]; then
        binary_path="$temp_dir/ainative-code/$BINARY_NAME"
    else
        log_error "Binary not found in archive"
        exit 1
    fi

    # Install binary
    install_binary "$binary_path" "$INSTALL_DIR"

    # Test installation
    test_installation

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
main
