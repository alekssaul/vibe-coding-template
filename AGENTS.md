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
│   └── store/            # SQLite data layer (sqlc + golang-migrate)
│       ├── db/           # Generated sqlc code
│       ├── migrations/   # golang-migrate .sql up/down files (embedded)
│       └── queries/      # sqlc .sql query files
├── flutter_app/          # Flutter frontend
│   ├── lib/
│   │   ├── models/       # Data models (from Go JSON)
│   │   ├── providers/    # Riverpod providers (using riverpod_generator)
│   │   └── screens/      # UI screens
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
| `make db-generate` | Generate sqlc Go code from SQL queries |
| `make migrate-add NAME=your_migration` | Create new golang-migrate migration files |

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

- State management: Riverpod (`flutter_riverpod`) with `riverpod_generator`. Strictly use `@riverpod` annotations.
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

### 1. Database (`sqlc` + `golang-migrate`)
1. Create a new SQL migration in `internal/store/migrations/`: `make migrate-add NAME=your_table`.
2. Define the schema `CREATE TABLE` / `DROP TABLE` in the `.up.sql` and `.down.sql` files.
3. Create a query file in `internal/store/queries/your_model.sql` with `-- name: YourQuery :one` etc.
4. Run `make db-generate` to generate the type-safe Go bindings in `internal/store/db/`.
5. Create `internal/store/your_model.go` to wrap the `s.queries` calls and map to `model.YourModel`.

### 2. Model (`internal/model/`)
1. Add model to `internal/model/<resource>.go`

### 3. Handlers (`internal/handler/`)
1. Add handlers to `internal/handler/<resource>.go` with swaggo annotations

### 4. Routing (`cmd/api/`)
1. Register routes in `cmd/api/main.go`

### 5. Flutter UI (`flutter_app/`)
1. Create a `lib/models/` file matching the Go JSON outputs using `.fromJson()`.
2. Create a `lib/providers/` file utilizing `@riverpod` and `riverpod_generator`. Run `dart run build_runner build -d` to generate the `.g.dart` file. Use `ApiClient.instance` within the provider instead of dependency injection. Handle mutation errors by catching them and using `SnackBarUtil.showError(e.toString())`.
3. Create the UI in `lib/screens/`. Hook up state with `ref.watch(yourProvider)`. Leverage `ErrorStateWidget` for the `error` state of `AsyncValue`.
4. Link the new screen in `lib/router/app_router.dart`.

### 6. Verification
1. Run `make verify` before closing the task
