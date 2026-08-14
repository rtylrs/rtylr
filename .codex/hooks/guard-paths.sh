#!/usr/bin/env bash
# Blocks writes that CLAUDE.md forbids. Exit 2 = block; stderr is fed back to Claude.
set -uo pipefail
FILE=$(node -e 'let s="";process.stdin.on("data",d=>s+=d).on("end",()=>{try{console.log(JSON.parse(s).tool_input?.file_path||"")}catch{console.log("")}})')
[ -z "$FILE" ] && exit 0
REL=${FILE#"$CLAUDE_PROJECT_DIR"/}

block () { echo "BLOCKED by .claude/hooks/guard-paths.sh: $1" >&2; exit 2; }

case "$REL" in
  packages/voxie/gen/*)     block "generated contract output. Change the .onk source in the voxie repo and regenerate." ;;
  .env|.env.local|.env.*)   block "env files are never edited by an agent." ;;
  pnpm-lock.yaml)           block "never hand-edit the lockfile. Use pnpm to add or remove a dependency." ;;
  */routes/AppRoutes.tsx)   block "AppRoutes.tsx is dead code. The live route tree is src/router.tsx." ;;
esac
exit 0
