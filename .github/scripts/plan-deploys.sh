#!/usr/bin/env bash
set -euo pipefail

# Picks the deploy targets for a release from .github/deploy-targets.json and
# writes `pages=<json>` and `ecs=<json>` to $GITHUB_OUTPUT (or stdout).
#
#   plan-deploys.sh files <file>        targets whose `paths` (minus `ignore`) match a changed file listed in <file>
#   plan-deploys.sh app <name|all>      one named target, or every target

targets="$(dirname "$0")/../deploy-targets.json"
mode="${1:?usage: plan-deploys.sh files <file> | app <name|all>}"
arg="${2:?missing argument}"

case "$mode" in
  files)
    changed="$(jq -R . "$arg" | jq -sc 'map(select(length > 0))')"
    selected="$(jq -c --argjson files "$changed" '
      map(select(. as $t
        | any($files[]; . as $f
            | any($t.paths[]; . as $p | $f | startswith($p))
              and (any(($t.ignore // [])[]; . as $i | $f | startswith($i)) | not))))' "$targets")"
    ;;
  app)
    if [ "$arg" = all ]; then
      selected="$(jq -c . "$targets")"
    else
      selected="$(jq -c --arg app "$arg" 'map(select(.app == $app))' "$targets")"
      [ "$selected" != '[]' ] || { echo "Unknown deploy target: $arg" >&2; exit 1; }
    fi
    ;;
  *)
    echo "Unknown mode: $mode" >&2
    exit 1
    ;;
esac

out="${GITHUB_OUTPUT:-/dev/stdout}"
for type in pages ecs; do
  echo "$type=$(jq -c --arg type "$type" 'map(select(.type == $type) | del(.paths, .ignore))' <<< "$selected")" >> "$out"
done
