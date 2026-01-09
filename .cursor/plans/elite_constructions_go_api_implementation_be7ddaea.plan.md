---
name: Elite Constructions Go API Implementation
overview: ""
todos:
  - id: phase1-init-dependencies
    content: Initialize Go module dependencies (Gin, sqlc, pgx, golang-migrate, validator, JWT, argon2, blurhash)
    status: completed
  - id: phase1-config-package
    content: Create internal/config package to load environment variables (DATABASE_URL, JWT_SECRET, PORT, STORAGE_PATH)
    status: completed
  - id: phase1-db-connection
    content: Setup internal/db/connection.go with pgxpool PostgreSQL connection
    status: completed
    dependencies:
      - phase1-config-package
  - id: phase1-db-migrations
    content: Setup internal/db/migrations.go with golang-migrate runner
    status: completed
    dependencies:
      - phase1-db-connection
  - id: phase1-env-example
    content: Create .env.example with minimal required environment variables
    status: completed
  - id: phase1-sqlc-config
    content: Create sqlc.yaml configuration file for code generation
    status: completed
---

# Elite Constructions Go REST API - Implementation Plan

## Architecture Decisions

**Stack:**

- Framework: **Gin** (lightweight, fast, well-documented)
- DB Access: **sqlc + pgx** (compile-time type safety, better performance than GORM)
- Migrations: **golang-migrate/migrate** (industry standard, supports both up/down)
- Validation: **go-playground/validator/v10**
- JWT: **github.com/golang-jwt/jwt/v5**
- Password: **golang.org/x/crypto/argon2** (Argon2id)
- File Storage: Local filesystem only

**Migrations Strategy:** Run migrations automatically on API startup (simpler for Docker). Alternative: one-shot `cmd/migrate` if preferred.

## Project Structure

```
elite-constructions-backend/
├── cmd/
│   ├── server/
│   │   └── main.go                 # API server entrypoint
│   ├── migrate-db/
│   │   └── main.go                 # MySQL -> PostgreSQL migration tool
│   └── migrate-files/
│       └── main.go                  # File migration tool
├── internal/
│   ├── config/
│   │   └── config.go               # Environment config loader
│   ├── db/
│   │   ├── connection.go           # PostgreSQL connection (pgxpool)
│   │   └── migrations.go           # Migration runner (golang-migrate)
│   ├── models/
│   │   └── models.go               # Domain models (structs)
│   ├── sqlc/
│   │   └── (generated)             # sqlc generated queries
│   ├── http/
│   │   ├── router.go               # Gin router setup
│   │   ├── handlers/
│   │   │   ├── public.go          # Public endpoints
│   │   │   ├── auth.go            # Login/logout/me
│   │   │   ├── projects.go        # Admin projects CRUD
│   │   │   ├── testimonials.go   # Admin testimonials CRUD
│   │   │   ├── users.go           # Admin users CRUD
│   │   │   ├── static_texts.go    # Admin static texts
│   │   │   ├── configs.go         # Admin configurations
│   │   │   └── visitor_messages.go # Admin visitor messages
│   │   └── responses.go           # Standardized JSON responses
│   ├── middleware/
│   │   ├── auth.go                # JWT validation middleware
│   │   ├── cors.go                # CORS middleware
│   │   └── errors.go              # Error handling middleware
│   ├── storage/
│   │   ├── local.go               # Local filesystem storage
│   │   └── blurhash.go            # Blurhash generation (synchronous)
│   └── auth/
│       ├── argon2.go              # Argon2id hashing
│       ├── jwt.go                 # JWT generation/validation
│       └── reset.go               # Password reset token flow
├── migrations/
│   ├── 000001_init_schema.up.sql
│   ├── 000001_init_schema.down.sql
│   └── (additional migrations)
├── queries/
│   └── (sqlc query files)
├── storage/
│   └── public/
│       └── img/                   # Uploaded project images
├── Dockerfile                      # Multi-stage build
├── docker-compose.prod.yaml        # Production compose (API + Postgres)
├── .env.example                    # Minimal env template
├── sqlc.yaml                       # sqlc configuration
├── migrate.yaml                    # golang-migrate config (optional)
└── README.md                       # Setup, migration, Docker docs
```

## Data Models

### Core Models

**User:**

```go
type User struct {
    ID                   int64     `json:"id"`
    Name                 string    `json:"name"`
    Email                string    `json:"email"`
    EmailVerifiedAt      *time.Time `json:"email_verified_at,omitempty"`
    Password             string    `json:"-"` // Argon2id hash
    PasswordResetRequired bool     `json:"password_reset_required"`
    ResetTokenHash       *string   `json:"-"` // SHA256 of reset token
    ResetTokenExpiresAt  *time.Time `json:"-"`
    RememberToken        *string   `json:"-"`
    CreatedAt            time.Time `json:"created_at"`
    UpdatedAt            time.Time `json:"updated_at"`
}
```

**Project:**

