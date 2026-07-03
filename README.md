# Nivas API Server

Go backend for the Nivas PG management app. Built with Gin, PostgreSQL, and a layered architecture.

## Structure

```
server/
├── cmd/api/              # Application entrypoint
├── internal/
│   ├── auth/             # JWT token service
│   ├── config/           # Environment configuration
│   ├── database/         # PostgreSQL pool
│   ├── domain/           # Entities and enums
│   ├── handler/          # HTTP handlers (Gin)
│   ├── middleware/       # Auth, CORS, logging, recovery
│   ├── repository/       # Data access interfaces
│   │   └── postgres/     # PostgreSQL implementations
│   ├── router/           # Route registration
│   └── service/          # Business logic
├── pkg/
│   ├── apperror/         # Typed errors with HTTP codes
│   ├── logger/           # Structured slog logger
│   └── response/         # Consistent JSON responses
└── migrations/           # SQL schema
```

## Quick start

```bash
# Start PostgreSQL
make docker-up

# Configure env
cp .env.example .env

# Run migrations
export DATABASE_URL=postgres://postgres:postgres@localhost:5432/nivas?sslmode=disable
make migrate

# Create the first owner account (hashes password in Go — same as login)
make bootstrap

# Run server
make run
```

API available at `http://localhost:8080/api/v1`.

Health check: `GET /health`

## Libraries

| Package | Purpose |
|---------|---------|
| `pkg/logger` | Structured JSON/text logging via `slog` |
| `pkg/apperror` | Typed errors with codes, HTTP status, details |
| `pkg/response` | Uniform `{ error: { code, message } }` responses |

## Logging

Set in `.env`:

```bash
LOG_LEVEL=debug   # debug | info | warn | error
LOG_FORMAT=text   # text for local dev, json for production
```

Logged automatically:

- HTTP requests with `request_id`, status, duration, `user_id` / `organization_id` when authenticated
- Handler errors with `error_code` (passwords are never logged)
- Auth login success/failure, staff invite, tenant/room/kitchen creates
- Database unique violations and Postgres errors

## Auth

- **Staff:** organization ID + email/password → JWT (`Authorization: Bearer <token>`)
- **Tenant:** organization ID + email/password → JWT (read-only access)
- JWT claims include `organization_id`, `type` (staff|tenant), and `user_id`
- Staff accounts are invite-only; no public registration

### Dev seed credentials

After `make migrate` and `make bootstrap`:

| Field | Value |
|-------|-------|
| Organization ID | `00000000-0000-0000-0000-000000000001` |
| Staff email | `owner@nivas.local` (override with `BOOTSTRAP_EMAIL`) |
| Staff password | `admin123` (override with `BOOTSTRAP_PASSWORD`) |

Passwords are hashed by the bootstrap command using the same bcrypt logic as login — not stored in SQL.
