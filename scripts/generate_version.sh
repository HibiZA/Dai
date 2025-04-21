#!/bin/bash

# Get the most recent tag
TAG=$(git describe --tags --abbrev=0 2>/dev/null)

# If no tag exists, use v0.0.0
if [ -z "$TAG" ]; then
  TAG="v0.0.0"
fi

# Remove the 'v' prefix if it exists
VERSION=${TAG#v}

# Replace the version in the version.go file
sed -i.bak "s/const Version = \"[^\"]*\"/const Version = \"$VERSION\"/" pkg/version/version.go
rm pkg/version/version.go.bak

echo "Version set to $VERSION" 