```go
type Project struct {
    ID         int64          `json:"id"`
    Status     int            `json:"status"`
    Name       string         `json:"name"`
    Category   *string        `json:"category,omitempty"`
    Client     *string        `json:"client,omitempty"`
    Order      int            `json:"order"`
    Highlighted bool          `json:"highlighted"`
    Images     []ProjectImage `json:"images,omitempty"`
    CreatedAt  time.Time      `json:"created_at"`
    UpdatedAt  time.Time      `json:"updated_at"`
}
```

**ProjectImage:**

```go
type ProjectImage struct {
    ID        int64     `json:"id"`
    Name      string    `json:"name"`
    URL       string    `json:"url"` // /storage/img/filename.jpg
    ProjectID int64     `json:"project_id"`
    Order     int       `json:"order"`
    BlurHash  *string   `json:"blur_hash,omitempty"` // data URL
    CreatedAt time.Time `json:"created_at"`
    UpdatedAt time.Time `json:"updated_at"`
}
```

**Testimonial:**

```go
type Testimonial struct {
    ID         int64     `json:"id"`
    FullName   string    `json:"full_name"`
    Profession string    `json:"profession"`
    Testimonial string   `json:"testimonial"`
    Status     string    `json:"status"` // "ready", "pending", etc.
    CreatedAt  time.Time `json:"created_at"`
    UpdatedAt  time.Time `json:"updated_at"`
}
```

**StaticText:**

```go
type StaticText struct {
    ID        int64     `json:"id"`
    Key       string    `json:"key"`
    Label     string    `json:"label"`
    Content   string    `json:"content"`
    CreatedAt time.Time `json:"created_at"`
    UpdatedAt time.Time `json:"updated_at"`
}
```

**Configuration:**

```go
type Configuration struct {
    ID        int64     `json:"id"`
    Key       string    `json:"key"`
    Value     string    `json:"value"`
    CreatedAt time.Time `json:"created_at"`
    UpdatedAt time.Time `json:"updated_at"`
}
```

**VisitorMessage:**

```go
type VisitorMessage struct {
    ID          int64     `json:"id"`
    Email       string    `json:"email"`
    Address     string    `json:"address"`
    Description string    `json:"description"`
    Seen        bool      `json:"seen"`
    CreatedAt   time.Time `json:"created_at"`
    UpdatedAt   time.Time `json:"updated_at"`
}
```

## Database Schema

### Migration: `000001_init_schema.up.sql`

```sql
-- Users table with Argon2id migration support
CREATE TABLE users (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    email_verified_at TIMESTAMP,
    password VARCHAR(255) NOT NULL, -- Argon2id hash
    password_reset_required BOOLEAN DEFAULT false,
    reset_token_hash VARCHAR(64), -- SHA256 hash of reset token
    reset_token_expires_at TIMESTAMP,
    remember_token VARCHAR(100),
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Projects
CREATE TABLE projects (
    id BIGSERIAL PRIMARY KEY,
    status SMALLINT NOT NULL DEFAULT 1,
    name VARCHAR(255) NOT NULL,
    category VARCHAR(255),
    client VARCHAR(255),
    "order" INTEGER NOT NULL DEFAULT 0,
    highlighted BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Project Images
CREATE TABLE project_images (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    url VARCHAR(500) NOT NULL, -- /storage/img/filename.jpg
    project_id BIGINT NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    "order" INTEGER NOT NULL DEFAULT 0,
    blur_hash TEXT, -- data URL format
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_project_images_project_id ON project_images(project_id);

-- Testimonials
CREATE TABLE testimonials (
    id BIGSERIAL PRIMARY KEY,
    full_name VARCHAR(255) NOT NULL,
    profession VARCHAR(255) NOT NULL,
    testimonial TEXT NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'ready',
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Static Texts
CREATE TABLE static_texts (
    id BIGSERIAL PRIMARY KEY,
    key VARCHAR(255) UNIQUE NOT NULL,
    label VARCHAR(255) NOT NULL,
    content TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Configurations
CREATE TABLE configurations (
    id BIGSERIAL PRIMARY KEY,
    key VARCHAR(255) UNIQUE NOT NULL,
    value TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Visitor Messages
CREATE TABLE visitor_messages (
    id BIGSERIAL PRIMARY KEY,
    email VARCHAR(255) NOT NULL,
    address TEXT NOT NULL,
    description TEXT NOT NULL,
    seen BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Seed configurations
INSERT INTO configurations (key, value) VALUES ('config.site.live', 'false');

-- Seed static texts (known keys)
INSERT INTO static_texts (key, label, content) VALUES
    ('hero.title', 'Hero Title', 'Welcome to Elite Constructions'),
    ('hero.subtitle', 'Hero Subtitle', 'Building Excellence Since...'),
    ('projects.read_more', 'Read More', 'Read More'),
    ('company.name', 'Company Name', 'Elite Constructions'),
    ('company.description', 'Company Description', '...'),
    ('footer.copyright', 'Footer Copyright', '© 2024 Elite Constructions');
```

