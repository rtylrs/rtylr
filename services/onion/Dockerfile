# Build stage
FROM golang:1.26-alpine AS builder

WORKDIR /app

# Layer 1: system tools — cached until this line changes
RUN apk add --no-cache curl bash

# Layer 2: sqlc pre-built binary — cached until version changes (~3s vs 5min from source)
RUN curl -sSL https://github.com/sqlc-dev/sqlc/releases/download/v1.29.0/sqlc_1.29.0_linux_amd64.tar.gz \
    | tar -xz -C /usr/local/bin sqlc

# Layer 3: deps — only reruns when go.mod / go.sum change
COPY go.mod go.sum ./
COPY voxie/go.mod voxie/go.sum* ./voxie/
RUN go mod download && go mod verify

# Layer 4: source — reruns on any code change but layers 1-3 are already cached
COPY cmd ./cmd
COPY internal ./internal
COPY scripts ./scripts
COPY voxie ./voxie

RUN chmod +x scripts/*.sh
RUN bash scripts/generate_sqlc.sh

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -ldflags="-s -w" -o /app/api ./cmd/api

# Final stage — minimal image
FROM alpine:latest

RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app

COPY --from=builder /app/api /app/api

EXPOSE 9999

CMD ["/app/api"]
