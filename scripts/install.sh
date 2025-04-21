#!/bin/bash

# Dai CLI Installation Script
# This script downloads and installs the Dai CLI tool for
# dependency management and vulnerability scanning.

set -e

# Default install directory
INSTALL_DIR="/usr/local/bin"
if [ ! -d "$INSTALL_DIR" ]; then
  INSTALL_DIR="$HOME/bin"
  mkdir -p "$INSTALL_DIR"
fi

# Detect the operating system and architecture
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

# Map architecture to standardized names
case $ARCH in
  x86_64)
    ARCH="amd64"
    ;;
  aarch64|arm64)
    ARCH="arm64"
    ;;
  i386|i686)
    ARCH="386"
    ;;
  *)
    echo "Error: Unsupported architecture: $ARCH"
    exit 1
    ;;
esac

# GitHub repo details
REPO="HibiZA/dai"
LATEST_RELEASE_URL="https://api.github.com/repos/$REPO/releases/latest"

# Get the latest release version
echo "Detecting latest version of Dai CLI..."
if command -v curl > /dev/null 2>&1; then
  LATEST_VERSION=$(curl -s $LATEST_RELEASE_URL | grep -o '"tag_name": "v[^"]*' | cut -d'"' -f4)
elif command -v wget > /dev/null 2>&1; then
  LATEST_VERSION=$(wget -q -O - $LATEST_RELEASE_URL | grep -o '"tag_name": "v[^"]*' | cut -d'"' -f4)
else
  echo "Error: curl or wget is required to download Dai CLI"
  exit 1
fi

if [ -z "$LATEST_VERSION" ]; then
  echo "Error: Could not determine the latest version"
  exit 1
fi

echo "Latest version: $LATEST_VERSION"

# Construct the download URL
DOWNLOAD_URL="https://github.com/$REPO/releases/download/$LATEST_VERSION/dai_${OS}_${ARCH}.tar.gz"
echo "Download URL: $DOWNLOAD_URL"

# Create a temporary directory
TMP_DIR=$(mktemp -d)
trap 'rm -rf "$TMP_DIR"' EXIT

# Download the tarball
echo "Downloading Dai CLI..."
if command -v curl > /dev/null 2>&1; then
  curl -L -o "$TMP_DIR/dai.tar.gz" "$DOWNLOAD_URL"
elif command -v wget > /dev/null 2>&1; then
  wget -O "$TMP_DIR/dai.tar.gz" "$DOWNLOAD_URL"
fi

# Extract the binary
echo "Extracting..."
tar -xzf "$TMP_DIR/dai.tar.gz" -C "$TMP_DIR"

# Install the binary
echo "Installing to $INSTALL_DIR..."
if [ -w "$INSTALL_DIR" ]; then
  mv "$TMP_DIR/dai" "$INSTALL_DIR/"
else
  sudo mv "$TMP_DIR/dai" "$INSTALL_DIR/"
fi

# Make sure the binary is executable
if [ -w "$INSTALL_DIR/dai" ]; then
  chmod +x "$INSTALL_DIR/dai"
else
  sudo chmod +x "$INSTALL_DIR/dai"
fi

# Verify installation
if [ -x "$INSTALL_DIR/dai" ]; then
  echo "Dai CLI installed successfully!"
  echo "Version: $("$INSTALL_DIR/dai" version 2>/dev/null || echo "unknown")"
  echo ""
  echo "To use Dai CLI, run:"
  echo "  dai scan                # Scan for vulnerabilities"
  echo "  dai upgrade [packages]  # Upgrade dependencies"
  echo ""
  echo "For more information, run: dai --help"
else
  echo "Error: Installation failed"
  exit 1
fi

# Check if PATH includes the install directory
if ! echo "$PATH" | grep -q "$INSTALL_DIR"; then
  echo ""
  echo "NOTE: Make sure $INSTALL_DIR is in your PATH."
  echo "You may need to add the following to your shell profile:"
  echo "  export PATH=\$PATH:$INSTALL_DIR"
fi 