## Endpoint Mapping

### Public Endpoints (No Auth)

| Method | Route | Handler | Notes |

| ------ | ------------------------------- | ------------------------------------- | ------------------------------------------- |

| GET | `/ping` | `handlers.Ping` | Returns `{"message": "Welcome to API 1.0"}` |

| GET | `/api/pub/projects` | `handlers.GetPublicProjects` | All projects (no pagination) |

| GET | `/api/pub/projects/highlighted` | `handlers.GetHighlightedProjects` | Filtered by highlighted=true |

| GET | `/api/pub/async/projects/page` | `handlers.GetPublicProjectsPaginated` | Query: `?page=1` (3 per page) |

| GET | `/api/pub/projects/:id` | `handlers.GetPublicProject` | Single project with images |

| GET | `/api/pub/testimonials` | `handlers.GetPublicTestimonials` | Only status='ready' |

| POST | `/api/pub/testimonials` | `handlers.CreatePublicTestimonial` | JSON body, sets status='pending' |

| GET | `/api/pub/static-texts` | `handlers.GetPublicStaticTexts` | All static texts |

| GET | `/api/pub/configs` | `handlers.GetPublicConfigs` | All configurations |

| POST | `/api/pub/visitor-messages` | `handlers.CreateVisitorMessage` | JSON body |

### Auth Endpoints

| Method | Route | Handler | Auth | Notes |

| ------ | ------------------------------ | -------------------------------- | ---- | --------------------------------------------------------------- |

| POST | `/api/login` | `handlers.Login` | No | Returns JWT OR `{"reset_required": true, "reset_token": "..."}` |

| POST | `/api/logout` | `handlers.Logout` | Yes | Stateless (optional, can be no-op) |

| GET | `/api/me` | `handlers.GetMe` | Yes | Current user info |

| POST | `/api/password-reset/complete` | `handlers.CompletePasswordReset` | No | Body: `{"reset_token": "...", "new_password": "..."}` |

### Admin Projects

| Method | Route | Handler | Auth | Notes |

| ------ | ------------------------------------ | -------------------------- | ---- | ---------------------------------------------------------------------- |

| GET | `/api/projects` | `handlers.GetProjects` | Yes | Query: `?page=1` (10 per page) |

| GET | `/api/projects/:id` | `handlers.GetProject` | Yes | Single project |

| POST | `/api/projects` | `handlers.CreateProject` | Yes | Multipart: name, category, client, order, files[], highlightImageIndex |

| PUT | `/api/projects/:id` | `handlers.UpdateProject` | Yes | Multipart: same fields, files[] can mix IDs + new files |

| PUT | `/api/projects/:id/highlight/toggle` | `handlers.ToggleHighlight` | Yes | Toggle highlighted boolean |

| DELETE | `/api/projects/:id` | `handlers.DeleteProject` | Yes | Cascade deletes images |

### Admin Testimonials

| Method | Route | Handler | Auth | Notes |

| ------ | ----------------------- | ---------------------------- | ---- | ------------------------------------------------ |

| GET | `/api/testimonials` | `handlers.GetTestimonials` | Yes | Query: `?page=1` (10 per page) |

| GET | `/api/testimonials/:id` | `handlers.GetTestimonial` | Yes | Single testimonial |

| POST | `/api/testimonials` | `handlers.CreateTestimonial` | Yes | JSON: full_name, profession, testimonial, status |

| PUT | `/api/testimonials/:id` | `handlers.UpdateTestimonial` | Yes | JSON: same fields |

| DELETE | `/api/testimonials/:id` | `handlers.DeleteTestimonial` | Yes | 400 if only 1 remains, else 204 |

### Admin Users

| Method | Route | Handler | Auth | Notes |

| ------ | ---------------- | --------------------- | ---- | -------------------------------------- |

| GET | `/api/users` | `handlers.GetUsers` | Yes | Query: `?page=1` (10 per page) |

| GET | `/api/users/:id` | `handlers.GetUser` | Yes | Single user |

| POST | `/api/users` | `handlers.CreateUser` | Yes | JSON: name, email, password (Argon2id) |

| PUT | `/api/users/:id` | `handlers.UpdateUser` | Yes | JSON: name, email (no password) |

| DELETE | `/api/users/:id` | `handlers.DeleteUser` | Yes | 400 if only 1 remains, else 204 |

### Admin Static Texts

| Method | Route | Handler | Auth | Notes |

| ------ | ----------------------- | --------------------------- | ---- | ----------------------------------- |

| GET | `/api/static-texts` | `handlers.GetStaticTexts` | Yes | Query: `?page=1` (10 per page) |

| GET | `/api/static-texts/:id` | `handlers.GetStaticText` | Yes | Single static text |

| PUT | `/api/static-texts/:id` | `handlers.UpdateStaticText` | Yes | JSON: content (key/label immutable) |

