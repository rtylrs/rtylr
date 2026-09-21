#!/usr/bin/env bash
#
# Promote a landing release tag to production.
#
# Production ships when apps/rtylr-landing.yaml in tryrtylr/deploy names a
# tag and that change lands on main: validate-and-dispatch.yml there sends
# deploy-release back to this repo. This script is that edit, done the same
# way every time, and it is the step the blog publisher calls once its own
# push has been tagged.
#
#   scripts/promote-landing.sh rtylr-landing-v0.2.1
#   scripts/promote-landing.sh rtylr-landing-v0.2.1 --dry-run
#
# --dry-run does everything except commit and push, and prints the diff. Use
# it to see what would ship.
#
# The tag must already exist on this repo's origin. No green Landing run means
# no tag, and no tag means this script refuses, which is the whole gate.

set -euo pipefail

DEPLOY_REPO="tryrtylr/deploy"
MANIFEST="apps/rtylr-landing.yaml"

die() { printf 'promote-landing: %s\n' "$1" >&2; exit 1; }

TAG=""
DRY_RUN=0
for arg in "$@"; do
  case "$arg" in
    --dry-run) DRY_RUN=1 ;;
    -*) die "unknown option: $arg" ;;
    *) [[ -z "$TAG" ]] || die "give exactly one tag"; TAG="$arg" ;;
  esac
done

[[ -n "$TAG" ]] || die "usage: scripts/promote-landing.sh <release_tag> [--dry-run]"

[[ "$TAG" =~ ^rtylr-landing-v([0-9]+\.[0-9]+\.[0-9]+)$ ]] ||
  die "not a landing release tag: $TAG (expected rtylr-landing-vMAJOR.MINOR.PATCH)"
VERSION="${BASH_REMATCH[1]}"

# The gate. A tag that is not on origin was never produced by a green run.
git rev-parse --git-dir >/dev/null 2>&1 || die "run this from inside the rtylr checkout"
git ls-remote --exit-code --tags origin "refs/tags/${TAG}" >/dev/null 2>&1 ||
  die "tag ${TAG} does not exist on origin. Nothing green produced it, so there is nothing to promote."

WORKDIR="$(mktemp -d)"
trap 'rm -rf "$WORKDIR"' EXIT

echo "promote-landing: cloning ${DEPLOY_REPO}"
if command -v gh >/dev/null 2>&1; then
  gh repo clone "$DEPLOY_REPO" "$WORKDIR/deploy" -- --quiet --depth 1 >/dev/null
else
  git clone --quiet --depth 1 "https://github.com/${DEPLOY_REPO}.git" "$WORKDIR/deploy"
fi

cd "$WORKDIR/deploy"
[[ -f "$MANIFEST" ]] || die "${MANIFEST} is not in ${DEPLOY_REPO}"

CURRENT_TAG="$(sed -nE 's/^release_tag: "(.*)"$/\1/p' "$MANIFEST")"
[[ -n "$CURRENT_TAG" ]] ||
  die "could not read release_tag from ${MANIFEST}. Its format changed; fix this script rather than the manifest."

if [[ "$CURRENT_TAG" == "$TAG" ]]; then
  echo "promote-landing: ${MANIFEST} already names ${TAG}. Nothing to do."
  exit 0
fi

sed -i.bak -E "s|^version: \".*\"$|version: \"${VERSION}\"|" "$MANIFEST"
sed -i.bak -E "s|^release_tag: \".*\"$|release_tag: \"${TAG}\"|" "$MANIFEST"
rm -f "${MANIFEST}.bak"

# A sed that matched nothing exits 0. Verify, or a format change here would
# report a successful promotion that shipped the previous tag.
grep -qxF "version: \"${VERSION}\"" "$MANIFEST" || die "version was not written to ${MANIFEST}"
grep -qxF "release_tag: \"${TAG}\"" "$MANIFEST" || die "release_tag was not written to ${MANIFEST}"

if [[ -f scripts/validate_manifests.rb ]]; then
  echo "promote-landing: validating manifests"
  ruby scripts/validate_manifests.rb
else
  die "scripts/validate_manifests.rb is missing from ${DEPLOY_REPO}. Refusing to push an unvalidated manifest."
fi

echo "promote-landing: ${CURRENT_TAG} -> ${TAG}"
git --no-pager diff -- "$MANIFEST"

if (( DRY_RUN )); then
  echo "promote-landing: dry run, nothing committed or pushed."
  exit 0
fi

git add "$MANIFEST"
git commit --quiet -m "Promote rtylr-landing to ${TAG}"
git push --quiet origin HEAD:main

echo "promote-landing: pushed. Watch deploy-release in tryrtylr/rtylr Actions."
