#!/usr/bin/env bash
# bootstrap.sh — wait for MindsDB to be ready, then apply every .sql file
# under local-env/mindsdb/init via its HTTP API. Idempotent (CREATE IF NOT EXISTS).
set -euo pipefail

MINDSDB_URL="${MINDSDB_URL:-http://localhost:47334}"
INIT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)/init"

echo "Waiting for MindsDB at $MINDSDB_URL ..."
for _ in $(seq 1 60); do
  if curl -fsS "$MINDSDB_URL/api/status" >/dev/null 2>&1; then
    echo "  ready."
    break
  fi
  sleep 2
done

for f in "$INIT_DIR"/*.sql; do
  [ -f "$f" ] || continue
  echo "  → $(basename "$f")"
  sql=$(cat "$f")
  payload=$(jq -Rn --arg q "$sql" '{query: $q}')
  curl -fsS -X POST "$MINDSDB_URL/api/sql/query" \
    -H "Content-Type: application/json" \
    -d "$payload" | jq -r '.type, (.error_message // empty)' || {
      echo "    ⚠ failed — see response above"
      exit 1
    }
done

echo "✅ MindsDB bootstrap complete. GUI: $MINDSDB_URL"
