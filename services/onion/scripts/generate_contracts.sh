#!/bin/bash
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
VOXIE_ROOT="$ROOT/voxie"
OUT_DIR="$VOXIE_ROOT/gen/go"
TEMPLATE_DIR="$(mktemp -d "${TMPDIR:-/tmp}/voxie.go.XXXXXX")"
TEMPLATE_FILE="$TEMPLATE_DIR/buf.gen.yaml"

cleanup() {
  rm -rf "$TEMPLATE_DIR"
}
trap cleanup EXIT

if [ ! -d "$VOXIE_ROOT/proto" ]; then
  echo "Error: voxie/proto not found. Run 'git submodule update --init --recursive' first."
  exit 1
fi

if ! command -v buf >/dev/null 2>&1; then
  echo "Error: buf is required to generate Voxie Go contracts."
  exit 1
fi

if ! command -v protoc-gen-go >/dev/null 2>&1; then
  echo "Error: protoc-gen-go is required. Install google.golang.org/protobuf/cmd/protoc-gen-go."
  exit 1
fi

ONEKIT_PLUGIN="${ONEKIT_GO_HTTP_PLUGIN:-protoc-gen-onekit-go-http}"
ONEKIT_REQUIRE="${ONION_REQUIRE_ONEKIT:-0}"

cat > "$TEMPLATE_FILE" <<EOF
version: v2
plugins:
  - local: protoc-gen-go
    out: $OUT_DIR
    opt:
      - paths=source_relative
EOF

if command -v "$ONEKIT_PLUGIN" >/dev/null 2>&1; then
  cat >> "$TEMPLATE_FILE" <<EOF
  - local: $ONEKIT_PLUGIN
    out: $OUT_DIR
    opt:
      - paths=source_relative
EOF
elif [ "$ONEKIT_REQUIRE" = "1" ] || [ "$ONEKIT_REQUIRE" = "true" ]; then
  echo "Error: $ONEKIT_PLUGIN is required when ONION_REQUIRE_ONEKIT is enabled."
  exit 1
else
  echo "Skipping oneKit HTTP generation; $ONEKIT_PLUGIN is not installed."
fi

rm -rf "$OUT_DIR"
mkdir -p "$OUT_DIR"
buf generate "$VOXIE_ROOT" --template "$TEMPLATE_FILE"
echo "Generated Voxie Go contracts in $OUT_DIR"
