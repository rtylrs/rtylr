import { test, expect, type APIRequestContext } from '@playwright/test';

/**
 * Real-backend app-shell check for the suite's main surfaces, opt-in via REAL_API.
 *
 * `dash`, `flow` and `crm` are behind `AppGuard`/`BusinessRoute`: they read a
 * `token` cookie, call `GET /hub/user/v1/users/me` to hydrate the user store,
 * and redirect to the auth app when that fails. Every other spec stubs the API,
 * so none of them prove the guarded shell actually opens against a real
 * backend for a real tenant session.
 *
 * This primes the `token` cookie with a genuine login and asserts each app
 * renders its shell for that tenant instead of bouncing to login. It runs on
 * the desktop project only (single-instance local API) and skips entirely
 * without the harness env, so a default `test:e2e` stays green.
 *
 * Usage:
 *   REAL_API=http://127.0.0.1:9999 \
 *   REAL_DASH_URL=http://localhost:5175 REAL_FLOW_URL=http://localhost:5183 \
 *   REAL_CRM_URL=http://localhost:5181 \
 *   REAL_OWNER_EMAIL=… REAL_OWNER_PASSWORD=… \
 *     npx playwright test e2e/app-shells-real.spec.ts
 */
const REAL_API = process.env.REAL_API;
const OWNER_EMAIL = process.env.REAL_OWNER_EMAIL;
const OWNER_PASSWORD = process.env.REAL_OWNER_PASSWORD;
const APPS = [
  { name: 'dash', url: process.env.REAL_DASH_URL ?? 'http://localhost:5175' },
  { name: 'flow', url: process.env.REAL_FLOW_URL ?? 'http://localhost:5183' },
  { name: 'crm', url: process.env.REAL_CRM_URL ?? 'http://localhost:5181' },
];
const enabled = Boolean(REAL_API && OWNER_EMAIL && OWNER_PASSWORD);

let token: string;

test.describe('App shells open for a real tenant session', () => {
  test.skip(!enabled, 'set REAL_API + REAL_OWNER_* to run against a live API');

  test.beforeAll(async ({ request }, testInfo) => {
    test.skip(
      testInfo.project.name !== 'chromium',
      'desktop project only; single-instance local API'
    );
    const res = await request.post(`${REAL_API}/common/auth/v1/login`, {
      data: { email: OWNER_EMAIL, password: OWNER_PASSWORD },
    });
    if (res.status() !== 200) throw new Error(`login failed: ${res.status()}`);
    token = (await res.json()).token;
  });

  for (const app of APPS) {
    test(`${app.name} renders its shell for a signed-in owner`, async ({ page, baseURL }) => {
      const target = app.url ?? baseURL!;
      // Prime the real session cookie on the app's own origin.
      const { hostname } = new URL(target);
      await page
        .context()
        .addCookies([
          {
            name: 'token',
            value: token,
            domain: hostname,
            path: '/',
            httpOnly: false,
            sameSite: 'Lax',
            secure: false,
          },
        ]);

      // Never let this reach production, even if an app misconfigures its API.
      await page.route(
        (url) => url.hostname.endsWith('voxire.com'),
        (route) => route.abort()
      );

      await page.goto(target, { waitUntil: 'domcontentloaded' });
      await page.waitForTimeout(1500);

      // The shell opened: the user store hydrated, so the guard let us in and
      // the app did not hard-redirect to the auth origin.
      expect(page.url(), `${app.name} redirected away from ${target}`).not.toContain(
        'sso.rtylr.com'
      );
      const body = await page.textContent('body');
      expect(body, `${app.name} rendered an empty shell`).toBeTruthy();
      expect(body, `${app.name} did not hydrate the signed-in user`).not.toContain(
        'undefined user'
      );
    });
  }
});
