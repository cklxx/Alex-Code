#!/bin/bash

# Alex CLI Tool Installation Script
# This script detects the OS and architecture, downloads the appropriate binary,
# and installs it to the user's PATH.

set -e

# 配置变量
BINARY_NAME="alex"
GITHUB_REPO="ckl/Alex-Code"
INSTALL_DIR="$HOME/.local/bin"
TMP_DIR="/tmp/alex-install"

# 颜色输出
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 日志函数
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# 检查命令是否存在
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# 检测操作系统和架构
detect_platform() {
    local os
    local arch
    
    # 检测操作系统
    case "$(uname -s)" in
        Linux*)
            os="linux"
            ;;
        Darwin*)
            os="darwin"
            ;;
        CYGWIN*|MINGW*|MSYS*)
            os="windows"
            ;;
        *)
            log_error "Unsupported operating system: $(uname -s)"
            exit 1
            ;;
    esac
    
    # 检测架构
    case "$(uname -m)" in
        x86_64|amd64)
            arch="amd64"
            ;;
        arm64|aarch64)
            arch="arm64"
            ;;
        *)
            log_error "Unsupported architecture: $(uname -m)"
            exit 1
            ;;
    esac
    
    echo "${os}-${arch}"
}

# 获取最新版本
get_latest_version() {
    log_info "Fetching latest version..."
    
    if command_exists curl; then
        latest_version=$(curl -s "https://api.github.com/repos/${GITHUB_REPO}/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')
    elif command_exists wget; then
        latest_version=$(wget -qO- "https://api.github.com/repos/${GITHUB_REPO}/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')
    else
        log_error "Neither curl nor wget is available. Please install one of them and try again."
        exit 1
    fi
    
    if [ -z "$latest_version" ]; then
        log_error "Failed to fetch latest version"
        exit 1
    fi
    
    echo "$latest_version"
}

# 下载文件
download_file() {
    local url="$1"
    local output="$2"
    
    log_info "Downloading from: $url"
    
    if command_exists curl; then
        curl -sSfL "$url" -o "$output"
    elif command_exists wget; then
        wget -q "$url" -O "$output"
    else
        log_error "Neither curl nor wget is available. Please install one of them and try again."
        exit 1
    fi
}

# 验证下载的文件
verify_binary() {
    local binary_path="$1"
    
    if [ ! -f "$binary_path" ]; then
        log_error "Downloaded binary not found: $binary_path"
        return 1
    fi
    
    if [ ! -x "$binary_path" ]; then
        log_error "Downloaded binary is not executable: $binary_path"
        return 1
    fi
    
    # 尝试运行 --version 检查
    if ! "$binary_path" --version >/dev/null 2>&1; then
        log_warning "Binary may not be working correctly (--version failed)"
        return 1
    fi
    
    return 0
}

# 安装到系统
install_binary() {
    local binary_path="$1"
    local platform="$2"
    
    # 确保安装目录存在
    mkdir -p "$INSTALL_DIR"
    
    # 复制二进制文件
    local target_path="$INSTALL_DIR/$BINARY_NAME"
    cp "$binary_path" "$target_path"
    chmod +x "$target_path"
    
    log_success "Binary installed to: $target_path"
    
    # 检查PATH
    if ! echo ":$PATH:" | grep -q ":$INSTALL_DIR:"; then
        log_warning "$INSTALL_DIR is not in your PATH"
        log_info "Add the following line to your shell profile (~/.bashrc, ~/.zshrc, etc.):"
        echo "export PATH=\"\$PATH:$INSTALL_DIR\""
        echo ""
        log_info "Or run the following command to add it temporarily:"
        echo "export PATH=\"\$PATH:$INSTALL_DIR\""
        echo ""
    fi
    
    # 验证安装
    if command_exists "$BINARY_NAME"; then
        log_success "Installation successful! You can now use '$BINARY_NAME'"
        "$BINARY_NAME" --version
    else
        log_warning "Installation completed, but '$BINARY_NAME' is not found in PATH"
        log_info "You may need to restart your shell or update your PATH"
    fi
}

# 主安装流程
main() {
    log_info "Starting Alex CLI installation..."
    
    # 检测平台
    platform=$(detect_platform)
    log_info "Detected platform: $platform"
    
    # 获取最新版本
    version=$(get_latest_version)
    log_info "Latest version: $version"
    
    # 构建下载URL
    binary_suffix=""
    case "$platform" in
        windows-*)
            binary_suffix=".exe"
            ;;
    esac
    
    binary_name="${BINARY_NAME}-${platform}${binary_suffix}"
    download_url="https://github.com/${GITHUB_REPO}/releases/download/${version}/${binary_name}"
    
    # 创建临时目录
    rm -rf "$TMP_DIR"
    mkdir -p "$TMP_DIR"
    
    # 下载二进制文件
    binary_path="$TMP_DIR/$binary_name"
    download_file "$download_url" "$binary_path"
    
    # 使二进制文件可执行
    chmod +x "$binary_path"
    
    # 验证二进制文件
    if ! verify_binary "$binary_path"; then
        log_error "Binary verification failed"
        exit 1
    fi
    
    # 安装到系统
    install_binary "$binary_path" "$platform"
    
    # 清理临时文件
    rm -rf "$TMP_DIR"
    
    log_success "Alex CLI has been successfully installed!"
    log_info "Run 'alex --help' to get started"
}

# 处理脚本参数
while [[ $# -gt 0 ]]; do
    case $1 in
        --version)
            VERSION="$2"
            shift 2
            ;;
        --repo)
            GITHUB_REPO="$2"
            shift 2
            ;;
        --install-dir)
            INSTALL_DIR="$2"
            shift 2
            ;;
        --help)
            echo "Alex CLI Installation Script"
            echo ""
            echo "Usage: $0 [OPTIONS]"
            echo ""
            echo "Options:"
            echo "  --version VERSION     Install specific version (default: latest)"
            echo "  --repo REPO          GitHub repository (default: $GITHUB_REPO)"
            echo "  --install-dir DIR    Installation directory (default: $INSTALL_DIR)"
            echo "  --help               Show this help message"
            echo ""
            echo "Examples:"
            echo "  $0                           # Install latest version"
            echo "  $0 --version v1.0.0         # Install specific version"
            echo "  $0 --install-dir /usr/local/bin  # Install to custom directory"
            exit 0
            ;;
        *)
            log_error "Unknown option: $1"
            echo "Use --help for usage information"
            exit 1
            ;;
    esac
done

# 运行主函数
main