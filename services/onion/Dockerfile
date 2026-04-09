# syntax=docker/dockerfile:1.4
# Build stage
FROM golang:1.25.0-alpine AS builder

WORKDIR /app

# Install build tools (cached layer — only reruns if this line changes)
RUN apk add --no-cache curl bash

# Download sqlc pre-built binary instead of compiling from source (~30s → ~3s)
RUN curl -sSL https://github.com/sqlc-dev/sqlc/releases/download/v1.29.0/sqlc_1.29.0_linux_amd64.tar.gz \
    | tar -xz -C /usr/local/bin sqlc

# Copy dependency manifests first for best cache hit
COPY go.mod go.sum ./

# Download modules with cache mount — survives across builds
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download && go mod verify

# Copy source
COPY cmd ./cmd
COPY internal ./internal
COPY scripts ./scripts

RUN chmod +x scripts/*.sh

# Generate SQLC with module cache mount
RUN --mount=type=cache,target=/go/pkg/mod \
    bash scripts/generate_sqlc.sh

# Build binary with both module cache and Go build cache mounts
# The build cache mount is the single biggest speedup — incremental recompilation
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -ldflags="-s -w" -o /app/api ./cmd/api

# Final stage — minimal image
FROM alpine:latest

RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app

COPY --from=builder /app/api /app/api

EXPOSE 9999

CMD ["/app/api"]
