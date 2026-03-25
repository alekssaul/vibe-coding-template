# Template

A production-ready CRUD application template using **Go** (backend API) + **Flutter** (frontend) + **SQLite** (database).

## Stack

| Layer | Technology |
|---|---|
| API | Go 1.25 · `net/http` · JSON logs (`log/slog`) |
| Database | SQLite via `modernc.org/sqlite` with `golang-migrate` embedded migrations |
| Auth | API key (`X-API-Key` header, read/write permissions) |
| Frontend | Flutter 3.41.4 · Riverpod · `go_router` · `flutter_dotenv` |
| OpenAPI | swaggo auto-generated from annotations |

## Quick Start

```bash
# 1. Clone and copy env
cp .env.example .env

# 2. Install Go tools
make install-tools

# 3. Generate OpenAPI docs
make docs

# 4. Run backend with hot-reload (prints default API key on first run)
make dev

# 5. In another terminal — run Flutter web
make flutter-run-web
```

## Initial Setup (rename from template)

```bash
make init PROJECT=myapp
```

This replaces `alekssaul/template` → `alekssaul/myapp` across all Go files and re-runs `go mod tidy`.

## Environment Variables

Copy `.env.example` → `.env` and edit:

| Variable | Default | Description |
|---|---|---|
| `API_PORT` | `8080` | Port the Go API listens on |
| `DB_PATH` | `data.db` | SQLite database file path |
| `ENV` | `development` | Runtime environment label |
| `CORS_ORIGINS` | `*` | Comma-separated allowed CORS origins |

Other environment templates exist: `.env.staging` and `.env.production`. They are safe to commit and copy over to `.env` depending on the deployed environment.

## API Reference

### Authentication

Pass `X-API-Key: <key>` on all `/v1/` requests.

- **Read key**: access to `GET` endpoints
- **Write key**: access to all endpoints

On first run, a default **write** key is printed to stdout — save it immediately.

### Endpoints

| Method | Path | Auth | Description |
|---|---|---|---|
| `GET` | `/health` | None | Health check + version |
| `GET` | `/docs/` | None | Swagger UI |
| `GET` | `/v1/items` | Read | List items (paginated) |
| `GET` | `/v1/items/{id}` | Read | Get item by ID |
| `POST` | `/v1/items` | Write | Create item |
| `PUT` | `/v1/items/{id}` | Write | Update item |
| `DELETE` | `/v1/items/{id}` | Write | Delete item |
| `GET` | `/v1/keys` | Write | List API keys |
| `POST` | `/v1/keys` | Write | Create API key |
| `DELETE` | `/v1/keys/{id}` | Write | Delete API key |

### Pagination

List endpoints accept `?limit=20&offset=0` (max limit: 100).

### Response Format

```json
// Single item
{ "data": { "id": 1, "name": "...", "description": "..." } }

// List
{ "data": [...], "total": 42, "limit": 20, "offset": 0 }

// Error
{ "error": "item not found", "code": "NOT_FOUND" }
```

## Make Targets

```bash
make build             # Build Go binary (embeds git SHA + build time)
make dev               # Hot-reload backend via air
make migrate-add NAME=foo # Creates a new golang-migrate .sql up/down template
make test              # Run Go tests
make lint              # Run golangci-lint
make fmt               # Format Go code
make docs              # Regenerate OpenAPI docs
make verify            # Full check: Go build + tests + Flutter analyze
make flutter-run-web   # Run Flutter app in Chrome
make flutter-build-web # Build Flutter for web
make install-tools     # Install swag + golangci-lint
make init PROJECT=foo  # Rename module from template to foo
```

## Project Structure

```
├── cmd/api/              # main.go + tests
├── internal/
│   ├── config/           # Config from env
│   ├── handler/          # HTTP handlers
│   ├── middleware/        # CORS, RequestID, API key auth
│   ├── model/            # Domain types
│   ├── response/         # HTTP response helpers
│   └── store/            # SQLite data layer
├── flutter_app/          # Flutter frontend
├── Makefile
├── .env.example
├── AGENTS.md             # AI agent context and conventions
└── README.md
```

## Adding a New Resource

See [`AGENTS.md`](./AGENTS.md) for the step-by-step guide.
