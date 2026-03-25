#!/bin/sh
# Pre-commit hook: enforces Go build and Flutter analysis pass before every commit.
# Install with: make install-hooks
set -e

echo "🔍 Pre-commit: verifying Go build..."
go build ./...

echo "🔍 Pre-commit: running Go tests..."
go test ./...

echo "🔍 Pre-commit: running Flutter analyze..."
cd flutter_app && flutter analyze

echo "✅ Pre-commit checks passed."
