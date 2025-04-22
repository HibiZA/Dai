# Dai CLI Installation Guide

This guide provides detailed instructions for installing Dai CLI on different platforms and configuring API keys for full functionality.

## Installation Methods

### macOS (Homebrew)

The recommended way to install Dai CLI on macOS is using Homebrew:

```bash
# Add the Dai CLI tap
brew tap HibiZA/dai

# Install Dai CLI
brew install dai-cli
```

Alternatively, you can use the single command:

```bash
brew install HibiZA/dai/dai-cli
```

To update Dai CLI to the latest version:

```bash
brew upgrade dai-cli
```

### Direct Installation Script (macOS/Linux)

For macOS and Linux systems, you can use the installation script:

```bash
# Install the latest stable release
curl -fsSL https://raw.githubusercontent.com/HibiZA/dai/main/scripts/install.sh | bash

# Install a specific version
curl -fsSL https://raw.githubusercontent.com/HibiZA/dai/main/scripts/install.sh | bash -s -- v0.1.0
```

The script installs Dai CLI to `/usr/local/bin/dai` by default.

### Manual Installation

#### Linux/macOS

1. Download the appropriate binary for your system from the [GitHub Releases page](https://github.com/HibiZA/dai/releases)
2. Extract the archive:
   ```bash
   tar -xzf dai_0.1.0_linux_amd64.tar.gz
   ```
3. Move the binary to a location in your PATH:
   ```bash
   sudo mv dai /usr/local/bin/
   ```
4. Make it executable:
   ```bash
   sudo chmod +x /usr/local/bin/dai
   ```

#### Windows

1. Download the Windows binary from the [GitHub Releases page](https://github.com/HibiZA/dai/releases)
2. Extract the ZIP file to a folder of your choice
3. Add the folder to your PATH environment variable:
   - Right-click on "This PC" or "My Computer" and select "Properties"
   - Click on "Advanced system settings"
   - Click on "Environment Variables"
   - Under "System variables", find and select "Path", then click "Edit"
   - Click "New" and add the path to the folder containing the dai.exe file
   - Click "OK" to close all dialogs

## Verifying Installation

After installation, verify that Dai CLI is installed correctly:

```bash
dai version
```

You should see output showing the current version of Dai CLI.

## API Key Configuration

Dai CLI requires API keys for full functionality:

### 1. OpenAI API Key

This is required for AI-powered features like generating upgrade rationales.

Obtain an API key from [OpenAI's platform](https://platform.openai.com/api-keys).

### 2. GitHub Token

This is required for GitHub integration features like creating pull requests.

Create a personal access token (classic) with the following scopes:
- `repo` (Full repository access)
- `workflow` (Optional, only if you want to create workflow dispatch events)

You can create a token at [GitHub Personal Access Tokens](https://github.com/settings/tokens).

## Setting Up API Keys

### Method 1: Using the config command

The easiest way to configure API keys is using the built-in config command:

```bash
# Set OpenAI API key
dai config --set openai --openai-key YOUR_API_KEY

# Set GitHub token
dai config --set github --github-token YOUR_GITHUB_TOKEN
```

This stores your API keys in a config file located at:
- Linux/macOS: `~/.config/dai/config.env`
- Windows: `%APPDATA%\dai\config.env`

### Method 2: Using environment variables

You can also set environment variables directly:

```bash
# For the current session
export DAI_OPENAI_API_KEY="your-openai-api-key"
export DAI_GITHUB_TOKEN="your-github-token"

# To make them permanent, add them to your shell profile (.bashrc, .zshrc, etc.)
echo 'export DAI_OPENAI_API_KEY="your-openai-api-key"' >> ~/.bashrc
echo 'export DAI_GITHUB_TOKEN="your-github-token"' >> ~/.bashrc
```

On Windows, you can set environment variables through system properties or in your terminal:

```powershell
# PowerShell
$env:DAI_OPENAI_API_KEY = "your-openai-api-key"
$env:DAI_GITHUB_TOKEN = "your-github-token"

# CMD
set DAI_OPENAI_API_KEY=your-openai-api-key
set DAI_GITHUB_TOKEN=your-github-token
```

### Method 3: Command-line flags

You can provide API keys directly when running commands:

```bash
dai upgrade --all --openai-key YOUR_API_KEY --github-token YOUR_GITHUB_TOKEN
```

### Verifying Configuration

To check your current configuration:

```bash
dai config --list
```

This will show which keys are set and their sources (environment variables or config file).

## Troubleshooting

### Common Issues

1. **Command not found**
   - Make sure the installation directory is in your PATH
   - Try reinstalling using one of the methods above

2. **Permission denied**
   - Ensure the binary is executable: `chmod +x /path/to/dai`

3. **API Key Issues**
   - Verify your API keys are correctly set using `dai config --list`
   - Try setting the keys using a different method

4. **GitHub Authentication Failures**
   - Ensure your GitHub token has the required scopes
   - Check if your token has expired and regenerate if necessary

### Getting Help

For any issues not covered here, please:

1. Check the help documentation: `dai --help`
2. Visit the [GitHub repository](https://github.com/HibiZA/dai) 
3. Open an issue if you encounter a bug or have a feature request

## Uninstalling

To uninstall Dai CLI:

### macOS (Homebrew)
```bash
brew uninstall dai-cli
```

### Linux/macOS (Manual Installation)
```bash
sudo rm /usr/local/bin/dai
```

### Windows
1. Delete the dai.exe file
2. Remove the path from your PATH environment variable 