### Admin Configurations

| Method | Route | Handler | Auth | Notes |

| ------ | ------------------- | ----------------------- | ---- | ------------------------ |

| PUT | `/api/configs/:key` | `handlers.UpdateConfig` | Yes | JSON: `{"value": "..."}` |

### Admin Visitor Messages

| Method | Route | Handler | Auth | Notes |

| ------ | --------------------------- | ------------------------------- | ---- | ------------------------------ |

| GET | `/api/visitor-messages` | `handlers.GetVisitorMessages` | Yes | Query: `?page=1` (10 per page) |

| DELETE | `/api/visitor-messages/:id` | `handlers.DeleteVisitorMessage` | Yes | Returns 204 |

## Implementation Phases

### Phase 1: Bootstrapping & Configuration

**Todos:**

- [ ] Initialize Go module dependencies (Gin, sqlc, pgx, golang-migrate, validator, JWT, argon2, blurhash)
- [ ] Create `internal/config` package (load from env: `DATABASE_URL`, `JWT_SECRET`, `PORT`, `STORAGE_PATH`)
- [ ] Setup `internal/db/connection.go` (pgxpool connection)
- [ ] Setup `internal/db/migrations.go` (golang-migrate runner)
- [ ] Create `.env.example` with minimal required vars
- [ ] Create `sqlc.yaml` configuration

**Details:**

1.  Initialize Go module dependencies

                                                                                                - Add dependencies: `github.com/gin-gonic/gin`, `github.com/kyleconroy/sqlc`, `github.com/jackc/pgx/v5`, `github.com/golang-migrate/migrate/v4`, `github.com/go-playground/validator/v10`, `github.com/golang-jwt/jwt/v5`, `golang.org/x/crypto/argon2`, `github.com/buckket/go-blurhash`

2.  Create `internal/config` package (load from env: `DATABASE_URL`, `JWT_SECRET`, `PORT`, `STORAGE_PATH`)

                                                                                                - Load and validate environment variables
                                                                                                - Provide typed config struct

3.  Setup `internal/db/connection.go` (pgxpool connection)

                                                                                                - Initialize pgxpool.Connect with DATABASE_URL
                                                                                                - Return connection pool for use across application

4.  Setup `internal/db/migrations.go` (golang-migrate runner)

                                                                                                - Initialize migrate instance pointing to migrations directory
                                                                                                - Implement Run() function to execute migrations on startup

5.  Create `.env.example` with minimal required vars

                                                                                                - DATABASE_URL, JWT_SECRET, PORT, STORAGE_PATH

6.  Create `sqlc.yaml` configuration

                                                                                                - Configure sqlc to generate code from queries directory
                                                                                                - Set PostgreSQL driver and output paths

### Phase 2: Database Schema & Migrations

1. Write `migrations/000001_init_schema.up.sql` (all tables)
2. Write `migrations/000001_init_schema.down.sql` (drop all)
3. Implement migration runner in `internal/db/migrations.go` (run on startup)
4. Test migrations up/down

### Phase 3: Models & SQL Queries (sqlc)

1. Create `queries/` directory with `.sql` files for each resource
2. Generate sqlc code: `sqlc generate`
3. Create `internal/models/models.go` with domain structs
4. Map sqlc types to domain models

### Phase 4: Authentication & Authorization

1.  Implement `internal/auth/argon2.go` (hash/verify Argon2id)
2.  Implement `internal/auth/jwt.go` (generate/validate JWT with HS256)
3.  Implement `internal/auth/reset.go` (reset token generation/validation)
4.  Create `internal/middleware/auth.go` (JWT validation middleware)
5.  Implement login flow in `handlers/auth.go`:

                                                                                                                                                                                                                                                                                                                                                                                                - Check if user has Argon2id password
                                                                                                                                                                                                                                                                                                                                                                                                - If not: generate reset token, return `{"reset_required": true, "reset_token": "..."}`
                                                                                                                                                                                                                                                                                                                                                                                                - If yes: verify password, return JWT

6.  Implement `handlers.CompletePasswordReset` (validate token, set new Argon2id password)

### Phase 5: Storage & File Handling

1.  Implement `internal/storage/local.go`:

                                                                                                                                                                                                                                                                                                                                                                                                - `SaveFile(file []byte, filename string) (string, error)` -> returns `/storage/img/...`
                                                                                                                                                                                                                                                                                                                                                                                                - `DeleteFile(url string) error`
                                                                                                                                                                                                                                                                                                                                                                                                - Validate image types (jpeg, png, webp)
                                                                                                                                                                                                                                                                                                                                                                                                - Generate unique filename (SHA1 of bytes + extension)

2.  Implement `internal/storage/blurhash.go`:

                                                                                                                                                                                                                                                                                                                                                                                                - `GenerateBlurHash(filePath string) (string, error)` -> returns data URL
                                                                                                                                                                                                                                                                                                                                                                                                - Use `github.com/buckket/go-blurhash` (synchronous)

