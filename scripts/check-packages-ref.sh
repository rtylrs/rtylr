#!/usr/bin/env bash
set -euo pipefail

# Git hooks export GIT_DIR, GIT_WORK_TREE and GIT_INDEX_FILE, and those override
# both `-C` and normal repository discovery. Without unsetting them, every
# `git -C "$repo_root/packages" ...` below silently answers from the PARENT
# repository, so this script compares rtylr's own HEAD against the recorded
# packages commit and can never match.
#
# The effect: this guard has never worked from inside a hook, only when run by
# hand from a shell. It went unnoticed while it was reachable only via
# `pnpm typecheck`. It is on the pre-push path now, where a false failure blocks
# every push and the printed sha is the parent's, which reads as nonsense.
unset GIT_DIR GIT_WORK_TREE GIT_INDEX_FILE

repo_root="$(git rev-parse --show-toplevel)"

recorded_packages="$(git -C "$repo_root" ls-files -s -- packages | awk '$1 == "160000" { print $2; exit }')"
if [[ -z "$recorded_packages" ]]; then
  echo "::error::rtylr/packages is not recorded as a git submodule." >&2
  exit 1
fi

# Confirm packages is its own repository BEFORE asking it for a HEAD.
#
# `git -C <dir> rev-parse HEAD` walks up to the enclosing repository when <dir>
# is not one itself, so an empty or uninitialised packages/ makes the check below
# report rtylr's own HEAD as the submodule's commit. That produced an error
# telling the reader that packages was "checked out at" an rtylr commit sha, and
# sent them hunting for a packages commit that does not exist. The remedy printed
# was right; the sha was nonsense.
if [[ "$(git -C "$repo_root/packages" rev-parse --show-toplevel 2>/dev/null)" != "$repo_root/packages" ]]; then
  echo "::error::rtylr/packages is not a git repository (empty or uninitialised submodule). Run git submodule update --init --recursive." >&2
  exit 1
fi

if ! actual_packages="$(git -C "$repo_root/packages" rev-parse --verify HEAD^{commit} 2>/dev/null)"; then
  echo "::error::rtylr/packages is not initialized at a commit. Run git submodule update --init --recursive." >&2
  exit 1
fi

if [[ "$actual_packages" != "$recorded_packages" ]]; then
  echo "::error::rtylr/packages is checked out at $actual_packages, but rtylr records $recorded_packages. Run git submodule update --init --recursive." >&2
  exit 1
fi

recorded_voxie="$(git -C "$repo_root/packages" ls-files -s -- voxie | awk '$1 == "160000" { print $2; exit }')"
if [[ -z "$recorded_voxie" ]]; then
  echo "::error::packages/voxie is not recorded as a nested git submodule." >&2
  exit 1
fi

if ! actual_voxie="$(git -C "$repo_root/packages/voxie" rev-parse --verify HEAD^{commit} 2>/dev/null)"; then
  echo "::error::packages/voxie is not initialized at a commit. Run git submodule update --init --recursive." >&2
  exit 1
fi

if [[ "$actual_voxie" != "$recorded_voxie" ]]; then
  echo "::error::packages/voxie is checked out at $actual_voxie, but packages records $recorded_voxie. Run git submodule update --init --recursive." >&2
  exit 1
fi

for generated_dir in gen/go gen/ts; do
  if [[ ! -d "$repo_root/packages/voxie/$generated_dir" ]]; then
    echo "::error::packages/voxie@$actual_voxie is missing $generated_dir output." >&2
    exit 1
  fi
done

echo "rtylr/packages is pinned to packages commit $actual_packages"
echo "packages/voxie is pinned to generated Voxie commit $actual_voxie"
