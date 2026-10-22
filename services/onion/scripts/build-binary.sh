#!/bin/bash
set -euo pipefail

# Compiles one onion command the way release images ship it:
#
#   scripts/build-binary.sh ./cmd/api /app/api
#
# Run from services/onion. The Dockerfiles' builder stages and the release
# workflow's native build both use this, so their binaries match. VERSION,
# COMMIT and BUILD_TIME are stamped into the binary, and TARGETARCH picks the
# architecture (buildx passes it; it defaults to the host's).

package="${1:?usage: build-binary.sh <package> <output>}"
output="${2:?usage: build-binary.sh <package> <output>}"
http_pkg=github.com/voxire/onion/internal/platform/server/http

CGO_ENABLED=0 GOOS=linux GOARCH="${TARGETARCH:-$(go env GOARCH)}" go build \
    -ldflags="-s -w \
      -X '${http_pkg}.Version=${VERSION:-dev}' \
      -X '${http_pkg}.Commit=${COMMIT:-unknown}' \
      -X '${http_pkg}.BuildTime=${BUILD_TIME:-unknown}'" \
    -o "$output" "$package"
