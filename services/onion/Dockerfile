# Build stage
# Build on the native runner arch and cross-compile to TARGETARCH — the builder
# downloads an amd64 sqlc and runs codegen, which can't run under arm64 emulation.
#
# Build from the repository root so the service and the contract module are both
# in context: docker build -f services/onion/Dockerfile .
FROM --platform=$BUILDPLATFORM golang:1.27-alpine3.24 AS builder

WORKDIR /src

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

# Layer 3: generated contract output is gitignored, so produce it from the
# pinned onek release before Go compiles against it.
COPY voxie/ ./voxie/
RUN GOBIN=/usr/local/bin go install github.com/1homsi/onekit/cmd/onek@v0.34.1 && \
    cd /src/voxie && onek build .

# Layer 4: the service, plus the SQL schemas and email templates it embeds.
COPY services/onion/ ./services/onion/
WORKDIR /src/services/onion
RUN set -eux; \
    test -f /src/voxie/go.mod; \
    test -f ./internal/modules/hr/sql/schema/hr_core/schema.sql; \
    test -d ./internal/platform/email/templates

RUN go mod download && go mod verify
RUN chmod +x scripts/*.sh
RUN bash scripts/generate_sqlc.sh

ARG VERSION=dev
ARG COMMIT=unknown
ARG BUILD_TIME=unknown
# TARGETARCH is provided by buildx; defaults to the build arch for plain builds
# so the hosted (amd64) path is unchanged while multi-arch publishes work.
ARG TARGETARCH

RUN CGO_ENABLED=0 GOOS=linux GOARCH=${TARGETARCH} \
    go build \
    -ldflags="-s -w \
      -X 'github.com/voxire/onion/internal/platform/server/http.Version=${VERSION}' \
      -X 'github.com/voxire/onion/internal/platform/server/http.Commit=${COMMIT}' \
      -X 'github.com/voxire/onion/internal/platform/server/http.BuildTime=${BUILD_TIME}'" \
    -o /app/api ./cmd/api

# Build the image/video transformation executable used by the API's upload
# routes. Keep it in the same Debian family as the runtime shared libraries.
FROM debian:bookworm-slim AS media-builder

RUN apt-get update && apt-get install -y --no-install-recommends \
    cmake \
    build-essential \
    pkg-config \
    libvips-dev \
    ffmpeg \
    libavcodec-dev \
    libavformat-dev \
    libavutil-dev \
    libswscale-dev \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /build
COPY services/onion/media-engine/CMakeLists.txt ./
COPY services/onion/media-engine/src/ ./src/
COPY services/onion/media-engine/include/ ./include/
RUN cmake -B build -DCMAKE_BUILD_TYPE=Release \
        -DMEDIA_ENGINE_BENCH=OFF \
        -DMEDIA_ENGINE_TESTS=OFF && \
    cmake --build build --target media-engine --parallel "$(nproc)"

# Final stage includes the media runtime libraries used by the API.
FROM debian:bookworm-slim

RUN apt-get update && apt-get install -y --no-install-recommends \
    ca-certificates \
    tzdata \
    wget \
    libvips42 \
    ffmpeg \
    && rm -rf /var/lib/apt/lists/* \
    && groupadd --system app \
    && useradd --system --gid app --no-create-home app

WORKDIR /app

COPY --from=builder /app/api /app/api
COPY --from=media-builder --chown=app:app /build/build/media-engine /app/media-engine/build/media-engine
COPY --chown=app:app services/onion/internal/ai/classifier/models/cua-s1-forms /app/models/cua-s1-forms

ENV FORM_MAPPING_CHECKPOINT_PATH=/app/models/cua-s1-forms
ENV MEDIA_ENGINE_PATH=/app/media-engine/build/media-engine

EXPOSE 9999

USER app

HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 CMD ["wget", "--spider", "--quiet", "http://127.0.0.1:9999/health"]

CMD ["/app/api"]
