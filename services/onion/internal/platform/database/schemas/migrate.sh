#!/bin/bash
# migrate.sh — forwarding entry point for the schema migrations.
#
# This file used to be a standalone copy of the retired schemas repository's
# runner. It drifted from the canonical runner (no DATABASE_URL support, no
# pinned utf8mb4 client charset, so flow/schema.sql's emoji defaults aborted a
# fresh install) and never applied the module-owned migrations, so it is now a
# one-line delegate instead of a second implementation.
#
# The canonical runner, its env vars and its coverage live at:
#   scripts/migrate.sh
#
# It accepts DB_HOST (default: localhost), DB_PORT (default: 3306), DB_USER,
# DB_PASSWORD, DB_NAME — or DATABASE_URL instead of the DB_* variables.

set -e

SCHEMAS_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
exec "$SCHEMAS_DIR/../../../../scripts/migrate.sh" "$@"
