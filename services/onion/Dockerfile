# Build stage
FROM golang:1.25.0-alpine AS builder

WORKDIR /app

# Install build tools and sqlc
RUN apk add --no-cache gcc musl-dev curl make bash

# Install sqlc via Go (architecture agnostic)
RUN go install github.com/sqlc-dev/sqlc/cmd/sqlc@v1.29.0

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download && go mod verify

# Copy source code
COPY . .

# Debug: Check if scripts directory exists and list contents
RUN ls -la scripts/ && echo "Scripts directory contents:"

# Make scripts executable
RUN chmod +x scripts/*.sh

# Generate SQL code (run script directly with bash)
RUN bash scripts/generate_sqlc.sh

# Build API binary with optimizations for the target platform
RUN CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -ldflags="-s -w" -o /app/api ./cmd/api/main.go

# Final stage
FROM alpine:latest

WORKDIR /app

# Copy API binary from builder
COPY --from=builder /app/api /app/api

# Expose API port
EXPOSE 9999

# Default to API service (backwards compatibility)
CMD ["/app/api"] 
