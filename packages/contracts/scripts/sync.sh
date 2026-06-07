#!/bin/bash
# Sync generated TypeScript types from voxie submodule into @voxire/contracts
# Run after `cd voxie && buf generate`

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
CONTRACTS_SRC="$SCRIPT_DIR/../src"
VOXIE_TS="$SCRIPT_DIR/../../../voxie/gen/ts"
VOXIE_VALIDATION="$SCRIPT_DIR/../../../voxie/gen/validation"

if [ ! -d "$VOXIE_TS" ]; then
  echo "Error: voxie/gen/ts not found. Run 'cd voxie && buf generate' first."
  exit 1
fi

if [ ! -d "$VOXIE_VALIDATION" ]; then
  echo "Error: voxie/gen/validation not found. Run 'cd voxie && make generate' first."
  exit 1
fi

# Clean and copy
rm -rf "$CONTRACTS_SRC"
mkdir -p "$CONTRACTS_SRC"
cp -r "$VOXIE_TS"/* "$CONTRACTS_SRC/"
mkdir -p "$CONTRACTS_SRC/validation"
cp -r "$VOXIE_VALIDATION"/* "$CONTRACTS_SRC/validation/"

echo "Synced $(find "$CONTRACTS_SRC" -name '*.ts' | wc -l | tr -d ' ') TypeScript files and validation metadata from voxie"
