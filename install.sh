#!/bin/bash

# SRS Installer Script
# This script downloads and installs the latest version of SRS to ~/.local/bin

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# GitHub repository
REPO="finnzink/srs"
INSTALL_DIR="$HOME/.local/bin"

echo -e "${GREEN}SRS Installer${NC}"
echo "Installing from: https://github.com/${REPO}"
echo "Install location: $INSTALL_DIR (no sudo required)"
echo

# Detect OS and architecture
OS="$(uname -s)"
ARCH="$(uname -m)"

case "$OS" in
    Linux*)
        case "$ARCH" in
            x86_64) PLATFORM="linux-amd64" ;;
            aarch64|arm64) PLATFORM="linux-arm64" ;;
            *) echo -e "${RED}Unsupported architecture: $ARCH${NC}"; exit 1 ;;
        esac
        ;;
    Darwin*)
        case "$ARCH" in
            x86_64) PLATFORM="darwin-amd64" ;;
            arm64) PLATFORM="darwin-arm64" ;;
            *) echo -e "${RED}Unsupported architecture: $ARCH${NC}"; exit 1 ;;
        esac
        ;;
    CYGWIN*|MINGW*|MSYS*)
        case "$ARCH" in
            x86_64) PLATFORM="windows-amd64"; BINARY_NAME="srs.exe" ;;
            *) echo -e "${RED}Unsupported architecture: $ARCH${NC}"; exit 1 ;;
        esac
        INSTALL_DIR="$HOME/bin"
        ;;
    *)
        echo -e "${RED}Unsupported operating system: $OS${NC}"
        exit 1
        ;;
esac

BINARY_NAME="${BINARY_NAME:-srs}"
echo "Detected platform: $PLATFORM"

# Get latest release
echo "Fetching latest release information..."
LATEST_URL="https://api.github.com/repos/${REPO}/releases/latest"
RELEASE_INFO=$(curl -s "$LATEST_URL")

if [ $? -ne 0 ]; then
    echo -e "${RED}Failed to fetch release information${NC}"
    exit 1
fi

# Extract tag name and download URL
TAG_NAME=$(echo "$RELEASE_INFO" | grep '"tag_name"' | cut -d'"' -f4)
if [ -z "$TAG_NAME" ]; then
    echo -e "${RED}Could not determine latest version${NC}"
    exit 1
fi

DOWNLOAD_URL="https://github.com/${REPO}/releases/download/${TAG_NAME}/srs-${PLATFORM}"
if [ "$BINARY_NAME" = "srs.exe" ]; then
    DOWNLOAD_URL="${DOWNLOAD_URL}.exe"
fi

echo "Latest version: $TAG_NAME"
echo "Download URL: $DOWNLOAD_URL"

# Create install directory
echo "Creating install directory: $INSTALL_DIR"
mkdir -p "$INSTALL_DIR"

# Create temporary file
TEMP_FILE=$(mktemp)
trap "rm -f $TEMP_FILE" EXIT

# Download binary
echo "Downloading..."
if ! curl -L -o "$TEMP_FILE" "$DOWNLOAD_URL"; then
    echo -e "${RED}Failed to download binary${NC}"
    exit 1
fi

# Verify download
if [ ! -s "$TEMP_FILE" ]; then
    echo -e "${RED}Downloaded file is empty${NC}"
    exit 1
fi

# Make executable and install
chmod +x "$TEMP_FILE"
mv "$TEMP_FILE" "$INSTALL_DIR/$BINARY_NAME"

echo -e "${GREEN}✅ Binary installed to $INSTALL_DIR/$BINARY_NAME${NC}"

# Check if install directory is in PATH
if [[ ":$PATH:" != *":$INSTALL_DIR:"* ]]; then
    echo
    echo -e "${YELLOW}⚠️  $INSTALL_DIR is not in your PATH${NC}"
    echo "To use srs from anywhere, add this to your shell profile:"
    echo
    echo -e "${BLUE}# Add to ~/.bashrc, ~/.zshrc, or ~/.profile${NC}"
    echo -e "${BLUE}export PATH=\"\$HOME/.local/bin:\$PATH\"${NC}"
    echo
    echo "Then restart your terminal or run:"
    echo -e "${BLUE}source ~/.bashrc${NC}  # or ~/.zshrc, ~/.profile"
    echo
    echo "Or run srs directly with: $INSTALL_DIR/$BINARY_NAME"
else
    echo -e "${GREEN}✅ Installation successful!${NC}"
    echo
    echo "Run 'srs version' to verify the installation"
    echo "Run 'srs --help' to get started"
fi

echo
echo "Quick start:"
echo "  mkdir my-deck"
echo "  echo '# What is 2 + 2?' > my-deck/math.md"
echo "  echo '' >> my-deck/math.md"
echo "  echo '---' >> my-deck/math.md"
echo "  echo '' >> my-deck/math.md"
echo "  echo '# 4' >> my-deck/math.md"
echo "  srs review my-deck"