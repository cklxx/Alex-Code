#!/bin/bash

# Alex CLI Tool Installation Script
# This script detects the OS and architecture, downloads the appropriate binary,
# and installs it to the user's PATH.

set -e

# Check if we're running under bash, if not, try to re-exec with bash
if [ -z "$BASH_VERSION" ]; then
    # Check if this is a piped execution (common case: curl | sh)
    if [ ! -t 0 ]; then
        # Script is being piped, we can't re-exec, so continue with POSIX-compatible mode
        echo "Info: Running in POSIX-compatible mode (piped execution detected)"
    elif [ -f "$0" ] && command -v bash >/dev/null 2>&1; then
        # Script is a file and bash is available, re-execute with bash
        exec bash "$0" "$@"
    else
        # Continue with current shell
        echo "Info: Continuing with POSIX-compatible shell"
    fi
fi

# 配置变量
BINARY_NAME="alex"
GITHUB_REPO="cklxx/Alex-Code"
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
    
    local api_url="https://api.github.com/repos/${GITHUB_REPO}/releases/latest"
    local response
    
    if command_exists curl; then
        response=$(curl -s --fail --connect-timeout 10 --max-time 30 "$api_url" 2>/dev/null)
        if [ $? -ne 0 ]; then
            log_error "Failed to connect to GitHub API. Please check your internet connection."
            log_info "You can also specify a version manually with --version flag"
            exit 1
        fi
    elif command_exists wget; then
        response=$(wget -qO- --timeout=30 --tries=2 "$api_url" 2>/dev/null)
        if [ $? -ne 0 ]; then
            log_error "Failed to connect to GitHub API. Please check your internet connection."
            log_info "You can also specify a version manually with --version flag"
            exit 1
        fi
    else
        log_error "Neither curl nor wget is available. Please install one of them and try again."
        exit 1
    fi
    
    # 更robust的JSON解析
    latest_version=$(echo "$response" | grep '"tag_name"' | head -n1 | sed 's/.*"tag_name".*:.*"\([^"]*\)".*/\1/')
    
    if [ -z "$latest_version" ]; then
        log_error "Failed to parse latest version from GitHub API response"
        log_info "The API response might be rate-limited or malformed"
        log_info "You can specify a version manually with --version flag"
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
        if ! curl -sSfL --connect-timeout 10 --max-time 300 "$url" -o "$output"; then
            log_error "Download failed with curl"
            return 1
        fi
    elif command_exists wget; then
        if ! wget -q --timeout=300 --tries=3 "$url" -O "$output"; then
            log_error "Download failed with wget"
            return 1
        fi
    else
        log_error "Neither curl nor wget is available. Please install one of them and try again."
        exit 1
    fi
    
    # 验证文件是否下载成功
    if [ ! -f "$output" ] || [ ! -s "$output" ]; then
        log_error "Downloaded file is empty or missing: $output"
        return 1
    fi
    
    return 0
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

# 安装系统依赖
install_dependencies() {
    log_info "Installing system dependencies..."
    
    # 检测并安装ripgrep
    if ! command_exists rg; then
        log_info "Installing ripgrep..."
        case "$(uname -s)" in
            Darwin*)
                if command_exists brew; then
                    if brew install ripgrep; then
                        log_success "ripgrep installed via Homebrew"
                    else
                        log_warning "Failed to install ripgrep via Homebrew"
                    fi
                elif command_exists port; then
                    if sudo port install ripgrep; then
                        log_success "ripgrep installed via MacPorts"
                    else
                        log_warning "Failed to install ripgrep via MacPorts"
                    fi
                else
                    log_warning "Neither brew nor port found. Please install ripgrep manually."
                    log_info "Install Homebrew: /bin/bash -c \"\$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)\""
                fi
                ;;
            Linux*)
                if command_exists apt; then
                    if sudo apt update && sudo apt install -y ripgrep; then
                        log_success "ripgrep installed via apt"
                    else
                        log_warning "Failed to install ripgrep via apt"
                    fi
                elif command_exists dnf; then
                    if sudo dnf install -y ripgrep; then
                        log_success "ripgrep installed via dnf"
                    else
                        log_warning "Failed to install ripgrep via dnf"
                    fi
                elif command_exists yum; then
                    if sudo yum install -y ripgrep; then
                        log_success "ripgrep installed via yum"
                    else
                        log_warning "Failed to install ripgrep via yum"
                    fi
                elif command_exists pacman; then
                    if sudo pacman -S --noconfirm ripgrep; then
                        log_success "ripgrep installed via pacman"
                    else
                        log_warning "Failed to install ripgrep via pacman"
                    fi
                elif command_exists zypper; then
                    if sudo zypper install -y ripgrep; then
                        log_success "ripgrep installed via zypper"
                    else
                        log_warning "Failed to install ripgrep via zypper"
                    fi
                else
                    log_warning "Package manager not found. Please install ripgrep manually."
                    log_info "Download from: https://github.com/BurntSushi/ripgrep/releases"
                fi
                ;;
            *)
                log_warning "Unsupported OS for automatic dependency installation. Please install ripgrep manually."
                ;;
        esac
        
        # 验证安装
        if command_exists rg; then
            log_success "ripgrep installed successfully"
        else
            log_warning "ripgrep installation may have failed"
        fi
    else
        log_info "ripgrep is already installed"
    fi
}

# 主安装流程
main() {
    log_info "Starting Alex CLI installation..."
    
    # 安装依赖
    install_dependencies
    
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
    if ! download_file "$download_url" "$binary_path"; then
        log_error "Failed to download binary from: $download_url"
        log_info "Please check:"
        log_info "1. Your internet connection"
        log_info "2. The release exists at: https://github.com/${GITHUB_REPO}/releases"
        log_info "3. Try specifying a different version with --version flag"
        exit 1
    fi
    
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
while [ $# -gt 0 ]; do
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