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

echo "🔐 Setting up cert-manager for SSL certificates..."
# Check if cert-manager is already installed
if kubectl get namespace cert-manager &> /dev/null; then
    echo "ℹ️  cert-manager namespace already exists, skipping installation"
else
    echo "📦 Installing cert-manager..."
    kubectl apply -f https://github.com/cert-manager/cert-manager/releases/download/v1.18.2/cert-manager.yaml

    echo "⏳ Waiting for cert-manager to be ready..."
    kubectl wait --for=condition=ready pod -l app.kubernetes.io/instance=cert-manager -n cert-manager --timeout=300s
fi

echo "📜 Creating Let's Encrypt ClusterIssuers..."
kubectl apply -f k8s/infrastructure/cluster-issuer-staging.yaml
kubectl apply -f k8s/infrastructure/cluster-issuer-production.yaml

echo "⏳ Waiting for ClusterIssuers to be ready..."
sleep 5
kubectl get clusterissuer letsencrypt-staging
kubectl get clusterissuer letsencrypt-production

echo "✅ Infrastructure setup completed!"

echo ""
echo "📊 Infrastructure Status:"
kubectl get pods -n $NAMESPACE -l 'app in (postgres,rabbitmq)'
kubectl get svc -n $NAMESPACE -l 'app in (postgres,rabbitmq)'

echo ""
echo "🔗 Access URLs:"
echo "  RabbitMQ Management: http://$(kubectl get nodes -o jsonpath='{.items[0].status.addresses[0].address}'):30672"

echo ""
echo "📋 SSL Certificate Configuration:"
echo "  ✓ cert-manager installed"
echo "  ✓ ClusterIssuers created (staging & production)"
echo ""
echo "⚠️  Next steps for HTTPS:"
echo "  1. Configure DNS A record for your domain pointing to VPS IP"
echo "  2. Update k8s/services/gateway/service.yaml with your domain"
echo "  3. Deploy gateway service to obtain SSL certificate"
