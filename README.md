# Greenlight API

A RESTful JSON API for managing movie records, built with Go and PostgreSQL. This project serves as a comprehensive reference implementation demonstrating modern Go API development patterns, database management, authentication, rate limiting, and performance profiling.

> **⚠️ Practice Repository Warning**
> This is a learning/practice repository. Use in production environments at your own risk.

## Features

- **RESTful API** with JSON responses
- **JWT Authentication** with user registration and activation
- **Role-based Access Control** with permissions system
- **Rate Limiting** to prevent abuse
- **Database Migrations** with PostgreSQL
- **Email Integration** for user activation and password reset
- **Performance Profiling** with pprof integration
- **CORS Support** for cross-origin requests
- **Comprehensive Logging** with structured JSON logs
- **Graceful Shutdown** handling
- **Docker Support** for database setup
- **Production Deployment** configuration

## API Endpoints

### Public Endpoints
- `GET /v1/healthcheck` - API health status
- `POST /v1/users` - User registration
- `POST /v1/tokens/authentication` - User login
- `POST /v1/tokens/password-reset` - Request password reset
- `POST /v1/tokens/activation` - Request activation token
- `PUT /v1/users/activated` - Activate user account

### Protected Endpoints (Require Authentication)
- `GET /v1/movies` - List movies with filtering and pagination
- `POST /v1/movies` - Create new movie (requires `movies:write` permission)
- `GET /v1/movies/:id` - Get movie by ID (requires `movies:read` permission)
- `PATCH /v1/movies/:id` - Update movie (requires `movies:write` permission)
- `DELETE /v1/movies/:id` - Delete movie (requires `movies:write` permission)
- `PUT /v1/users/password` - Update user password

### Debug Endpoints
- `GET /debug/vars` - Runtime metrics and statistics

## Quick Start

### Prerequisites
- Go 1.24+
- PostgreSQL 12+
- migrate CLI tool
- make

### 1. Clone and Setup
```bash
git clone <repository-url>
cd greenlight
```

### 2. Environment Configuration
Create a `.envrc` file:
```bash
export GREENLIGHT_DB_DSN="postgres://username:password@localhost/greenlight?sslmode=disable" # ssl can be removed if you are running postgres locally/using ssl certs with Docker
export POSTGRES_USER="your_db_user"
export POSTGRES_PASSWORD="your_db_password"
```

### 3. Database Setup
```bash
# Start PostgreSQL container
make db/start

# Run migrations
make db/migrations/up
```

### 4. Install Dependencies
```bash
make vendor
```

### 5. Run the API
```bash
make run/api
```

The API will be available at `http://localhost:4000`

## Development

### Available Make Commands

#### Development
- `make run/api` - Start the API server
- `make run/api/profiling` - Start API with profiling enabled
- `make audit` - Format, vet, and test code
- `make vendor` - Tidy and vendor dependencies

#### Database
- `make db/start` - Start PostgreSQL container
- `make db/stop` - Stop PostgreSQL container
- `make db/psql` - Connect to database
- `make db/migrations/new name=migration_name` - Create new migration
- `make db/migrations/up` - Apply migrations

#### Build
- `make build/api` - Build production binary

#### Production (NOT RECOMMENDED UNLESS YOU KNOW WHAT YOU ARE DOING!)
- `make production/connect` - SSH to production server
- `make production/deploy/api` - Deploy to production

### Configuration Options

The API supports various configuration flags, pick whichever for whatever scenario you want to mess about with:

```bash
go run ./cmd/api \
  -port=4000 \
  -env=development \
  -db-dsn="postgres://..." \
  -db-max-open-conns=25 \
  -db-max-idle-conns=25 \
  -db-max-idle-time=15m \
  -limiter-rps=2 \
  -limiter-burst=4 \
  -limiter-enabled=true \
  -smtp-host=smtp.mailtrap.io \
  -smtp-port=2525 \
  -cors-trusted-origins="http://localhost:3000"
```

## Performance Profiling

The API includes comprehensive profiling capabilities. See [PROFILING.md](PROFILING.md) for detailed instructions.

### Quick Profiling
```bash
# Start with profiling enabled
make run/api/profiling

# Capture CPU profile
make profile/cpu

# View metrics
make profile/metrics

# Capture all profiles
make profile/all
```

## Database Schema

The API uses PostgreSQL with the following main tables:

- **movies** - Movie records with title, year, runtime, genres
- **users** - User accounts with email, password hash, activation status
- **tokens** - Authentication and activation tokens
- **permissions** - Role-based access control
- **users_permissions** - User permission assignments

## Authentication & Authorization

### User Registration Flow
1. `POST /v1/users` - Register with email/password
2. Activation email sent with token
3. `PUT /v1/users/activated` - Activate account with token

### Authentication Flow
1. `POST /v1/tokens/authentication` - Login with email/password
2. Receive JWT token
3. Include `Authorization: Bearer <token>` header in subsequent requests

### Permissions System
- `movies:read` - Read movie data
- `movies:write` - Create, update, delete movies

## Rate Limiting

- Default: 2 requests per second with burst of 4
- Per-IP tracking with automatic cleanup
- Configurable via command line flags

## CORS Configuration

Configure trusted origins for cross-origin requests:
```bash
-cors-trusted-origins="http://localhost:3000 https://mydomain.com"
```

## Deployment

### Docker Compose
```bash
# Start database
docker-compose up -d postgres
```

### Production Deployment
That's for you to figure out the rest!
Won't be responsible for you doing something dumb, and I can't help with that till I comb through the cli docs of whatever vps I want to settle with, hence the incomplete setup.
Why push to repo?
So that me seeing it pushes me to see the whole process to the end (plus that remote file is basic at most)

## Contributing

This is a practice repository for learning purposes. The code follows patterns from "Let's Go Further" by Alex Edwards.

## License

This project is for educational purposes. Use at your own risk in production environments.

## Reference & Learning

Based on "Let's Go Further" by Alex Edwards. The codebase serves as a reference implementation for:
- RESTful API design patterns
- Go web development best practices
- Database integration and migrations
- Authentication and authorization
- Performance monitoring and profiling
- Production deployment strategies

For web application development, consider "Let's Go" by the same author.
