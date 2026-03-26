# ── Build Stage ────────────────────────────────────────────────────────────────
FROM golang:1.23-alpine AS builder

WORKDIR /app

# Download dependencies first (Docker layer caching).
COPY go.mod go.sum ./
RUN go mod download

# Copy source and build.
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /api ./cmd/api

# ── Run Stage ─────────────────────────────────────────────────────────────────
FROM alpine:3.21

RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app
COPY --from=builder /api /app/api
COPY --from=builder /app/internal/store/migrations /app/internal/store/migrations

EXPOSE 8080
ENV API_PORT=8080

ENTRYPOINT ["/app/api"]
