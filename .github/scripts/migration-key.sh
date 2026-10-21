#!/usr/bin/env bash
set -euo pipefail

# Prints a key for the migrations services/onion/scripts/migrate.sh applies at
# <commit>: the schema registry, every file it lists, the module-owned
# migration_*.sql files it discovers, and migrate.sh itself (it holds the skip
# list). Two commits with the same key apply exactly the same migrations.
# Exits non-zero, printing nothing, when the commit cannot be read.
#
#   migration-key.sh [commit]   (default HEAD)

commit="${1:-HEAD}"
root=services/onion
registry="$root/internal/platform/database/schema_registry.tsv"
listing="$(mktemp)"
trap 'rm -f "$listing"' EXIT

sha256() {
  if command -v sha256sum >/dev/null 2>&1; then sha256sum; else shasum -a 256; fi
}

git ls-tree --full-tree "$commit" -- "$registry" "$root/scripts/migrate.sh" > "$listing"
git show "$commit:$registry" | sed '/^#/d; /^$/d' | cut -f2 | tr '|' '\n' |
  sed "s|^|$root/|" | xargs -r git ls-tree --full-tree "$commit" -- >> "$listing"
git ls-tree -r --full-tree "$commit" -- "$root/internal/modules" |
  awk -F '\t' '$2 ~ /\/sql\/migration_.*\.sql$/' >> "$listing"

# ls-tree succeeds silently for a missing path; never key an empty listing.
if ! awk -F '\t' -v path="$registry" '$2 == path { found = 1 } END { exit !found }' "$listing"; then
  echo "No schema registry at $commit" >&2
  exit 1
fi

sort -u "$listing" | sha256 | cut -c1-16
