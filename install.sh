#!/usr/bin/env bash

# ==============================================================================
# inceptools-cli Installation Script
#
# Usage:
#   curl -fsSL https://raw.githubusercontent.com/IncepTools/inceptools-cli/main/install.sh | bash
#
# Description:
#   This script automates the installation, update, and uninstallation of
#   inceptools-cli.
# ==============================================================================

# No 'set -e' - we handle errors manually for better reporting and robustness.

# Configuration
REPO="IncepTools/inceptools-cli"
GITHUB_API_RELEASES="https://api.github.com/repos/$REPO/releases"
GITHUB_URL="https://github.com/$REPO"
BINARY_NAME="inceptools"
INSTALL_DIR="/usr/local/bin"

# Colors and Styling
RED='\033[0;31m'
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
BOLD='\033[1m'
NC='\033[0m' # No Color

# ------------------------------------------------------------------------------
# Utility Functions
# ------------------------------------------------------------------------------

info() { echo -e "${BLUE}info:${NC} $1"; }
success() { echo -e "${GREEN}success:${NC} $1"; }
warn() { echo -e "${YELLOW}warning:${NC} $1"; }
error() { echo -e "${RED}error:${NC} $1"; exit 1; }

header() {
    echo -e "${BOLD}"
    echo "===================================================="
    echo "       IncepTools CLI - Installation Wizard         "
    echo "===================================================="
    echo -e "${NC}"
}

preflight() {
    deps=("curl" "grep" "sed" "tr" "uname")
    for dep in "${deps[@]}"; do
        if ! command -v "$dep" >/dev/null 2>&1; then
            error "Required dependency '$dep' is not installed."
        fi
    done
}

# Detection Logic
detect_system() {
    OS_TYPE=$(uname -s | tr '[:upper:]' '[:lower:]')
    ARCH_TYPE=$(uname -m)

    case "$ARCH_TYPE" in
        x86_64) ARCH="amd64" ;;
        aarch64|arm64) ARCH="arm64" ;;
        *) error "Unsupported architecture: $ARCH_TYPE" ;;
    esac

    case "$OS_TYPE" in
        linux) OS="linux" ;;
        darwin) OS="darwin" ;;
        msys*|cygwin*|mingw*) OS="windows" ;;
        *) error "Unsupported OS: $OS_TYPE" ;;
    esac

    ASSET_NAME="${BINARY_NAME}-${OS}-${ARCH}"
    [ "$OS" = "windows" ] && ASSET_NAME="${ASSET_NAME}.exe"
}

check_existing() {
    if command -v "$BINARY_NAME" >/dev/null 2>&1; then
        EXISTING_PATH=$(command -v "$BINARY_NAME")
        # Extract version safely, masking exit code of the CLI itself
        EXISTING_VERSION=$("$BINARY_NAME" --version 2>&1 | head -n 1 | awk '{print $NF}')
        # Handle cases where --version isn't implemented or fails
        if [[ "$EXISTING_VERSION" == "subcommands" || -z "$EXISTING_VERSION" ]]; then
            EXISTING_VERSION="installed"
        fi
        return 0
    fi
    return 1
}

# Menu Functions
do_uninstall() {
    if check_existing; then
        info "Uninstalling $BINARY_NAME from $EXISTING_PATH..."
        [ ! -w "$(dirname "$EXISTING_PATH")" ] && SUDO="sudo" || SUDO=""
        if $SUDO rm -f "$EXISTING_PATH"; then
            success "$BINARY_NAME has been uninstalled."
        else
            error "Failed to remove $EXISTING_PATH. Check permissions."
        fi
    else
        warn "$BINARY_NAME is not currently installed."
    fi
}

