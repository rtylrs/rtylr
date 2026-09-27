# Rtylr 1:1 — real-backend verification harness

This directory contains opt-in Playwright specs that exercise rtylr against a
**real** API, real MySQL schema and real bcrypt/JWT auth. Every other spec in
this folder stubs `page.route`, which is right for CI and wrong for proving
that authentication, the JWT tenant claim, and the users↔business↔app
authorization boundary actually work end to end.

## What these specs prove

| Spec                      | Checks                                                                                                                                                                                                                                                                                                                                                                                            |
| ------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `auth-real-login.spec.ts` | a seeded user signs in through the real form; the JWT session carries the right business; a wrong password is rejected (401); one tenant cannot read another's business; one tenant's CRM customer list contains none of the other tenant's rows; **every licensed app (pos, erp, hr, finance, insights, recruit, menu, crm, flow) opens its data plane for a licensed owner** — none returns 403 |
| `app-shells-real.spec.ts` | the `dash`, `crm`, `flow`, `hr`, `erp`, `finance`, `menu`, `recruit` and `insights` shells open for a genuinely signed-in owner (they sit behind `AppGuard`/`BusinessRoute` and redirect to the auth app when the session is not real). POS has no Pages target, so it is covered at the API layer instead                                                                                        |

## Coverage notes

The shell checks assert that a real session hydrates and the guard lets the app
through; they are a session/guard check, **not** an exhaustive per-button walk of
every screen. Object storage (S3/floci), image transforms (media-engine), email
delivery, realtime/websockets and Redis cache behaviour are **not** exercised —
they need services this harness does not stand up. See
`services/onion/docs/real-backend-setup.md`.

## Running them

They **skip by default**, so the normal `pnpm test:e2e` run is unaffected. To
run them you need a live API with a migrated, seeded database and the apps
serving on their dev ports.

```sh
# 1. A live API + seeded DB. One way (no Docker required) is described in
#    docs/real-backend-setup.md.
# 2. Sign-in as two users from different tenants, then:
REAL_API=http://127.0.0.1:9999 \
AUTH_URL=http://127.0.0.1:5176 \
REAL_OWNER_EMAIL=owner@example.test  REAL_OWNER_PASSWORD=… \
REAL_OTHER_EMAIL=other@example.test  REAL_OTHER_PASSWORD=… \
  npx playwright test e2e/auth-real-login.spec.ts
```

`app-shells-real.spec.ts` additionally reads `REAL_DASH_URL`, `REAL_FLOW_URL`
and `REAL_CRM_URL` (defaults `5175`, `5183`, `5181`) — the ports the Playwright
`webServer` array already boots.

## Design notes

- **They import `test` from `@playwright/test`, not `./fixtures`.** The shared
  fixture installs a context route that returns `TEST_LEAK` for anything that
  reaches `localhost:9999`. That guard is correct for the stubbed suite and
  would abort the very request a real-backend test is trying to make, so these
  specs re-establish isolation themselves by blocking any non-local host. The
  property that matters — never contact production — is preserved.
- **They log in once in `beforeAll` and run on the desktop project only.** The
  login route is rate limited per IP (5/min) and a local backend is a single
  instance, so logging in per-test (or under both the desktop and 390px
  projects) trips the limit.
- **No credentials are committed.** Every user, password and URL comes from the
  environment.

See `docs/real-backend-setup.md` for how to bring up MySQL + the API without
Docker, including the migrations, the seeds, and the tenant/authorization rows
these specs need.
