#!/bin/bash

# Dai CLI Quick Install Script
# This is a minimal installer for use in documentation examples 
# and one-line curl | bash style installation commands.

set -e

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
  *)
    echo "Error: Unsupported architecture: $ARCH"
    exit 1
    ;;
esac

# Map OS to standardized names
case $OS in
  darwin)
    OS="macos"
    ;;
  linux)
    OS="linux"
    ;;
  *)
    echo "Error: Unsupported operating system: $OS"
    exit 1
    ;;
esac

# GitHub repo details
REPO="HibiZA/dai"
LATEST_RELEASE_URL="https://api.github.com/repos/$REPO/releases/latest"

# Get the latest release version
echo "Finding latest release..."
if command -v curl > /dev/null 2>&1; then
  LATEST_VERSION=$(curl -s $LATEST_RELEASE_URL | grep -o '"tag_name": "v[^"]*' | cut -d'"' -f4)
elif command -v wget > /dev/null 2>&1; then
  LATEST_VERSION=$(wget -q -O - $LATEST_RELEASE_URL | grep -o '"tag_name": "v[^"]*' | cut -d'"' -f4)
else
  echo "Error: curl or wget is required"
  exit 1
fi

if [ -z "$LATEST_VERSION" ]; then
  LATEST_VERSION="v0.1.0"  # Fallback version
fi

# Construct the download URL
DOWNLOAD_URL="https://github.com/$REPO/releases/download/$LATEST_VERSION/dai_${OS}_${ARCH}.tar.gz"

# Installation directory
INSTALL_DIR="/usr/local/bin"
if [ ! -w "$INSTALL_DIR" ]; then
  INSTALL_DIR="$HOME/bin"
  mkdir -p "$INSTALL_DIR"
fi

# Download and install
echo "Downloading Dai CLI ${LATEST_VERSION}..."
if command -v curl > /dev/null 2>&1; then
  curl -L "$DOWNLOAD_URL" | tar xz -C "$INSTALL_DIR"
elif command -v wget > /dev/null 2>&1; then
  wget -O- "$DOWNLOAD_URL" | tar xz -C "$INSTALL_DIR"
fi

echo "Dai CLI installed successfully in $INSTALL_DIR!"
echo "Run 'dai --help' to get started."

# Check if PATH includes the install directory
if ! echo "$PATH" | grep -q "$INSTALL_DIR"; then
  echo ""
  echo "NOTE: Make sure $INSTALL_DIR is in your PATH."
fi 