do_install() {
    detect_system
    info "Detected system: ${OS}/${ARCH}"

    # Fetch versions from GitHub
    info "Fetching available versions from GitHub..."
    RELEASES_JSON=$(curl -sL "$GITHUB_API_RELEASES") || error "Failed to connect to GitHub API."

    # Safely extract versions
    VERSIONS=$(echo "$RELEASES_JSON" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')

    if [ -z "$VERSIONS" ]; then
        warn "Could not list versions from GitHub (you might be rate-limited)."
        echo -ne "Please enter the version tag manually (e.g., v1.0.0): "
        read -r VERSION < /dev/tty
        if [ -z "$VERSION" ]; then error "Version is required."; fi
    else
        echo -e "\nAvailable versions:"
        echo "$VERSIONS" | nl -w 2 -s '. '
        LATEST_VERSION=$(echo "$VERSIONS" | head -n 1)
        echo -ne "Enter version number or tag [Default: $LATEST_VERSION]: "
        read -r SELECTION < /dev/tty

        if [ -z "$SELECTION" ]; then
            VERSION="$LATEST_VERSION"
        elif [[ "$SELECTION" =~ ^[0-9]+$ ]]; then
            VERSION=$(echo "$VERSIONS" | sed -n "${SELECTION}p")
            [ -z "$VERSION" ] && error "Invalid Selection Number."
        else
            VERSION="$SELECTION"
        fi
    fi

    info "Selected version: $VERSION"

    # Preparation
    TMP_DIR=$(mktemp -d -p "/tmp" 2>/dev/null || mktemp -d)
    trap 'rm -rf "$TMP_DIR"' EXIT

    DOWNLOAD_URL="${GITHUB_URL}/releases/download/${VERSION}/${ASSET_NAME}"
    CHECKSUM_URL="${GITHUB_URL}/releases/download/${VERSION}/checksums.txt"

    info "Downloading ${ASSET_NAME} from ${VERSION}..."
    CURL_STATUS=$(curl -sL -w "%{http_code}" "$DOWNLOAD_URL" -o "${TMP_DIR}/${ASSET_NAME}")
    if [ "$CURL_STATUS" -ne 200 ]; then
        error "Download failed (HTTP $CURL_STATUS). Asset for $VERSION on $OS/$ARCH may not exist."
    fi

    info "Checking for checksums..."
    curl -sL "$CHECKSUM_URL" -o "${TMP_DIR}/checksums.txt"

    # Verification logic
    pushd "$TMP_DIR" > /dev/null
    EXPECTED_HASH=$(grep "${ASSET_NAME}" checksums.txt | cut -d ' ' -f 1 | tr -d '\r')

    if [ -n "$EXPECTED_HASH" ]; then
        info "Verifying integrity..."
        if command -v sha256sum >/dev/null 2>&1; then ACTUAL_HASH=$(sha256sum "${ASSET_NAME}" | cut -d ' ' -f 1)
        elif command -v shasum >/dev/null 2>&1; then ACTUAL_HASH=$(shasum -a 256 "${ASSET_NAME}" | cut -d ' ' -f 1)
        fi
        if [ "$EXPECTED_HASH" != "$ACTUAL_HASH" ]; then error "Checksum mismatch!"; fi
        success "Integrity verified."
    else
        warn "No hash found in checksums.txt. Skipping verification."
    fi

    # Installation logic
    if [ "$OS" = "windows" ]; then
        success "Download complete. Windows binary saved: ${TMP_DIR}/${ASSET_NAME}"
        info "Manually move the .exe to a folder in your PATH."
    else
        [ ! -w "$INSTALL_DIR" ] && SUDO="sudo" || SUDO=""
        info "Installing to ${INSTALL_DIR}/${BINARY_NAME}..."
        if $SUDO mv "${ASSET_NAME}" "${INSTALL_DIR}/${BINARY_NAME}" && $SUDO chmod +x "${INSTALL_DIR}/${BINARY_NAME}"; then
            success "$BINARY_NAME updated successfully!"
        else
            error "Installation failed. Check sudo permissions."
        fi
    fi
    popd >/dev/null
}

# ------------------------------------------------------------------------------
# Main UI
# ------------------------------------------------------------------------------

preflight
header

# Package Manager Suggestions
if [[ "$OS" == "darwin" ]] && command -v brew >/dev/null 2>&1; then
    info "Suggestion: You have Homebrew installed. You can also install via:"
    echo -e "  ${BOLD}brew tap IncepTools/inceptools${NC}"
    echo -e "  ${BOLD}brew install inceptools${NC}\n"
elif [[ "$OS" == "linux" ]] && command -v apt >/dev/null 2>&1; then
    info "Suggestion: You are on a Debian-based system. Check our releases for .deb packages."
fi

if check_existing; then
    echo -e "Existing installation found:"
    echo -e "  - Path:    $EXISTING_PATH"
    echo -e "  - Version: $EXISTING_VERSION"
    echo -e "\nWhat would you like to do?"
    echo "  1. Update/Reinstall $BINARY_NAME"
    echo "  2. Uninstall $BINARY_NAME"
    echo "  3. Exit"
    echo -ne "\nOption [1-3]: "
    read -r CHOICE < /dev/tty

    case $CHOICE in
        1) do_install ;;
        2) do_uninstall ;;
        *) info "Execution aborted."; exit 0 ;;
    esac
else
    info "No installation detected."
    do_install
fi

info "Type '${BINARY_NAME}' to get started."
