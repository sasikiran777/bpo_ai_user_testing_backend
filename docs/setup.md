# Backend Setup Notes

This repository contains a minimal, production-friendly Go API skeleton intended to match the architecture direction in [architecture-backend.md](file:///Users/sasikiran/Web%20Projects/BPO%20Admin/AI%20User%20Testing/backend/architecture-backend.md).

It provides:

- A Gin HTTP server entrypoint (`cmd/api`)
- `.env`-based configuration loading (with sensible defaults)
- Structured JSON logging using Go `slog`
- Request ID and request logging middleware
- A `GET /health` endpoint
- A versioned API group under `/api/v1`
- A module-style auth package under `internal/modules/auth`
- Shared response/validation helpers under `internal/shared`
- Bun + Postgres migrations runner (`cmd/migrate`)

## Repository Layout (Current)

- `cmd/api`: API binary entrypoint
- `cmd/migrate`: migrations runner entrypoint (bun/migrate)
- `internal/config`: config loader (`.env` + env vars)
- `internal/log`: slog logger initialization
- `internal/db`: Bun Postgres connection
- `internal/http`: router setup
- `internal/http/middleware`: middleware (request-id, request-logging)
- `internal/shared`: shared helpers/responses/models/validator helpers
- `internal/modules/auth`: auth module (dto/handler/service/middleware/routes/validator)
- `internal/modules/users`: users module skeleton (models + service/repo skeleton)
- `migrations`: Go migrations (bun/migrate registry + migration files)
- `docs`: developer documentation (this folder)

## Quick Start (Local)

1. Create local env file:

```bash
cp .env.example .env
```

2. Run the API:

```bash
go run ./cmd/api
```

3. Health check:

```bash
curl http://localhost:8080/health
```

## Hot Reload (Air)

Air is configured via [.air.toml](file:///Users/sasikiran/Web%20Projects/BPO%20Admin/AI%20User%20Testing/backend/.air.toml) to rebuild and restart the API on code changes.

If you are using Devbox:

```bash
devbox run dev
```

If you installed Air globally:

```bash
air
```

## Quick Start (Docker + Postgres)

1. Create local env file:

```bash
cp .env.example .env
```

2. Start services:

```bash
docker compose up --build
```

- API: `http://localhost:${PORT}` (defaults to 8080)
- Postgres: `localhost:5432`

## Migrations

Migrations use Bun migrate and are executed via `cmd/migrate`.
Table creation uses Bun schema builder (`db.NewCreateTable().Model(...).Exec(ctx)`), and the migration enables `pgcrypto` for UUID defaults (`gen_random_uuid()`).

Commands:

```bash
go run ./cmd/migrate up
go run ./cmd/migrate down
go run ./cmd/migrate status
```

Current tables created by migrations:

- `users`

## Configuration

Config is loaded from (in order):

- `.env` (if present)
- process environment variables
- defaults in code (if not set)

File: [internal/config/config.go](file:///Users/sasikiran/Web%20Projects/BPO%20Admin/AI%20User%20Testing/backend/internal/config/config.go)

### Required / Common Variables

- `ENV`
  - Values: `development` (default), `dev`, `local`, `production`
  - Used to toggle Gin release mode and logger behavior (source locations in dev).
- `PORT`
  - Default: `8080`
  - API bind address is `:${PORT}` when running locally via `go run`.
- `LOG_LEVEL`
  - Default: `info`
  - Values: `debug`, `info`, `warn`, `error`
- `DATABASE_URL`
  - Postgres DSN used by both `cmd/migrate` and the API.

### Postgres Variables (for Docker Compose)

These are currently only used by `docker-compose.yml` and `.env.example` to bootstrap the database container:

- `POSTGRES_DB`
- `POSTGRES_USER`
- `POSTGRES_PASSWORD`

## API Entrypoint and Graceful Shutdown

File: [cmd/api/main.go](file:///Users/sasikiran/Web%20Projects/BPO%20Admin/AI%20User%20Testing/backend/cmd/api/main.go)

Main flow:

- `config.Load()` reads `.env` + env vars into a small `Config` struct.
- `log.New(cfg)` creates a JSON logger (Go `slog`).
- `http.NewRouter(logger, db, &cfg)` constructs the Gin engine and registers middleware and routes.
- An `http.Server` is started in a goroutine.
- A signal-aware context listens for `SIGINT`/`SIGTERM` to trigger graceful shutdown.
- On shutdown, `Server.Shutdown()` is called with a 10s timeout to allow in-flight requests to finish.

## Router and Middleware

File: [internal/http/router.go](file:///Users/sasikiran/Web%20Projects/BPO%20Admin/AI%20User%20Testing/backend/internal/http/router.go)

Installed middleware (in order):

- `gin.Recovery()` to prevent panics from crashing the process and to return consistent 500 responses.
- `RequestID()` generates/propagates `X-Request-Id`.
  - File: [request_id.go](file:///Users/sasikiran/Web%20Projects/BPO%20Admin/AI%20User%20Testing/backend/internal/http/middleware/request_id.go)
- `RequestLogger(logger)` logs one structured record per request after it completes.
  - File: [request_logger.go](file:///Users/sasikiran/Web%20Projects/BPO%20Admin/AI%20User%20Testing/backend/internal/http/middleware/request_logger.go)

Routes:

- `GET /health` returns `{ "ok": true }`.
- `POST /api/v1/auth/login` authenticates and returns a JWT.
- `POST /api/v1/auth/register` exists but is currently a stub (service not implemented yet).

Router notes:

- Routes are grouped under `/api/v1`.
- The router creates a public group and a protected group; JWT middleware is applied only to the protected group (for future protected endpoints).
- Auth routes are registered by [auth_module.go](file:///Users/sasikiran/Web%20Projects/BPO%20Admin/AI%20User%20Testing/backend/internal/modules/auth/bootstrap/auth_module.go).

## Logging

File: [internal/log/log.go](file:///Users/sasikiran/Web%20Projects/BPO%20Admin/AI%20User%20Testing/backend/internal/log/log.go)

- Uses `slog.NewJSONHandler(os.Stdout, ...)` for JSON logs (container-friendly).
- Adds base fields to every log line:
  - `service=api`
  - `env=<ENV>`

Request logs include:

- `requestId` (propagated via `X-Request-Id`)
- `userId` (when available, e.g. for authenticated routes)
- `method`, `path`, `status`, `latencyMs`

## Auth (JWT)

This repo includes a minimal JWT setup intended to unblock development early. It is currently DB-backed (login checks `users.email` + `users.password`).

Implementation lives under: [internal/modules/auth](file:///Users/sasikiran/Web%20Projects/BPO%20Admin/AI%20User%20Testing/backend/internal/modules/auth)

Environment variables:

- `JWT_SECRET` (required)
- `JWT_TTL_MINUTES` (default: 1440)

Endpoints:

- `POST /api/v1/auth/login`
  - Body: `{ "email": "...", "password": "..." }`
  - Response: `{ token, firstName }`

- `POST /api/v1/auth/register`
  - Body: `{ "first_name": "...", "last_name": "...", "phone": "...", "email": "...", "password": "...", "total_exp_months": 0, "skills": ["..."], "past_job_title": "...", "company": "..." }`
  - Status: currently returns success with empty data until implemented

Validation pattern:

- Validators are Gin middleware that bind and validate the request body using `binding` tags in DTOs.
- On validation errors, they return a consistent 422 response via shared helpers.
- Validated payloads are stored in Gin context and accessed in handlers using the shared payload helper.

## Users (Module + Tables)

Users are introduced as a module skeleton and a DB migration.

Module folder:

- [internal/modules/users](file:///Users/sasikiran/Web%20Projects/BPO%20Admin/AI%20User%20Testing/backend/internal/modules/users)

Models:

- [user.go](file:///Users/sasikiran/Web%20Projects/BPO%20Admin/AI%20User%20Testing/backend/internal/modules/users/model/user.go)
- [user_profile.go](file:///Users/sasikiran/Web%20Projects/BPO%20Admin/AI%20User%20Testing/backend/internal/modules/users/model/user_profile.go)

Notes:

- `skills` is stored as `jsonb` (JSON array of strings).
- Experience is stored as `total_exp_months` (int).
- The API does not expose user endpoints yet (routes are intentionally empty for now).

## Devbox

File: [devbox.json](file:///Users/sasikiran/Web%20Projects/BPO%20Admin/AI%20User%20Testing/backend/devbox.json)

Devbox is used to make the development toolchain consistent across machines. It declares:

- Go toolchain
- `gopls`
- Go tools
- `golangci-lint`
- `air` (hot reload)

Usage depends on your Devbox installation. Common pattern:

- `devbox shell` to enter a shell with the declared tools available.
- `devbox run dev` to run the API with hot reload.

## Docker and Compose

Files:

- [Dockerfile](file:///Users/sasikiran/Web%20Projects/BPO%20Admin/AI%20User%20Testing/backend/Dockerfile)
- [docker-compose.yml](file:///Users/sasikiran/Web%20Projects/BPO%20Admin/AI%20User%20Testing/backend/docker-compose.yml)
- [.dockerignore](file:///Users/sasikiran/Web%20Projects/BPO%20Admin/AI%20User%20Testing/backend/.dockerignore)

Dockerfile:

- Multi-stage build: compiles the API binary then copies it into a small Alpine runtime image.
- Runs as a non-root user.

docker-compose:

- `db`: Postgres container with a persisted volume.
- `api`: builds the API image and injects environment variables from `.env`.
  - Overrides `DATABASE_URL` to point to `db` inside the compose network.

## Next Steps (Planned)

As implementation grows, expected additions (per architecture doc):

- `internal/domain`: entities/enums
- `internal/repos`: data access
- `internal/services`: business logic (attempt lifecycle, submissions, results)
- `internal/http/handlers`: route handlers (thin)
- `cmd/worker`: grading worker
