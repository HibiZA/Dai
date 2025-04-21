#!/bin/bash

# Dai CLI Release Script
# This script builds Dai for multiple platforms and creates distribution packages

set -e

# Check if goreleaser is installed
if ! command -v goreleaser &> /dev/null; then
    echo "Error: goreleaser is required but not installed."
    echo "Please install it: https://goreleaser.com/install/"
    exit 1
fi

# Version from argument or git tag
VERSION=${1:-$(git describe --tags --abbrev=0 2>/dev/null || echo "v0.1.0")}
VERSION=${VERSION#v}  # Remove 'v' prefix if present

echo "Building Dai CLI release version v$VERSION"

# Create a temporary .goreleaser.yml file
cat > .goreleaser.yml << EOF
project_name: dai
before:
  hooks:
    - go mod tidy
builds:
  - env:
      - CGO_ENABLED=0
    goos:
      - linux
      - windows
      - darwin
    goarch:
      - amd64
      - arm64
    ignore:
      - goos: windows
        goarch: arm64
    ldflags:
      - -s -w -X github.com/HibiZA/dai/cmd.Version=v${VERSION}
archives:
  - format_overrides:
      - goos: windows
        format: zip
    replacements:
      darwin: macos
    files:
      - LICENSE
      - README.md
checksum:
  name_template: 'checksums.txt'
snapshot:
  name_template: "{{ .Tag }}-next"
changelog:
  sort: asc
  filters:
    exclude:
      - '^docs:'
      - '^test:'
      - '^Merge'
EOF

# Create version.go file to store version info
mkdir -p cmd
cat > cmd/version.go << EOF
package cmd

// Version contains the current version of Dai CLI
var Version = "v${VERSION}"

// VersionCommand returns the version of Dai CLI
func VersionCommand() string {
    return Version
}
EOF

# Run goreleaser
echo "Building packages..."
if [ "$2" == "--snapshot" ]; then
    goreleaser build --snapshot --rm-dist
else
    goreleaser release --rm-dist
fi

# Update Homebrew formula
echo "Updating Homebrew formula..."
FORMULA_FILE="scripts/homebrew/dai.rb"
mkdir -p scripts/homebrew

# Get checksums
DARWIN_AMD64_SHA=$(grep darwin_amd64 dist/checksums.txt | awk '{print $1}')
DARWIN_ARM64_SHA=$(grep darwin_arm64 dist/checksums.txt | awk '{print $1}')
LINUX_AMD64_SHA=$(grep linux_amd64 dist/checksums.txt | awk '{print $1}')
LINUX_ARM64_SHA=$(grep linux_arm64 dist/checksums.txt | awk '{print $1}')

# Update formula template
cat > $FORMULA_FILE << EOF
class Dai < Formula
  desc "AI-backed dependency upgrade advisor for package.json projects"
  homepage "https://github.com/HibiZA/dai"
  license "MIT"
  version "${VERSION}"

  if OS.mac?
    if Hardware::CPU.arm?
      url "https://github.com/HibiZA/dai/releases/download/v${VERSION}/dai_macos_arm64.tar.gz"
      sha256 "${DARWIN_ARM64_SHA}"
    else
      url "https://github.com/HibiZA/dai/releases/download/v${VERSION}/dai_macos_amd64.tar.gz"
      sha256 "${DARWIN_AMD64_SHA}"
    end
  elsif OS.linux?
    if Hardware::CPU.intel?
      url "https://github.com/HibiZA/dai/releases/download/v${VERSION}/dai_linux_amd64.tar.gz"
      sha256 "${LINUX_AMD64_SHA}"
    else
      url "https://github.com/HibiZA/dai/releases/download/v${VERSION}/dai_linux_arm64.tar.gz"
      sha256 "${LINUX_ARM64_SHA}"
    end
  end

  def install
    bin.install "dai"
  end

  test do
    system "#{bin}/dai", "--version"
  end

  def caveats
    <<~EOS
      To use Dai CLI, you may need to configure your API keys:
      
      For OpenAI integration:
      $ export DAI_OPENAI_API_KEY=your_api_key
      
      For GitHub integration:
      $ export DAI_GITHUB_TOKEN=your_github_token
      
      You can add these to your shell profile for permanent use.
    EOS
  end
end
EOF

echo "Release v$VERSION created successfully!"
echo "dist/ directory contains all release artifacts"
echo "Homebrew formula updated at $FORMULA_FILE" 