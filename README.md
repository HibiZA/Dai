# Dai CLI Homebrew Tap

This repository contains Homebrew formulae for [Dai CLI](https://github.com/HibiZA/dai).

## Usage

### Using Homebrew (macOS)

> **Note:** Before using Homebrew, make sure the Homebrew tap repository has been created at https://github.com/HibiZA/homebrew-dai

```bash
# Install from Homebrew
brew tap HibiZA/dai
brew install dai
```

### Installing from a Local Formula (Development)

During development or before an official release is published:

```bash
# Install from the local formula
cd ~/Documents/Project/Dai
brew install --build-from-source $(pwd)/scripts/homebrew/dai.rb
```

### Direct Installation (macOS/Linux)

## Local Development

For local development and testing, you can use:

```bash
# Install from a local tap
brew install --build-from-source Formula/dai.rb
```

## Available Formulae

- `dai`: AI-backed dependency upgrade advisor for package.json projects

## Development

To update the formula after a new release:

1. Update the version in `Formula/dai.rb`
2. Update the SHA256 checksums for each platform
3. Commit and push the changes

This is typically handled automatically by the GitHub Actions release workflow.

## License

This repository is available under the same license as [Dai CLI](https://github.com/HibiZA/dai).