3.  Setup static file serving in `router.go` (serve `/storage/img/*` from `./storage/public/img/`)

### Phase 6: Public Endpoints

1.  Implement `handlers/public.go`:

                                                                                                                                                                                                                                                                                                                                                                                                - `Ping`, `GetPublicProjects`, `GetHighlightedProjects`, `GetPublicProjectsPaginated`
                                                                                                                                                                                                                                                                                                                                                                                                - `GetPublicProject`, `GetPublicTestimonials`, `CreatePublicTestimonial`
                                                                                                                                                                                                                                                                                                                                                                                                - `GetPublicStaticTexts`, `GetPublicConfigs`, `CreateVisitorMessage`

2.  Wire up public routes in `router.go` (no auth middleware)

### Phase 7: Admin Endpoints

1.  Implement `handlers/projects.go`:

                                                                                                                                                                                                                                                                                                                                                                                                - `GetProjects` (paginated, 10 per page)
                                                                                                                                                                                                                                                                                                                                                                                                - `GetProject` (single with images)
                                                                                                                                                                                                                                                                                                                                                                                                - `CreateProject` (multipart: parse form, save files, generate blurhash, create DB records)
                                                                                                                                                                                                                                                                                                                                                                                                - `UpdateProject` (multipart: handle existing image IDs + new files, delete removed images)
                                                                                                                                                                                                                                                                                                                                                                                                - `ToggleHighlight` (toggle boolean)
                                                                                                                                                                                                                                                                                                                                                                                                - `DeleteProject` (cascade deletes images via FK)

2.  Implement `handlers/testimonials.go`:

                                                                                                                                                                                                                                                                                                                                                                                                - `GetTestimonials` (paginated)
                                                                                                                                                                                                                                                                                                                                                                                                - `GetTestimonial` (single)
                                                                                                                                                                                                                                                                                                                                                                                                - `CreateTestimonial` (JSON, status required)
                                                                                                                                                                                                                                                                                                                                                                                                - `UpdateTestimonial` (JSON)
                                                                                                                                                                                                                                                                                                                                                                                                - `DeleteTestimonial` (check count, return 400 if only 1 remains)

3.  Implement `handlers/users.go`:

                                                                                                                                                                                                                                                                                                                                                                                                - `GetUsers` (paginated)
                                                                                                                                                                                                                                                                                                                                                                                                - `GetUser` (single, no password)
                                                                                                                                                                                                                                                                                                                                                                                                - `CreateUser` (JSON, hash password with Argon2id)
                                                                                                                                                                                                                                                                                                                                                                                                - `UpdateUser` (JSON, name/email only)
                                                                                                                                                                                                                                                                                                                                                                                                - `DeleteUser` (check count, return 400 if only 1 remains)

4.  Implement `handlers/static_texts.go`:

                                                                                                                                                                                                                                                                                                                                                                                                - `GetStaticTexts` (paginated)
                                                                                                                                                                                                                                                                                                                                                                                                - `GetStaticText` (single)
                                                                                                                                                                                                                                                                                                                                                                                                - `UpdateStaticText` (JSON, content only)

5.  Implement `handlers/configs.go`:

                                                                                                                                                                                                                                                                                                                                                                                                - `UpdateConfig` (JSON, key from URL param)

6.  Implement `handlers/visitor_messages.go`:

                                                                                                                                                                                                                                                                                                                                                                                                - `GetVisitorMessages` (paginated)
                                                                                                                                                                                                                                                                                                                                                                                                - `DeleteVisitorMessage` (204)

7.  Wire up admin routes in `router.go` (with JWT auth middleware)

### Phase 8: Migration Tools

1.  **cmd/migrate-db/main.go**:

                                                                                                                                                                                                                                                                                                                                                                                                - CLI flags: `--mysql-dsn`, `--postgres-dsn`, `--dry-run`
                                                                                                                                                                                                                                                                                                                                                                                                - Connect to both databases
                                                                                                                                                                                                                                                                                                                                                                                                - Migration order:

                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                1. Create schema if not exists (or run migrations first)
                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                2. Migrate `users` (set `password_reset_required=true`, leave `password` empty or placeholder)
                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                3. Migrate `projects`
                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                4. Migrate `project_images`
                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                5. Migrate `testimonials`
                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                6. Migrate `static_texts`
                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                7. Migrate `configurations`
                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                8. Migrate `visitor_messages`

                                                                                                                                                                                                                                                                                                                                                                                                - Preserve IDs (use `SET IDENTITY_INSERT` equivalent or direct INSERT with IDs)
                                                                                                                                                                                                                                                                                                                                                                                                - Reset sequences: `SELECT setval('users_id_seq', (SELECT MAX(id) FROM users));` (repeat for all tables)
                                                                                                                                                                                                                                                                                                                                                                                                - Validate: row counts match, FK integrity checks
                                                                                                                                                                                                                                                                                                                                                                                                - Output: migration report (rows migrated per table, errors)

