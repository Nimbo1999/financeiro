# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is a microservices-based finance tracking application built with Go. The system allows users to upload bank transaction CSV files, analyze their financial data, and gain insights into their spending patterns. The architecture consists of three main services that communicate via gRPC and HTTP, with RabbitMQ for event-driven messaging.

## Architecture

### Services

1. **Users Service** (port 8081 HTTP, 9091 gRPC)
   - Manages user data and profiles
   - Exposes both HTTP REST and gRPC APIs
   - PostgreSQL database on port 15432

2. **Authentication Service** (port 8082 HTTP, 9092 gRPC)
   - Handles authentication and authorization
   - Issues JWT tokens using RSA256 signing
   - Uses RabbitMQ for publishing authentication events
   - Implements circuit breaker pattern for resilient gRPC client calls to Users service
   - PostgreSQL database on port 25432

3. **Gateway Service** (port 8083)
   - HTTP API Gateway that routes requests to backend services
   - Implements CORS and Circuit Breaker middleware (configurable via env vars)
   - Proxies HTTP requests to Users and Authentication services
   - Communicates with Authentication service via gRPC for token validation

### Key Architectural Patterns

- **Microservices**: Each service is independently deployable with its own database
- **gRPC Communication**: Services expose gRPC APIs for inter-service communication (defined in `/proto` directories)
- **Circuit Breaker**: Implemented in authentication service for gRPC calls, gateway middleware for HTTP proxying
- **Event-Driven Messaging**: RabbitMQ used for publishing events (e.g., authentication events)
- **API Gateway Pattern**: Gateway handles routing, authentication middleware, and cross-cutting concerns

### Module Dependencies

Services use Go module `replace` directives for local development:
- `authentication` depends on `users` (imports gRPC client from users/pkg/grpc)
- `gateway` depends on `authentication` (imports gRPC client from authentication/pkg/grpc)

## Development Commands

### Infrastructure

Start all dependencies (PostgreSQL, RabbitMQ, pgAdmin):
```bash
make compose-up
```

Stop and clean up containers:
```bash
make compose-down
```

View logs:
```bash
make compose-logs
```

Access PostgreSQL databases:
```bash
make psql-auth    # Authentication DB
make psql-user    # Users DB
```

### Service-Specific Commands

Each service has its own Makefile. Navigate to the service directory first.

#### Users Service

```bash
cd users

# Run database migrations
make migrate-up

# Run service locally
make run

# Build binary
make build

# Generate gRPC code from proto files
make generate-proto
```

#### Authentication Service

```bash
cd authentication

# Run database migrations
make migrate-up

# Run service locally
make run

# Build binary
make build

# Run tests with coverage
make unit-test

# Generate RSA key pairs for JWT signing
make g-cert-files
```

#### Gateway Service

```bash
cd gateway

# Run service locally
make run

# Build binary
make build

# Run tests with coverage
make unit-test
```

### Testing

Run tests for a specific service:
```bash
cd <service>
make unit-test  # Generates coverage.out, coverage.html, and prints total coverage
```

Run a single test file:
```bash
go test -v ./path/to/package -run TestName
```

## Database Migrations

Create a new migration:
```bash
# Use sequential numbering instead of timestamps
migrate create -seq -dir migrations -ext sql [MIGRATION_NAME]
```

Run migrations:
```bash
cd <service>
make migrate-up
```

## Project Structure

Each service follows a similar internal structure:

- `cmd/`: Entry points (main.go files)
- `internal/`: Private application code
  - `app/`: Application initialization and server setup
  - `handler/`: HTTP/gRPC request handlers
  - `services/`: Business logic layer
  - `repository/`: Database access layer
  - `models/`: Data models
  - `config/`: Configuration management (environment variables)
  - `grpc/`: gRPC server setup and interceptors (logging)
  - `middleware/`: HTTP middleware (authentication service only)
  - `clients/`: gRPC client wrappers (authentication service)
  - `messaging/`: RabbitMQ connection and publisher (authentication service)
  - `utils/`: Utility functions
- `pkg/grpc/`: Generated gRPC code (exported for other services to consume)
- `proto/`: Protocol buffer definitions
- `migrations/`: SQL migration files

## gRPC Integration

Services expose gRPC clients in their `pkg/grpc/<service>/v1/` directories. Other services import these packages to make gRPC calls:

- Users service exposes `pkg/grpc/users/v1/`
- Authentication service exposes `pkg/grpc/auth/v1/`

To regenerate proto files after modifying `.proto` definitions:
```bash
cd users
./scripts/generate_proto.sh
# OR
make generate-proto
```

## Environment Configuration

Each service reads configuration from environment variables. Key variables:

**Users:**
- `HTTP_PORT`, `GRPC_PORT`, `POSTGRES_CONNECTION_STRING`

**Authentication:**
- `HTTP_PORT`, `GRPC_PORT`, `POSTGRES_CONNECTION_STRING`
- `RABBITMQ_URL`, `USER_GRPC_ADDRESS`

