# Infrastructure guide

## Overview

Infrastructure components:

- **3 instances PostgreSQL** (users, authentication, notification)
- **1 instance RabbitMQ** (with management UI)
- **Persistent Volumes** for persistent data
- **cert-manager** for automatic SSL certificate management
- **Let's Encrypt ClusterIssuers** (staging and production)

## Architecture

```
┌─────────────────────────────────────────────────────┐
│ Finance App Namespace                               │
│                                                     │
│ ┌──────────────┐ ┌──────────────┐ ┌───────────┐     │
│ │ postgres-    │ │ postgres-    │ │ postgres- │     │
│ │ users        │ │ auth         │ │ notif.    │     │
│ │ :5432        │ │ :5432        │ │ :5432     │     │
│ └──────┬───────┘ └──────┬───────┘ └─────┬─────┘     │
│        │                │               │           │
│ ┌──────▼────────────────▼───────────────▼────────┐  │
│ │ Application Services                           │  │
│ │ (users, authentication, notification)          │  │
│ └──────────────────────┬─────────────────────────┘  │
│                        │                            │
│ ┌──────────────────────▼───────┐                    │
│ │ RabbitMQ                     │                    │
│ │ :5672                        │                    │
│ │ :15672 (UI)                  │                    │
│ └──────────────────────────────┘                    │
└─────────────────────────────────────────────────────┘
```

## Recursos de Infraestrutura

### PostgreSQL Instances

#### Users Database

- **PVC**: `postgres-users-pvc` (3Gi)
- **Image**: `postgres:15-alpine3.22`

#### Authentication Database

- **PVC**: `postgres-auth-pvc` (10Gi)
- **Image**: `postgres:15-alpine3.22`

#### Notification Database

- **PVC**: `postgres-notification-pvc` (10Gi)
- **Image**: `postgres:15-alpine3.22`

### RabbitMQ

- **Management UI**: `rabbitmq:15672`
- **External Access**: `NodePort 30672`
- **PVC**: `rabbitmq-pvc` (5Gi)
- **Image**: `rabbitmq:3-management-alpine`

### cert-manager

- **Namespace**: `cert-manager`
- **Version**: `v1.18.2`
- **Purpose**: Automatic SSL/TLS certificate management from Let's Encrypt
- **ClusterIssuers**:
  - `letsencrypt-staging`: For testing certificates (avoid rate limits)
  - `letsencrypt-production`: For production certificates

## Utility commands

### Status verification

```bash
# General infrastructure status
kubectl get all -n financeiro -l 'app in (postgres,rabbitmq)'

# Database status
kubectl get pods -n financeiro -l app=postgres

# RabbitMQ status
kubectl get pods -n financeiro -l app=rabbitmq

# Verify PVCs
kubectl get pvc -n financeiro

# Verify storage usage
kubectl get pvc -n financeiro -o custom-columns=NAME:.metadata.name,CAPACITY:.spec.resources.requests.storage,USED:.status.capacity.storage

# cert-manager status
kubectl get pods -n cert-manager

# ClusterIssuers status
kubectl get clusterissuer

# Certificate status
kubectl get certificate -n financeiro

# Check certificate details
kubectl describe certificate <certificate-name> -n financeiro
```

### Logs

```bash
# Logs Users PostgreSQL
kubectl logs -f deployment/postgres-users -n financeiro

# Logs Auth PostgreSQL
kubectl logs -f deployment/postgres-auth -n financeiro

# Logs Notification PostgreSQL
kubectl logs -f deployment/postgres-notification -n financeiro

# Logs RabbitMQ
kubectl logs -f deployment/rabbitmq -n financeiro

# Logs recentes (últimas 50 linhas)
kubectl logs deployment/postgres-users -n financeiro --tail=50

# cert-manager logs
kubectl logs -n cert-manager -l app=cert-manager --tail=50

# cert-manager logs (follow)
kubectl logs -n cert-manager -l app=cert-manager -f
```

### Database access

```bash
# Connect to Users PostgreSQL
kubectl exec -it deployment/postgres-users -n financeiro -- \
  psql -U <USERNAME> -d <DATABASE_NAME>

# Connect to  Auth PostgreSQL
kubectl exec -it deployment/postgres-auth -n financeiro -- \
  psql -U <USERNAME> -d <DATABASE_NAME>

# Conectar ao PostgreSQL Notification
kubectl exec -it deployment/postgres-notification -n financeiro -- \
  psql -U <USERNAME> -d <DATABASE_NAME>

# Executar query específica
kubectl exec -it deployment/postgres-users -n financeiro -- \
  psql -U <USERNAME> -d <DATABASE_NAME> -c "SELECT * FROM users LIMIT 5;"
```

### Backup de Banco de Dados

