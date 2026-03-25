# AGENTS.md

This file provides context for AI agents (Claude, Gemini, etc.) working in this repository.

## Project Overview

A CRUD application template with:

- **Backend**: Go 1.25 · `net/http` · `modernc.org/sqlite` · JSON structured logging (`log/slog`)
- **Frontend**: Flutter 3.41.4 · Riverpod · `flutter_dotenv`
- **Auth**: API key via `X-API-Key` header (read/write permissions)
- **OpenAPI**: Auto-generated from swaggo annotations (`make docs`)

## Repository Structure

```
├── cmd/api/              # Go entrypoint (main.go) + boot tests
├── internal/
│   ├── config/           # Env-based config loading
│   ├── handler/          # HTTP handlers (health, items, apikeys)
│   ├── middleware/        # CORS, RequestID, API key auth
│   ├── model/            # Domain types (Item, APIKey, etc.)
│   ├── response/         # HTTP response helpers (WriteJSON, WriteError, WriteList)
│   └── store/            # SQLite data layer (items, apikeys)
│       └── migrations/   # golang-migrate .sql up/down files (embedded)
├── flutter_app/          # Flutter frontend (Riverpod, flutter_dotenv, go_router)
├── docs/                 # Generated OpenAPI/Swagger (gitignored — run `make docs`)
├── Makefile
├── .env.example
└── AGENTS.md
```

## ⚠️ Mandatory Build Verification

**After ANY Go code change:**
1. Run `make verify-go` — MUST succeed before declaring work done
2. Run `make test` — ALL tests must pass

**After ANY Flutter code change:**
1. Run `make flutter-analyze` — zero errors required
2. Run `make verify-flutter` — build must succeed

Do not declare a task complete until these pass.

## Key Make Targets

| Command | Description |
|---|---|
| `make build` | Build Go binary with git SHA |
| `make dev` | Run backend with hot-reload (`air`) |
| `make test` | Run Go tests |
| `make verify` | Full verification (Go + tests + Flutter analyze) |
| `make docs` | Regenerate OpenAPI docs |
| `make flutter-analyze` | Lint Flutter/Dart code |
| `make install-tools` | Install swag + golangci-lint |
| `make init PROJECT=myapp` | Rename template module |

## API Conventions

**All responses are JSON.** Standard shapes:

```json
// Success (single)
{ "data": { ... } }

// Success (list)
{ "data": [...], "total": 42, "limit": 20, "offset": 0 }

// Error
{ "error": "human readable", "code": "SNAKE_CASE_CODE" }
```

**Authentication**: Pass `X-API-Key: <key>` on all `/v1/` routes.
- `GET` routes require `read` permission (write keys also work)
- `POST`, `PUT`, `DELETE` require `write` permission

**Pagination**: `?limit=20&offset=0` on list endpoints. Max limit is 100.

## Go Conventions

- Logger: `log/slog` with `JSONHandler` — always use `logger.InfoContext(r.Context(), ...)` to propagate request ID
- Errors: return `sql.ErrNoRows` from store → handler maps to 404
- Error codes: `UNAUTHORIZED`, `FORBIDDEN`, `NOT_FOUND`, `BAD_REQUEST`, `VALIDATION_ERROR`, `INTERNAL_ERROR`
- Module: `github.com/alekssaul/template` (rename with `make init PROJECT=myapp`)

## Flutter Conventions

- State management: Riverpod (`flutter_riverpod`)
- Routing: `go_router` (configured in `lib/router/app_router.dart`)
- Design System: Use `AppTheme` (in `lib/theme/app_theme.dart`) for standard spacing and colors.
- Forms: Use `flutter_form_builder` for complex inputs
- API base URL: loaded from `.env` via `flutter_dotenv` as `API_BASE_URL`
- API key: loaded from `.env` as `API_KEY`
- All API calls via `lib/services/api_client.dart`

## Environment Variables

See `.env.example`. Copy to `.env` before running.

Key vars: `API_PORT` (default 8080), `DB_PATH` (default data.db), `CORS_ORIGINS` (default `*`), `ENV`

## Adding a New Resource

1. Add model to `internal/model/<resource>.go`
2. Generate migration: `make migrate-add NAME=create_<resource>` and write the `.sql` files 
3. Add store methods to `internal/store/<resource>.go`
4. Add handlers to `internal/handler/<resource>.go` with swaggo annotations
4. Register routes in `cmd/api/main.go`
5. Add Flutter model, provider, and screen
6. Run `make verify` before closing the task
