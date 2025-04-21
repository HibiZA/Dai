#!/bin/bash

# GitHub PR Creation End-to-End Test Script
# This script tests the end-to-end workflow for creating GitHub PRs
# with dependency updates using the Dai CLI.

set -e # Exit on error

# Check if required environment variables are set
if [ -z "$GITHUB_TOKEN" ]; then
  echo "Error: GITHUB_TOKEN environment variable must be set."
  echo "Usage: GITHUB_TOKEN=<token> OPENAI_API_KEY=<key> ./github_pr_test.sh"
  exit 1
fi

if [ -z "$OPENAI_API_KEY" ]; then
  echo "Warning: OPENAI_API_KEY environment variable not set."
  echo "AI-generated descriptions will not be available."
fi

echo "=== GitHub PR Creation E2E Test ==="
echo "Creating a temporary test directory..."

# Create temporary directory for testing
TEST_DIR=$(mktemp -d)
echo "Using temporary directory: $TEST_DIR"

# Create a simple package.json file for testing
cat > $TEST_DIR/package.json << EOF
{
  "name": "dai-e2e-test",
  "version": "1.0.0",
  "description": "Test package for Dai E2E testing",
  "dependencies": {
    "express": "4.17.1",
    "lodash": "4.17.20"
  },
  "devDependencies": {
    "jest": "26.6.3"
  }
}
EOF

echo "Created test package.json file."

# Track current location to find Dai binary
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
PROJECT_ROOT="$( cd "$SCRIPT_DIR/../.." && pwd )"
DAI_BIN="$PROJECT_ROOT/dai"

# Check if Dai binary exists, build if necessary
if [ ! -f "$DAI_BIN" ]; then
  echo "Building Dai CLI..."
  cd $PROJECT_ROOT
  go build -o dai
  chmod +x dai
fi

echo "Using Dai binary: $DAI_BIN"

# Run Dai upgrade command with GitHub PR creation
echo "Running Dai upgrade with PR creation..."
cd $TEST_DIR
$DAI_BIN upgrade --all --apply --pr --github-token $GITHUB_TOKEN

# Verify the PR was created successfully
if [ $? -eq 0 ]; then
  echo "✅ E2E test passed: PR created successfully!"
else
  echo "❌ E2E test failed: PR creation failed!"
  exit 1
fi

# Clean up the temporary directory
echo "Cleaning up temporary directory..."
rm -rf $TEST_DIR

echo "Test completed successfully." 