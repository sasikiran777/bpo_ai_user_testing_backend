# Backend Architecture (Go + Gin + Postgres + Bun + LangChainGo)

## Goals

- Support the current UI flows: login → dashboard → single-attempt test flow (instructions → writing → reading → speaking) → results
- Enforce one attempt per user per test type (future retake policies can be added later)
- Store submissions, proctoring events, grading jobs, and final results
- Run grading asynchronously using LangChainGo

## High-Level Components

- API service (Gin): authentication, attempt lifecycle, section submissions, results retrieval
- Database (Postgres + Bun ORM): source of truth for attempts, submissions, proctoring, grading jobs, results
- Grading worker (Go process): pulls jobs from Postgres, runs LangChainGo chains, writes results
- Audio storage (recommended): store speaking audio in S3-compatible storage (S3/MinIO); store only metadata in Postgres

## Repository Layout

- cmd/api: Gin HTTP server entrypoint
- cmd/worker: grading worker entrypoint
- internal/config: env/config loading
- internal/log: structured logger initialization
- internal/db: Bun initialization and migrations wiring
- internal/domain: domain entities and enums
- internal/repos: DB access layer (Bun queries)
- internal/services: business logic (attempt state machine, submissions, results, grading enqueue)
- internal/http: router setup and middleware
- internal/http/handlers: route handlers (thin; validate/parse and call services)
- internal/grading: LangChainGo chains, prompts, scoring rules
- internal/storage: audio storage interface (S3/local) and implementations
- migrations: SQL migrations
- docs: architecture and API documentation

## Core Domain Concepts

- User: authenticated user record
- Test type: `english` now; designed to support more later (e.g., `agentic_ai`)
- Attempt: single-attempt entity per user+testType with a state machine
  - `not_started` → `in_progress` → `grading` → `completed`
  - `failed` (terminal)
- Submissions: writing, reading, speaking
- Proctoring events: focus/visibility changes, captured with timestamps and metadata
- Results: final scores and breakdown
- Grading jobs: async job rows processed by worker

## Database Schema (Suggested)

### users

- id (uuid pk)
- email (unique)
- password_hash
- created_at

### test_attempts

- id (uuid pk)
- user_id (fk)
- test_type (text)
- status (text)
- current_section (text nullable)
- started_at (timestamptz nullable)
- submitted_at (timestamptz nullable)
- failed_reason (text nullable)
- created_at, updated_at (timestamptz)

Constraints

- unique(user_id, test_type) to enforce single attempt

### writing_submissions

- attempt_id (pk/fk)
- about_me (text)
- location (text)
- experience (text)
- roles (text)
- responsibilities (text)
- other (text)
- started_at, submitted_at (timestamptz)

### reading_submissions

- attempt_id (pk/fk)
- reading_set_id (text)
- answers (jsonb)
- started_at, submitted_at (timestamptz)

### speaking_submissions

- attempt_id (pk/fk)
- audio_object_key (text) or audio_url (text)
- duration_sec (int)
- started_at, submitted_at (timestamptz)

### proctoring_events

- id (uuid pk)
- attempt_id (fk)
- section (text)
- event_type (text)
- event_at (timestamptz)
- meta (jsonb)

### grading_jobs

- id (uuid pk)
- attempt_id (unique fk)
- status (text: queued|running|done|failed)
- locked_at (timestamptz nullable)
- locked_by (text nullable)
- error (text nullable)
- created_at, updated_at (timestamptz)

### test_results

- attempt_id (pk/fk)
- status (text: grading|completed|failed)
- overall_score (int)
- overall_max (int)
- writing_score (int)
- writing_max (int)
- reading_score (int)
- reading_max (int)
- speaking_score (int)
- speaking_max (int)
- details (jsonb)
- updated_at (timestamptz)

## API Endpoints (Suggested)

### Auth

- POST /auth/login

### Tests Catalog

- GET /tests

### Attempt Lifecycle

- POST /tests/:testType/attempt (get or create attempt)
- POST /tests/:testType/start (transition `not_started` → `in_progress`)
- POST /tests/:testType/fail (mark failed; used for leave/refresh handling)

### Content

- GET /tests/:testType/instructions
- GET /tests/:testType/reading-set
- GET /tests/:testType/speaking-topic

### Section Submissions

- POST /tests/:testType/writing/submit
- POST /tests/:testType/reading/submit
- POST /tests/:testType/speaking/submit (multipart/form-data for audio)

### Finalize + Results

- POST /tests/:testType/submit (transition to grading + enqueue job)
- GET /tests/:testType/results

## Server-Side Rules

- Block submissions when attempt is `completed` or `failed`
- Enforce section order: writing → reading → speaking; once submitted, cannot resubmit
- Apply time-limit validation using server time (accept client timestamps but validate against server-side rules)
- Always scope attempt access by `user_id` in all queries

## Middleware (Gin)

- Auth (JWT) middleware: validate token, set user identity in context
- Request ID: `X-Request-Id` propagation
- Structured logging: method/path/status/latency/requestId/userId (never log secrets)
- Recovery: panic recovery and consistent JSON errors
- CORS: allow the UI origin(s)

## Grading Worker (LangChainGo)

### Enqueue

On finalize (POST /submit):

- transaction:
  - set attempt status to `grading`
  - insert `grading_jobs(status=queued)`

### Process

- lock a queued job atomically (update with a predicate, returning the locked row)
- load attempt + submissions
- run grading chains:
  - writing rubric scoring
  - reading correctness scoring
  - speaking rubric scoring (can start as fixed score and evolve later)
- write `test_results`
- mark job `done`
- set attempt `completed`

### Failure Handling

- if grading fails: mark job `failed` with error, and decide whether attempt stays `grading` (retry) or becomes `failed` (policy-driven)

## Audio Storage

- Keep Postgres free of large blobs
- Store speaking audio in object storage (S3/MinIO)
- Store `audio_object_key` (and optionally duration/codec) in `speaking_submissions`

## Logging

- JSON logs for both api and worker
- Standard fields: ts, level, service, requestId, userId, attemptId, testType, route, status, latencyMs
- Never log JWTs, passwords, or raw audio URLs containing signed query strings

## Configuration

- DATABASE_URL
- JWT_SECRET
- CORS_ORIGINS
- LOG_LEVEL
- STORAGE_DRIVER (s3|local)
- S3_ENDPOINT, S3_BUCKET, S3_ACCESS_KEY, S3_SECRET_KEY
- LLM_PROVIDER, LLM_API_KEY, LLM_MODEL

## Suggested Bootstrap Order

- Initialize go module and repo layout
- Add migrations + Bun connection
- Implement auth (login) + JWT middleware
- Implement attempts with unique(user_id, test_type)
- Implement section submissions + proctoring events
- Implement finalize → grading job enqueue
- Implement worker → LangChainGo grading → results
- Add API documentation (OpenAPI or docs/api.md)

