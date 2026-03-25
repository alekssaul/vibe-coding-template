PROJECT_NAME ?= template
MODULE_PATH   = github.com/alekssaul/$(PROJECT_NAME)
BINARY        = $(PROJECT_NAME)
BIN_DIR       = bin

GIT_SHA   := $(shell git rev-parse --short HEAD 2>/dev/null || echo "dev")
BUILD_TIME := $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
LDFLAGS   := -ldflags "-X main.gitSHA=$(GIT_SHA) -X main.buildTime=$(BUILD_TIME)"

# Auto-load .env if present
ifneq (,$(wildcard .env))
  include .env
  export
endif

.DEFAULT_GOAL := help

.PHONY: help
help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-22s\033[0m %s\n", $$1, $$2}'

# ── Go ────────────────────────────────────────────────────────────────────────

.PHONY: build
build: ## Build the Go binary with git SHA + build time
	@mkdir -p $(BIN_DIR)
	go build $(LDFLAGS) -o $(BIN_DIR)/$(BINARY) ./cmd/api

.PHONY: run
run: build ## Build and run the API server
	./$(BIN_DIR)/$(BINARY)

.PHONY: dev
dev: ## Run without building a binary (hot-suitable for development)
	go run $(LDFLAGS) ./cmd/api

.PHONY: test
test: ## Run all tests
	go test ./...

.PHONY: test-v
test-v: ## Run all tests with verbose output
	go test -v ./...

.PHONY: lint
lint: ## Run golangci-lint
	golangci-lint run ./...

.PHONY: fmt
fmt: ## Format Go source files
	gofmt -w .

.PHONY: docs
docs: ## Generate OpenAPI docs via swag
	swag init -g cmd/api/main.go -o docs/

.PHONY: verify-go
verify-go: ## Verify Go code compiles (run after every Go change)
	go build ./...

.PHONY: clean
clean: ## Remove build artifacts
	rm -rf $(BIN_DIR)/

# ── Flutter ───────────────────────────────────────────────────────────────────

.PHONY: flutter-run-web
flutter-run-web: ## Run Flutter app in Chrome
	cd flutter_app && flutter run -d chrome

.PHONY: flutter-run-android
flutter-run-android: ## Run Flutter app on connected Android device
	cd flutter_app && flutter run -d android

.PHONY: flutter-build-web
flutter-build-web: ## Build Flutter web release
	cd flutter_app && flutter build web

.PHONY: flutter-build-android
flutter-build-android: ## Build Flutter Android APK
	cd flutter_app && flutter build apk

.PHONY: flutter-analyze
flutter-analyze: ## Analyze Flutter/Dart code
	cd flutter_app && flutter analyze

.PHONY: verify-flutter
verify-flutter: ## Verify Flutter builds (run after every Flutter change)
	cd flutter_app && flutter build web --debug

# ── Combined ──────────────────────────────────────────────────────────────────

.PHONY: verify
verify: verify-go test flutter-analyze ## Run all verifications (Go build + tests + Flutter analyze)

# ── Setup ─────────────────────────────────────────────────────────────────────

.PHONY: install-tools
install-tools: ## Install required Go CLI tools
	go install github.com/swaggo/swag/cmd/swag@latest
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

.PHONY: init
init: ## Rename template to your project: make init PROJECT=myapp
	@if [ -z "$(PROJECT)" ]; then echo "Usage: make init PROJECT=myapp"; exit 1; fi
	@echo "Renaming template → $(PROJECT)..."
	@find . -type f \( -name "*.go" -o -name "*.mod" -o -name "Makefile" \) \
	  | grep -v vendor \
	  | xargs sed -i '' 's|alekssaul/template|alekssaul/$(PROJECT)|g'
	@sed -i '' 's|^name: flutter_app|name: $(PROJECT)|g' flutter_app/pubspec.yaml
	@sed -i '' 's|^description: "A new Flutter project."|description: "$(PROJECT) — a CRUD app"|g' flutter_app/pubspec.yaml
	@go mod tidy
	@[ -f .env ] || cp .env.example .env && echo "  ✓ created .env from .env.example"
	@[ -f flutter_app/.env ] || cp flutter_app/.env.example flutter_app/.env && echo "  ✓ created flutter_app/.env from .env.example"
	@echo "✅ Done! Project is now: $(PROJECT)"
	@echo "   → Edit .env and flutter_app/.env with your real values, then: make dev"

.PHONY: install-hooks
install-hooks: ## Install git pre-commit hook (runs Go build + tests + Flutter analyze)
	@cp scripts/pre-commit.sh .git/hooks/pre-commit
	@chmod +x .git/hooks/pre-commit
	@echo "✅ Pre-commit hook installed"

