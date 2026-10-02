#!/usr/bin/env bash
# Verifies that the generated Voxie contracts match voxie/api/*.onk.
#
# Contract output is generated, not committed (voxie/.gitignore excludes gen/,
# and git ls-files voxie/gen is empty), so there is no committed tree to diff
# against: regenerating to a temp dir and diffing would compare the tree against
# itself. Staleness therefore has to be decided from content, not timestamps.
#
# A present-but-stale tree is the case this exists for. Advancing across commits
# leaves the old voxie/gen and clients/packages/contracts/src in place, and an
# existence check accepts that silently. Here every service declared in a
# service.onk must register exactly the same operations in the mirrored
# generated server.gen.go, and every generated server.gen.go must correspond to
# a declared service.onk. An orphan in either direction is an endpoint no
# committed contract defines, or a contract with no generated server, and either
# one reaches Go handlers and TypeScript consumers as if it were contracted.
#
# The parity check needs no Go toolchain and no network, so it runs first and
# always. Regeneration is only attempted when output is missing or parity
# actually failed, which keeps the expensive toolchain off the common path while
# still self-healing a stale tree.
#
# mtime is deliberately NOT a staleness signal: a fresh clone gives every file
# the same checkout time, and checking out an older .onk leaves a *newer* mtime
# on generated output, so an mtime rule both false-fails and false-passes. It is
# weaker than content parity, which is why parity is the gate.
set -euo pipefail

EXIT_OK=0
EXIT_STALE=1
EXIT_LAYOUT=2
EXIT_TOOLING=3
EXIT_GENERATE=4

usage() {
  cat <<'EOF'
Usage: ensure-contracts.sh [--no-generate] [--help]

Ensures voxie/gen and clients/packages/contracts/src match voxie/api/*.onk.

Nothing generated is committed, so a checkout that advanced across commits
keeps output generated for an older revision. This compares the generated tree
against the .onk sources directly, per service, and fails if they disagree.

Options:
  --no-generate  verify only; never run `make -C voxie generate`. Use this to
                see the real drift without touching the toolchain.
  --help        show this help.

Exit codes:
  0  generated contracts match voxie/api/*.onk
  1  contracts are stale, or an operation/service is missing or orphaned
  2  the in-tree voxie contracts are missing or malformed
  3  a required tool is unavailable; the check did not run and was NOT
     silently skipped
  4  `make -C voxie generate` failed

How to fix a failure:
  make -C voxie clean generate

That removes voxie/gen and clients/packages/contracts/src and rebuilds them
from voxie/api/*.onk. It needs a Go toolchain and network access for
`go install onek`, so run it yourself rather than trusting this check to do it
silently. This script will also run it for you when it is needed and the
toolchain is present; when it is not, it fails with exit 3 rather than passing.

Never commit voxie/gen or clients/packages/contracts/src - only the .onk sources
under voxie/api are tracked.
EOF
}

generate_ok=1
for arg in ${1+"$@"}; do
  case "$arg" in
    --no-generate)
      generate_ok=0
      ;;
    -h | --help)
      usage
      exit "$EXIT_OK"
      ;;
    *)
      echo "::error::unknown argument '$arg' (try --help)" >&2
      exit "$EXIT_LAYOUT"
      ;;
  esac
done

fail() {
  local code="$1"
  shift
  echo "::error::$*" >&2
  exit "$code"
}

remedy() {
  cat >&2 <<'EOF'

How to fix:
  make -C voxie clean generate

That removes voxie/gen and clients/packages/contracts/src and rebuilds them
from voxie/api/*.onk. It needs a Go toolchain and network access for
`go install onek`, so run it yourself rather than expecting this check to do it.
To see the drift without regenerating: ./scripts/ensure-contracts.sh --no-generate
voxie/gen and clients/packages/contracts/src are gitignored output: never
commit them, only the .onk sources under voxie/api.
EOF
}

repo_root="$(git rev-parse --show-toplevel 2>/dev/null || true)"
if [[ -z "$repo_root" || ! -f "$repo_root/voxie/onekit.toml" ]]; then
  # Git hooks can export worktree variables that make nested commands report
  # their invocation directory as the root. Resolve the monorepo by its
  # contract manifest when that reported root does not contain the source.
  repo_root="$PWD"
  while [[ "$repo_root" != / && ! -f "$repo_root/voxie/onekit.toml" ]]; do
    repo_root="$(dirname "$repo_root")"
  done
fi
if [[ ! -f "$repo_root/voxie/onekit.toml" ]]; then
  fail "$EXIT_LAYOUT" "not inside a git work tree or RTYLR checkout; cannot locate voxie/onekit.toml"
fi

voxie_dir="$repo_root/voxie"
contracts_src="$repo_root/clients/packages/contracts/src"
gen_dir="$voxie_dir/gen"
gen_go="$gen_dir/go"

if [[ ! -f "$voxie_dir/onekit.toml" ]]; then
  fail "$EXIT_LAYOUT" "in-tree contracts missing at $voxie_dir"
fi
if [[ ! -d "$voxie_dir/api" ]]; then
  fail "$EXIT_LAYOUT" "contract sources missing at $voxie_dir/api"
fi

# Every tool below is required, including tr (orphan/missing report), cat
# (--help and remedy text) and rm (the temp-dir cleanup trap). A missing one
# aborts with EXIT_TOOLING: a check that quietly skips itself is the same defect
# as a check that cannot fail.
for tool in find sort comm mktemp grep sed tr cat rm; do
  if ! command -v "$tool" >/dev/null 2>&1; then
    fail "$EXIT_TOOLING" "'$tool' is not on PATH; the contracts check cannot run and was not silently skipped"
  fi
done

work="$(mktemp -d)"
trap 'rm -rf "$work"' EXIT

# Operation declarations read `  getSecurityState(Req) -> Resp @get("/state")`.
# Operation names are compared rather than HTTP routes because the route prefix
# is derived by the generator (the package path, or the service's own base_path),
# which this script must not have to reimplement.
declared_ops() {
  grep -oE '^[[:space:]]+[a-zA-Z_][a-zA-Z0-9_]*\([^)]*\)[[:space:]]*->' "$1" |
    sed -E 's/^[[:space:]]+([a-zA-Z_][a-zA-Z0-9_]*)\(.*/\1/' | LC_ALL=C sort -u || true
}

