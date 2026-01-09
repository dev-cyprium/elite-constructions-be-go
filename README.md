# Elite Constructions Backend API

Go REST API backend for Elite Constructions built with Gin, PostgreSQL, and sqlc.

## Features

- RESTful API with JWT authentication
- Argon2id password hashing
- PostgreSQL database with migrations
- Local filesystem storage for images
- Blurhash generation for images
- Docker support for production

## Setup

### Prerequisites

- Go 1.21+
- PostgreSQL 15+
- Docker and Docker Compose (for production)

### Local Development

1. Clone the repository
2. Copy `.env.example` to `.env` and configure:
   ```env
   DATABASE_URL=postgres://user:password@localhost:5432/elite_constructions?sslmode=disable
   JWT_SECRET=your-secret-key-min-32-chars-change-in-production
   PORT=8080
   STORAGE_PATH=./storage
   ```

3. Run migrations (automatically on server startup)
4. Start the server:
   ```bash
   go run cmd/server/main.go
   ```

## Docker Production

### Build and Run

```bash
docker-compose -f docker-compose.prod.yaml up -d
```

### Environment Variables

Set the following environment variables:
- `POSTGRES_USER` (default: `elite`)
- `POSTGRES_PASSWORD` (required)
- `POSTGRES_DB` (default: `elite_constructions`)
- `JWT_SECRET` (required, min 32 characters)

## Migration Tools

### Database Migration (MySQL → PostgreSQL)

Migrate data from MySQL to PostgreSQL:

```bash
go run cmd/migrate-db/main.go \
  --mysql-dsn="user:password@tcp(localhost:3306)/database" \
  --postgres-dsn="postgres://user:password@localhost:5432/database?sslmode=disable"
```

Options:
- `--dry-run`: Perform a dry run without making changes

### File Migration

Migrate files from existing storage to new location:

```bash
go run cmd/migrate-files/main.go \
  --source-dir="/path/to/existing/storage" \
  --target-dir="./storage/public" \
  --postgres-dsn="postgres://user:password@localhost:5432/database?sslmode=disable"
```

## API Endpoints

### Public Endpoints

- `GET /ping` - Health check
- `GET /api/pub/projects` - List all projects
- `GET /api/pub/projects/highlighted` - List highlighted projects
- `GET /api/pub/async/projects/page?page=1` - Paginated projects (3 per page)
- `GET /api/pub/projects/:id` - Get project by ID
- `GET /api/pub/testimonials` - List testimonials (status='ready')
- `POST /api/pub/testimonials` - Create testimonial
- `GET /api/pub/static-texts` - List static texts
- `GET /api/pub/configs` - List configurations
- `POST /api/pub/visitor-messages` - Create visitor message

### Authentication

- `POST /api/login` - Login (returns JWT or reset token)
- `POST /api/password-reset/complete` - Complete password reset
- `POST /api/logout` - Logout
- `GET /api/me` - Get current user (requires auth)

### Admin Endpoints (require JWT)

See the plan document for complete endpoint documentation.

## Project Structure

```
elite-constructions-backend/
├── cmd/
│   ├── server/          # API server
│   ├── migrate-db/      # Database migration tool
│   └── migrate-files/   # File migration tool
├── internal/
│   ├── config/          # Configuration
│   ├── db/              # Database connection and migrations
│   ├── models/          # Domain models
│   ├── sqlc/            # Generated SQL queries
│   ├── http/            # HTTP handlers and router
│   ├── middleware/      # Middleware (auth, CORS, errors)
│   ├── storage/         # File storage and blurhash
│   └── auth/            # Authentication (JWT, Argon2id, reset)
├── migrations/          # Database migrations
├── queries/             # SQL queries for sqlc
└── storage/             # Local file storage
```

## Development

### Generate sqlc Code

```bash
sqlc generate
```

### Run Tests

```bash
go test ./...
```

## License

Copyright © 2024 Elite Constructions
