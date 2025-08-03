#!/bin/bash
set -e

# SRS Installation Script
# Usage: curl -sSL https://raw.githubusercontent.com/USER/REPO/main/install.sh | bash

REPO="YOUR_GITHUB_USERNAME/srs"  # Update this when you push to GitHub
INSTALL_DIR="/usr/local/bin"
BINARY_NAME="srs"

# Detect platform
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

case $ARCH in
    x86_64) ARCH="amd64" ;;
    aarch64|arm64) ARCH="arm64" ;;
    *) echo "Unsupported architecture: $ARCH"; exit 1 ;;
esac

case $OS in
    linux) PLATFORM="linux" ;;
    darwin) PLATFORM="darwin" ;;
    *) echo "Unsupported OS: $OS"; exit 1 ;;
esac

BINARY_FILE="${BINARY_NAME}-${PLATFORM}-${ARCH}"
if [ "$OS" = "windows" ]; then
    BINARY_FILE="${BINARY_FILE}.exe"
fi

echo "Installing SRS for ${PLATFORM}-${ARCH}..."

# Get latest release URL
LATEST_RELEASE=$(curl -s "https://api.github.com/repos/${REPO}/releases/latest")
DOWNLOAD_URL=$(echo "$LATEST_RELEASE" | grep -o "https://github.com/${REPO}/releases/download/[^\"]*/${BINARY_FILE}" | head -1)

if [ -z "$DOWNLOAD_URL" ]; then
    echo "Error: Could not find download URL for ${BINARY_FILE}"
    echo "Available releases:"
    echo "$LATEST_RELEASE" | grep -o "https://github.com/${REPO}/releases/download/[^\"]*" | head -5
    exit 1
fi

echo "Downloading from: $DOWNLOAD_URL"

# Create temporary file
TMP_FILE=$(mktemp)

# Download binary
curl -L "$DOWNLOAD_URL" -o "$TMP_FILE"

# Make executable
chmod +x "$TMP_FILE"

# Install to system
if [ -w "$INSTALL_DIR" ]; then
    mv "$TMP_FILE" "${INSTALL_DIR}/${BINARY_NAME}"
else
    echo "Installing to ${INSTALL_DIR} (requires sudo)..."
    sudo mv "$TMP_FILE" "${INSTALL_DIR}/${BINARY_NAME}"
fi

echo "✅ SRS installed successfully!"
echo ""
echo "Quick start:"
echo "  mkdir my-deck"
echo "  echo '# What is 2 + 2?' > my-deck/math.md"
echo "  echo '' >> my-deck/math.md"
echo "  echo '---' >> my-deck/math.md"
echo "  echo '' >> my-deck/math.md"
echo "  echo '# 4' >> my-deck/math.md"
echo "  srs review my-deck"
echo ""
echo "Run 'srs --help' for more information."