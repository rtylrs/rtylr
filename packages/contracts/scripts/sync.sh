#!/bin/bash
# Sync generated TypeScript types from voxie submodule into @voxire/contracts
# Run after `cd voxie && buf generate`

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
CONTRACTS_SRC="$SCRIPT_DIR/../src"
VOXIE_TS="$SCRIPT_DIR/../../../voxie/gen/ts"

if [ ! -d "$VOXIE_TS" ]; then
  echo "Error: voxie/gen/ts not found. Run 'cd voxie && buf generate' first."
  exit 1
fi

# Clean and copy
rm -rf "$CONTRACTS_SRC"
mkdir -p "$CONTRACTS_SRC"
cp -r "$VOXIE_TS"/* "$CONTRACTS_SRC/"

echo "Synced $(find "$CONTRACTS_SRC" -name '*.ts' | wc -l | tr -d ' ') TypeScript files from voxie"
