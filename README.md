# Vibe Coding Template (Go + Flutter)

A lightning-fast, production-ready full-stack template optimized for AI agents (Cursor, Copilot, Claude). Builds an instant backend API with typed SQLite, paired with a robust Flutter web/mobile interface.

## Tech Stack

- **Backend:** Go 1.25+, pure Go SQLite (`modernc.org/sqlite`), `sqlc` for type-safe code generation, `golang-migrate`
- **Frontend:** Flutter 3.41+, `go_router`, `riverpod` + `riverpod_generator`, Material 3 design system with manual light/dark themes
- **Auth:** Static, SHA-256 hashed API Keys stored in SQLite with middleware validation
- **Dev-Ex:** Built-in `Makefile` automation, `air` for hot-reloading, `golangci-lint`, VSCode compound debug templates
- **API Spec:** Automatic OpenAPI (Swagger) generation via `swag`

## Quick Start

### 1. Clone & Environment
```bash
git clone https://github.com/alekssaul/vibe-coding-template.git
cd vibe-coding-template
```

Copy the environment files for local development:
```bash
cp .env.example .env
cp flutter_app/.env.example flutter_app/.env
```

### 2. Rename & Initialize (Optional but recommended)
Run the initialization script. This will rename the go module, the flutter package, and update text references automatically.
```bash
make init PROJECT=your_new_app_name
```

### 3. Install Tools & Generate Code
Install all required CLI tools (`air`, `migrate`, `sqlc`, `swag`) and generate database + Flutter code:
```bash
make install-tools
make db-generate
cd flutter_app && dart run build_runner build -d
cd ..
```

### 4. Seed the Database
Seed the database with dummy Items and developer API keys (prints the plaintext keys you need for Flutter):
```bash
make seed
```

> **Important**: Copy one of the seeded API keys into your `flutter_app/.env` (`API_KEY=...`) to allow the mobile app to hit the backend.

### 5. Run the Stack

Run the backend API (with hot-reloading via `air`):
```bash
make dev
```

In another terminal — run the Flutter frontend:
```bash
cd flutter_app
flutter run -d chrome
```

## AI Agent Integration (`AGENTS.md`)

This repository is strictly designed for **Context-Driven Development (CDD)** with AI agents.

Before asking your AI to write code, tell it to read `AGENTS.md`. It contains strict rules for:
- How to write new `sqlc` queries and migrations.
- Which libraries to use (e.g., `riverpod_generator`).
- How to structure handlers, models, and UI screens.
- Pre-commit verifications (always ask the AI to run `make verify` before concluding a task).

## Scripts & Tools

| Command | Description |
|---|---|
| `make dev` | Start backend with `air` hot-reloader |
| `make db-generate` | Regenerate Go types from `internal/store/queries/` |
| `make seed` | Auto-migrate and seed dummy data and API keys |
| `make scaffold RESOURCE=... FIELDS=...` | Generate full-stack CRUD (9 files + auto-inject routes) |
| `make verify` | Run Go build, Go tests, and Flutter analyzer |
| `make test` | Run Go unit/integration tests |
| `make init PROJECT=` | Rename the template across all files |
| `make migrate-add NAME=` | Scaffold a new `.sql` migration file |
| `make docs` | Generate Swagger/OpenAPI docs via `swag` |

## Docker Deployment

```bash
docker compose up --build
```

The API will be available at `http://localhost:8080`. SQLite data persists via a Docker volume.

## License

MIT
