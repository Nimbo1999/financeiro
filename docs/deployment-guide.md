# Production Deployment Guide

## Prerequisites

- VPS with K3S installed
- GitHub Actions configured
- Secrets configured in GitHub
- SSH access to the VPS

## Initial Deployment (First Time)

### 1. Infrastructure Setup

```bash
# Connect to the VPS
ssh -i ~/.ssh/key <USERNAME>@<VPS_IP>

# Clone repository
git clone https://github.com/Nimbo1999/financeiro.git
cd financeiro

# Run infrastructure setup
bash .github/scripts/setup-infrastructure.sh

# Check status
kubectl get all -n financeiro
```

### 2. Application Deployment

```bash
# On your local machine
git tag -a v1.0.0 -m "Production release v1.0.0"
git push --tags

# Wait for pipeline to complete (5-10 minutes)
# Monitor at: https://github.com/Nimbo1999/financeiro/actions
```

### 3. Post-Deployment Verification

```bash
# Check all pods
ssh <USERNAME>@<VPS_IP> "kubectl get pods -n financeiro"

# All should be Running/Completed

# Access services
curl http://VPS-IP/health  # Gateway
curl http://VPS-IP:30672   # RabbitMQ UI
```

## Update Deployment

```bash
# 1. Make code changes
git add .
git commit -m "feat: new feature"
git push origin main

# 2. Create release tag
git tag -a v1.1.0 -m "Release v1.1.0"
git push origin v1.1.0

# 3. Wait for automatic deployment
```

## Rollback

### Automatic

If health checks fail, rollback is automatic.

### Manual

```bash
ssh <USERNAME>@<VPS_IP>

# Rollback specific service
kubectl rollout undo deployment/SERVICE_NAME -n financeiro

# Check status
kubectl rollout status deployment/SERVICE_NAME -n financeiro
```

## Monitoring

```bash
# View real-time logs
kubectl logs -f deployment/SERVICE_NAME -n financeiro

# View recent events
kubectl get events -n financeiro --sort-by='.lastTimestamp' | tail -20

# Use monitoring script
bash .github/scripts/monitor-infrastructure.sh
```

## Backup and Restore

Automatic backups run daily at 2 AM.

### Manual Backup

```bash
kubectl create job --from=cronjob/database-backup manual-backup-$(date +%s) -n financeiro
```

### Restore

See [Infrastructure Guide](./infrastructure-guide.md#restore-from-specific-backup)

## Troubleshooting

See [Operations Runbook](./operations-runbook.md#troubleshooting)

## Support Contacts

- GitHub Issues: https://github.com/Nimbo1999/financeiro/issues
- Email: matlopes1999@gmail.com
