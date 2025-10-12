#!/bin/bash

set -e

CURRENT_TAG=$1
SERVICES=("gateway" "users" "authentication" "notification" "frontend")

# Find the previous tag
PREVIOUS_TAG=$(git describe --tags --abbrev=0 "$CURRENT_TAG^" 2>/dev/null || echo "")

if [[ -z "$PREVIOUS_TAG" ]]; then
    echo "⚠️  No previous tag found. Will deploy all services."
    CHANGED_SERVICES=("${SERVICES[@]}")
else
    echo "📊 Comparing $PREVIOUS_TAG → $CURRENT_TAG"

    CHANGED_SERVICES=()
    for service in "${SERVICES[@]}"; do
        if git diff --name-only "$PREVIOUS_TAG" "$CURRENT_TAG" | grep -q "^${service}/"; then
            echo "  ✓ $service - CHANGED"
            CHANGED_SERVICES+=("$service")
        else
            echo "  - $service - no changes"
        fi
    done
fi

# Convert array to JSON array
SERVICES_JSON=$(printf '%s\n' "${CHANGED_SERVICES[@]}" | jq -R . | jq -s -c .)

if [[ ${#CHANGED_SERVICES[@]} -eq 0 ]]; then
    echo "⚠️  No services changed, but will deploy all to be safe"
    SERVICES_JSON='["gateway","users","authentication","notification"]'
fi

echo "services=$SERVICES_JSON" >> $GITHUB_OUTPUT
echo "🎯 Services to deploy: $SERVICES_JSON"
