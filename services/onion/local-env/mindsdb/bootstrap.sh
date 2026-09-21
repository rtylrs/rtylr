#!/usr/bin/env bash
# bootstrap.sh — wait for MindsDB to be ready, then apply every .sql file
# under local-env/mindsdb/init via its HTTP API. Idempotent (CREATE IF NOT EXISTS).
set -euo pipefail

MINDSDB_URL="${MINDSDB_URL:-http://localhost:47334}"
INIT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)/init"

echo "Waiting for MindsDB at $MINDSDB_URL ..."
ready=false
for _ in $(seq 1 60); do
  if curl -fsS "$MINDSDB_URL/api/status" >/dev/null 2>&1; then
    echo "  ready."
    ready=true
    break
  fi
  sleep 2
done

if [ "$ready" != true ]; then
  echo "  ❌ timed out waiting for MindsDB at $MINDSDB_URL"
  exit 1
fi

for f in "$INIT_DIR"/*.sql; do
  [ -f "$f" ] || continue
  echo "  → $(basename "$f")"
  sql=$(cat "$f")
  payload=$(jq -Rn --arg q "$sql" '{query: $q}')
  response=$(curl -fsS -X POST "$MINDSDB_URL/api/sql/query" \
    -H "Content-Type: application/json" \
    -d "$payload") || {
      echo "    ⚠ failed — request to MindsDB failed"
      exit 1
    }
  # MindsDB answers 200 even when the statement is rejected, so the payload
  # must be inspected: {"type":"error","error_message":"..."} is a failure.
  qtype=$(echo "$response" | jq -r '.type // empty' 2>/dev/null) || {
    echo "    ⚠ failed — unparseable response from MindsDB"
    exit 1
  }
  if [ "$qtype" = "error" ]; then
    echo "    ⚠ failed — $(echo "$response" | jq -r '.error_message // "unknown error"')"
    exit 1
  fi
  if [ -z "$qtype" ]; then
    echo "    ⚠ failed — response has no type: $response"
    exit 1
  fi
  echo "    $qtype"
done

echo "✅ MindsDB bootstrap complete. GUI: $MINDSDB_URL"
