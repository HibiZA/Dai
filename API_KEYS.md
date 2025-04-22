# Dai CLI API Keys Guide

This document provides detailed information about the API keys required for Dai CLI, how to obtain them, configure them, and use them securely.

## Required API Keys

Dai CLI requires two main API keys for full functionality:

1. **OpenAI API Key**: Used for AI-generated upgrade rationales and PR descriptions
2. **GitHub Token**: Used for GitHub integration (creating PRs, branches, etc.)

## 1. OpenAI API Key

### Why it's needed

The OpenAI API key is used for:
- Generating detailed upgrade rationales for dependency updates
- Analyzing potential breaking changes
- Creating smart PR descriptions
- Providing AI-backed recommendations for package upgrades

### How to obtain an OpenAI API Key

1. Visit [OpenAI's platform](https://platform.openai.com/api-keys)
2. Create or sign in to your OpenAI account
3. Navigate to API keys section
4. Click "Create new secret key"
5. Give your key a name (e.g., "Dai CLI")
6. Copy the API key immediately (it won't be shown again)

### Usage Costs

Be aware that using OpenAI's API incurs costs based on your usage:
- The cost depends on the models used and the number of tokens processed
- Dai CLI is designed to be efficient with token usage to minimize costs
- You can monitor your usage on the [OpenAI usage dashboard](https://platform.openai.com/usage)

## 2. GitHub Token

### Why it's needed

The GitHub token is used for:
- Creating pull requests with dependency upgrades
- Creating branches for changes
- Fetching repository information
- Interacting with GitHub API services

### How to obtain a GitHub Token

1. Visit [GitHub's Personal Access Tokens page](https://github.com/settings/tokens)
2. Click "Generate new token" (classic)
3. Give your token a descriptive name (e.g., "Dai CLI")
4. Select the required scopes:
   - `repo` (Full control of private repositories)
   - `workflow` (Optional, only if you want to trigger workflow runs)
5. Click "Generate token"
6. Copy the token immediately (it won't be shown again)

### Token Scopes

For Dai CLI, the following scopes are required:
- `repo`: Allows access to repositories (creating branches, PRs, etc.)
  - `repo:status`: Access commit status
  - `repo_deployment`: Access deployment status
  - `public_repo`: Access public repositories
  - `repo:invite`: Access repository invitations
  - `security_events`: Read security events

Optional scopes:
- `workflow`: Update GitHub Action workflows

## Configuring API Keys in Dai CLI

You have multiple options for configuring your API keys:

### Method 1: Config Command (Recommended)

The most user-friendly approach is using the built-in config command:

```bash
# Configure OpenAI API key
dai config --set openai --openai-key sk-your-openai-api-key

# Configure GitHub token
dai config --set github --github-token ghp_your-github-token
```

This stores your API keys in a secure configuration file located at:
- Linux/macOS: `~/.config/dai/config.env`
- Windows: `%APPDATA%\dai\config.env`

### Method 2: Environment Variables

You can set environment variables directly:

```bash
# For current shell session
export DAI_OPENAI_API_KEY="sk-your-openai-api-key"
export DAI_GITHUB_TOKEN="ghp_your-github-token"
```

To make these permanent, add them to your shell profile:

```bash
# For bash users (~/.bashrc)
echo 'export DAI_OPENAI_API_KEY="sk-your-openai-api-key"' >> ~/.bashrc
echo 'export DAI_GITHUB_TOKEN="ghp_your-github-token"' >> ~/.bashrc

# For zsh users (~/.zshrc)
echo 'export DAI_OPENAI_API_KEY="sk-your-openai-api-key"' >> ~/.zshrc
echo 'export DAI_GITHUB_TOKEN="ghp_your-github-token"' >> ~/.zshrc
```

For Windows PowerShell, add to your profile:

```powershell
# Add to PowerShell profile
Add-Content -Path $PROFILE -Value '$env:DAI_OPENAI_API_KEY = "sk-your-openai-api-key"'
Add-Content -Path $PROFILE -Value '$env:DAI_GITHUB_TOKEN = "ghp_your-github-token"'
```

### Method 3: Command-line Arguments

For one-time use, you can provide the keys directly via command-line flags:

```bash
dai upgrade --all --openai-key sk-your-openai-api-key --github-token ghp_your-github-token
```

This approach is useful for CI/CD environments or when you don't want to store the keys permanently.

## Verifying Configuration

To check your current configuration:

```bash
dai config --list
```

This will show:
- Whether each key is set
- A masked version of the keys for security
- The source of each key (environment variable or config file)

## API Key Security Best Practices

1. **Never commit API keys to version control**
   - Always use environment variables or the config command
   - Add config files to .gitignore

2. **Use appropriate scopes for GitHub tokens**
   - Only grant the permissions needed for Dai CLI
   - Create a dedicated token for Dai CLI rather than reusing tokens

3. **Rotate keys periodically**
   - Regenerate your GitHub tokens regularly
   - Update your configuration when keys change

4. **Monitor usage**
   - Regularly check your OpenAI usage dashboard
   - Watch for unexpected GitHub token usage

5. **Set expiration for GitHub tokens**
   - When creating GitHub tokens, set an expiration date
   - This limits the impact of a leaked token

## Working in CI/CD Environments

For CI/CD pipelines, use secrets or environment variables:

### GitHub Actions Example

```yaml
name: Dependency Scan

on:
  schedule:
    - cron: '0 0 * * 1'  # Weekly on Mondays

jobs:
  scan:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      
      - name: Install Dai CLI
        run: curl -fsSL https://raw.githubusercontent.com/HibiZA/dai/main/scripts/install.sh | bash
      
      - name: Scan Dependencies
        run: dai scan
        env:
          DAI_OPENAI_API_KEY: ${{ secrets.OPENAI_API_KEY }}
          DAI_GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
```

### GitLab CI Example

```yaml
dependency_scan:
  image: ubuntu:latest
  script:
    - curl -fsSL https://raw.githubusercontent.com/HibiZA/dai/main/scripts/install.sh | bash
    - dai scan
  environment:
    name: production
  variables:
    DAI_OPENAI_API_KEY: ${OPENAI_API_KEY}
    DAI_GITHUB_TOKEN: ${GITHUB_TOKEN}
```

## Troubleshooting API Key Issues

### Common Problems and Solutions

1. **Authentication Failed (OpenAI)**
   - Verify your API key is correct and not expired
   - Check if your OpenAI account has billing enabled
   - Try regenerating the API key

2. **GitHub API Rate Limit Exceeded**
   - Use a personal access token with appropriate scopes
   - Wait for rate limit to reset
   - Consider using a different GitHub account

3. **Permission Denied (GitHub)**
   - Ensure your token has the necessary scopes
   - Verify you have access to the repository
   - Check if the token has expired

4. **Config File Permission Issues**
   - Check permissions on the config directory (chmod 700)
   - Ensure the user running Dai CLI has write access to the config file

### Debugging API Key Configuration

If you're experiencing issues, you can run with debug output:

```bash
dai upgrade --debug --all
```

This will show additional information about:
- Which API keys are being used
- API request details (with sensitive information redacted)
- Error responses from APIs

## Removing API Keys

To remove configured API keys:

1. **Delete the config file**
   - Linux/macOS: `rm ~/.config/dai/config.env`
   - Windows: Delete `%APPDATA%\dai\config.env`

2. **Unset environment variables**
   ```bash
   unset DAI_OPENAI_API_KEY
   unset DAI_GITHUB_TOKEN
   ```

3. **Revoke tokens at the source**
   - [Revoke OpenAI API keys](https://platform.openai.com/api-keys)
   - [Revoke GitHub tokens](https://github.com/settings/tokens) 