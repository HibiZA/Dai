# Migration Guide for Dai CLI Configuration

In version v0.1.1, Dai CLI has updated its configuration file locations to follow platform standards. If you were using a previous version, you might need to migrate your configuration.

## Configuration Path Change

The configuration file has moved from:
- `~/.dai/config.env` to `~/.config/dai/config.env` on Linux/macOS
- `%USERPROFILE%\.dai\config.env` to `%APPDATA%\dai\config.env` on Windows

## How to Migrate Your Configuration

### Option 1: Automatic Migration

The next time you use the `dai config` command, your configuration will be automatically saved to the new location.

### Option 2: Manual Migration

1. **Create the new directory:**
   ```bash
   # Linux/macOS
   mkdir -p ~/.config/dai
   
   # Windows (in PowerShell)
   mkdir -Force $env:APPDATA\dai
   ```

2. **Copy your existing configuration:**
   ```bash
   # Linux/macOS
   cp ~/.dai/config.env ~/.config/dai/config.env
   
   # Windows (in PowerShell)
   Copy-Item "$env:USERPROFILE\.dai\config.env" -Destination "$env:APPDATA\dai\config.env"
   ```

3. **Verify your configuration:**
   ```bash
   dai config --list
   ```

### Option 3: Reconfigure

If you prefer to start fresh, you can reconfigure your API keys:

```bash
# Set OpenAI API key
dai config --set openai --openai-key YOUR_API_KEY

# Set GitHub token
dai config --set github --github-token YOUR_GITHUB_TOKEN
```

## Environment Variables

The environment variable names have also been standardized. Please use:
- `DAI_OPENAI_API_KEY` instead of `OPENAI_API_KEY`
- `DAI_GITHUB_TOKEN` instead of `GITHUB_TOKEN`
- `DAI_NVD_API_KEY` instead of `NVD_API_KEY`

For backward compatibility, the application will still check for the old variable names if the new ones are not set.

## Why This Change?

This change brings Dai CLI in line with standard platform conventions:
- Using the XDG Base Directory specification on Linux/macOS
- Using %APPDATA% on Windows
- Standardizing on environment variable naming

These changes make Dai CLI more robust, predictable, and aligned with platform best practices. 