#!/bin/bash

set -e

TAG=$1

if [[ -z "$TAG" ]]; then
    echo "❌ Error: No tag provided"
    exit 1
fi

# Validates SEMVER format: v1.2.3 or v1.2.3-beta.1
if [[ ! "$TAG" =~ ^v[0-9]+\.[0-9]+\.[0-9]+(-[a-zA-Z0-9.]+)?$ ]]; then
    echo "❌ Error: Tag '$TAG' does not follow SEMVER format (vX.Y.Z)"
    echo "Examples: v1.0.0, v2.1.3, v1.0.0-beta.1"
    exit 1
fi

echo "✅ Tag '$TAG' is valid SEMVER"

# Extract version without the 'v'
VERSION=${TAG#v}
echo "version=$VERSION" >> $GITHUB_OUTPUT

# Extract components
IFS='.' read -r MAJOR MINOR PATCH <<< "$VERSION"
PATCH=${PATCH%%-*}  # Remove pre-release suffix

echo "major=$MAJOR" >> $GITHUB_OUTPUT
echo "minor=$MINOR" >> $GITHUB_OUTPUT
echo "patch=$PATCH" >> $GITHUB_OUTPUT

echo "📦 Version components: $MAJOR.$MINOR.$PATCH"
