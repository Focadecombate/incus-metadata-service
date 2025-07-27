# Vendor Data Tests

This directory contains comprehensive tests for the vendor data API endpoints.

## Test Files

### `vendordata_test.go`

Unit tests for the individual handler methods:

- `UpdateVendorData` - Tests for updating existing vendor data
- `CreateVendorData` - Tests for creating new vendor data
- `GetVendorData` - Tests for retrieving vendor data

Each method includes tests for:

- Successful operations
- Validation errors (missing parameters, invalid JSON)
- Database errors
- Edge cases (nil data, empty strings)

### `vendordata_integration_test.go`

Integration tests that test the handlers with actual router setup:

- Complete workflow tests (Create → Get → Update)
- Edge case validation tests
- Error handling tests
- Large payload tests

## Test Coverage

The tests cover the following scenarios:

### Success Cases

- Creating vendor data with valid payload
- Retrieving existing vendor data
- Updating existing vendor data
- Creating vendor data with nil/empty data (valid)
- Complex nested JSON data structures

### Error Cases

- Missing required parameters (vendor_name)
- Invalid JSON payloads
- Database connection errors
- Non-existent vendor data
- Data parsing errors
- Vendor already exists (for create operations)

### Edge Cases

- Large JSON payloads (1000+ keys)
- Special characters in vendor names
- Empty vendor names
- Nil data handling
- Complex nested data structures

## Running Tests

```bash
# Run all tests in the package
go test ./metadata-service/internal/api/internal/ -v

# Run tests with coverage
go test ./metadata-service/internal/api/internal/ -v -cover

# Run specific test
go test ./metadata-service/internal/api/internal/ -v -run TestUpdateVendorData_Success
```

## Test Architecture

### Mock Structure

- `MockQuerier` - Implements the `db.Querier` interface for database operations
- `Handler` - Uses the original production handlers with dependency injection via the `db.Querier` interface

### Helper Functions

- `setupTestHandler()` - Creates a test handler with mock database using the original Handler struct
- `setupTestContext()` - Creates a Gin test context with request/response
- `setupRouter()` - Creates a test router with actual route definitions

### Dependency Injection

The tests use dependency injection through the `db.Querier` interface, allowing us to:

- Test the actual production handler code (88.7% code coverage)
- Mock database operations cleanly
- Maintain separation between business logic and data access

## Test Philosophy

These tests follow the AAA pattern (Arrange, Act, Assert):

1. **Arrange** - Set up test data, mocks, and expectations
2. **Act** - Execute the function being tested
3. **Assert** - Verify the results and mock expectations

The tests use testify/mock for mocking database operations, ensuring that we test the handler logic in isolation while verifying that the correct database operations are called with the expected parameters. By using the original handlers with dependency injection, we achieve high code coverage and confidence that our tests reflect the actual production behavior.
