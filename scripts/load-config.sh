#!/bin/bash

# Dai CLI Configuration Loader
# This script sources the configuration from the config file

# Find config file
CONFIG_DIR="${HOME}/.dai"
CONFIG_FILE="${CONFIG_DIR}/config.env"

if [ -f "$CONFIG_FILE" ]; then
  echo "Loading Dai CLI configuration from ${CONFIG_FILE}"
  
  # Source the config file
  set -a
  source "$CONFIG_FILE"
  set +a
  
  # Print loaded variables (masked)
  if [ -n "$DAI_OPENAI_API_KEY" ]; then
    MASKED_KEY="${DAI_OPENAI_API_KEY:0:3}...${DAI_OPENAI_API_KEY: -3}"
    echo "Loaded DAI_OPENAI_API_KEY: $MASKED_KEY"
  fi
  
  if [ -n "$DAI_GITHUB_TOKEN" ]; then
    MASKED_TOKEN="${DAI_GITHUB_TOKEN:0:3}...${DAI_GITHUB_TOKEN: -3}"
    echo "Loaded DAI_GITHUB_TOKEN: $MASKED_TOKEN"
  fi
  
  echo "Configuration loaded successfully!"
else
  echo "No configuration file found at ${CONFIG_FILE}"
  echo "Run 'dai config --set openai --openai-key YOUR_API_KEY' to create one"
fi

# Reminder to use in shell profile
echo ""
echo "To automatically load this configuration in your shell:"
echo "  Add this line to your shell profile (~/.bashrc, ~/.zshrc, etc.):"
echo "  source ${CONFIG_FILE}" 