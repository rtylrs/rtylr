#!/usr/bin/env bash
cat <<'TXT'
CLAUDE.md non-negotiables (re-injected):
- Gates before claiming done: pnpm typecheck && pnpm lint && pnpm test
- NEVER edit apps/hr/**, packages/voxie/gen/**, .env*, or pnpm-lock.yaml by hand
- src/router.tsx is the live route tree; routes/AppRoutes.tsx is dead
- No `any`, no @ts-ignore, explicit return types, Zod at every boundary
- All user-facing strings via t(); no hardcoded hex; mobile-first only
- Never hardcode a business rule - read it from company config
- Never ship a control that does not work (no toast-only handlers)
- Branch from latest main, PR explains WHY. Never push to main.
- Read apps/<app>/CLAUDE.md before working inside an app.
TXT
