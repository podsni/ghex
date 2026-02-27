#!/bin/bash
# GHEX Installer Script
# Usage: curl -sSL https://raw.githubusercontent.com/dwirx/ghex/main/scripts/install.sh | bash

set -e

REPO="dwirx/ghex"
BINARY_NAME="ghex"
INSTALL_DIR="/usr/local/bin"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

print_banner() {
    echo -e "${BLUE}"
    echo "  ██████╗ ██╗  ██╗███████╗██╗  ██╗"
    echo " ██╔════╝ ██║  ██║██╔════╝╚██╗██╔╝"
    echo " ██║  ███╗███████║█████╗   ╚███╔╝ "
    echo " ██║   ██║██╔══██║██╔══╝   ██╔██╗ "
    echo " ╚██████╔╝██║  ██║███████╗██╔╝ ██╗"
    echo "  ╚═════╝ ╚═╝  ╚═╝╚══════╝╚═╝  ╚═╝"
    echo -e "${NC}"
    echo "GitHub Account Switcher & Universal Downloader"
    echo ""
}

info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

error() {
    echo -e "${RED}[ERROR]${NC} $1"
    exit 1
}

detect_os() {
    OS=$(uname -s | tr '[:upper:]' '[:lower:]')
    case "$OS" in
        linux*)  OS="linux" ;;
        darwin*) OS="darwin" ;;
        mingw*|msys*|cygwin*) OS="windows" ;;
        *) error "Unsupported OS: $OS" ;;
    esac
    echo "$OS"
}

detect_arch() {
    ARCH=$(uname -m)
    case "$ARCH" in
        x86_64|amd64) ARCH="amd64" ;;
        aarch64|arm64) ARCH="arm64" ;;
        *) error "Unsupported architecture: $ARCH" ;;
    esac
    echo "$ARCH"
}

get_latest_version() {
    if command -v jq &> /dev/null; then
        curl -sSL "https://api.github.com/repos/${REPO}/releases/latest" | jq -r '.tag_name'
    else
        curl -sSL "https://api.github.com/repos/${REPO}/releases/latest" | \
            grep '"tag_name":' | \
            sed 's/.*"tag_name": *"\([^"]*\)".*/\1/'
    fi
}

download_and_install() {
    local version=$1
    local os=$2
    local arch=$3
    
    local ext="tar.gz"
    if [ "$os" = "windows" ]; then
        ext="zip"
    fi
    
    local filename="ghex-${os}-${arch}.${ext}"
    local url="https://github.com/${REPO}/releases/download/${version}/${filename}"
    
    info "Downloading ${filename}..."
    
    local tmp_dir
    tmp_dir=$(mktemp -d)
    
    if ! curl -sSL -o "${tmp_dir}/${filename}" "$url"; then
        rm -rf "$tmp_dir"
        error "Failed to download from $url"
    fi
    
    info "Extracting..."
    if [ "$ext" = "tar.gz" ]; then
        tar -xzf "${tmp_dir}/${filename}" -C "$tmp_dir"
    else
        unzip -q "${tmp_dir}/${filename}" -d "$tmp_dir"
    fi
    
    # Binary name from goreleaser is just "ghex" (or "ghex.exe" on windows)
    local binary="ghex"
    if [ "$os" = "windows" ]; then
        binary="ghex.exe"
    fi
    
    # Debug: show extracted files
    info "Extracted files:"
    ls -la "$tmp_dir"
    
    if [ ! -f "${tmp_dir}/${binary}" ]; then
        rm -rf "$tmp_dir"
        error "Binary '$binary' not found after extraction. Files in archive: $(ls -1 "$tmp_dir")"
    fi
    
    info "Installing to ${INSTALL_DIR}..."
    
    if [ "$os" = "windows" ]; then
        # On Windows (Git Bash/MSYS2), install to user's local app data
        local win_install_dir="${LOCALAPPDATA:-$USERPROFILE/AppData/Local}/ghex"
        mkdir -p "$win_install_dir"
        mv "${tmp_dir}/${binary}" "$win_install_dir/$BINARY_NAME"
        # Add to PATH via user profile if not already there
        if ! echo "$PATH" | grep -q "$win_install_dir"; then
            echo "export PATH=\"\$PATH:$win_install_dir\"" >> "$HOME/.bashrc"
            info "Added $win_install_dir to PATH in ~/.bashrc"
            info "Please restart your terminal or run: source ~/.bashrc"
        fi
        INSTALL_DIR="$win_install_dir"
    elif [ -w "$INSTALL_DIR" ]; then
        mv "${tmp_dir}/${binary}" "${INSTALL_DIR}/${BINARY_NAME}"
        chmod +x "${INSTALL_DIR}/${BINARY_NAME}"
    else
        sudo mv "${tmp_dir}/${binary}" "${INSTALL_DIR}/${BINARY_NAME}"
        sudo chmod +x "${INSTALL_DIR}/${BINARY_NAME}"
    fi
    
    rm -rf "$tmp_dir"
}

verify_installation() {
    if command -v ghex &> /dev/null; then
        local installed_version=$(ghex version 2>/dev/null || echo "unknown")
        success "GHEX installed successfully!"
        echo ""
        echo "  Version: ${installed_version}"
        echo "  Location: $(which ghex)"
        echo ""
        echo "Run 'ghex --help' to get started."
        echo ""
        echo "To update GHEX later, run: ghex update"
    else
        warn "Installation completed but 'ghex' not found in PATH"
        echo "You may need to add ${INSTALL_DIR} to your PATH"
    fi
}

main() {
    print_banner
    
    local os=$(detect_os)
    local arch=$(detect_arch)
    
    info "Detected OS: ${os}, Architecture: ${arch}"
    
    local version=${1:-$(get_latest_version)}
    if [ -z "$version" ]; then
        error "Could not determine latest version"
    fi
    
    info "Installing GHEX ${version}..."
    
    download_and_install "$version" "$os" "$arch"
    verify_installation
}

main "$@"
