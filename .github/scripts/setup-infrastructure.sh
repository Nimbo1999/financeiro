#!/bin/bash

set -e

echo "🏗️  Setting up infrastructure in K3S..."

NAMESPACE="financeiro"

# Create host directories for persistent volumes
echo "📁 Creating volume directories..."
sudo mkdir -p /mnt/data/postgres-users
sudo mkdir -p /mnt/data/postgres-auth
sudo mkdir -p /mnt/data/postgres-notification
sudo mkdir -p /mnt/data/rabbitmq

# Apply infrastructure manifests
echo "📦 Creating namespace..."
kubectl apply -f k8s/base/namespace.yaml

echo "🔐 Creating secrets..."
kubectl apply -f k8s/infrastructure/postgres-secret.yaml
kubectl apply -f k8s/infrastructure/rabbitmq-secret.yaml
kubectl apply -f k8s/base/db-secrets.yaml
kubectl apply -f k8s/base/rabbitmq-secrets.yaml

echo "💾 Creating persistent volumes..."
kubectl apply -f k8s/infrastructure/postgres-pv.yaml
kubectl apply -f k8s/infrastructure/rabbitmq-pv.yaml

echo "🗄️  Deploying PostgreSQL instances..."
kubectl apply -f k8s/infrastructure/postgres-users.yaml
kubectl apply -f k8s/infrastructure/postgres-auth.yaml
kubectl apply -f k8s/infrastructure/postgres-notification.yaml

echo "🐰 Deploying RabbitMQ..."
kubectl apply -f k8s/infrastructure/rabbitmq.yaml

echo "⏳ Waiting for databases to be ready..."
kubectl wait --for=condition=ready pod -l app=postgres,database=users -n $NAMESPACE --timeout=300s
kubectl wait --for=condition=ready pod -l app=postgres,database=authentication -n $NAMESPACE --timeout=300s
kubectl wait --for=condition=ready pod -l app=postgres,database=notification -n $NAMESPACE --timeout=300s

echo "⏳ Waiting for RabbitMQ to be ready..."
kubectl wait --for=condition=ready pod -l app=rabbitmq -n $NAMESPACE --timeout=300s

echo "✅ Infrastructure setup completed!"

echo ""
echo "📊 Infrastructure Status:"
kubectl get pods -n $NAMESPACE -l 'app in (postgres,rabbitmq)'
kubectl get svc -n $NAMESPACE -l 'app in (postgres,rabbitmq)'

echo ""
echo "🔗 Access URLs:"
echo "  RabbitMQ Management: http://$(kubectl get nodes -o jsonpath='{.items[0].status.addresses[0].address}'):30672"
