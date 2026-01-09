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

**Projects:**

- `GET /api/projects?page=1` - List projects (10 per page)
- `GET /api/projects/:id` - Get project by ID
- `POST /api/projects` - Create project (multipart: name, category, client, order, files[], highlightImageIndex)
- `PUT /api/projects/:id` - Update project (multipart: same fields, files[] can mix IDs + new files)
- `PUT /api/projects/:id/highlight/toggle` - Toggle highlighted boolean
- `DELETE /api/projects/:id` - Delete project (cascade deletes images)

**Testimonials:**

- `GET /api/testimonials?page=1` - List testimonials (10 per page)
- `GET /api/testimonials/:id` - Get testimonial by ID
- `POST /api/testimonials` - Create testimonial (JSON: full_name, profession, testimonial, status)
- `PUT /api/testimonials/:id` - Update testimonial (JSON: same fields)
- `DELETE /api/testimonials/:id` - Delete testimonial (400 if only 1 remains)

**Users:**

- `GET /api/users?page=1` - List users (10 per page)
- `GET /api/users/:id` - Get user by ID
- `POST /api/users` - Create user (JSON: name, email, password)
- `PUT /api/users/:id` - Update user (JSON: name, email)
- `DELETE /api/users/:id` - Delete user (400 if only 1 remains)

**Static Texts:**

- `GET /api/static-texts?page=1` - List static texts (10 per page)
- `GET /api/static-texts/:id` - Get static text by ID
- `PUT /api/static-texts/:id` - Update static text (JSON: content)

**Configurations:**

- `PUT /api/configs/:key` - Update configuration (JSON: {"value": "..."})

**Visitor Messages:**

- `GET /api/visitor-messages?page=1` - List visitor messages (10 per page)
- `DELETE /api/visitor-messages/:id` - Delete visitor message

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

## CLI

To createa new admin user, run the following command:

```bash
go run ./cmd/create-admin --name "Admin User" --email "admin@example.com" --password "securepassword"
```

## License

Copyright © 2024 Elite Constructions
