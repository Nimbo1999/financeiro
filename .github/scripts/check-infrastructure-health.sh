#!/bin/bash

set -e

NAMESPACE="financeiro"

echo "🏥 Checking infrastructure health..."

# Cores para output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

check_resource() {
    local resource_type=$1
    local resource_name=$2
    local label=$3

    if kubectl get $resource_type $resource_name -n $NAMESPACE &> /dev/null; then
        echo -e "${GREEN}✓${NC} $resource_type/$resource_name exists"
        return 0
    else
        echo -e "${RED}✗${NC} $resource_type/$resource_name not found"
        return 1
    fi
}

check_pod_ready() {
    local label=$1
    local name=$2

    if kubectl wait --for=condition=ready pod -l $label -n $NAMESPACE --timeout=10s &> /dev/null; then
        echo -e "${GREEN}✓${NC} $name is ready"
        return 0
    else
        echo -e "${YELLOW}⚠${NC} $name is not ready"
        return 1
    fi
}

# Check namespace
echo ""
echo "📦 Namespace:"
check_resource "namespace" $NAMESPACE

# Check secrets
echo ""
echo "🔐 Secrets:"
check_resource "secret" "postgres-credentials"
check_resource "secret" "rabbitmq-credentials"
check_resource "secret" "db-secrets"
check_resource "secret" "rabbitmq-secrets"

# Check PostgreSQL
echo ""
echo "🗄️  PostgreSQL Databases:"
check_resource "deployment" "postgres-users"
check_resource "deployment" "postgres-auth"
check_resource "deployment" "postgres-notification"
check_resource "service" "postgres-users"
check_resource "service" "postgres-auth"
check_resource "service" "postgres-notification"

echo ""
echo "PostgreSQL Pod Status:"
check_pod_ready "app=postgres,database=users" "postgres-users"
check_pod_ready "app=postgres,database=authentication" "postgres-auth"
check_pod_ready "app=postgres,database=notification" "postgres-notification"

# Check RabbitMQ
echo ""
echo "🐰 RabbitMQ:"
check_resource "deployment" "rabbitmq"
check_resource "service" "rabbitmq"
check_resource "service" "rabbitmq-management"

echo ""
echo "RabbitMQ Pod Status:"
check_pod_ready "app=rabbitmq" "rabbitmq"

# Check PVCs
echo ""
echo "💾 Persistent Volume Claims:"
check_resource "pvc" "postgres-users-pvc"
check_resource "pvc" "postgres-auth-pvc"
check_resource "pvc" "postgres-notification-pvc"
check_resource "pvc" "rabbitmq-pvc"

# Connection tests
echo ""
echo "🔌 Connection Tests:"

# Test PostgreSQL connections
# for db in users auth notification; do
#     if kubectl run psql-test-$db --image=postgres:15-alpine3.22 -i --rm --restart=Never -n $NAMESPACE -- \
#         psql postgresql://<USERNAME>:<PASSWORD>@<HOST>:5432/<DATABASE_NAME> -c "SELECT 1;" &> /dev/null; then
#         echo -e "${GREEN}✓${NC} PostgreSQL $db connection successful"
#     else
#         echo -e "${RED}✗${NC} PostgreSQL $db connection failed"
#     fi
# done

# Test RabbitMQ connection
# if kubectl run rabbitmq-test --image=curlimages/curl -i --rm --restart=Never -n $NAMESPACE -- \
#     curl -s -u <USERNAME>:<PASSWORD> http://rabbitmq:15672/api/overview &> /dev/null; then
#     echo -e "${GREEN}✓${NC} RabbitMQ connection successful"
# else
#     echo -e "${RED}✗${NC} RabbitMQ connection failed"
# fi

echo ""
echo "💾 Storage Usage:"
echo "─────────────────────────────────────────────────────────────"
kubectl get pvc -n $NAMESPACE -o custom-columns=\
  NAME:.metadata.name,\
  STATUS:.status.phase,\
  CAPACITY:.spec.resources.requests.storage

echo ""
echo "📈 Resource Usage:"
echo "─────────────────────────────────────────────────────────────"
kubectl top pods -n $NAMESPACE --no-headers 2>/dev/null || echo "Metrics server not available"

echo ""
echo "Last updated: $(date)"
echo "✅ Infrastructure health check completed!"