2.  **cmd/migrate-files/main.go**:

                                                                                                                                                                                                                                                                                                                                                                                                - CLI flags: `--source-dir`, `--target-dir`, `--postgres-dsn`
                                                                                                                                                                                                                                                                                                                                                                                                - Connect to PostgreSQL
                                                                                                                                                                                                                                                                                                                                                                                                - Query `project_images.url` from DB
                                                                                                                                                                                                                                                                                                                                                                                                - Parse URLs:
                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                - If starts with `/storage/img/`, extract filename
                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                - If absolute external URL, skip
                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                - If relative local path, resolve to source dir
                                                                                                                                                                                                                                                                                                                                                                                                - Copy files to `target-dir/img/` preserving filename
                                                                                                                                                                                                                                                                                                                                                                                                - Validate: file exists after copy
                                                                                                                                                                                                                                                                                                                                                                                                - Output: report (copied count, missing files list, skipped external URLs)

### Phase 9: Docker Setup

1.  **Dockerfile** (multi-stage):

    ```dockerfile
    # Build stage
    FROM golang:1.21-alpine AS builder
    WORKDIR /app
    COPY go.mod go.sum ./
    RUN go mod download
    COPY . .
    RUN CGO_ENABLED=0 GOOS=linux go build -o /app/server ./cmd/server

    # Runtime stage
    FROM alpine:latest
    RUN apk --no-cache add ca-certificates
    WORKDIR /app
    COPY --from=builder /app/server .
    COPY --from=builder /app/migrations ./migrations
    EXPOSE 8080
    CMD ["./server"]
    ```

2.  **docker-compose.prod.yaml**:

    ```yaml
    version: "3.8"
    services:
      api:
        build: .
        ports:
          - "8080:8080"
        environment:
          - DATABASE_URL=postgres://user:pass@postgres:5432/elite_constructions?sslmode=disable
          - JWT_SECRET=${JWT_SECRET}
          - PORT=8080
          - STORAGE_PATH=/app/storage
        volumes:
          - storage_data:/app/storage/public
        depends_on:
          postgres:
            condition: service_healthy
        healthcheck:
          test:
            [
              "CMD",
              "wget",
              "--quiet",
              "--tries=1",
              "--spider",
              "http://localhost:8080/ping",
            ]
          interval: 30s
          timeout: 10s
          retries: 3

      postgres:
        image: postgres:15-alpine
        environment:
          - POSTGRES_USER=${POSTGRES_USER:-elite}
          - POSTGRES_PASSWORD=${POSTGRES_PASSWORD}
          - POSTGRES_DB=${POSTGRES_DB:-elite_constructions}
        volumes:
          - postgres_data:/var/lib/postgresql/data
        healthcheck:
          test: ["CMD-SHELL", "pg_isready -U ${POSTGRES_USER:-elite}"]
          interval: 10s
          timeout: 5s
          retries: 5

    volumes:
      postgres_data:
      storage_data:
    ```

3.  Update `cmd/server/main.go` to:

                                                                                                                                                                                                                                                                                                                                                                                                - Run migrations on startup (call `internal/db/migrations.Run()`)
                                                                                                                                                                                                                                                                                                                                                                                                - Serve static files: `router.Static("/storage/img", "./storage/public/img")`

### Phase 10: Testing & Documentation

1.  Create `README.md` with:

                                                                                                                                                                                                                                                                                                                                                                                                - Setup instructions (local dev)
                                                                                                                                                                                                                                                                                                                                                                                                - Docker production setup
                                                                                                                                                                                                                                                                                                                                                                                                - Migration tools usage (`migrate-db`, `migrate-files`)
                                                                                                                                                                                                                                                                                                                                                                                                - Environment variables reference
                                                                                                                                                                                                                                                                                                                                                                                                - API endpoint documentation (or link to OpenAPI/Swagger if added)

2.  Add basic integration tests (optional but recommended):

                                                                                                                                                                                                                                                                                                                                                                                                - Test login flow (normal + reset required)
                                                                                                                                                                                                                                                                                                                                                                                                - Test password reset completion
                                                                                                                                                                                                                                                                                                                                                                                                - Test project CRUD with file uploads
                                                                                                                                                                                                                                                                                                                                                                                                - Test pagination
                                                                                                                                                                                                                                                                                                                                                                                                - Test validation errors

## Migration Strategy Details

### Database Migration (MySQL → PostgreSQL)

**Order of Operations:**

1.  **Schema Creation**: Run PostgreSQL migrations first (creates empty tables with correct structure)

