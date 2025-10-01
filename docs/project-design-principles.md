# Project Design Principles

This document outlines the core design principles and architectural patterns used in this project. We follow **SOLID principles** and industry best practices to create maintainable, testable, and scalable microservices.

## Table of Contents

1. [SOLID Principles Overview](#solid-principles-overview)
2. [Single Responsibility Principle (SRP)](#single-responsibility-principle-srp)
3. [Open/Closed Principle (OCP)](#openclosed-principle-ocp)
4. [Liskov Substitution Principle (LSP)](#liskov-substitution-principle-lsp)
5. [Interface Segregation Principle (ISP)](#interface-segregation-principle-isp)
6. [Dependency Inversion Principle (DIP)](#dependency-inversion-principle-dip)
7. [Additional Design Patterns](#additional-design-patterns)
8. [Project Structure](#project-structure)
9. [Best Practices](#best-practices)

## SOLID Principles Overview

SOLID is an acronym for five design principles that make software designs more understandable, flexible, and maintainable:

- **S** - Single Responsibility Principle
- **O** - Open/Closed Principle
- **L** - Liskov Substitution Principle
- **I** - Interface Segregation Principle
- **D** - Dependency Inversion Principle

## Single Responsibility Principle (SRP)

> A class/module should have only one reason to change.

Each component in our system has a single, well-defined responsibility.

### Example: Layered Architecture

We separate concerns into distinct layers:

#### Repository Layer
**Responsibility**: Data persistence and retrieval

```go
// authentication/internal/repository/auth_code_repository.go
type AuthCodeRepository interface {
    Create(ctx context.Context, authCode *models.AuthCode) error
    FindByUserID(ctx context.Context, userId string) (*models.AuthCode, error)
    MarkAsUsed(ctx context.Context, id string) error
    CleanExpired(ctx context.Context) error
}
```

The repository only handles database operations. It doesn't contain business logic, HTTP handling, or event publishing.

#### Service Layer
**Responsibility**: Business logic and orchestration

```go
// authentication/internal/services/auth_service.go
type AuthService interface {
    RequestAuthCode(ctx context.Context, email string) (*AuthCodeResult, error)
    VerifyAuthCode(ctx context.Context, email, code string) (*AuthResult, error)
    RefreshTokens(ctx context.Context, refreshToken string) (*TokenPair, error)
    CleanExpiredCodes(ctx context.Context) error
}
```

The service layer orchestrates business logic but doesn't handle HTTP requests/responses or database queries directly.

#### Handler Layer
**Responsibility**: HTTP request/response handling

```go
// authentication/internal/handler/auth_handler.go
type AuthHandler struct {
    authService services.AuthService
}

func (h *AuthHandler) RequestCodeHandler(w http.ResponseWriter, r *http.Request) {
    // 1. Parse and validate request
    // 2. Call service
    // 3. Format and return response
}
```

Handlers only deal with HTTP concerns: parsing requests, calling services, and formatting responses.

#### Messaging Layer
**Responsibility**: Event publishing and messaging

```go
// authentication/internal/messaging/publisher.go
type Publisher interface {
    PublishEvent(ctx context.Context, event Event) error
    PublishWithRetry(ctx context.Context, event Event, maxRetries int) error
    Close() error
    IsHealthy() bool
}
```

The messaging layer is solely responsible for publishing events to message brokers.

### Why This Matters

- **Easier Testing**: Each layer can be tested in isolation with mocks
- **Maintainability**: Changes to one concern don't affect others
- **Clarity**: Each component's purpose is immediately clear

## Open/Closed Principle (OCP)

> Software entities should be open for extension but closed for modification.

We use interfaces and composition to allow extending functionality without modifying existing code.

### Example: Multiple Repository Implementations

```go
// Interface definition (closed for modification)
type AuthCodeRepository interface {
    Create(ctx context.Context, authCode *models.AuthCode) error
    FindByUserID(ctx context.Context, userId string) (*models.AuthCode, error)
    MarkAsUsed(ctx context.Context, id string) error
    CleanExpired(ctx context.Context) error
}

// PostgreSQL implementation
type PostgresAuthCodeRepository struct {
    db *gorm.DB
}

// Future: MongoDB implementation (extension without modification)
type MongoAuthCodeRepository struct {
    client *mongo.Client
}

// Future: Redis implementation (extension without modification)
type RedisAuthCodeRepository struct {
    client *redis.Client
}
```

The `AuthService` depends on the `AuthCodeRepository` interface, not the concrete implementation. You can add new storage backends without changing the service code.

### Example: Service Extensions

```go
// Base service interface
type AuthService interface {
    RequestAuthCode(ctx context.Context, email string) (*AuthCodeResult, error)
    VerifyAuthCode(ctx context.Context, email, code string) (*AuthResult, error)
}

// Enhanced service with additional features (extension)
type EnhancedAuthService interface {
    AuthService // Embeds base interface
    RequestAuthCodeWithSMS(ctx context.Context, email, phone string) (*AuthCodeResult, error)
    VerifyAuthCodeWithBiometrics(ctx context.Context, email, code, biometric string) (*AuthResult, error)
}
```

### Why This Matters

- **Flexibility**: Add new features without breaking existing code
- **Reduced Risk**: Existing functionality remains untouched
- **Backward Compatibility**: Old code continues to work with new implementations

## Liskov Substitution Principle (LSP)

> Objects of a superclass should be replaceable with objects of its subclasses without breaking the application.

All implementations of an interface must fulfill the contract defined by that interface.

### Example: Repository Implementations

```go
// Any implementation of AuthCodeRepository can be substituted
func NewAuthService(
    authRepo repository.AuthCodeRepository,  // Any implementation works
    jwtService JWTService,
    userServiceClient clients.UserServiceClient,
    publisher messaging.Publisher,
    config *AuthConfig,
) AuthService {
    return &authService{
        authRepo:          authRepo,
        jwtService:        jwtService,
        userServiceClient: userServiceClient,
        publisher:         publisher,
        config:            config,
        rateLimiter:       make(map[string]*rateLimitEntry),
    }
}
```

Whether you pass `PostgresAuthCodeRepository`, `MongoAuthCodeRepository`, or a mock repository for testing, the service works correctly because all implementations respect the interface contract.

### Example: Client Implementations

```go
// Interface contract
type UserServiceClient interface {
    GetUserByEmail(ctx context.Context, email string) (*userv1.User, bool, error)
    GetUserById(ctx context.Context, userID string) (*userv1.User, bool, error)
    HealthCheck(ctx context.Context) (userv1.HealthCheckResponse_Status, string, error)
    Close() error
}

// gRPC implementation
type userServiceClient struct {
    conn   *grpc.ClientConn
    client userv1.UserServiceClient
}

// Mock implementation for testing
type MockUserServiceClient struct {
    mock.Mock
}
```

Both implementations can be used interchangeably without modifying the service that depends on `UserServiceClient`.

### Why This Matters

- **Predictability**: Implementations behave as expected
- **Testability**: Easy to create mock implementations
- **Reliability**: No surprises when swapping implementations

## Interface Segregation Principle (ISP)

> Clients should not be forced to depend on interfaces they don't use.

We create focused, minimal interfaces rather than large, monolithic ones.

### Example: Separated Service Interfaces

Instead of one large interface:

```go
// ❌ BAD: Monolithic interface
type AuthenticationSystem interface {
    // Auth code methods
    RequestAuthCode(ctx context.Context, email string) (*AuthCodeResult, error)
    VerifyAuthCode(ctx context.Context, email, code string) (*AuthResult, error)
    CleanExpiredCodes(ctx context.Context) error

    // JWT methods
    GenerateTokenPair(ctx context.Context, userID, email string) (*TokenPair, error)
    ValidateAccessToken(ctx context.Context, tokenString string) (*UserContext, error)
    ValidateRefreshToken(ctx context.Context, tokenString string) (*UserContext, error)

    // User methods
    GetUserByEmail(ctx context.Context, email string) (*User, bool, error)
    GetUserById(ctx context.Context, userID string) (*User, bool, error)

    // Messaging methods
    PublishEvent(ctx context.Context, event Event) error
}
```

We split into focused interfaces:

```go
// ✅ GOOD: Focused interfaces
type AuthService interface {
    RequestAuthCode(ctx context.Context, email string) (*AuthCodeResult, error)
    VerifyAuthCode(ctx context.Context, email, code string) (*AuthResult, error)
    RefreshTokens(ctx context.Context, refreshToken string) (*TokenPair, error)
    CleanExpiredCodes(ctx context.Context) error
}

type JWTService interface {
    GenerateTokenPair(ctx context.Context, userID, email string) (*TokenPair, error)
    ValidateAccessToken(ctx context.Context, tokenString string) (*UserContext, error)
    ValidateRefreshToken(ctx context.Context, tokenString string) (*UserContext, error)
    RefreshTokens(ctx context.Context, refreshToken string) (*TokenPair, error)
    GetPublicKey() *rsa.PublicKey
}

type UserServiceClient interface {
    GetUserByEmail(ctx context.Context, email string) (*User, bool, error)
    GetUserById(ctx context.Context, userID string) (*User, bool, error)
    HealthCheck(ctx context.Context) (HealthCheckResponse_Status, string, error)
    Close() error
}

type Publisher interface {
    PublishEvent(ctx context.Context, event Event) error
    PublishWithRetry(ctx context.Context, event Event, maxRetries int) error
    Close() error
    IsHealthy() bool
}
```

### Example: Event Interface

Events implement a minimal interface:

```go
type Event interface {
    GetType() EventType
    GetID() string
    GetTimestamp() time.Time
    ToJSON() ([]byte, error)
    GetRoutingKey() string
}
```

Publishers only need these methods. They don't need to know about event-specific data structures.

### Why This Matters

- **Flexibility**: Components only depend on what they actually use
- **Easier Mocking**: Smaller interfaces are easier to mock in tests
- **Reduced Coupling**: Changes to unused methods don't affect clients

## Dependency Inversion Principle (DIP)

> High-level modules should not depend on low-level modules. Both should depend on abstractions.

We depend on interfaces (abstractions) rather than concrete implementations.

### Example: Service Dependencies

```go
// High-level module (AuthService) depends on abstractions, not concrete types
type authService struct {
    authRepo          repository.AuthCodeRepository    // Interface, not *PostgresAuthCodeRepository
    jwtService        JWTService                       // Interface, not *jwtService
    userServiceClient clients.UserServiceClient        // Interface, not *userServiceClient
    publisher         messaging.Publisher              // Interface, not *publisher
    config            *AuthConfig
    rateLimiter       map[string]*rateLimitEntry
}

func NewAuthService(
    authRepo repository.AuthCodeRepository,    // Accept interfaces
    jwtService JWTService,                     // Accept interfaces
    userServiceClient clients.UserServiceClient, // Accept interfaces
    publisher messaging.Publisher,             // Accept interfaces
    config *AuthConfig,
) AuthService {
    // ...
}
```

### Example: Dependency Injection in Application Initialization

```go
// authentication/internal/app/app.go
func (a *App) RunHTTP(config *config.Config, jwtService services.JWTService) error {
    // Create concrete implementations
    authCodeRepository := repository.NewPostgresAuthCodeRepository(a.db)

    userServiceClient, err := clients.NewUserServiceClient(clients.UserServiceConfig{
        Address: config.UserGRPCAddress,
    })
    if err != nil {
        return fmt.Errorf("failed to create user service client: %w", err)
    }

    rabbitMqConnection := messaging.NewRabbitMQConnection(messaging.RabbitMQConfig{
        URL: config.RabbitMQURL,
    })

    publisher, err := messaging.NewPublisher(rabbitMqConnection, queueManager, config)
    if err != nil {
        return fmt.Errorf("failed to create RabbitMQ publisher: %w", err)
    }

    // Inject dependencies (as interfaces)
    authCodeService := services.NewAuthService(
        authCodeRepository,    // Injected
        jwtService,           // Injected
        userServiceClient,    // Injected
        publisher,            // Injected
        nil,                  // Config
    )

    authHandler := handler.NewAuthHandler(authCodeService) // Injected

    // ...
}
```

### Benefits of Dependency Inversion

#### 1. **Testability**

```go
// In tests, inject mocks instead of real implementations
func (suite *AuthServiceTestSuite) SetupTest() {
    suite.mockRepo = new(MockAuthCodeRepository)
    suite.mockJWTService = new(MockJWTService)
    suite.mockUserClient = new(MockUserServiceClient)
    suite.mockPublisher = new(MockPublisher)

    // Service doesn't care if these are mocks or real implementations
    suite.service = services.NewAuthService(
        suite.mockRepo,
        suite.mockJWTService,
        suite.mockUserClient,
        suite.mockPublisher,
        &services.AuthConfig{},
    )
}
```

#### 2. **Flexibility**

```go
// Easy to switch implementations
// Development: Use PostgreSQL
authRepo := repository.NewPostgresAuthCodeRepository(db)

// Production: Use distributed cache + PostgreSQL
authRepo := repository.NewCachedAuthCodeRepository(
    repository.NewPostgresAuthCodeRepository(db),
    redisClient,
)

// Testing: Use in-memory mock
authRepo := &MockAuthCodeRepository{}

// Service code remains unchanged!
authService := services.NewAuthService(authRepo, ...)
```

#### 3. **Decoupling**

The `AuthService` doesn't know or care:
- Which database is used (PostgreSQL, MongoDB, MySQL)
- How JWT tokens are signed (RSA, HMAC, ECDSA)
- How events are published (RabbitMQ, Kafka, SQS)
- Where user data comes from (gRPC, REST, database)

It only knows the interface contracts.

### Why This Matters

- **Maintainability**: Change implementations without touching business logic
- **Testability**: Easy to inject mocks for unit testing
- **Flexibility**: Swap dependencies based on environment or requirements
- **Parallel Development**: Teams can work on different implementations simultaneously

## Additional Design Patterns

### 1. Constructor Pattern with Configuration

We use dedicated configuration structs with sensible defaults:

```go
type AuthConfig struct {
    CodeLength         int           // Length of auth code (default: 6)
    CodeExpiryDuration time.Duration // How long codes are valid (default: 5 minutes)
    RateLimitWindow    time.Duration // Rate limiting window (default: 1 hour)
    MaxCodesPerEmail   int           // Max codes per email in window (default: 3)
    MaxRequestsPerIP   int           // Max requests per IP in window (default: 5)
    CleanupInterval    time.Duration // How often to clean expired codes (default: 1 hour)
}

func NewAuthService(..., config *AuthConfig) AuthService {
    if config == nil {
        config = &AuthConfig{}
    }

    // Set defaults
    if config.CodeLength == 0 {
        config.CodeLength = 6
    }
    if config.CodeExpiryDuration == 0 {
        config.CodeExpiryDuration = 5 * time.Minute
    }
    // ... other defaults
}
```

**Benefits**:
- Easy to customize behavior without changing code
- Sensible defaults for quick setup
- Configuration is explicit and documented

### 2. Error Handling with Sentinel Errors

We define package-level error variables for common errors:

```go
// Service layer errors
var (
    ErrEmailInvalid           = errors.New("email is invalid")
    ErrCodeExpired            = errors.New("auth code has expired")
    ErrCodeAlreadyUsed        = errors.New("auth code has already been used")
    ErrRateLimitExceeded      = errors.New("rate limit exceeded")
    ErrUserNotFound           = errors.New("user not found")
)

// Repository layer errors
var (
    ErrAuthCodeNotFound = errors.New("auth code not found")
    ErrAuthCodeExpired  = errors.New("auth code expired")
    ErrAuthCodeUsed     = errors.New("auth code already used")
)
```

**Usage**:

```go
// Check for specific errors
authCode, err := s.authRepo.FindByUserID(ctx, userID)
if err != nil {
    if errors.Is(err, repository.ErrAuthCodeNotFound) {
        return nil, ErrInvalidAuthCode
    }
    return nil, fmt.Errorf("failed to find auth code: %w", err)
}
```

**Benefits**:
- Errors can be checked programmatically
- Clear error propagation with `errors.Is()` and `errors.As()`
- Wrapping errors preserves context

### 3. Context Propagation

We always pass `context.Context` as the first parameter:

```go
func (s *authService) RequestAuthCode(ctx context.Context, email string) (*AuthCodeResult, error) {
    // Use context for timeouts
    userCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
    defer cancel()

    user, found, err := s.userServiceClient.GetUserByEmail(userCtx, email)
    // ...
}
```

**Benefits**:
- Timeout and cancellation propagation
- Request-scoped values (trace IDs, etc.)
- Graceful shutdown support

### 4. Validation at Boundaries

Input validation happens at service boundaries:

```go
func (s *authService) RequestAuthCode(ctx context.Context, email string) (*AuthCodeResult, error) {
    // Validate at service entry point
    if email == "" {
        return nil, ErrEmailInvalid
    }

    // Normalize input
    email = strings.ToLower(strings.TrimSpace(email))

    // Validate format
    if !utils.IsValidEmail(email) {
        return nil, ErrEmailInvalid
    }

    // Proceed with business logic
    // ...
}
```

**Benefits**:
- Fail fast with clear error messages
- Prevent invalid data from entering the system
- Consistent validation logic

### 5. Async Operations with Goroutines

Non-critical operations run asynchronously:

```go
func (s *authService) RequestAuthCode(ctx context.Context, email string) (*AuthCodeResult, error) {
    // ... business logic ...

    // Publish event asynchronously (don't block on event publishing)
    go s.publishAuthCodeRequestedEvent(
        context.Background(), // New context, not tied to request
        email,
        code,
        authCode.ID,
        authCode.ExpiresAt,
    )

    return &AuthCodeResult{
        CodeID:    authCode.ID,
        ExpiresAt: authCode.ExpiresAt,
        Success:   true,
    }, nil
}
```

**Benefits**:
- Non-blocking operations improve response times
- Event publishing failures don't fail the main request
- Better user experience

### 6. Resource Cleanup with defer

Always clean up resources:

```go
func NewUserServiceClient(config UserServiceConfig) (UserServiceClient, error) {
    conn, err := grpc.NewClient(config.Address, options...)
    if err != nil {
        return nil, fmt.Errorf("failed to create client: %w", err)
    }

    ctx, cancel := context.WithTimeout(context.Background(), config.ConnectTimeout)
    defer cancel() // Always cleanup

    // ... connection logic ...
}
```

**Benefits**:
- No resource leaks
- Guaranteed cleanup even on errors
- Clear resource lifecycle

## Project Structure

Our project follows a clean architecture with clear separation of concerns:

```
authentication/
├── cmd/
│   └── server/          # Application entry point
│       └── main.go
├── internal/
│   ├── app/            # Application initialization and wiring
│   │   └── app.go
│   ├── clients/        # External service clients
│   │   └── user_service_client.go
│   ├── config/         # Configuration management
│   │   └── config.go
│   ├── handler/        # HTTP handlers (presentation layer)
│   │   ├── auth_handler.go
│   │   ├── health_handler.go
│   │   └── middleware.go
│   ├── messaging/      # Event publishing and messaging
│   │   ├── events.go
│   │   ├── publisher.go
│   │   └── rabbitmq.go
│   ├── models/         # Domain models
│   │   └── auth_code.go
│   ├── repository/     # Data access layer
│   │   ├── auth_code_repository.go
│   │   └── postgres_auth_code_repository.go
│   ├── services/       # Business logic layer
│   │   ├── auth_service.go
│   │   └── jwt_service.go
│   └── utils/          # Shared utilities
│       └── validation.go
└── pkg/                # Public packages (if any)
```

### Layer Responsibilities

| Layer | Responsibility | Depends On |
|-------|---------------|------------|
| **Handler** | HTTP request/response handling | Services |
| **Services** | Business logic and orchestration | Repository, Clients, Messaging |
| **Repository** | Data persistence and retrieval | Database |
| **Clients** | Communication with external services | gRPC/HTTP |
| **Messaging** | Event publishing | Message Broker |
| **Models** | Domain entities | Nothing (pure data) |

### Dependency Flow

```
Handler → Service → Repository → Database
                 → Clients → External Services
                 → Messaging → Message Broker
```

Dependencies always point inward (toward business logic), never outward.

## Best Practices

### 1. Interface-First Design

Define interfaces before implementations:

```go
// 1. Define interface
type AuthCodeRepository interface {
    Create(ctx context.Context, authCode *models.AuthCode) error
    FindByUserID(ctx context.Context, userId string) (*models.AuthCode, error)
}

// 2. Implement interface
type PostgresAuthCodeRepository struct {
    db *gorm.DB
}

func (r *PostgresAuthCodeRepository) Create(ctx context.Context, authCode *models.AuthCode) error {
    // Implementation
}
```

### 2. Constructor Functions

Always use constructor functions that return interfaces:

```go
// Return interface, not concrete type
func NewAuthService(...) AuthService {
    return &authService{...}
}

func NewPostgresAuthCodeRepository(db *gorm.DB) AuthCodeRepository {
    return &PostgresAuthCodeRepository{db: db}
}
```

### 3. Explicit Dependencies

Make dependencies explicit in constructors:

```go
// ✅ GOOD: Explicit dependencies
func NewAuthService(
    authRepo repository.AuthCodeRepository,
    jwtService JWTService,
    userServiceClient clients.UserServiceClient,
    publisher messaging.Publisher,
    config *AuthConfig,
) AuthService {
    // ...
}

// ❌ BAD: Hidden dependencies (global variables, singletons)
func NewAuthService(config *AuthConfig) AuthService {
    authRepo := repository.GetGlobalRepository()  // Hidden dependency!
    // ...
}
```

### 4. Error Wrapping

Always wrap errors with context:

```go
if err := s.authRepo.Create(ctx, authCode); err != nil {
    return nil, fmt.Errorf("failed to store auth code: %w", err)
}
```

### 5. Defensive Programming

Check for nil and invalid inputs:

```go
func NewAuthService(..., config *AuthConfig) AuthService {
    if config == nil {
        config = &AuthConfig{}
    }
    // Set defaults...
}

func (r *PostgresAuthCodeRepository) Create(ctx context.Context, authCode *models.AuthCode) error {
    if authCode == nil {
        return errors.New("auth code cannot be nil")
    }
    // ...
}
```

### 6. Separation of Concerns in Tests

Test each layer independently:

```go
// Repository tests: Use sqlmock (no service logic)
func (suite *RepositoryTestSuite) TestCreate_Success() {
    suite.mock.ExpectExec("INSERT INTO").WillReturnResult(...)
    // ...
}

// Service tests: Use mocked repository (no database)
func (suite *ServiceTestSuite) TestRequestAuthCode_Success() {
    suite.mockRepo.On("Create", mock.Anything, mock.Anything).Return(nil)
    // ...
}

// Handler tests: Use mocked service (no business logic)
func (suite *HandlerTestSuite) TestRequestCodeHandler_Success() {
    suite.mockService.On("RequestAuthCode", mock.Anything, mock.Anything).Return(result, nil)
    // ...
}
```

### 7. Configuration Over Hard-coding

Use configuration for values that might change:

```go
// ✅ GOOD: Configurable
type AuthConfig struct {
    CodeExpiryDuration time.Duration
    MaxCodesPerEmail   int
}

// ❌ BAD: Hard-coded
func (s *authService) RequestAuthCode(...) {
    expiresAt := time.Now().Add(5 * time.Minute)  // Hard-coded!
    maxCodes := 3  // Hard-coded!
}
```

### 8. Graceful Degradation

Handle failures gracefully:

```go
func (s *authService) RequestAuthCode(ctx context.Context, email string) (*AuthCodeResult, error) {
    // ... business logic ...

    // Don't fail the main flow if event publishing fails
    go s.publishAuthCodeRequestedEvent(context.Background(), email, code, authCode.ID, authCode.ExpiresAt)

    return result, nil
}

func (s *authService) publishAuthCodeRequestedEvent(...) {
    if err := s.publisher.PublishEvent(ctx, event); err != nil {
        // Log error but don't fail the main flow
        fmt.Printf("Failed to publish event: %v\n", err)
    }
}
```

### 9. Retry Logic for External Services

Implement retry logic for transient failures:

```go
func (p *publisher) PublishWithRetry(ctx context.Context, event Event, maxRetries int) error {
    var lastErr error

    for attempt := 0; attempt <= maxRetries; attempt++ {
        err := p.publishSingle(ctx, event)
        if err == nil {
            return nil
        }

        lastErr = err

        if attempt < maxRetries {
            time.Sleep(p.config.RetryDelay)
            if !p.IsHealthy() {
                p.setupChannel() // Re-establish connection
            }
        }
    }

    return fmt.Errorf("failed after %d attempts: %w", maxRetries+1, lastErr)
}
```

### 10. Health Checks

Implement health checks for external dependencies:

```go
type Publisher interface {
    PublishEvent(ctx context.Context, event Event) error
    IsHealthy() bool  // Health check
    Close() error
}

func (p *publisher) IsHealthy() bool {
    return p.isOpen &&
        p.connection.IsConnected() &&
        p.channel != nil &&
        !p.channel.IsClosed()
}
```

## Summary

By following SOLID principles and these design patterns, we achieve:

✅ **Maintainable Code**: Easy to understand and modify
✅ **Testable Code**: Each component can be tested in isolation
✅ **Flexible Architecture**: Easy to add new features or swap implementations
✅ **Scalable System**: Clean separation allows for independent scaling
✅ **Reliable Services**: Proper error handling and graceful degradation

### Key Takeaways

1. **Depend on abstractions** (interfaces), not concrete implementations
2. **Single responsibility**: Each component does one thing well
3. **Inject dependencies** through constructors
4. **Validate at boundaries** to fail fast
5. **Handle errors explicitly** with proper wrapping
6. **Test each layer independently** with appropriate mocks
7. **Design for failure** with retries and graceful degradation
8. **Make configuration explicit** to support different environments

These principles guide all development in this project and should be followed when adding new features or refactoring existing code.
