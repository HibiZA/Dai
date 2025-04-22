# Dai CLI Usage Guide

This guide provides detailed examples and usage instructions for all Dai CLI commands.

## Table of Contents

- [Getting Started](#getting-started)
- [Command Overview](#command-overview)
- [Scan Command](#scan-command)
- [Upgrade Command](#upgrade-command)
- [Config Command](#config-command)
- [Working with Package.json Projects](#working-with-packagejson-projects)
- [Automating with CI/CD](#automating-with-cicd)
- [Best Practices](#best-practices)

## Getting Started

After [installing Dai CLI](INSTALLATION.md) and [configuring your API keys](API_KEYS.md), you can start using the tool:

```bash
# Check the installed version
dai version

# See available commands
dai --help
```

## Command Overview

Dai CLI provides several commands to manage your project dependencies:

- `scan`: Scan dependencies for outdated packages and security vulnerabilities
- `upgrade`: Upgrade dependencies with AI-generated rationales
- `config`: Configure API keys and settings
- `version`: Display the current version

## Scan Command

The `scan` command checks your project for outdated packages and security vulnerabilities.

### Basic Scan

```bash
# Navigate to your project directory containing package.json
cd my-project

# Run a basic scan
dai scan
```

This will:
1. Parse your package.json file
2. Analyze dependencies for security vulnerabilities
3. Display a report with vulnerability details and severity

### Scan Options

```bash
# Scan only production dependencies (exclude dev dependencies)
dai scan --dev=false

# Output in table format
dai scan --format table
```

### Understanding Scan Results

The scan results include:
- Package name and version
- Vulnerability ID and description
- Severity level (low, medium, high, critical)
- Recommended version to upgrade to
- CVSS score when available

Example output:
```
🔍 Scanning project dependencies for vulnerabilities...
--------------------------------------------------------------------------------

Project: my-app@1.0.0

✓ react@17.0.2: No vulnerabilities found
✓ react-dom@17.0.2: No vulnerabilities found
⚠️ lodash@4.17.15: 3 vulnerabilities found
  [1] CVE-2020-8203 - Prototype pollution in lodash (High)
  [2] CVE-2021-23337 - Command injection vulnerability (Critical)
  [3] CVE-2019-10744 - Prototype pollution in defaultsDeep (Medium)
  Recommendation: Upgrade to lodash@4.17.21
```

## Upgrade Command

The `upgrade` command analyzes and upgrades your dependencies with AI-generated rationales.

### Upgrading Specific Packages

```bash
# Upgrade a single package
dai upgrade react

# Upgrade multiple packages (comma-separated, no spaces)
dai upgrade react,react-dom,lodash
```

### Upgrading All Packages

```bash
# Upgrade all dependencies
dai upgrade --all
```

### Preview Before Applying

```bash
# Show what would be upgraded without making changes
dai upgrade --all --dry-run
```

### Applying Upgrades

```bash
# Apply upgrades directly to package.json
dai upgrade --all --apply
```

### Creating Pull Requests

```bash
# Create a PR with the changes (requires GitHub token)
dai upgrade --all --apply --pr
```

### Understanding Upgrade Output

The upgrade command will show:
1. Current and latest available versions
2. Semver compatibility assessment
3. Breaking changes analysis
4. AI-generated upgrade rationale
5. Diff preview of changes

Example output:
```
🔄 Checking upgrades for 3 packages...
--------------------------------------------------------------------------------

📦 lodash: 4.17.15 → 4.17.21 (patch)
✅ Semver compatible: Yes (patch update)
🔒 Security: Fixes 3 vulnerabilities
💬 AI Rationale: This patch update addresses three security vulnerabilities 
   including prototype pollution (CVE-2020-8203) and command injection 
   (CVE-2021-23337). No breaking changes expected.

📦 react: 17.0.2 → 18.2.0 (major)
⚠️ Semver compatible: No (major update)
🔒 Security: No known vulnerabilities
💬 AI Rationale: React 18 introduces concurrent rendering and automatic batching.
   Breaking changes include:
   - Suspense behavior changes
   - Stricter hydration errors
   - No more implicit batching for setTimeout, etc.
   Migration steps required: Update ReactDOM.render to createRoot API.
```

### Advanced Upgrade Options

```bash
# Specify a custom npm registry
dai upgrade --all --registry https://registry.npmjs.org

# Enable debug output
dai upgrade --debug --all

# Test AI content quality without making changes
dai upgrade --test-ai react
```

## Config Command

The `config` command manages API keys and settings.

### Setting API Keys

```bash
# Set OpenAI API key
dai config --set openai --openai-key YOUR_API_KEY

# Set GitHub token
dai config --set github --github-token YOUR_GITHUB_TOKEN
```

### Viewing Current Configuration

```bash
# List current configuration
dai config --list
```

Example output:
```
Current Configuration:
--------------------------------------------------------------------------------
OpenAI API Key: Set ✓
GitHub Token: Set ✓

Environment Variables:
  DAI_OPENAI_API_KEY=sk-***key
  DAI_GITHUB_TOKEN=ghp_***ken

Config File:
  /Users/username/.config/dai/config.env
```

## Working with Package.json Projects

Dai CLI is designed to work with Node.js projects using package.json for dependency management.

### Project Types Supported

- Node.js applications
- React/Vue/Angular applications
- NPM packages
- Any project with a package.json file

### Finding Dependencies to Upgrade

To identify which dependencies should be upgraded:

1. Scan for security vulnerabilities:
   ```bash
   dai scan
   ```

2. Check for outdated packages:
   ```bash
   dai upgrade --all --dry-run
   ```

3. Prioritize upgrades based on:
   - Security vulnerabilities (critical and high first)
   - Breaking changes impact
   - Dependencies with the most outdated versions

### Workflow Example

Typical workflow for maintaining dependencies:

1. **Scan project**
   ```bash
   dai scan
   ```

2. **Preview potential upgrades**
   ```bash
   dai upgrade --all --dry-run
   ```

3. **Apply safe upgrades**
   ```bash
   dai upgrade --all --apply
   ```

4. **Test your application**
   ```bash
   npm test
   ```

5. **Create PR for breaking changes**
   ```bash
   dai upgrade react --apply --pr
   ```

## Automating with CI/CD

Dai CLI can be integrated into CI/CD pipelines for automated dependency maintenance.

### GitHub Actions Example

```yaml
name: Weekly Dependency Check

on:
  schedule:
    - cron: '0 0 * * 1'  # Every Monday at midnight

jobs:
  dependency-check:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      
      - name: Install Dai CLI
        run: curl -fsSL https://raw.githubusercontent.com/HibiZA/dai/main/scripts/install.sh | bash
      
      - name: Scan Dependencies
        run: dai scan
        env:
          DAI_OPENAI_API_KEY: ${{ secrets.OPENAI_API_KEY }}
          
      - name: Create Dependency Update PR
        if: success()
        run: |
          git config --global user.name "Dependency Bot"
          git config --global user.email "bot@example.com"
          dai upgrade --all --apply --pr
        env:
          DAI_OPENAI_API_KEY: ${{ secrets.OPENAI_API_KEY }}
          DAI_GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
```

### CI/CD Best Practices

1. **Schedule regular scans**
   - Weekly for most projects
   - Daily for security-critical projects

2. **Create separate PRs for different update types**
   - Security updates (`dai upgrade --all --apply --pr`)
   - Non-breaking updates only (`dai upgrade --all --apply --pr`)
   - Breaking changes individually (`dai upgrade packagename --apply --pr`)

3. **Add automated testing**
   - Run tests after applying upgrades
   - Only create PRs if tests pass

## Best Practices

### Security-First Approach

1. **Prioritize security vulnerabilities**
   - Update packages with critical/high vulnerabilities first
   - Run regular security scans (`dai scan`)

2. **Keep dependencies minimal**
   - Review dependencies regularly
   - Remove unused packages

### Effective Upgrading

1. **Use semantic versioning wisely**
   - Patch updates are usually safe to apply automatically
   - Review major version upgrades carefully
   - Test thoroughly after upgrades

2. **Batch similar updates**
   - Group related packages (e.g., all React packages)
   - Create separate PRs for framework updates vs utility libraries

3. **Review AI rationales**
   - The AI provides context about changes
   - Pay attention to breaking changes and migration steps
   - Use the information to plan testing approach

### Team Workflows

1. **Document upgrade decisions**
   - Use the AI-generated rationales in PR descriptions
   - Add additional context about testing performed

2. **Establish a regular upgrade schedule**
   - Weekly or bi-weekly dependency reviews
   - Immediate updates for critical security issues

3. **Assign dependency maintenance rotation**
   - Share responsibility across team
   - Use Dai CLI to make the process efficient 