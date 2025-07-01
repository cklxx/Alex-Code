#!/bin/bash

# Installation script for Deep Coding Agent
# This script downloads and installs the latest release

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

print_status() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Configuration
REPO="your-org/deep-coding"  # Replace with actual repository
BINARY_NAME="deep-coding-agent"
INSTALL_DIR="${INSTALL_DIR:-/usr/local/bin}"
FORCE_INSTALL="${FORCE_INSTALL:-false}"

# Detect platform
detect_platform() {
    local os
    local arch
    
    case "$(uname -s)" in
        Linux*) os="linux" ;;
        Darwin*) os="darwin" ;;
        CYGWIN*|MINGW*|MSYS*) os="windows" ;;
        *) 
            print_error "Unsupported operating system: $(uname -s)"
            exit 1
            ;;
    esac
    
    case "$(uname -m)" in
        x86_64|amd64) arch="amd64" ;;
        arm64|aarch64) arch="arm64" ;;
        *)
            print_error "Unsupported architecture: $(uname -m)"
            exit 1
            ;;
    esac
    
    PLATFORM="${os}-${arch}"
    print_status "Detected platform: $PLATFORM"
}

# Get latest version from GitHub releases
get_latest_version() {
    print_status "Fetching latest version..."
    
    if command -v curl &> /dev/null; then
        VERSION=$(curl -s "https://api.github.com/repos/$REPO/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')
    elif command -v wget &> /dev/null; then
        VERSION=$(wget -qO- "https://api.github.com/repos/$REPO/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')
    else
        print_error "Neither curl nor wget is available. Please install one of them."
        exit 1
    fi
    
    if [ -z "$VERSION" ]; then
        print_error "Failed to get latest version"
        exit 1
    fi
    
    print_status "Latest version: $VERSION"
}

# Download and install binary
install_binary() {
    local download_url
    local temp_file
    local binary_name
    
    if [[ "$PLATFORM" == "windows"* ]]; then
        binary_name="${BINARY_NAME}-${VERSION}-${PLATFORM}.exe.zip"
        download_url="https://github.com/$REPO/releases/download/$VERSION/$binary_name"
    else
        binary_name="${BINARY_NAME}-${VERSION}-${PLATFORM}.tar.gz"
        download_url="https://github.com/$REPO/releases/download/$VERSION/$binary_name"
    fi
    
    print_status "Downloading $binary_name..."
    
    temp_file=$(mktemp)
    
    if command -v curl &> /dev/null; then
        curl -sL "$download_url" -o "$temp_file"
    elif command -v wget &> /dev/null; then
        wget -q "$download_url" -O "$temp_file"
    else
        print_error "Neither curl nor wget is available"
        exit 1
    fi
    
    # Verify download
    if [ ! -f "$temp_file" ] || [ ! -s "$temp_file" ]; then
        print_error "Download failed or file is empty"
        exit 1
    fi
    
    print_status "Extracting binary..."
    
    # Create temporary directory for extraction
    temp_dir=$(mktemp -d)
    cd "$temp_dir"
    
    if [[ "$PLATFORM" == "windows"* ]]; then
        unzip -q "$temp_file"
        extracted_binary="${BINARY_NAME}-${VERSION}-${PLATFORM}.exe"
    else
        tar -xzf "$temp_file"
        extracted_binary="${BINARY_NAME}-${VERSION}-${PLATFORM}"
    fi
    
    if [ ! -f "$extracted_binary" ]; then
        print_error "Extracted binary not found: $extracted_binary"
        ls -la
        exit 1
    fi
    
    # Install binary
    print_status "Installing to $INSTALL_DIR..."
    
    # Check if install directory exists and is writable
    if [ ! -d "$INSTALL_DIR" ]; then
        print_status "Creating install directory: $INSTALL_DIR"
        sudo mkdir -p "$INSTALL_DIR"
    fi
    
    if [ ! -w "$INSTALL_DIR" ]; then
        print_status "Using sudo to install to $INSTALL_DIR"
        sudo cp "$extracted_binary" "$INSTALL_DIR/$BINARY_NAME"
        sudo chmod +x "$INSTALL_DIR/$BINARY_NAME"
    else
        cp "$extracted_binary" "$INSTALL_DIR/$BINARY_NAME"
        chmod +x "$INSTALL_DIR/$BINARY_NAME"
    fi
    
    # Cleanup
    cd /
    rm -rf "$temp_dir" "$temp_file"
    
    print_success "Installation completed successfully!"
}

# Verify installation
verify_installation() {
    print_status "Verifying installation..."
    
    if command -v "$BINARY_NAME" &> /dev/null; then
        local installed_version
        installed_version=$("$BINARY_NAME" version 2>/dev/null | head -1 || echo "unknown")
        print_success "✓ $BINARY_NAME is installed and working"
        print_status "Installed version: $installed_version"
    else
        print_warning "Binary installed but not found in PATH"
        print_warning "You may need to add $INSTALL_DIR to your PATH"
        print_warning "Or run the binary directly: $INSTALL_DIR/$BINARY_NAME"
    fi
}

# Check if already installed
check_existing() {
    if command -v "$BINARY_NAME" &> /dev/null && [ "$FORCE_INSTALL" != "true" ]; then
        local existing_version
        existing_version=$("$BINARY_NAME" version 2>/dev/null | head -1 || echo "unknown")
        print_warning "$BINARY_NAME is already installed: $existing_version"
        echo -n "Do you want to reinstall? (y/N): "
        read -r response
        if [[ ! $response =~ ^[Yy]$ ]]; then
            print_status "Installation cancelled"
            exit 0
        fi
    fi
}

# Show usage
show_usage() {
    cat << EOF
Deep Coding Agent Installation Script

Usage: $0 [options]

Options:
    -d, --dir DIR       Install directory (default: /usr/local/bin)
    -f, --force         Force installation even if already installed
    -h, --help          Show this help message
    -v, --version VER   Install specific version (default: latest)

Environment Variables:
    INSTALL_DIR         Install directory
    FORCE_INSTALL       Force installation (true/false)

Examples:
    $0                          # Install latest version to /usr/local/bin
    $0 -d ~/.local/bin          # Install to custom directory
    $0 -f                       # Force reinstall
    $0 -v v1.2.3               # Install specific version

EOF
}

# Parse command line arguments
parse_args() {
    while [[ $# -gt 0 ]]; do
        case $1 in
            -d|--dir)
                INSTALL_DIR="$2"
                shift 2
                ;;
            -f|--force)
                FORCE_INSTALL="true"
                shift
                ;;
            -v|--version)
                VERSION="$2"
                shift 2
                ;;
            -h|--help)
                show_usage
                exit 0
                ;;
            *)
                print_error "Unknown option: $1"
                show_usage
                exit 1
                ;;
        esac
    done
}

