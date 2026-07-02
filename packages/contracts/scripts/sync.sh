#!/bin/bash
# Generate TypeScript contracts from the local voxie submodule.

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
CONTRACTS_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
FRONTEND_ROOT="$(cd "$CONTRACTS_ROOT/../.." && pwd)"
CONTRACTS_SRC="$CONTRACTS_ROOT/src"
VOXIE_ROOT="$FRONTEND_ROOT/voxie"
TEMPLATE_DIR="$(mktemp -d "${TMPDIR:-/tmp}/voxie.ts.XXXXXX")"
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
  echo "Error: buf is required to generate contracts."
  exit 1
fi

export PATH="$FRONTEND_ROOT/node_modules/.bin:$CONTRACTS_ROOT/node_modules/.bin:$PATH"
if ! command -v protoc-gen-ts_proto >/dev/null 2>&1; then
  echo "Error: protoc-gen-ts_proto is required. Run 'pnpm install' first."
  exit 1
fi

ONEKIT_TS_PLUGIN="${ONEKIT_TS_CLIENT_PLUGIN:-protoc-gen-onekit-ts-client}"
ONEKIT_REQUIRE="${VOXIRE_REQUIRE_ONEKIT:-0}"
ONEKIT_CLIENTS_DIR="$CONTRACTS_SRC/onekit"

cat > "$TEMPLATE_FILE" <<EOF
version: v2
plugins:
  - local: protoc-gen-ts_proto
    out: $CONTRACTS_SRC
    opt:
      - esModuleInterop=true
      - outputServices=false
      - env=browser
      - useExactTypes=false
      - exportCommonSymbols=false
      - onlyTypes=true
EOF

if command -v "$ONEKIT_TS_PLUGIN" >/dev/null 2>&1; then
  cat >> "$TEMPLATE_FILE" <<EOF
  - local: $ONEKIT_TS_PLUGIN
    out: $ONEKIT_CLIENTS_DIR
    opt:
      - paths=source_relative
EOF
elif [ "$ONEKIT_REQUIRE" = "1" ] || [ "$ONEKIT_REQUIRE" = "true" ]; then
  echo "Error: $ONEKIT_TS_PLUGIN is required when VOXIRE_REQUIRE_ONEKIT is enabled."
  exit 1
else
  echo "Skipping oneKit TypeScript client generation; $ONEKIT_TS_PLUGIN is not installed."
fi

echo "Generating contracts from voxie..."
rm -rf "$CONTRACTS_SRC"
mkdir -p "$CONTRACTS_SRC"
buf generate "$VOXIE_ROOT" --template "$TEMPLATE_FILE"
python3 "$SCRIPT_DIR/generate_validation_metadata.py" \
  --proto-root "$VOXIE_ROOT/proto" \
  --out-dir "$CONTRACTS_SRC/validation"

if command -v pnpm >/dev/null 2>&1; then
  (
    cd "$FRONTEND_ROOT"
    pnpm exec prettier --write "packages/contracts/src/**/*.ts" "packages/contracts/src/validation/*.json" >/dev/null
  )
fi

echo "Generated $(find "$CONTRACTS_SRC" -name '*.ts' | wc -l | tr -d ' ') TypeScript files and validation metadata"