2.  **Data Migration** (preserve IDs):

                                                                                                                                                                                                                                                                                                                                                                                                - **Users**:
                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                - Copy: `id`, `name`, `email`, `email_verified_at`, `remember_token`, `created_at`, `updated_at`
                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                - Set: `password_reset_required = true`
                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                - Leave: `password = ''` (or use placeholder like `'MIGRATION_PENDING'`)
                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                - Do NOT copy bcrypt password (cannot verify, force reset)

                                                                                                                                                                                                                                                                                                                                                                                                - **Projects**: Direct copy (all fields map 1:1, `status` as SMALLINT)

                                                                                                                                                                                                                                                                                                                                                                                                - **Project Images**: Direct copy, preserve `url` paths

                                                                                                                                                                                                                                                                                                                                                                                                - **Testimonials**: Direct copy, `status` as VARCHAR

                                                                                                                                                                                                                                                                                                                                                                                                - **Static Texts**: Direct copy

                                                                                                                                                                                                                                                                                                                                                                                                - **Configurations**: Direct copy

                                                                                                                                                                                                                                                                                                                                                                                                - **Visitor Messages**: Direct copy

3.  **Sequence Reset** (after all inserts):

    ```sql
    SELECT setval('users_id_seq', (SELECT MAX(id) FROM users));
    SELECT setval('projects_id_seq', (SELECT MAX(id) FROM projects));
    -- ... repeat for all tables
    ```

4.  **Validation**:

                                                                                                                                                                                                                                                                                                                                                                                                - Row count comparison (MySQL vs PostgreSQL)
                                                                                                                                                                                                                                                                                                                                                                                                - FK integrity: `SELECT COUNT(*) FROM project_images WHERE project_id NOT IN (SELECT id FROM projects);` (should be 0)
                                                                                                                                                                                                                                                                                                                                                                                                - Check for NULL in required fields

**Encoding Considerations:**

- MySQL `utf8mb4` → PostgreSQL `UTF8` (default)
- Timestamps: MySQL `DATETIME` → PostgreSQL `TIMESTAMP` (both timezone-naive, store as-is)
- Booleans: MySQL `TINYINT(1)` → PostgreSQL `BOOLEAN` (convert 0/1 to false/true)

### File Migration Strategy

**Path Parsing Rules:**

1.  **Database URLs** (from `project_images.url`):

                                                                                                                                                                                                                                                                                                                                                                                                - Pattern: `/storage/img/filename.jpg` → extract `filename.jpg`
                                                                                                                                                                                                                                                                                                                                                                                                - Pattern: `https://example.com/image.jpg` → skip (external)
                                                                                                                                                                                                                                                                                                                                                                                                - Pattern: `../public/img/file.jpg` → resolve relative to source dir

2.  **File Copy Logic**:

                                                                                                                                                                                                                                                                                                                                                                                                - Source: `{source-dir}/img/filename.jpg` (or resolved relative path)
                                                                                                                                                                                                                                                                                                                                                                                                - Target: `{target-dir}/img/filename.jpg`
                                                                                                                                                                                                                                                                                                                                                                                                - Preserve filename exactly (no renaming)
                                                                                                                                                                                                                                                                                                                                                                                                - Create target directory if missing

3.  **Validation**:

                                                                                                                                                                                                                                                                                                                                                                                                - Check file exists in source before copy
                                                                                                                                                                                                                                                                                                                                                                                                - Verify file exists in target after copy (file size match)
                                                                                                                                                                                                                                                                                                                                                                                                - Report missing files (referenced in DB but not found in source)

4.  **Edge Cases**:

                                                                                                                                                                                                                                                                                                                                                                                                - Duplicate filenames: preserve as-is (DB references are unique by ID)
                                                                                                                                                                                                                                                                                                                                                                                                - Missing files: log warning, continue migration
                                                                                                                                                                                                                                                                                                                                                                                                - Invalid URLs: skip with warning

## Docker Deliverables Summary

**Dockerfile Approach:**

- Multi-stage build (smaller final image)
- Alpine-based runtime (security + size)
- Copy migrations directory for startup execution
- Non-root user (optional, add if security-critical)

**Compose Services:**

- `api`: Built from Dockerfile, exposes port 8080
- `postgres`: Official PostgreSQL 15 image
- Healthchecks on both services

**Volumes:**

- `postgres_data`: Persistent PostgreSQL data
- `storage_data`: Persistent uploaded images (`/app/storage/public`)

**Startup/Migrations Strategy:**

- Migrations run automatically in `cmd/server/main.go` on startup
- Alternative: separate `migrate` command (not chosen for simplicity)
- If migration fails, server exits with error (prevents running with bad schema)

**Static File Serving:**

- Gin static file middleware: `router.Static("/storage/img", "./storage/public/img")`
- Volume mount ensures files persist across container restarts

## .env.example Contents

```env
# Database
DATABASE_URL=postgres://user:password@localhost:5432/elite_constructions?sslmode=disable

# JWT
JWT_SECRET=your-secret-key-min-32-chars-change-in-production

# Server
PORT=8080

# Storage
STORAGE_PATH=./storage
```

**Notes:**

- Minimal and self-contained
- No unused sections (removed AWS/S3, queue configs, etc.)
- All required for basic operation

## Acceptance Checklist

### Manual Verification

