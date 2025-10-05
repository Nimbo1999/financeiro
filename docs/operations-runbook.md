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
kubectl rollout undo deployment/SERVICE_NAME -n finaceiro

# Verify rollback
kubectl rollout status deployment/SERVICE_NAME -n finaceiro
```

### Viewing Logs

```bash
# Real-time logs
kubectl logs -f deployment/SERVICE_NAME -n finaceiro

# Last 100 lines
kubectl logs deployment/SERVICE_NAME -n finaceiro --tail=100

# All pods with specific label
kubectl logs -l app=SERVICE_NAME -n finaceiro
```

### Common Issues

#### Pod Not Starting

```bash
kubectl describe pod POD_NAME -n finaceiro
kubectl logs POD_NAME -n finaceiro
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
kubectl get pods -n finaceiro

# Check resource usage
kubectl top pods -n finaceiro
kubectl top nodes

# Check events
kubectl get events -n finaceiro --sort-by='.lastTimestamp'
```

### Database Migrations

Migrations run automatically on service startup. To run manually:

```bash
kubectl exec -it deployment/SERVICE_NAME -n finaceiro -- \
  /app/migrate -path /app/migrations -database "postgres://..." up
```
````