```bash
# Backup Users DB
kubectl exec -i deployment/postgres-users -n financeiro -- \
  pg_dump -U <USERNAME> <DATABASE_NAME> > backup-users-$(date +%Y%m%d).sql

# Backup Auth DB
kubectl exec -i deployment/postgres-auth -n financeiro -- \
  pg_dump -U <USERNAME> <DATABASE_NAME> > backup-auth-$(date +%Y%m%d).sql

# Backup Notification DB
kubectl exec -i deployment/postgres-notification -n financeiro -- \
  pg_dump -U <USERNAME> <DATABASE_NAME> > backup-notification-$(date +%Y%m%d).sql

# Restaurar backup
kubectl exec -i deployment/postgres-users -n financeiro -- \
  psql -U <USERNAME> <DATABASE_NAME> < backup-users-20250104.sql
```

### Restore from specific backup

```bash
# 1. List available backups
kubectl exec deployment/postgres-users -n financeiro -- ls -lh /backups

# 2. Copy backup to local
kubectl cp financeiro/postgres-users-<POD-NAME>:/backups/backup-20250104_020000.tar.gz ./backup.tar.gz

# 3. Extrair backup
tar -xzf backup.tar.gz

# 4. Restaurar Users DB
kubectl exec -i deployment/postgres-users -n financeiro -- \
  pg_restore -U <USERNAME> -d <DATABASE_NAME> -c < 20250104_020000/users.dump

# 5. Restaurar Auth DB
kubectl exec -i deployment/postgres-auth -n financeiro -- \
  pg_restore -U <USERNAME> -d <DATABASE_NAME> -c < 20250104_020000/auth.dump

# 6. Restaurar Notification DB
kubectl exec -i deployment/postgres-notification -n financeiro -- \
  pg_restore -U <USERNAME> -d <DATABASE_NAME> -c < 20250104_020000/notification.dump
```

### Emergency restorations

If you need to restore the database fast:

```bash
# 1. scale services to zero (avoid data writes during the restoration)
kubectl scale deployment/users --replicas=0 -n financeiro
kubectl scale deployment/authentication --replicas=0 -n financeiro
kubectl scale deployment/notification --replicas=0 -n financeiro

# 2. Execute restouration (commands above)

# 3. Verify restored data
kubectl exec -it deployment/postgres-users -n financeiro -- \
  psql -U <USERNAME> -d <DATABASE> -c "SELECT COUNT(*) FROM users;"

# 4. Scale services back
kubectl scale deployment/users --replicas=2 -n financeiro
kubectl scale deployment/authentication --replicas=2 -n financeiro
kubectl scale deployment/notification --replicas=1 -n financeiro
```

### RabbitMQ Management

```bash
# Verify queues using API
kubectl exec deployment/rabbitmq -n financeiro -- \
  rabbitmqctl list_queues

# Verify exchanges
kubectl exec deployment/rabbitmq -n financeiro -- \
  rabbitmqctl list_exchanges

# Verify conexões
kubectl exec deployment/rabbitmq -n financeiro -- \
  rabbitmqctl list_connections

# Clear all the queues (CAUTION!!!!!!)
kubectl exec deployment/rabbitmq -n financeiro -- \
  rabbitmqctl purge_queue <QUEUE_NAME>
```

## Maintenance

### Restart services

```bash
# Restart Users PostgreSQL
kubectl rollout restart deployment/postgres-users -n financeiro

# Restart Auth PostgreSQL
kubectl rollout restart deployment/postgres-auth -n financeiro

# Restart Notification PostgreSQL
kubectl rollout restart deployment/postgres-notification -n financeiro

# Restart RabbitMQ
kubectl rollout restart deployment/rabbitmq -n financeiro

# Restart cert-manager
kubectl rollout restart deployment/cert-manager -n cert-manager
kubectl rollout restart deployment/cert-manager-webhook -n cert-manager
kubectl rollout restart deployment/cert-manager-cainjector -n cert-manager
```

### Scaling Resources

```bash
# Increase memory of Users PostgreSQL
kubectl set resources deployment/postgres-users -n financeiro \
  --limits=memory=1Gi --requests=memory=512Mi

# Verify current resource usage
kubectl top pods -n financeiro -l app=postgres
kubectl top pods -n financeiro -l app=rabbitmq
```

### Increase Storage

```bash
# Edit PVC (if storage class supports expansion)
kubectl edit pvc postgres-users-pvc -n financeiro

# Modify the field `spec.resources.requests.storage`
# Example: from 5Gi to 10Gi
```

## Troubleshooting

### Database is not responding

```bash
# Verify pod status
kubectl get pod -l app=postgres,database=users -n financeiro

# View events
kubectl describe pod -l app=postgres,database=users -n financeiro

# View logs
kubectl logs -l app=postgres,database=users -n financeiro --tail=100

# Verify if we can connect
kubectl exec -it deployment/postgres-users -n financeiro -- \
  pg_isready -U <USERNAME>
```