**Gateway:**
- `GATEWAY_PORT`, `USER_SERVICE_URL`, `AUTH_SERVICE_URL`, `AUTH_SERVICE_GRPC_URL`
- `GATEWAY_ENABLE_CORS`, `GATEWAY_ENABLE_CIRCUIT_BREAKER`

Default values are defined in each service's Makefile.

## Starting Services Locally

Services must be started in dependency order:

1. Start infrastructure: `make compose-up`
2. Run migrations for each service: `cd <service> && make migrate-up`
3. Generate JWT certificates (first time only): `cd authentication && make g-cert-files`
4. Start users service: `cd users && make run`
5. Start authentication service: `cd authentication && make run`
6. Start gateway service: `cd gateway && make run`

## Port Reference

- **5672**: RabbitMQ AMQP
- **15672**: RabbitMQ Management UI
- **15432**: Users PostgreSQL
- **25432**: Authentication PostgreSQL
- **80**: pgAdmin UI
- **8081**: Users HTTP API
- **8082**: Authentication HTTP API
- **8083**: Gateway HTTP API
- **9091**: Users gRPC
- **9092**: Authentication gRPC

## Design Principles

This project follows **SOLID principles** and clean architecture patterns. For detailed guidance, see `docs/project-design-principles.md`.

### Key Principles

**Layered Architecture**: Each service separates concerns into distinct layers:
- **Handler Layer**: HTTP/gRPC request/response handling
- **Service Layer**: Business logic and orchestration
- **Repository Layer**: Data persistence
- **Clients Layer**: External service communication
- **Messaging Layer**: Event publishing

**Dependency Flow**: Dependencies always point inward (toward business logic). Services depend on interfaces, not concrete implementations.

**Interface-First Design**: Define interfaces before implementations. All dependencies are injected through constructors that accept interfaces:

```go
// Services depend on interfaces
func NewAuthService(
    authRepo repository.AuthCodeRepository,    // Interface
    jwtService JWTService,                     // Interface
    userClient clients.UserServiceClient,      // Interface
    publisher messaging.Publisher,             // Interface
    config *AuthConfig,
) AuthService {
    return &authService{...}
}
```

**Error Handling**: Use sentinel errors for common cases, wrap errors with context using `fmt.Errorf("context: %w", err)`, and handle failures gracefully.

**Validation**: Validate inputs at service boundaries, fail fast with clear error messages.

**Async Operations**: Run non-critical operations (like event publishing) asynchronously using goroutines to avoid blocking main request flow.

## Testing Patterns

This project follows Test-Driven Development (TDD) using testify's suite pattern. For detailed testing guide, see `docs/how-to-write-tests.md`.

### Testing Libraries

- `github.com/stretchr/testify`: Assertions, mocks, and test suites
- `github.com/DATA-DOG/go-sqlmock`: Mock database operations
- `net/http/httptest`: Test HTTP handlers

### Test Structure

All tests use testify's test suite pattern:

```go
type YourTestSuite struct {
    suite.Suite
    // Add dependencies and mocks
}

func (suite *YourTestSuite) SetupTest() {
    // Initialize mocks before each test
}

func (suite *YourTestSuite) TearDownTest() {
    // Assert mock expectations after each test
}

func TestYourTestSuite(t *testing.T) {
    suite.Run(t, new(YourTestSuite))
}
```

### Testing by Layer

**Repository Tests**: Use `sqlmock` to mock database interactions. Test CRUD operations, validation errors, and edge cases.

**Service Tests**: Mock all dependencies (repositories, clients, publishers). Test business logic, validation, error scenarios, and async operations.

**Handler Tests**: Use `httptest` to simulate HTTP requests. Mock service layer. Test request parsing, response formatting, and HTTP status codes.

**Event Tests**: Test event creation, serialization, and interface implementation.

### Test Naming Convention

Format: `Test<MethodName>_<Scenario>`

```go
func (suite *YourTestSuite) TestCreate_Success() {}
func (suite *YourTestSuite) TestCreate_ValidationError() {}
func (suite *YourTestSuite) TestCreate_DatabaseError() {}
func (suite *YourTestSuite) TestFindByID_NotFound() {}
```

### Arrange-Act-Assert Pattern

```go
func (suite *YourTestSuite) TestMethod() {
    // Arrange: Setup test data and mocks
    input := &Input{Field: "value"}
    suite.mock.On("Method", mock.Anything).Return(nil)

    // Act: Execute the method being tested
    result, err := suite.service.Method(context.Background(), input)

    // Assert: Verify the results
    assert.NoError(suite.T(), err)
    assert.NotNil(suite.T(), result)
}
```

### Test Coverage Requirements

Ensure tests cover:
- ✅ Success scenarios
- ✅ Validation errors (nil inputs, empty strings, invalid formats)
- ✅ Not found scenarios
- ✅ Database/repository errors
- ✅ External service errors
- ✅ Edge cases

### Running Specific Tests

```bash
# Run all tests
go test ./...

# Run with verbose output
go test ./... -v

# Run specific test suite
go test ./internal/services -v

# Run specific test case
go test ./internal/services -v -run TestAuthServiceTestSuite/TestRequestAuthCode_Success
```
