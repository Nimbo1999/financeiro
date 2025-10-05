# Operations Runbook

## Deployment Process

### Creating a New Release

1. Ensure all changes are merged to main
2. Create and push a semver tag:
   ```bash
   git tag -a vX.Y.Z -m "Release version X.Y.Z"
   git push origin vX.Y.Z
   ```

````

3. Monitor GitHub Actions: https://github.com/Nimbo1999/financeiro/actions
4. Verify deployment on VPS

### Rollback Procedure

#### Automated Rollback

If health checks fail, the pipeline automatically rolls back.

#### Manual Rollback

```bash
# Connect to VPS
ssh -i ~/.ssh/key <USERNAME>@<VPS_IP>

# Rollback specific service
kubectl rollout undo deployment/SERVICE_NAME -n financeiro

# Verify rollback
kubectl rollout status deployment/SERVICE_NAME -n financeiro
```

### Viewing Logs

```bash
# Real-time logs
kubectl logs -f deployment/SERVICE_NAME -n financeiro

# Last 100 lines
kubectl logs deployment/SERVICE_NAME -n financeiro --tail=100

# All pods with specific label
kubectl logs -l app=SERVICE_NAME -n financeiro
```

### Common Issues

#### Pod Not Starting

```bash
kubectl describe pod POD_NAME -n financeiro
kubectl logs POD_NAME -n financeiro
```

#### ImagePullBackOff

- Verify registry credentials
- Check image exists in registry
- Verify network connectivity

#### CrashLoopBackOff

- Check application logs
- Verify environment variables
- Check database connectivity

### Monitoring

```bash
# Check pod status
kubectl get pods -n financeiro

# Check resource usage
kubectl top pods -n financeiro
kubectl top nodes

# Check events
kubectl get events -n financeiro --sort-by='.lastTimestamp'
```

### Database Migrations

Migrations run automatically on service startup. To run manually:

```bash
kubectl exec -it deployment/SERVICE_NAME -n financeiro -- \
  /app/migrate -path /app/migrations -database "postgres://..." up
```
````