# Main installation flow
main() {
    print_status "Deep Coding Agent Installation Script"
    print_status "======================================"
    
    parse_args "$@"
    detect_platform
    check_existing
    
    if [ -z "$VERSION" ]; then
        get_latest_version
    else
        print_status "Using specified version: $VERSION"
    fi
    
    install_binary
    verify_installation
    
    echo ""
    print_success "🎉 Deep Coding Agent installation completed!"
    echo ""
    print_status "Quick start:"
    print_status "  $BINARY_NAME --help              # Show help"
    print_status "  $BINARY_NAME version             # Check version"
    print_status "  $BINARY_NAME analyze .           # Analyze current directory"
    print_status "  $BINARY_NAME generate 'API' go   # Generate Go code"
    echo ""
    print_status "Documentation: https://github.com/$REPO"
}

# Check dependencies
check_dependencies() {
    local missing_deps=()
    
    if ! command -v curl &> /dev/null && ! command -v wget &> /dev/null; then
        missing_deps+=("curl or wget")
    fi
    
    if ! command -v tar &> /dev/null; then
        missing_deps+=("tar")
    fi
    
    if [[ "$PLATFORM" == "windows"* ]] && ! command -v unzip &> /dev/null; then
        missing_deps+=("unzip")
    fi
    
    if [ ${#missing_deps[@]} -ne 0 ]; then
        print_error "Missing required dependencies:"
        for dep in "${missing_deps[@]}"; do
            print_error "  - $dep"
        done
        exit 1
    fi
}

# Run main function with error handling
if [ "${BASH_SOURCE[0]}" = "${0}" ]; then
    check_dependencies
    main "$@"
fi