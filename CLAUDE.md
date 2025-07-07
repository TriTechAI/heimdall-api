# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Heimdall API is a high-performance blog system backend built with Go-Zero framework, using microservices architecture with unified module management. The project is structured as a monorepo with shared code packages.

## Architecture

### Services Structure
- **admin-api** (Port 8080): Management service for blog administrators, editors, and authors
- **public-api** (Port 8081): Public service for blog visitors, search engines, and third-party applications  
- **common/**: Shared code packages used by both services

### Key Technologies
- **Backend Framework**: Go-Zero
- **Database**: MongoDB
- **Cache**: Redis
- **Authentication**: JWT
- **Testing**: GoConvey + Mockey

## Common Development Commands

### Building and Running
```bash
# Build all services
make build

# Start admin service (port 8080)
make admin

# Start public service (port 8081)
make public

# Run all tests
make test

# Install dependencies
make deps
```

### Code Generation
```bash
# Regenerate API code for both services
make generate

# Generate admin API code only
cd admin-api/admin && goctl api go -api admin.api -dir . --style=gozero

# Generate public API code only
cd public-api/public && goctl api go -api public.api -dir . --style=gozero
```

### Development Tools
```bash
# Format code
make fmt

# Run linter (requires golangci-lint)
make lint

# Generate Swagger documentation
make swagger

# Install development tools
make install-tools
```

## Development Workflow

### API-First Development
1. **Define**: All API definitions must be in `.api` files (admin.api, public.api)
2. **Generate**: Run `make generate` to regenerate handler/types/routes code
3. **Implement**: Write business logic in `internal/logic/` files
4. **Test**: Write tests for logic layer

### File Organization
- **Never modify generated files**: handler/, types/, routes/ are auto-generated
- **Business logic**: Implement in `internal/logic/` directory
- **Data access**: Use `common/dao/` for database operations
- **Models**: Define in `common/model/` with MongoDB integration
- **Shared code**: Place in `common/` directory

## Configuration

### Environment Files
- `admin-api/admin/etc/admin-api.yaml`: Admin service configuration
- `public-api/public/etc/public-api.yaml`: Public service configuration

### Key Configuration Sections
- **MongoDB**: Database connection settings
- **Redis**: Cache configuration for JWT blacklist, sessions, rate limiting
- **JWT**: Authentication token settings
- **CORS**: Cross-origin resource sharing settings
- **Rate Limiting**: API request rate limiting
- **Security**: Password encryption, login security, content filtering

## Database and Models

### MongoDB Collections
- **users**: User accounts and authentication
- **posts**: Blog posts with rich metadata
- **pages**: Static pages with custom templates
- **tags**: Post categorization
- **login_logs**: Security audit logs

### Data Layer Pattern
- **Models** (`common/model/`): Struct definitions with validation methods
- **DAOs** (`common/dao/`): Database access interfaces and implementations
- **Logic** (`internal/logic/`): Business logic that orchestrates DAO calls

## Testing

### Test Structure
- Logic tests: `*_test.go` files alongside logic files
- DAO tests: `common/dao/*_test.go` for database operations
- Model tests: `common/model/*_test.go` for validation and methods

### Running Tests
```bash
# Run all tests with verbose output
go test ./... -v -gcflags="all=-N -l"

# Run specific test file
go test ./admin-api/admin/internal/logic -v

# Run tests in a specific package
go test ./common/dao -v
```

## Code Style and Conventions

### File Naming
- Use lowercase letters only: `userlogic.go`, `postdao.go`
- No underscores or hyphens in filenames
- Test files: `*_test.go`

### Function Guidelines
- Maximum 40 lines per function (ideally 15-25 lines)
- Single responsibility principle
- Clear, descriptive function names
- Avoid deep nesting (max 4 levels)

### Constants Management
- Shared constants: `common/constants/`
- Service-specific constants: `internal/constants/`
- Group by business domain (user, post, tag, etc.)

### Configuration Guidelines
- Never hardcode configuration values in code
- Use YAML configuration files in `etc/` directories
- Avoid naming conflicts with go-zero built-in config fields (`Log`, `Timeout`, `Host`, `Port`)
- Use prefixed names like `LogConfig`, `TimeoutConfig`, or `ApplicationLog`

## Security Considerations

### Authentication
- JWT tokens with configurable expiration
- Refresh token mechanism
- JWT blacklist for logout/revocation
- Rate limiting on authentication endpoints

### User Management
- Role-based access control (Owner, Admin, Editor, Author)
- Account locking after failed login attempts
- Password strength requirements
- Login attempt logging and monitoring

## API Design Guidelines

### RESTful Principles
- **Resource-oriented**: URLs represent resources (`/users`, `/posts`, `/posts/{postId}/comments`)
- **HTTP methods**: Use standard verbs (GET, POST, PUT, PATCH, DELETE)
- **Stateless**: Each request contains all necessary information (JWT tokens)

### URL Structure
- **Plural nouns**: Use `/users`, `/posts` for collections
- **Path variables**: `/users/{userId}` for individual resources
- **Lowercase**: All URL paths in lowercase
- **Version control**: All APIs prefixed with `/api/v1`

### Request/Response Format
- **JSON only**: All requests and responses use JSON format
- **camelCase**: JSON fields use camelCase naming (`userId`, `firstName`, `createdAt`)
- **Content-Type**: Always `application/json`

### Response Codes
- **GET**: `200 OK` with resource or array (empty array `[]` if no results)
- **POST**: `201 Created` with created resource
- **PUT/PATCH**: `200 OK` with updated resource
- **DELETE**: `204 No Content` with empty body

### Error Response Format
```json
{
  "code": "unique_error_code",
  "msg": "Human-readable error message",
  "details": {
    "field": "Specific field that caused error"
  }
}
```

### Common HTTP Status Codes
- `400`: Bad Request (validation errors)
- `401`: Unauthorized (missing/invalid token)
- `403`: Forbidden (insufficient permissions)
- `404`: Not Found (resource doesn't exist)
- `409`: Conflict (resource already exists)
- `500`: Internal Server Error

### Authentication
- **JWT tokens**: Use `Authorization: Bearer <token>` header
- **go-zero annotation**: Use `@server(jwt: Auth)` in .api files

### Pagination
- **Query parameters**: `page` (default 1), `limit` (default 10, max 100)
- **Response format**:
```json
{
  "code": 200,
  "message": "Success",
  "data": {
    "list": [...],
    "pagination": {
      "page": 2,
      "limit": 20,
      "total": 156,
      "totalPages": 8,
      "hasNext": true,
      "hasPrev": true
    }
  },
  "timestamp": "2024-01-01T12:00:00Z"
}
```

## API Documentation

### Swagger Generation
```bash
# Generate all API documentation
make swagger

# View documentation
open docs/swagger/index.html
```

### API Endpoints
- **Admin API**: `/api/v1/admin/*` - Management operations
- **Public API**: `/api/v1/public/*` - Public content access

## Common Patterns

### Error Handling
- Use predefined errors from `common/errors/`
- Return errors from logic layer to handler
- Global error handling via middleware

### Response Format
- Consistent JSON response structure
- Pagination for list endpoints
- Standard error response format

### Caching Strategy
- Redis caching for frequently accessed data
- JWT blacklist caching
- Rate limiting state caching
- Session management

## Development Guidelines

### Before Making Changes
1. Read existing code patterns in similar files
2. Follow Go-Zero conventions and project structure
3. Implement business logic in logic layer, not handlers
4. Use existing DAOs and models where possible
5. Add appropriate tests for new functionality

### Code Review Checklist
- API definitions updated in `.api` files
- Business logic in appropriate layer
- Error handling implemented
- Tests written and passing
- Documentation updated if needed
- No hardcoded values (use configuration)

### Performance Considerations
- Use MongoDB indexes for query optimization
- Implement pagination for list endpoints
- Cache frequently accessed data
- Use Redis for session and rate limiting
- Monitor database query performance

## Testing Framework and Best Practices

### Testing Tools
- **GoConvey** (`github.com/smartystreets/goconvey/convey`): BDD-style testing framework
- **Mockey** (`github.com/bytedance/mockey`): Runtime patching for mocking without interfaces
- **MongoDB Testing** (`go.mongodb.org/mongo-driver/mongo/integration/mtest`): For database integration tests
- **Redis Testing** (`github.com/alicebob/miniredis/v2`): In-memory Redis for cache testing

### TDD Approach
- Follow Red-Green-Refactor cycle
- Write tests first, then implement logic
- Use mocking to isolate units under test
- Separate unit tests from integration tests

### Test Organization
- Logic tests: `*_test.go` files alongside logic files  
- DAO tests: `common/dao/*_test.go` for database operations
- Model tests: `common/model/*_test.go` for validation methods
- Service isolation: Each service's tests must run independently

### Mocking Strategy
- Use `mockey` for runtime patching - no interfaces required
- Mock DAO methods in logic layer tests
- Use in-memory databases for integration tests
- Ensure proper cleanup with `defer Unpatch()`