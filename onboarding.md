# Backend Onboarding Guide

Welcome to the KTH AI Society backend project! This guide will help you understand the project structure, tech stack, and how to get started.

## 📋 Table of Contents

- [Tech Stack](#tech-stack)
- [Project Structure](#project-structure)
- [Getting Started](#getting-started)
- [Architecture Overview](#architecture-overview)
- [Key Features](#key-features)
- [Development Guidelines](#development-guidelines)

## 🛠 Tech Stack

### Core Technologies

- **Go 1.23.6** - Primary programming language
- **Gin Framework** - Web framework for routing and middleware
- **GORM** - ORM for database operations
- **PostgreSQL 17** - Primary database (postgres:17-alpine)
- **Redis** - Caching and session storage
- **Docker** - Containerization and local development

### Key Dependencies

#### Web Framework & Routing

- `github.com/gin-gonic/gin` - HTTP web framework
- `github.com/gin-contrib/cors` - CORS middleware
- `github.com/gin-contrib/sessions` - Session management
- `github.com/gin-contrib/sessions/cookie` - Cookie-based sessions

#### Database & ORM

- `gorm.io/gorm` - ORM library
- `gorm.io/driver/postgres` - PostgreSQL driver
- `github.com/redis/go-redis/v9` - Redis client

#### Authentication

- `github.com/markbates/goth` - OAuth provider integration
- Google OAuth2 authentication

#### Utilities

- `github.com/joho/godotenv` - Environment variable management
- `github.com/google/uuid` - UUID generation

## 📁 Project Structure

```
backend/
├── cmd/
│   └── api/
│       └── main.go              # Application entry point
├── internal/
│   ├── config/
│   │   └── config.go           # Configuration management
│   ├── database/
│   │   └── redis.go            # Redis client setup
│   ├── handlers/               # HTTP request handlers
│   │   ├── interfaces.go       # Handler interface
│   │   ├── auth_handler.go     # Authentication endpoints
│   │   ├── event_handler.go    # Event management
│   │   ├── profile_handler.go  # User profiles
│   │   ├── registration_handler.go
│   │   ├── alumni_handler.go
│   │   ├── startup_handler.go
│   │   ├── speaker_handler.go
│   │   └── sponsor_handler.go
│   ├── middleware/             # HTTP middleware
│   │   ├── auth.go            # Authentication middleware
│   │   ├── admin.go           # Admin authorization
│   │   └── rate_limit.go      # Rate limiting
│   └── models/                # Data models
│       ├── user.go
│       ├── profile.go
│       ├── event.go
│       ├── registration.go
│       ├── team_member.go
│       ├── alumni.go
│       ├── startup.go
│       ├── speaker.go
│       └── sponsor.go
├── docs/                       # API documentation
├── docker-compose.yml          # Local development services
├── Dockerfile                  # Production container build
├── go.mod                      # Go module dependencies
├── go.sum                      # Dependency checksums
└── README.md                   # Project documentation
```

### Directory Explanations

#### `cmd/api/`

Contains the application entry point (`main.go`). This is where:

- Database connections are established
- Middleware is configured
- Routes are registered
- The server is started

#### `internal/`

Houses all internal application code that is not meant to be imported by other projects.

- **`config/`** - Configuration loading and management

  - Environment variables
  - Database configuration
  - OAuth settings
  - CORS settings

- **`handlers/`** - HTTP request handlers organized by domain

  - Each handler implements the `Handler` interface
  - Responsible for request validation and response formatting
  - Business logic for their respective domains

- **`middleware/`** - HTTP middleware functions

  - `auth.go` - Session-based authentication
  - `admin.go` - Admin role verification
  - `rate_limit.go` - Request rate limiting

- **`models/`** - Database models using GORM

  - Struct definitions with JSON and GORM tags
  - Relationships between entities
  - Database constraints

- **`database/`** - Database client implementations
  - Redis client setup and configuration

## 🚀 Getting Started

### Prerequisites

- Go 1.23.6 or higher
- Docker and Docker Compose
- Git

### Setup Steps

1. **Clone the repository**

   ```bash
   git clone https://github.com/kthaisociety/backend.git
   cd backend
   ```

2. **Set up environment variables**
   Create a `.env` file in the project root:

   ```env
   # Database
   DB_HOST=localhost
   DB_PORT=5432
   DB_USER=postgres
   DB_PASSWORD=password
   DB_NAME=kthais
   DB_SSLMODE=disable

   # Server
   SERVER_PORT=8080

   # Redis
   REDIS_HOST=localhost
   REDIS_PORT=6379
   REDIS_PASSWORD=

   # OAuth
   GOOGLE_CLIENT_ID=your_google_client_id
   GOOGLE_CLIENT_SECRET=your_google_client_secret

   # Security
   SESSION_KEY=base64_encoded_32_byte_key

   # CORS
   ALLOWED_ORIGINS=http://localhost:3000,http://localhost:3001

   # Backend URL
   BACKEND_URL=http://localhost:8080

   # Development Mode
   DEVELOPMENT=true
   ```

3. **Start the database services**

   ```bash
   docker compose up -d
   ```

   This starts:

   - PostgreSQL on port 5432
   - Redis on port 6379

4. **Install dependencies**

   ```bash
   go mod download
   ```

5. **Run the application**

   ```bash
   go run cmd/api/main.go
   ```

6. **Verify it's running**
   ```bash
   curl http://localhost:8080/api/v1/health
   ```
   Should return: `{"status":"ok"}`

## 🏗 Architecture Overview

### Handler Pattern

The project uses a handler-based architecture where each domain has its own handler:

```go
type Handler interface {
    Register(router *gin.RouterGroup)
}
```

Each handler:

1. Implements the `Handler` interface
2. Encapsulates domain-specific logic
3. Registers its own routes
4. Has access to the database via dependency injection

### Authentication Flow

1. User initiates OAuth flow via `/api/v1/auth/google`
2. Google redirects back to `/api/v1/auth/google/callback`
3. Session is created with user information
4. Protected routes use `AuthRequired()` middleware
5. Admin routes use both `AuthRequired()` and `AdminRequired()` middleware

### Database Migrations

GORM auto-migration runs on startup:

```go
db.AutoMigrate(
    &models.User{},
    &models.Profile{},
    &models.Event{},
    // ... other models
)
```

### CORS Configuration

CORS is configured to allow requests from specified origins:

- Allowed origins loaded from `ALLOWED_ORIGINS` env var
- Credentials support enabled for cookies/sessions
- Pre-flight requests handled automatically

## 🔑 Key Features

### 1. Authentication System

- Google OAuth2 integration
- Session-based authentication
- Secure cookie handling with HttpOnly and SameSite flags
- Development/production mode support

### 2. User Management

- User registration and profiles
- Email-based user identification
- OAuth provider tracking

### 3. Event Management

- Create, read, update, delete events
- Event types: lectures, workshops, hackathons, job fairs, etc.
- Registration tracking and limits
- ICS calendar file support

### 4. Community Features

- Alumni directory
- Startup showcase
- Speaker management
- Sponsor management

### 5. Security Features

- Rate limiting middleware
- Admin authorization
- Session management
- CORS protection

## 📝 Development Guidelines

### Adding a New Feature

1. **Create the model** in `internal/models/`

   ```go
   type NewFeature struct {
       gorm.Model
       Name string `gorm:"not null" json:"name"`
       // ... other fields
   }
   ```

2. **Create the handler** in `internal/handlers/`

   ```go
   type NewFeatureHandler struct {
       db *gorm.DB
   }

   func (h *NewFeatureHandler) Register(router *gin.RouterGroup) {
       router.GET("/features", h.List)
       router.POST("/features", h.Create)
       // ... other routes
   }
   ```

3. **Register the handler** in `cmd/api/main.go`

   ```go
   allHandlers := []handlers.Handler{
       // ... existing handlers
       handlers.NewFeatureHandler(db),
   }
   ```

4. **Add to auto-migration** in `cmd/api/main.go`
   ```go
   db.AutoMigrate(
       // ... existing models
       &models.NewFeature{},
   )
   ```

### Code Style

- Follow [Effective Go](https://golang.org/doc/effective_go) guidelines
- Use meaningful variable and function names
- Add comments for exported functions
- Keep functions focused and single-purpose

### Git Workflow

This project uses [Conventional Commits](https://www.conventionalcommits.org/):

**Commit format:**

```
<type>(<scope>): <description>

[optional body]

[optional footer]
```

**Types:**

- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation changes
- `style`: Code style changes (formatting)
- `refactor`: Code refactoring
- `test`: Adding tests
- `chore`: Maintenance tasks

**Example:**

```
feat(events): add recurring events support

- Add recurrence pattern to Event model
- Implement RRULE parsing
- Update event creation endpoint

fixes #42
```

### Testing

- Write unit tests for handlers
- Use table-driven tests where appropriate
- Mock database interactions for unit tests
- Integration tests should use a test database

### API Versioning

All routes are versioned under `/api/v1/`:

- Ensures backward compatibility
- Allows gradual migration to new versions
- Clear API evolution path

## 📚 Additional Resources

- [Gin Documentation](https://gin-gonic.com/docs/)
- [GORM Documentation](https://gorm.io/docs/)
- [Go Best Practices](https://go.dev/doc/effective_go)
- [Conventional Commits](https://www.conventionalcommits.org/)

## 🤝 Getting Help

- Check existing issues on GitHub
- Ask in the team Mattermost channel
- Review the main [README.md](README.md) for project-specific information
- Contact the Head of IT for architectural questions

## 📄 License

This project is licensed under the MIT License. See [LICENSE](LICENSE) for details.

**Note:** Logos, icons, and images are property of KTH AI Society and are not covered by the MIT license.

---

## 🚧 Known Issues & TODO

### Critical Issues

#### 1. **Authentication System Broken**

- **Google OAuth Redirect Issue**: After Google confirmation page, the redirect flow is not working correctly
- **Magic Link Sign-up**: Not implemented yet - needs to be added as an alternative authentication method
- **Frontend Auth Handling**: Frontend is not properly handling authentication state and redirects
- **Action Items**:
  - [ ] Fix Google OAuth callback redirect logic in `internal/handlers/auth_handler.go`
  - [ ] Implement magic link email-based authentication system
  - [ ] Review and fix session management between backend and frontend
  - [ ] Ensure CORS and cookie settings are properly configured for auth flow
  - [ ] Add proper error handling and user feedback for failed auth attempts

#### 2. **Event Management - Switch to Luma API**

- **Current State**: Custom event management implementation
- **Goal**: Integrate with Luma API for event management
- **Action Items**:
  - [ ] Research Luma API endpoints and authentication
  - [ ] Design integration architecture (replace or wrap existing event handlers?)
  - [ ] Implement Luma API client in `internal/handlers/event_handler.go`
  - [ ] Update event models to match Luma's data structure
  - [ ] Migrate existing events or create sync mechanism
  - [ ] Update registration flow to use Luma
  - [ ] Review all event-related routes (`/api/v1/events/*`) for compatibility

#### 3. **Validation Issues**

- **Problem**: Poor or missing validation across the codebase
- **Action Items**:
  - [ ] Audit all handler functions for input validation
  - [ ] Add proper validation for:
    - Event creation/updates
    - User registration data
    - Profile updates
    - Alumni/Startup/Speaker/Sponsor submissions
  - [ ] Implement consistent error responses for validation failures
  - [ ] Consider using a validation library or middleware
  - [ ] Add validation tests

#### 4. **Poor Error Handling**

- **Problem**: Inconsistent or missing error handling throughout the codebase
- **Action Items**:
  - [ ] Audit all handlers for proper error handling
  - [ ] Implement consistent error response format (status codes, error messages, error types)
  - [ ] Add error logging and monitoring
  - [ ] Handle database errors gracefully (connection failures, query errors, etc.)
  - [ ] Improve error messages for debugging and user feedback
  - [ ] Add error recovery mechanisms where appropriate
  - [ ] Document error handling patterns in the codebase
  - [ ] Add error handling middleware for common cases

### Route Review Needed

Review all API routes for proper implementation:

- [ ] `/api/v1/events/*` - Event management (priority: Luma integration)
- [ ] `/api/v1/auth/*` - Authentication flows (priority: fix redirects)
- [ ] `/api/v1/profile/*` - User profiles
- [ ] `/api/v1/registration/*` - Event registrations
- [ ] `/api/v1/alumni/*` - Alumni directory
- [ ] `/api/v1/startups/*` - Startup showcase
- [ ] `/api/v1/speakers/*` - Speaker management
- [ ] `/api/v1/sponsors/*` - Sponsor management

### Frontend-Backend Integration

- [ ] Verify auth state synchronization between frontend and backend
- [ ] Ensure proper error handling for all API responses
- [ ] Test session persistence across browser refreshes
- [ ] Validate CORS configuration for production environment
