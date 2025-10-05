## Comprehencive docs about project topics

- [How to write tests on golang services](docs/how-to-write-tests.md)

## Infrastructure

### Components

- **PostgreSQL 15 Alpine** (3 instances)
  - Users Database
  - Authentication Database
  - Notification Database
- **RabbitMQ 3 Management Alpine** (1 instance)

### Complete documentation

- [Infrastructure guide](docs/infrastructure-guide.md)
- [Operational guide](docs/operations-runbook.md)

### Initial Setup

```bash
# 1. Deploy Infrastructure on locally with make and Docker (dev)
make compose-up

# 2. Setup Infrastructure on K3S (prod)
ssh <USERNAME>@<VPS_IP> "bash -s" < .github/scripts/setup-infrastructure.sh

# 3. Verify infrastructure healthy
ssh <USERNAME>@<VPS_IP> "bash -s" < .github/scripts/check-infrastructure-health.sh
```

### Common commands

#### Create Migration file

```(bash)
# -seq
#   Use sequential numbers instead of timestamp
# -dir
#   Directory to place the file in
# -ext
#   File extension
# [NAME]
#   Migration name that would be added to the sql file
migrate create -seq -dir migrations -ext sql [NAME]
```