### RabbitMQ is not accessible

```bash
# Verify status
kubectl get pods -l app=rabbitmq -n financeiro

# View logs
kubectl logs -l app=rabbitmq -n financeiro --tail=100

# Verify diagnostics
kubectl exec deployment/rabbitmq -n financeiro -- \
  rabbitmq-diagnostics status

# Test connectivity
kubectl run test-rabbitmq --image=curlimages/curl -i --rm --restart=Never -n financeiro -- \
  curl -u <USERNAME>:<PASSWORD> http://rabbitmq:15672/api/overview
```

### SSL Certificate not issuing

```bash
# Check certificate status
kubectl describe certificate <certificate-name> -n financeiro

# Check certificate request
kubectl get certificaterequest -n financeiro
kubectl describe certificaterequest <request-name> -n financeiro

# Check ACME challenge
kubectl get challenge -n financeiro
kubectl describe challenge <challenge-name> -n financeiro

# Check cert-manager logs
kubectl logs -n cert-manager -l app=cert-manager --tail=100

# Verify ClusterIssuer status
kubectl describe clusterissuer letsencrypt-production

# Test DNS resolution (from within cluster)
kubectl run test-dns --image=busybox -i --rm --restart=Never -- nslookup <your-domain>
```

### Problems with Storage

```bash
# Verify PVCs
kubectl get pvc -n financeiro

# View specific PVC details
kubectl describe pvc postgres-users-pvc -n financeiro

# View PVs
kubectl get pv

# View disc usage on node
ssh <USERNAME>@<VPS_IP> "df -h /mnt/data"
```

## Monitoring

### Important metrics

```bash
# Active PostgresSQL connections
kubectl exec deployment/postgres-users -n financeiro -- \
  psql -U <USERNAME> -d <DATABASE_NAME> -c "SELECT count(*) FROM pg_stat_activity;"

# Database size
kubectl exec deployment/postgres-users -n financeiro -- \
  psql -U <USERNAME> -d <DATABASE_NAME> -c "SELECT pg_size_pretty(pg_database_size('<DATABASE_NAME>'));"

# RabbitMQ Messages
kubectl exec deployment/rabbitmq -n financeiro -- \
  rabbitmqctl list_queues name messages messages_ready messages_unacknowledged
```

## Security

### Password rotation

```bash
# 1. Generate new passwords
NEW_PASSWORD=$(openssl rand -base64 16)

# 2. Update Postgres user's password
kubectl exec -it deployment/postgres-users -n financeiro -- \
  psql -U <USERNAME> -d <DATABASE_NAME> -c "ALTER USER <USERNAME> WITH PASSWORD '$NEW_PASSWORD';"

# 3. Update secret
kubectl create secret generic db-secrets \
  --from-literal=users-db-url="postgresql://<USERNAME>:$NEW_PASSWORD@postgres-users:5432/<DATABASE_NAME>?sslmode=disable" \
  --dry-run=client -o yaml | kubectl apply -f -

# 4. Restart services that rely on secret
kubectl rollout restart deployment/users -n financeiro
```

## Disaster Recovery

### Complete backups

```bash
# Automatic backup script
#!/bin/bash
DATE=$(date +%Y%m%d_%H%M%S)
BACKUP_DIR="/backups/$DATE"
mkdir -p $BACKUP_DIR

# PostgreSQL Backups
for db in users auth notification; do
  kubectl exec deployment/postgres-$db -n financeiro -- \
    pg_dump -U ${db}user ${db}db > $BACKUP_DIR/postgres-$db.sql
done

# RabbitMQ definitions Backup
kubectl exec deployment/rabbitmq -n financeiro -- \
  rabbitmqctl export_definitions /tmp/rabbitmq-definitions.json

kubectl cp financeiro/$(kubectl get pod -l app=rabbitmq -n financeiro -o jsonpath='{.items[0].metadata.name}'):/tmp/rabbitmq-definitions.json \
  $BACKUP_DIR/rabbitmq-definitions.json

# Compact
tar -czf backup-$DATE.tar.gz $BACKUP_DIR
```

### Restoration

```bash
# Restore Users PostgreSQL
kubectl exec -i deployment/postgres-users -n financeiro -- \
  psql -U <USERNAME> <DATABASE_NAME> < backups/postgres-users.sql

# Restore RabbitMQ
kubectl cp backups/rabbitmq-definitions.json \
  financeiro/$(kubectl get pod -l app=rabbitmq -n financeiro -o jsonpath='{.items[0].metadata.name}'):/tmp/

kubectl exec deployment/rabbitmq -n financeiro -- \
  rabbitmqctl import_definitions /tmp/rabbitmq-definitions.json
```
