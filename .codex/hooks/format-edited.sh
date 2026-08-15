#!/usr/bin/env bash
set -uo pipefail
FILE=$(node -e 'let s="";process.stdin.on("data",d=>s+=d).on("end",()=>{try{console.log(JSON.parse(s).tool_input?.file_path||"")}catch{console.log("")}})')
case "$FILE" in
  *.ts|*.tsx|*.js|*.jsx|*.json|*.css|*.md) npx --no-install prettier --write "$FILE" >/dev/null 2>&1 || true ;;
esac
exit 0
