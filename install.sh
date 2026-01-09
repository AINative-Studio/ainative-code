#!/bin/bash

# AINative Code Installation Script
# This script installs the latest version of AINative Code for Linux and macOS

# Note: We don't use 'set -e' to allow graceful error handling and fallback installation

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
        if ! curl -fsSL "$url" -o "$output"; then
            print_error "Failed to download from $url"
            return 1
        fi
    elif command -v wget >/dev/null 2>&1; then
        if ! wget -qO "$output" "$url"; then
            print_error "Failed to download from $url"
            return 1
        fi
    else
        print_error "Neither curl nor wget is available"
        print_error "Please install curl or wget and try again"
        return 1
    fi
    return 0
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
        print_success "Installation complete!"
        return 0
    else
        print_warning "Installing to $INSTALL_DIR requires root privileges"

        # Try sudo installation
        if sudo -n true 2>/dev/null; then
            # Passwordless sudo available
            if sudo cp "$binary_path" "$install_path" && sudo chmod +x "$install_path"; then
                print_success "Installation complete!"
                return 0
            fi
        else
            # Need password for sudo
            print_info "Please enter your password for sudo access..."
            if sudo cp "$binary_path" "$install_path" 2>/dev/null && sudo chmod +x "$install_path" 2>/dev/null; then
                print_success "Installation complete!"
                return 0
            fi
        fi

        # Sudo failed or was cancelled - fall back to user directory
        print_warning "Sudo installation failed or was cancelled"
        return 1
    fi
}

install_to_user_directory() {
    local binary_path="$1"

    # Try ~/.local/bin first (XDG standard), then ~/bin
    local user_dirs=("$HOME/.local/bin" "$HOME/bin")

    for user_dir in "${user_dirs[@]}"; do
        print_info "Attempting fallback installation to $user_dir..."

        # Create directory if it doesn't exist
        if [ ! -d "$user_dir" ]; then
            if mkdir -p "$user_dir" 2>/dev/null; then
                print_info "Created directory $user_dir"
            else
                print_warning "Could not create directory $user_dir"
                continue
            fi
        fi

        # Check if directory is writable
        if [ -w "$user_dir" ]; then
            local install_path="$user_dir/ainative-code"
            if cp "$binary_path" "$install_path" && chmod +x "$install_path"; then
                print_success "Successfully installed to $install_path"
                INSTALL_DIR="$user_dir"
                return 0
            else
                print_warning "Failed to install to $user_dir"
            fi
        else
            print_warning "Directory $user_dir is not writable"
        fi
    done

    # All fallback attempts failed
    print_error "Could not install to any location"
    print_error "Tried: /usr/local/bin, ~/.local/bin, ~/bin"
    print_error ""
    print_error "You can manually install by running:"
    print_error "  mkdir -p ~/.local/bin"
    print_error "  cp $binary_path ~/.local/bin/ainative-code"
    print_error "  chmod +x ~/.local/bin/ainative-code"
    print_error "  export PATH=\"\$HOME/.local/bin:\$PATH\""
    return 1
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

check_path() {
    # Check if INSTALL_DIR is in PATH
    case ":$PATH:" in
        *":$INSTALL_DIR:"*)
            return 0
            ;;
        *)
            return 1
            ;;
    esac
}

detect_shell_rc() {
    # Detect which shell rc file to use
    if [ -n "$BASH_VERSION" ]; then
        if [ -f "$HOME/.bashrc" ]; then
            echo "$HOME/.bashrc"
        elif [ -f "$HOME/.bash_profile" ]; then
            echo "$HOME/.bash_profile"
        else
            echo "$HOME/.profile"
        fi
    elif [ -n "$ZSH_VERSION" ]; then
        echo "$HOME/.zshrc"
    else
        # Try to detect from SHELL environment variable
        case "$SHELL" in
            */zsh)
                echo "$HOME/.zshrc"
                ;;
            */bash)
                if [ -f "$HOME/.bashrc" ]; then
                    echo "$HOME/.bashrc"
                elif [ -f "$HOME/.bash_profile" ]; then
                    echo "$HOME/.bash_profile"
                else
                    echo "$HOME/.profile"
                fi
                ;;
            *)
                echo "$HOME/.profile"
                ;;
        esac
    fi
}

setup_path() {
    # Only set up PATH if INSTALL_DIR is not /usr/local/bin (which is usually in PATH)
    if [ "$INSTALL_DIR" = "/usr/local/bin" ]; then
        return 0
    fi

    # Check if already in PATH
    if check_path; then
        print_info "Installation directory is already in PATH"
        return 0
    fi

    # Detect shell rc file
    local shell_rc
    shell_rc=$(detect_shell_rc)

    echo ""
    print_warning "Installation directory $INSTALL_DIR is not in your PATH"
    echo ""
    print_info "To use ainative-code, you need to add it to your PATH"
    echo ""
    echo "Run these commands to add it automatically:"
    echo ""
    echo "  echo 'export PATH=\"$INSTALL_DIR:\$PATH\"' >> $shell_rc"
    echo "  source $shell_rc"
    echo ""
    echo "Or add this line to your $shell_rc manually and restart your terminal:"
    echo ""
    echo "  export PATH=\"$INSTALL_DIR:\$PATH\""
    echo ""
    echo "After updating your PATH, verify the installation with:"
    echo "  ainative-code version"
    echo ""
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
    if ! download_file "$download_url" "$archive_path"; then
        print_error "Failed to download AINative Code archive"
        exit 1
    fi

    # Download and verify checksum
    local checksums_path="${TEMP_DIR}/checksums.txt"
    if ! download_file "$checksums_url" "$checksums_path"; then
        print_warning "Failed to download checksums file, skipping verification"
    else
        local expected_checksum
        expected_checksum=$(grep "$archive_name" "$checksums_path" 2>/dev/null | awk '{print $1}')

        if [ -z "$expected_checksum" ]; then
            print_warning "Could not find checksum for $archive_name. Skipping verification."
        else
            verify_checksum "$archive_path" "$expected_checksum"
        fi
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
    if ! install_binary "$binary_path"; then
        # Primary installation failed, try user directory fallback
        print_info ""
        print_info "Falling back to user directory installation..."
        if ! install_to_user_directory "$binary_path"; then
            exit 1
        fi
    fi

    # Verify installation
    print_info "Verifying installation..."
    if command -v ainative-code >/dev/null 2>&1; then
        local installed_version
        installed_version=$(ainative-code version --short 2>/dev/null || echo "unknown")
        print_success "AINative Code $installed_version installed successfully!"

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
    else
        echo ""
        echo "╔════════════════════════════════════════════╗"
        echo "║           Installation Complete!           ║"
        echo "╚════════════════════════════════════════════╝"
        echo ""

        # Set up PATH instructions
        setup_path

        echo "After setting up your PATH, get started with:"
        echo "  ainative-code version    # Show version"
        echo "  ainative-code chat       # Start a chat session"
        echo ""
        echo "For more information, visit:"
        echo "  https://github.com/${REPO}"
        echo ""
    fi
}

# Run main function
main "$@"