registered_ops() {
  grep -oE 'RequestMetadata\{Service: "[^"]*", Method: "[^"]*"' "$1" |
    sed -E 's/.*Method: "([^"]*)"/\1/' | LC_ALL=C sort -u || true
}

# Compares declared and generated services. Echoes nothing; sets `parity_failed`
# and `parity_services`.
parity_failed=0
parity_services=0
parity_generated=0
check_parity() {
  local service_listing gen_listing src pkg_dir gen_server
  service_listing="$(find "$voxie_dir/api" -type f -name 'service.onk' | LC_ALL=C sort)"
  if [[ -z "$service_listing" ]]; then
    fail "$EXIT_LAYOUT" "no service.onk contracts under $voxie_dir/api"
  fi
  gen_listing="$(find "$gen_go" -type f -name 'server.gen.go' | LC_ALL=C sort)"

  while IFS= read -r src; do
    parity_services=$((parity_services + 1))
    pkg_dir="${src#"$voxie_dir/api"/}"
    pkg_dir="${pkg_dir%/service.onk}"
    gen_server="$gen_go/$pkg_dir/server.gen.go"

    if [[ ! -f "$gen_server" ]]; then
      echo "::error::$gen_server is missing; $src declares a service that was never generated" >&2
      parity_failed=1
      continue
    fi
    parity_generated=$((parity_generated + 1))

    declared_ops "$src" >"$work/declared"
    registered_ops "$gen_server" >"$work/registered"

    local orphans missing
    orphans="$(comm -13 "$work/declared" "$work/registered")"
    missing="$(comm -23 "$work/declared" "$work/registered")"
    if [[ -n "$orphans" || -n "$missing" ]]; then
      echo "::error::generated server for $pkg_dir does not match $src" >&2
      if [[ -n "$orphans" ]]; then
        printf '  registered in generated output but declared by no .onk (orphan route): %s\n' \
          "$(echo "$orphans" | tr '\n' ' ')" >&2
      fi
      if [[ -n "$missing" ]]; then
        printf '  declared in .onk but absent from generated output: %s\n' \
          "$(echo "$missing" | tr '\n' ' ')" >&2
      fi
      parity_failed=1
    fi
  done <<<"$service_listing"

  # Reverse direction: a generated server with no committed .onk is the same
  # orphan defect seen from the other side, and the forward loop cannot see it.
  find "$voxie_dir/api" -type f -name 'service.onk' |
    sed -e "s#^$voxie_dir/api/##" -e 's#/service\.onk$##' | LC_ALL=C sort >"$work/declared_pkgs"
  if [[ -n "$gen_listing" ]]; then
    echo "$gen_listing" |
      sed -e "s#^$gen_go/##" -e 's#/server\.gen\.go$##' | LC_ALL=C sort >"$work/generated_pkgs"
    local unbacked
    unbacked="$(comm -13 "$work/declared_pkgs" "$work/generated_pkgs")"
    if [[ -n "$unbacked" ]]; then
      while IFS= read -r pkg_dir; do
        [[ -n "$pkg_dir" ]] || continue
        echo "::error::$gen_go/$pkg_dir/server.gen.go is generated but no service.onk declares it" >&2
        parity_failed=1
      done <<<"$unbacked"
    fi
  fi
}

output_present=0
if [[ -d "$contracts_src" && -d "$gen_go" ]]; then
  output_present=1
fi

if [[ $output_present -eq 1 ]]; then
  check_parity
fi

# Regenerate only when something is actually wrong. Doing it unconditionally
# would put `go install onek` and a full rebuild in front of every typecheck.
if [[ $parity_failed -ne 0 || $output_present -eq 0 ]]; then
  if [[ $output_present -eq 0 ]]; then
    echo "::error::generated contracts are missing; expected $gen_go and $contracts_src" >&2
  fi

  if [[ $generate_ok -eq 0 ]]; then
    remedy
    exit "$EXIT_STALE"
  fi

  for tool in make go python3; do
    if ! command -v "$tool" >/dev/null 2>&1; then
      fail "$EXIT_TOOLING" "'$tool' is not on PATH; the contracts are stale and cannot be regenerated here, so the check did not run and was not silently skipped"
    fi
  done

  echo "Generating Voxie contracts..." >&2
  if ! make -C "$voxie_dir" generate; then
    fail "$EXIT_GENERATE" "'make -C $voxie_dir generate' failed; see its output above"
  fi
  if [[ ! -d "$contracts_src" || ! -d "$gen_go" ]]; then
    fail "$EXIT_GENERATE" "'make -C $voxie_dir generate' succeeded but left no generated output at $gen_go / $contracts_src"
  fi

  # Re-verify against the sources. Generation is not proof: it proves the
  # generator ran, not that what it emitted matches the committed contracts.
  parity_failed=0
  parity_services=0
  parity_generated=0
  check_parity
  if [[ $parity_failed -ne 0 ]]; then
    echo "::error::regeneration did not resolve the drift above" >&2
    remedy
    exit "$EXIT_STALE"
  fi
fi

echo "Voxie contracts match voxie/api ($parity_services services, $parity_generated generated servers compared)."
exit "$EXIT_OK"
