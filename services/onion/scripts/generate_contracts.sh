#!/bin/bash
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
VOXIE_ROOT="$ROOT/voxie"
OUT_DIR="$VOXIE_ROOT/gen/go"
TEMPLATE_FILE="$(mktemp "${TMPDIR:-/tmp}/voxie.go.XXXXXX.yaml")"

cleanup() {
  rm -f "$TEMPLATE_FILE"
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

cat > "$TEMPLATE_FILE" <<EOF
version: v2
plugins:
  - local: protoc-gen-go
    out: $OUT_DIR
    opt:
      - paths=source_relative
EOF

rm -rf "$OUT_DIR"
mkdir -p "$OUT_DIR"
buf generate "$VOXIE_ROOT" --template "$TEMPLATE_FILE"
echo "Generated Voxie Go contracts in $OUT_DIR"
