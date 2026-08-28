# Build stage
# Build on the native runner arch and cross-compile to TARGETARCH — the builder
# downloads an amd64 sqlc and runs codegen, which can't run under arm64 emulation.
FROM --platform=$BUILDPLATFORM golang:1.26-alpine3.22 AS builder

WORKDIR /app

# Layer 1: system tools — cached until this line changes
RUN apk add --no-cache curl bash ca-certificates

# Layer 2: sqlc binary — use the fast release asset, with Go-module fallback
# when GitHub's release-assets endpoint is unavailable.
RUN set -eux; \
    if curl --fail --location --retry 5 --retry-all-errors --retry-delay 2 \
        --output /tmp/sqlc.tar.gz \
        https://github.com/sqlc-dev/sqlc/releases/download/v1.29.0/sqlc_1.29.0_linux_amd64.tar.gz; then \
      tar -xzf /tmp/sqlc.tar.gz -C /usr/local/bin sqlc; \
    else \
      GOBIN=/usr/local/bin go install github.com/sqlc-dev/sqlc/cmd/sqlc@v1.29.0; \
    fi; \
    rm -f /tmp/sqlc.tar.gz; \
    sqlc version

# Layer 3: copy source tree. The build context must contain the pinned
# submodules recorded by the checkout; cloning a moving branch here would make
# identical source commits produce different binaries.
COPY . .
RUN set -eux; \
    test -f ./voxie/go.mod; \
    test -f ./internal/platform/database/schemas/hr_core/schema.sql; \
    test -d ./internal/platform/email/templates

RUN go mod download && go mod verify
RUN chmod +x scripts/*.sh
RUN bash scripts/generate_sqlc.sh

ARG VERSION=dev
ARG COMMIT=unknown
ARG BUILD_TIME=unknown
ARG LICENSE_VENDOR_PUBLIC_KEY
ARG DEPLOYMENT_MODE=hosted
# TARGETARCH is provided by buildx; defaults to the build arch for plain builds
# so the hosted (amd64) path is unchanged while multi-arch publishes work.
ARG TARGETARCH

RUN CGO_ENABLED=0 GOOS=linux GOARCH=${TARGETARCH} \
    go build \
    -ldflags="-s -w \
      -X 'github.com/voxire/onion/internal/platform/server/http.Version=${VERSION}' \
      -X 'github.com/voxire/onion/internal/platform/server/http.Commit=${COMMIT}' \
      -X 'github.com/voxire/onion/internal/platform/server/http.BuildTime=${BUILD_TIME}' \
      -X 'github.com/voxire/onion/internal/platform/license.VendorPublicKey=${LICENSE_VENDOR_PUBLIC_KEY}' \
      -X 'github.com/voxire/onion/internal/platform/license.BuildDeploymentMode=${DEPLOYMENT_MODE}'" \
    -o /app/api ./cmd/api

# Final stage — minimal image
FROM alpine:3.22

RUN apk add --no-cache ca-certificates tzdata
RUN addgroup -S app && adduser -S -G app app

WORKDIR /app

COPY --from=builder /app/api /app/api

EXPOSE 9999

USER app

HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 CMD ["wget", "--spider", "--quiet", "http://127.0.0.1:9999/health"]

CMD ["/app/api"]