**Setup & Configuration:**

- [ ] `.env.example` is minimal and complete
- [ ] Server starts and runs migrations successfully
- [ ] Docker compose builds and runs (API + Postgres healthy)

**Authentication:**

- [ ] `POST /api/login` with Argon2id password returns JWT
- [ ] `POST /api/login` with user having no Argon2id returns `{"reset_required": true, "reset_token": "..."}`
- [ ] `POST /api/password-reset/complete` with valid token sets new password
- [ ] After reset, normal login works with new password
- [ ] `GET /api/me` requires valid JWT
- [ ] Invalid JWT returns 401

**Public Endpoints:**

- [ ] `GET /ping` returns `{"message": "Welcome to API 1.0"}`
- [ ] `GET /api/pub/projects` returns all projects
- [ ] `GET /api/pub/projects/highlighted` returns only highlighted
- [ ] `GET /api/pub/async/projects/page?page=1` returns 3 per page
- [ ] `GET /api/pub/projects/:id` returns single project with images
- [ ] `GET /api/pub/testimonials` returns only status='ready'
- [ ] `POST /api/pub/testimonials` creates with status='pending'
- [ ] `GET /api/pub/static-texts` returns all
- [ ] `GET /api/pub/configs` returns all
- [ ] `POST /api/pub/visitor-messages` creates message

**Admin Projects:**

- [ ] `GET /api/projects?page=1` returns 10 per page (requires auth)
- [ ] `POST /api/projects` (multipart) creates project with images, generates blurhash
- [ ] `PUT /api/projects/:id` updates, handles existing image IDs + new files
- [ ] `PUT /api/projects/:id` deletes removed images from filesystem
- [ ] `PUT /api/projects/:id/highlight/toggle` toggles boolean
- [ ] `DELETE /api/projects/:id` cascade deletes images
- [ ] Images accessible at `/storage/img/filename.jpg`

**Admin Testimonials:**

- [ ] `GET /api/testimonials?page=1` paginated (requires auth)
- [ ] `POST /api/testimonials` creates with status
- [ ] `PUT /api/testimonials/:id` updates
- [ ] `DELETE /api/testimonials/:id` returns 400 if only 1 remains, else 204

**Admin Users:**

- [ ] `GET /api/users?page=1` paginated (requires auth)
- [ ] `POST /api/users` creates with Argon2id hashed password
- [ ] `PUT /api/users/:id` updates name/email only
- [ ] `DELETE /api/users/:id` returns 400 if only 1 remains, else 204

**Admin Static Texts:**

- [ ] `GET /api/static-texts?page=1` paginated (requires auth)
- [ ] `PUT /api/static-texts/:id` updates content only

**Admin Configurations:**

- [ ] `PUT /api/configs/:key` updates value

**Admin Visitor Messages:**

- [ ] `GET /api/visitor-messages?page=1` paginated (requires auth)
- [ ] `DELETE /api/visitor-messages/:id` returns 204

**Error Handling:**

- [ ] All errors return `{"error": "message", "details": ...}` format
- [ ] Validation errors return 400 with details
- [ ] 404 for not found resources
- [ ] 401 for unauthorized
- [ ] 500 for server errors (with error message in dev, generic in prod)

**Pagination:**

- [ ] All paginated endpoints return consistent format (e.g., `{"data": [...], "page": 1, "per_page": 10, "total": 100}`)

**Timestamps:**

- [ ] All timestamps in UTC RFC3339 format

### Migration Tools

**Database Migration:**

- [ ] `migrate-db --mysql-dsn=... --postgres-dsn=...` migrates all tables
- [ ] IDs preserved correctly
- [ ] Sequences reset to MAX(id)
- [ ] Users have `password_reset_required=true`, empty password
- [ ] Row counts match MySQL
- [ ] FK integrity validated (no orphaned records)

**File Migration:**

- [ ] `migrate-files --source-dir=... --target-dir=... --postgres-dsn=...` copies files
- [ ] Files copied to correct location (`./storage/public/img/`)
- [ ] External URLs skipped
- [ ] Missing files reported in output
- [ ] Report shows copied count and missing list

### Docker Production

- [ ] `docker-compose -f docker-compose.prod.yaml up` starts both services
- [ ] Healthchecks pass
- [ ] Migrations run on API startup
- [ ] Static files served at `/storage/img/*`
- [ ] Volumes persist data across restarts
- [ ] API accessible on port 8080

### Key Automated Tests (if implemented)

- [ ] Login with Argon2id password → JWT returned
- [ ] Login with reset required → reset token returned
- [ ] Password reset completion → new password works
- [ ] Project creation with file upload → blurhash generated
- [ ] Project update with mixed existing IDs + new files → correct behavior
- [ ] Testimonial deletion with only 1 remaining → 400 error
- [ ] User deletion with only 1 remaining → 400 error
- [ ] Pagination returns correct page size
- [ ] JWT middleware rejects invalid tokens
