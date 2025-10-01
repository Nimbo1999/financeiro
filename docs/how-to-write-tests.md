# How to Write Tests in Go - TDD Approach

This guide outlines the testing patterns and practices used in this project. We follow Test-Driven Development (TDD) principles and use industry-standard testing libraries to ensure robust, maintainable test coverage.

## Table of Contents

1. [Testing Libraries](#testing-libraries)
2. [Test File Structure](#test-file-structure)
3. [Repository Layer Testing](#repository-layer-testing)
4. [Service Layer Testing](#service-layer-testing)
5. [Handler Layer Testing](#handler-layer-testing)
6. [Event/Messaging Testing](#eventmessaging-testing)
7. [TDD Workflow](#tdd-workflow)
8. [Best Practices](#best-practices)
9. [Common Patterns](#common-patterns)

## Testing Libraries

We use the following libraries for testing:

- **`github.com/stretchr/testify`**: Assertion library and test suites
  - `assert`: For making test assertions
  - `mock`: For creating mock objects
  - `suite`: For organizing tests into test suites

- **`github.com/DATA-DOG/go-sqlmock`**: For mocking database operations in repository tests

- **`net/http/httptest`**: For testing HTTP handlers

## Test File Structure

### General Rules

1. **File Naming**: Test files should be named `<filename>_test.go` alongside the file being tested
2. **Package**: Use the same package name as the code being tested
3. **Test Suite Pattern**: Use testify's suite pattern for organizing related tests

### Basic Structure

```go
package yourpackage

import (
    "testing"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/suite"
)

// Test Suite Definition
type YourTestSuite struct {
    suite.Suite
    // Add fields for dependencies, mocks, etc.
}

// Setup method - runs before each test
func (suite *YourTestSuite) SetupTest() {
    // Initialize mocks and dependencies
}

// Teardown method - runs after each test
func (suite *YourTestSuite) TearDownTest() {
    // Assert all mock expectations were met
    // Clean up resources
}

// Test suite runner
func TestYourTestSuite(t *testing.T) {
    suite.Run(t, new(YourTestSuite))
}

// Individual test methods
func (suite *YourTestSuite) TestMethodName_Scenario() {
    // Arrange
    // Act
    // Assert
}
```

## Repository Layer Testing

Repository tests use `go-sqlmock` to mock database interactions without requiring a real database.

### Example: Testing a Repository Method

```go
type PostgresYourRepositoryTestSuite struct {
    suite.Suite
    db   *gorm.DB
    mock sqlmock.Sqlmock
    repo YourRepository
}

func (suite *PostgresYourRepositoryTestSuite) SetupTest() {
    var err error
    var sqlDB *sql.DB

    // Create mock database
    sqlDB, suite.mock, err = sqlmock.New()
    suite.Require().NoError(err)

    // Initialize GORM with mock
    suite.db, err = gorm.Open(postgres.New(postgres.Config{
        Conn: sqlDB,
    }), &gorm.Config{})
    suite.Require().NoError(err)

    // Create repository instance
    suite.repo = NewPostgresYourRepository(suite.db)
}

func (suite *PostgresYourRepositoryTestSuite) TearDownTest() {
    suite.mock.ExpectationsWereMet()
}
```

### Testing CRUD Operations

#### Create Operation

```go
func (suite *PostgresYourRepositoryTestSuite) TestCreate_Success() {
    entity := &models.YourEntity{
        Field1: "value1",
        Field2: "value2",
        CreatedAt: time.Now(),
    }

    // Expect transaction begin
    suite.mock.ExpectBegin()

    // Expect INSERT query
    suite.mock.ExpectExec(`INSERT INTO "your_table"`).
        WithArgs(entity.Field1, entity.Field2, sqlmock.AnyArg()).
        WillReturnResult(sqlmock.NewResult(1, 1))

    // Expect transaction commit
    suite.mock.ExpectCommit()

    err := suite.repo.Create(context.Background(), entity)

    assert.NoError(suite.T(), err)
}

func (suite *PostgresYourRepositoryTestSuite) TestCreate_DatabaseError() {
    entity := &models.YourEntity{
        Field1: "value1",
        Field2: "value2",
    }

    suite.mock.ExpectBegin()
    suite.mock.ExpectExec(`INSERT INTO "your_table"`).
        WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg()).
        WillReturnError(errors.New("database connection failed"))
    suite.mock.ExpectRollback()

    err := suite.repo.Create(context.Background(), entity)

    assert.Error(suite.T(), err)
    assert.Contains(suite.T(), err.Error(), "database connection failed")
}
```

#### Read Operation

```go
func (suite *PostgresYourRepositoryTestSuite) TestFindByID_Success() {
    id := "test-id"
    expectedEntity := models.YourEntity{
        ID:     id,
        Field1: "value1",
        Field2: "value2",
    }

    // Create rows mock
    rows := sqlmock.NewRows([]string{"id", "field1", "field2"}).
        AddRow(expectedEntity.ID, expectedEntity.Field1, expectedEntity.Field2)

    suite.mock.ExpectQuery(`SELECT \* FROM "your_table" WHERE id = \$1`).
        WithArgs(id).
        WillReturnRows(rows)

    result, err := suite.repo.FindByID(context.Background(), id)

    assert.NoError(suite.T(), err)
    assert.NotNil(suite.T(), result)
    assert.Equal(suite.T(), expectedEntity.ID, result.ID)
    assert.Equal(suite.T(), expectedEntity.Field1, result.Field1)
}

func (suite *PostgresYourRepositoryTestSuite) TestFindByID_NotFound() {
    id := "non-existent-id"

    suite.mock.ExpectQuery(`SELECT \* FROM "your_table" WHERE id = \$1`).
        WithArgs(id).
        WillReturnError(gorm.ErrRecordNotFound)

    result, err := suite.repo.FindByID(context.Background(), id)

    assert.Error(suite.T(), err)
    assert.Nil(suite.T(), result)
    assert.Equal(suite.T(), ErrEntityNotFound, err)
}
```

#### Update Operation

```go
func (suite *PostgresYourRepositoryTestSuite) TestUpdate_Success() {
    id := "test-id"

    suite.mock.ExpectBegin()
    suite.mock.ExpectExec(`UPDATE "your_table" SET "field1"=\$1 WHERE id = \$2`).
        WithArgs(sqlmock.AnyArg(), id).
        WillReturnResult(sqlmock.NewResult(0, 1))
    suite.mock.ExpectCommit()

    err := suite.repo.Update(context.Background(), id, "new-value")

    assert.NoError(suite.T(), err)
}
```

#### Delete Operation

```go
func (suite *PostgresYourRepositoryTestSuite) TestDelete_Success() {
    id := "test-id"

    suite.mock.ExpectBegin()
    suite.mock.ExpectExec(`DELETE FROM "your_table" WHERE id = \$1`).
        WithArgs(id).
        WillReturnResult(sqlmock.NewResult(0, 1))
    suite.mock.ExpectCommit()

    err := suite.repo.Delete(context.Background(), id)

    assert.NoError(suite.T(), err)
}
```

### Validation Tests

Always test input validation:

```go
func (suite *PostgresYourRepositoryTestSuite) TestCreate_NilEntity() {
    err := suite.repo.Create(context.Background(), nil)

    assert.Error(suite.T(), err)
    assert.Contains(suite.T(), err.Error(), "cannot be nil")
}

func (suite *PostgresYourRepositoryTestSuite) TestFindByID_EmptyID() {
    result, err := suite.repo.FindByID(context.Background(), "")

    assert.Error(suite.T(), err)
    assert.Nil(suite.T(), result)
    assert.Equal(suite.T(), ErrIDEmpty, err)
}
```

## Service Layer Testing

Service layer tests use mocks for all dependencies (repositories, other services, clients).

### Creating Mocks

```go
type MockYourRepository struct {
    mock.Mock
}

func (m *MockYourRepository) Create(ctx context.Context, entity *models.YourEntity) error {
    args := m.Called(ctx, entity)
    return args.Error(0)
}

func (m *MockYourRepository) FindByID(ctx context.Context, id string) (*models.YourEntity, error) {
    args := m.Called(ctx, id)
    if args.Get(0) == nil {
        return nil, args.Error(1)
    }
    return args.Get(0).(*models.YourEntity), args.Error(1)
}

type MockExternalClient struct {
    mock.Mock
}

func (m *MockExternalClient) GetData(ctx context.Context, id string) (*Data, bool, error) {
    args := m.Called(ctx, id)
    if args.Get(0) == nil {
        return nil, args.Bool(1), args.Error(2)
    }
    return args.Get(0).(*Data), args.Bool(1), args.Error(2)
}
```

### Test Suite Setup

```go
type YourServiceTestSuite struct {
    suite.Suite
    service      YourService
    mockRepo     *MockYourRepository
    mockClient   *MockExternalClient
    config       *ServiceConfig
}

func (suite *YourServiceTestSuite) SetupTest() {
    suite.mockRepo = new(MockYourRepository)
    suite.mockClient = new(MockExternalClient)
    suite.config = &ServiceConfig{
        Timeout: 5 * time.Second,
    }
    suite.service = NewYourService(suite.mockRepo, suite.mockClient, suite.config)
}

func (suite *YourServiceTestSuite) TearDownTest() {
    suite.mockRepo.AssertExpectations(suite.T())
    suite.mockClient.AssertExpectations(suite.T())
}
```

### Testing Success Scenarios

```go
func (suite *YourServiceTestSuite) TestCreateEntity_Success() {
    input := &CreateEntityInput{
        Field1: "value1",
        Field2: "value2",
    }

    // Mock external client call
    externalData := &Data{ID: "external-123"}
    suite.mockClient.On("GetData", mock.Anything, "key").
        Return(externalData, true, nil)

    // Mock repository call
    suite.mockRepo.On("Create", mock.Anything, mock.MatchedBy(func(entity *models.YourEntity) bool {
        return entity.Field1 == input.Field1 &&
            entity.Field2 == input.Field2
    })).Return(nil)

    result, err := suite.service.CreateEntity(context.Background(), input)

    assert.NoError(suite.T(), err)
    assert.NotNil(suite.T(), result)
    assert.True(suite.T(), result.Success)
}
```

### Testing Error Scenarios

```go
func (suite *YourServiceTestSuite) TestCreateEntity_ValidationError() {
    input := &CreateEntityInput{
        Field1: "", // Invalid: empty field
        Field2: "value2",
    }

    result, err := suite.service.CreateEntity(context.Background(), input)

    assert.Error(suite.T(), err)
    assert.Nil(suite.T(), result)
    assert.Equal(suite.T(), ErrInvalidInput, err)
}

func (suite *YourServiceTestSuite) TestCreateEntity_ExternalServiceError() {
    input := &CreateEntityInput{
        Field1: "value1",
        Field2: "value2",
    }

    // Mock external service failure
    suite.mockClient.On("GetData", mock.Anything, "key").
        Return(nil, false, errors.New("service unavailable"))

    result, err := suite.service.CreateEntity(context.Background(), input)

    assert.Error(suite.T(), err)
    assert.Nil(suite.T(), result)
    assert.Equal(suite.T(), ErrExternalServiceUnavailable, err)
}

func (suite *YourServiceTestSuite) TestCreateEntity_RepositoryError() {
    input := &CreateEntityInput{
        Field1: "value1",
        Field2: "value2",
    }

    externalData := &Data{ID: "external-123"}
    suite.mockClient.On("GetData", mock.Anything, "key").
        Return(externalData, true, nil)

    suite.mockRepo.On("Create", mock.Anything, mock.Anything).
        Return(errors.New("database error"))

    result, err := suite.service.CreateEntity(context.Background(), input)

    assert.Error(suite.T(), err)
    assert.Nil(suite.T(), result)
    assert.Contains(suite.T(), err.Error(), "database error")
}
```

### Testing Async Operations

For services that publish events asynchronously (using goroutines):

```go
func (suite *YourServiceTestSuite) TestCreateEntity_EventPublished() {
    input := &CreateEntityInput{
        Field1: "value1",
        Field2: "value2",
    }

    suite.mockClient.On("GetData", mock.Anything, "key").
        Return(&Data{ID: "123"}, true, nil)
    suite.mockRepo.On("Create", mock.Anything, mock.Anything).Return(nil)

    // Use .Maybe() for async calls since they might not be verified immediately
    suite.mockPublisher.On("PublishEvent", mock.Anything,
        mock.AnythingOfType("*messaging.EntityCreatedEvent")).
        Return(nil).Maybe()

    result, err := suite.service.CreateEntity(context.Background(), input)

    assert.NoError(suite.T(), err)
    assert.NotNil(suite.T(), result)
    // Note: Don't assert on publisher expectations in TearDownTest for async calls
}
```

## Handler Layer Testing

HTTP handler tests use `httptest` to simulate HTTP requests and responses.

### Mock Service Setup

```go
type MockYourService struct {
    mock.Mock
}

func (m *MockYourService) DoSomething(ctx context.Context, input *Input) (*Result, error) {
    args := m.Called(ctx, input)
    if args.Get(0) == nil {
        return nil, args.Error(1)
    }
    return args.Get(0).(*Result), args.Error(1)
}
```

### Test Suite Setup

```go
type YourHandlerTestSuite struct {
    suite.Suite
    handler     *YourHandler
    mockService *MockYourService
}

func (suite *YourHandlerTestSuite) SetupTest() {
    suite.mockService = new(MockYourService)
    suite.handler = NewYourHandler(suite.mockService)
}

func (suite *YourHandlerTestSuite) TearDownTest() {
    suite.mockService.AssertExpectations(suite.T())
}
```

### Testing Success Responses

```go
func (suite *YourHandlerTestSuite) TestYourHandler_Success() {
    expectedResult := &Result{
        ID:      "123",
        Success: true,
    }

    suite.mockService.On("DoSomething", mock.Anything, mock.Anything).
        Return(expectedResult, nil)

    // Create request
    reqBody := YourRequest{Field: "value"}
    body, _ := json.Marshal(reqBody)
    req := httptest.NewRequest(http.MethodPost, "/your-endpoint", bytes.NewBuffer(body))
    req.Header.Set("Content-Type", "application/json")

    // Create response recorder
    w := httptest.NewRecorder()

    // Call handler
    suite.handler.YourHandler(w, req)

    // Assert response
    assert.Equal(suite.T(), http.StatusOK, w.Code)

    var response APIResponse
    err := json.Unmarshal(w.Body.Bytes(), &response)
    assert.NoError(suite.T(), err)
    assert.True(suite.T(), response.Success)
    assert.NotNil(suite.T(), response.Data)
}
```

### Testing Validation Errors

```go
func (suite *YourHandlerTestSuite) TestYourHandler_InvalidInput() {
    // Create request with invalid data
    reqBody := YourRequest{Field: ""} // Invalid: empty field
    body, _ := json.Marshal(reqBody)
    req := httptest.NewRequest(http.MethodPost, "/your-endpoint", bytes.NewBuffer(body))
    req.Header.Set("Content-Type", "application/json")

    w := httptest.NewRecorder()

    suite.handler.YourHandler(w, req)

    assert.Equal(suite.T(), http.StatusBadRequest, w.Code)

    var response APIResponse
    err := json.Unmarshal(w.Body.Bytes(), &response)
    assert.NoError(suite.T(), err)
    assert.False(suite.T(), response.Success)
    assert.NotNil(suite.T(), response.Error)
}

func (suite *YourHandlerTestSuite) TestYourHandler_InvalidJSON() {
    req := httptest.NewRequest(http.MethodPost, "/your-endpoint",
        bytes.NewBuffer([]byte("invalid json")))
    req.Header.Set("Content-Type", "application/json")

    w := httptest.NewRecorder()
    suite.handler.YourHandler(w, req)

    assert.Equal(suite.T(), http.StatusBadRequest, w.Code)

    var response APIResponse
    err := json.Unmarshal(w.Body.Bytes(), &response)
    assert.NoError(suite.T(), err)
    assert.False(suite.T(), response.Success)
}
```

### Testing Service Errors

```go
func (suite *YourHandlerTestSuite) TestYourHandler_ServiceError() {
    suite.mockService.On("DoSomething", mock.Anything, mock.Anything).
        Return(nil, ErrRateLimitExceeded)

    reqBody := YourRequest{Field: "value"}
    body, _ := json.Marshal(reqBody)
    req := httptest.NewRequest(http.MethodPost, "/your-endpoint", bytes.NewBuffer(body))
    req.Header.Set("Content-Type", "application/json")

    w := httptest.NewRecorder()

    suite.handler.YourHandler(w, req)

    assert.Equal(suite.T(), http.StatusTooManyRequests, w.Code)

    var response APIResponse
    err := json.Unmarshal(w.Body.Bytes(), &response)
    assert.NoError(suite.T(), err)
    assert.False(suite.T(), response.Success)
    assert.Equal(suite.T(), "RATE_LIMIT_EXCEEDED", response.Error.Code)
}
```

## Event/Messaging Testing

Testing event creation and serialization.

### Testing Event Creation

```go
func (suite *EventsTestSuite) TestNewYourEvent() {
    field1 := "value1"
    field2 := "value2"

    event := NewYourEvent(field1, field2)

    assert.NotNil(suite.T(), event)
    assert.Equal(suite.T(), EventTypeYourEvent, event.Type)
    assert.Equal(suite.T(), "your-service", event.Source)
    assert.Equal(suite.T(), "1.0", event.Version)
    assert.NotEmpty(suite.T(), event.ID)
    assert.WithinDuration(suite.T(), time.Now(), event.Timestamp, time.Second)

    assert.Equal(suite.T(), field1, event.Data.Field1)
    assert.Equal(suite.T(), field2, event.Data.Field2)
}
```

### Testing Event Interface Implementation

```go
func (suite *EventsTestSuite) TestYourEvent_EventInterface() {
    event := NewYourEvent("value1", "value2")

    assert.Equal(suite.T(), EventTypeYourEvent, event.GetType())
    assert.NotEmpty(suite.T(), event.GetID())
    assert.WithinDuration(suite.T(), time.Now(), event.GetTimestamp(), time.Second)
    assert.Equal(suite.T(), string(EventTypeYourEvent), event.GetRoutingKey())

    jsonData, err := event.ToJSON()
    assert.NoError(suite.T(), err)
    assert.NotEmpty(suite.T(), jsonData)

    // Verify JSON can be unmarshaled
    var unmarshaled YourEvent
    err = json.Unmarshal(jsonData, &unmarshaled)
    assert.NoError(suite.T(), err)
    assert.Equal(suite.T(), event.ID, unmarshaled.ID)
    assert.Equal(suite.T(), event.Type, unmarshaled.Type)
}
```

### Testing Event Serialization

```go
func TestEventSerialization(t *testing.T) {
    testCases := []struct {
        name  string
        event Event
    }{
        {"YourEvent", NewYourEvent("value1", "value2")},
        {"AnotherEvent", NewAnotherEvent("data")},
    }

    for _, tc := range testCases {
        t.Run(tc.name, func(t *testing.T) {
            jsonData, err := tc.event.ToJSON()
            assert.NoError(t, err)
            assert.NotEmpty(t, jsonData)

            // Verify it's valid JSON
            var result map[string]interface{}
            err = json.Unmarshal(jsonData, &result)
            assert.NoError(t, err)

            // Check common fields
            assert.Equal(t, tc.event.GetID(), result["id"])
            assert.Equal(t, string(tc.event.GetType()), result["type"])
            assert.Equal(t, "your-service", result["source"])
            assert.Equal(t, "1.0", result["version"])
        })
    }
}
```

## TDD Workflow

Follow these steps when implementing new features using TDD:

### 1. Write the Test First (Red)

```go
func (suite *YourServiceTestSuite) TestNewFeature_Success() {
    input := &NewFeatureInput{
        Field: "value",
    }

    suite.mockRepo.On("DoSomething", mock.Anything, mock.Anything).
        Return(&models.Entity{ID: "123"}, nil)

    result, err := suite.service.NewFeature(context.Background(), input)

    assert.NoError(suite.T(), err)
    assert.NotNil(suite.T(), result)
    assert.Equal(suite.T(), "123", result.ID)
}
```

### 2. Run the Test (Fails)

```bash
go test ./... -v
# Test should fail because NewFeature doesn't exist yet
```

### 3. Write Minimal Code to Pass (Green)

```go
func (s *yourService) NewFeature(ctx context.Context, input *NewFeatureInput) (*Result, error) {
    entity, err := s.repo.DoSomething(ctx, input)
    if err != nil {
        return nil, err
    }
    return &Result{ID: entity.ID}, nil
}
```

### 4. Run the Test Again (Passes)

```bash
go test ./... -v
# Test should now pass
```

### 5. Refactor

Clean up the code while ensuring tests still pass.

### 6. Add More Test Cases

Add edge cases, error scenarios, etc:

```go
func (suite *YourServiceTestSuite) TestNewFeature_InvalidInput() {
    // Test validation
}

func (suite *YourServiceTestSuite) TestNewFeature_RepositoryError() {
    // Test error handling
}
```

## Best Practices

### 1. Test Naming Conventions

```go
// Format: Test<MethodName>_<Scenario>
func (suite *YourTestSuite) TestCreate_Success() {}
func (suite *YourTestSuite) TestCreate_ValidationError() {}
func (suite *YourTestSuite) TestCreate_DatabaseError() {}
func (suite *YourTestSuite) TestFindByID_NotFound() {}
```

### 2. Arrange-Act-Assert Pattern

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

### 3. Test Coverage

Ensure you test:
- ✅ Success scenarios
- ✅ Validation errors (nil inputs, empty strings, invalid formats)
- ✅ Not found scenarios
- ✅ Database/repository errors
- ✅ External service errors
- ✅ Context cancellation
- ✅ Edge cases

### 4. Mock Expectations

```go
// Use specific matchers when possible
suite.mock.On("Method", ctx, "specific-id").Return(result, nil)

// Use mock.Anything for flexible matching
suite.mock.On("Method", mock.Anything, mock.Anything).Return(result, nil)

// Use custom matchers for complex validation
suite.mock.On("Create", mock.Anything, mock.MatchedBy(func(entity *Entity) bool {
    return entity.Field1 == "expected" && entity.Field2 != ""
})).Return(nil)

// Use Maybe() for async operations
suite.mock.On("PublishEvent", mock.Anything, mock.Anything).Return(nil).Maybe()
```

### 5. Context Testing

```go
func (suite *YourTestSuite) TestContextCancellation() {
    ctx, cancel := context.WithCancel(context.Background())
    cancel() // Cancel immediately

    err := suite.service.Method(ctx, input)

    // Verify context cancellation is handled
    suite.ErrorAs(err, &context.Canceled)
}
```

### 6. Table-Driven Tests

For testing multiple scenarios with similar logic:

```go
func (suite *YourTestSuite) TestValidation_TableDriven() {
    testCases := []struct {
        name        string
        input       *Input
        setupMock   func()
        expectError bool
        errorMsg    string
    }{
        {
            name:        "Valid Input",
            input:       &Input{Field: "valid"},
            setupMock:   func() { /* setup mocks */ },
            expectError: false,
        },
        {
            name:        "Invalid Input",
            input:       &Input{Field: ""},
            setupMock:   func() {},
            expectError: true,
            errorMsg:    "field cannot be empty",
        },
    }

    for _, tc := range testCases {
        suite.Run(tc.name, func() {
            tc.setupMock()

            result, err := suite.service.Method(context.Background(), tc.input)

            if tc.expectError {
                assert.Error(suite.T(), err)
                if tc.errorMsg != "" {
                    assert.Contains(suite.T(), err.Error(), tc.errorMsg)
                }
            } else {
                assert.NoError(suite.T(), err)
                assert.NotNil(suite.T(), result)
            }
        })
    }
}
```

### 7. Benchmark Tests

Include benchmarks for performance-critical code:

```go
func BenchmarkYourMethod(b *testing.B) {
    // Setup
    service := setupService()

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        _, _ = service.Method(context.Background(), input)
    }
}
```

## Common Patterns

### Testing Time-Dependent Code

```go
func (suite *YourTestSuite) TestExpiration() {
    entity := &models.Entity{
        ExpiresAt: time.Now().Add(5 * time.Minute),
    }

    // Test within time window
    assert.True(suite.T(), entity.ExpiresAt.After(time.Now()))

    // Test with tolerance
    assert.WithinDuration(suite.T(), time.Now(), entity.CreatedAt, time.Second)
}
```

### Testing UUID/Random Generation

```go
func (suite *YourTestSuite) TestGeneration() {
    result1 := generateID()
    time.Sleep(1 * time.Millisecond)
    result2 := generateID()

    assert.NotEmpty(suite.T(), result1)
    assert.NotEmpty(suite.T(), result2)
    assert.NotEqual(suite.T(), result1, result2)
}
```

### Testing Error Types

```go
func (suite *YourTestSuite) TestErrorHandling() {
    _, err := suite.service.Method(ctx, invalidInput)

    assert.Error(suite.T(), err)
    assert.Equal(suite.T(), ErrSpecificError, err)
    // or
    assert.ErrorIs(suite.T(), err, ErrSpecificError)
    // or
    assert.Contains(suite.T(), err.Error(), "expected message")
}
```

### Testing SQL Queries

Use regex patterns for SQL query matching:

```go
suite.mock.ExpectQuery(`SELECT \* FROM "table" WHERE id = \$1`).
    WithArgs(id).
    WillReturnRows(rows)

suite.mock.ExpectExec(`UPDATE "table" SET "field"=\$1 WHERE id = \$2`).
    WithArgs(value, id).
    WillReturnResult(sqlmock.NewResult(0, 1))
```

## Running Tests

```bash
# Run all tests
go test ./...

# Run tests with verbose output
go test ./... -v

# Run tests with coverage
go test ./... -cover

# Run tests with coverage report
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out

# Run specific test suite
go test ./internal/services -v

# Run specific test
go test ./internal/services -v -run TestAuthServiceTestSuite/TestRequestAuthCode_Success

# Run benchmarks
go test ./... -bench=.
```

## Checklist for New Tests

When writing tests for a new feature, ensure:

- [ ] Test file created with `_test.go` suffix
- [ ] Test suite struct defined with all necessary dependencies
- [ ] SetupTest initializes all mocks and dependencies
- [ ] TearDownTest asserts mock expectations
- [ ] Success scenario tested
- [ ] All validation errors tested
- [ ] Error scenarios tested (not found, database errors, etc.)
- [ ] Edge cases covered
- [ ] Context cancellation tested (if applicable)
- [ ] All mocks properly configured
- [ ] Assertions are clear and specific
- [ ] Test names follow naming convention
- [ ] Code follows Arrange-Act-Assert pattern
- [ ] Tests are independent and can run in any order
- [ ] No hard-coded values that might cause flaky tests
