#!/bin/bash

set -e

SERVICE=$1
NAMESPACE=${2:-default}

if [[ -z "$SERVICE" ]]; then
    echo "❌ Error: Service name required"
    exit 1
fi

echo "🔄 Rolling back $SERVICE in namespace $NAMESPACE..."

kubectl rollout undo deployment/$SERVICE -n $NAMESPACE

echo "⏳ Waiting for rollback to complete..."
kubectl rollout status deployment/$SERVICE -n $NAMESPACE --timeout=5m

echo "✅ Rollback completed successfully"

# Display the current pods for the service
kubectl get pods -n $NAMESPACE -l app=$SERVICE
