# Dai CLI - AI‑Backed Dependency Upgrade Advisor

Dai CLI automates dependency maintenance by scanning your project, detecting outdated packages, and using AI to draft upgrade rationales and PRs.

## Installation

### macOS (Homebrew)

```bash
brew install HibiZA/dai/dai-cli
```

### Direct Installation (macOS/Linux)

```bash
# For latest stable release
curl -fsSL https://raw.githubusercontent.com/HibiZA/dai/main/scripts/install.sh | bash

# For specific version
curl -fsSL https://raw.githubusercontent.com/HibiZA/dai/main/scripts/install.sh | bash -s -- v0.1.0
```

### Windows

Download the latest release from [GitHub Releases](https://github.com/HibiZA/dai/releases) and add the executable to your PATH.

## Configuration

### Setting API Keys

Dai CLI requires API keys for full functionality:

#### Using the config command

```bash
# Set OpenAI API key
dai config --set openai --openai-key YOUR_API_KEY

# Set GitHub token
dai config --set github --github-token YOUR_GITHUB_TOKEN
```

#### Using environment variables

```bash
# Set environment variables permanently in your shell profile (.bashrc, .zshrc, etc.)
export DAI_OPENAI_API_KEY="your-openai-api-key"
export DAI_GITHUB_TOKEN="your-github-token"
```

#### View current configuration

```bash
dai config --list
```

## Usage

### Scanning for Vulnerabilities

```bash
# Scan all dependencies (including dev)
dai scan

# Scan only production dependencies
dai scan --dev=false

# Output in table format
dai scan --format table
```

### Upgrading Dependencies

```bash
# Upgrade a specific package
dai upgrade react

# Upgrade multiple packages
dai upgrade react,react-dom,redux

# Upgrade all dependencies
dai upgrade --all

# Preview upgrades without applying changes
dai upgrade --all --dry-run

# Apply upgrades and create a PR
dai upgrade --all --apply --pr
```

## Command Reference

### General Commands

```bash
# Show version
dai version

# Show help
dai --help
```

### Scan Command

```bash
# Show scan command help
dai scan --help
```

### Upgrade Command

```bash
# Show upgrade command help
dai upgrade --help
```

### Config Command

```bash
# Show config command help
dai config --help
```

## Getting Help

For more detailed information on any command, use the `--help` flag:

```bash
dai --help
dai scan --help
dai upgrade --help
dai config --help
```

## License

[MIT License](LICENSE)