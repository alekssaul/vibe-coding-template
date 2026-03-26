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
| `make scaffold RESOURCE=name FIELDS="f:type,..."` | Generate full-stack CRUD (9 files) + auto-inject routes |
| `make seed` | Seed database with dummy data and dev API keys |

> **Scaffold rule**: Never manually write Go model/handler/store/SQL/Flutter model/provider/screen for a new CRUD resource. Always run `make scaffold` first. Routes are auto-injected into both `cmd/api/main.go` and `flutter_app/lib/router/app_router.dart`. Only deviate when the resource has non-standard patterns.

## Search & Pagination

All `List*` endpoints support `?limit=`, `?offset=`, and `?search=` query params. Search matches against `name` and `description` columns.

## Deployment

```bash
docker compose up --build   # Local Docker
```

CI runs automatically on push/PR to `main` via `.github/workflows/ci.yml`.

## Setup Screen

The Flutter app gates on API key presence. On first launch, users see a Setup screen where they paste their key. The key is validated against `/health` and stored in `shared_preferences`.

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

### Quick Path (Recommended — use the scaffold generator)

```bash
make scaffold RESOURCE=product FIELDS="price:float,active:bool"
make db-generate
cd flutter_app && dart run build_runner build -d && cd ..
```

Routes are **automatically injected** into `cmd/api/main.go` at the `// scaffold:routes` marker. The only remaining manual step is adding the new screen to `flutter_app/lib/router/app_router.dart`.

Supported field types: `string`, `int`, `float`, `bool`. Every resource automatically gets `name` and `description` fields plus `id`, `created_at`, `updated_at`.

### Manual Path (for custom patterns)

1. **Database:** Create migration (`make migrate-add NAME=...`), write `.sql` queries in `internal/store/queries/`, run `make db-generate`.
2. **Model:** Add struct to `internal/model/<resource>.go` with `validate` tags.
3. **Store:** Wrap sqlc calls in `internal/store/<resource>.go`.
4. **Handler:** Add CRUD handlers in `internal/handler/<resource>.go` with swaggo annotations.
5. **Routes:** Register in `cmd/api/main.go`.
6. **Flutter:** Create model, `@riverpod` provider (use `ApiClient.instance`), and screen. Run `dart run build_runner build -d`.
7. **Verify:** Run `make verify`.
