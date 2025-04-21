# Dai GitHub Actions

This directory contains GitHub Actions for automating dependency management using the Dai CLI.

## Dependency Check Workflow

The `dependency-check.yml` file defines a GitHub Action that automatically checks for dependency updates and security vulnerabilities in your project.

### Features

- **Scheduled Runs**: Automatically runs every Monday at midnight
- **Manual Triggering**: Can be manually triggered from the GitHub Actions tab
- **Vulnerability Scanning**: Checks for known security vulnerabilities in dependencies
- **Dependency Updates**: Identifies available updates for dependencies
- **PR Creation**: Can create pull requests with AI-generated descriptions for dependency updates

### How to Use

1. **Scheduled Checks**: The workflow runs automatically every Monday, producing reports that can be viewed as artifacts in the workflow run.

2. **Manual Trigger**:
   - Go to the "Actions" tab in your GitHub repository
   - Select the "Dependency Check" workflow
   - Click "Run workflow"
   - The workflow will scan dependencies and create artifacts

3. **Creating PRs**:
   - When manually triggered, the workflow can create pull requests with dependency updates
   - This requires the `GITHUB_TOKEN` secret to be available (automatically provided by GitHub)
   - For AI-generated descriptions, you need to add an `OPENAI_API_KEY` secret to your repository

### Configuration

To set up the required secrets:

1. Go to your repository settings
2. Navigate to "Secrets and variables" > "Actions"
3. Add a new repository secret:
   - Name: `OPENAI_API_KEY`
   - Value: Your OpenAI API key

The `GITHUB_TOKEN` is automatically provided by GitHub and does not need to be manually configured.

### Workflow Steps

1. **Checkout code**: Fetches the repository content
2. **Set up Go**: Installs Go programming language
3. **Install Dai CLI**: Builds and installs the Dai CLI tool
4. **Scan for vulnerabilities**: Runs `dai scan` and saves the report
5. **Check for dependency updates**: Runs `dai upgrade --simulate` and saves the report
6. **Create PR for dependency updates** (manual trigger only): Runs `dai upgrade --apply --pr`
7. **Upload reports as artifacts**: Makes reports available for download

### Customization

You can customize the workflow by editing the `.github/workflows/dependency-check.yml` file:

- Change the schedule by modifying the `cron` expression
- Adjust the Go version in the setup-go action
- Modify the command parameters for each step

### Troubleshooting

If the workflow fails, check:

1. The workflow run logs for specific error messages
2. That your repository has the required secrets configured
3. That the Dai CLI can be built successfully from your code 