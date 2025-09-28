### GRPC requests with grpcurl

The following are commands that can be used with the grpcurl CLI for sending requests to the
user's gRPC server

#### List available services

```bash
grpcurl -plaintext localhost:9091 list
```

#### List available methods of a service

```bash
grpcurl -plaintext localhost:9091 list users.v1.UserService
```

#### Get user by ID

```bash
grpcurl -plaintext -d '{"id":"<USER_ID>"}' localhost:9091 users.v1.UserService.GetUserById
```

#### Get user by email

```bash
grpcurl -plaintext -d '{"email": "<USER_EMAIL>"}' localhost:9091 users.v1.UserService.GetUserByEmail
```

#### Health check

```bash
grpcurl -plaintext localhost:9091 users.v1.UserService.HealthCheck
```
