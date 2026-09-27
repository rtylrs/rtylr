import { test, expect } from '@playwright/test';

/**
 * Real-backend app-shell check for every deployable app, opt-in via REAL_API.
 *
 * `dash`, `flow`, `crm`, `hr`, `erp`, `finance`, `menu`, `recruit` and
 * `insights` are all behind `AppGuard`/`BusinessRoute`: they read a `token`
 * cookie, call `GET /hub/user/v1/users/me` to hydrate the user store, and
 * redirect to the auth app when that fails. Every other spec in this folder
 * stubs the API, so none of them prove a guarded shell actually opens against
 * a real backend for a real tenant session.
 *
 * This primes the `token` cookie with a genuine login and asserts each app
 * renders its shell for that tenant instead of bouncing to login. POS has no
 * Pages target in the Playwright `webServer` array, so it is covered by
 * `auth-real-login.spec.ts` (its API) rather than here.
 *
 * It runs on the desktop project only (single-instance local API) and skips
 * without the harness env, so a default `test:e2e` stays green.
 *
 * Usage:
 *   REAL_API=http://127.0.0.1:9999 REAL_OWNER_EMAIL=… REAL_OWNER_PASSWORD=… \
 *     npx playwright test e2e/app-shells-real.spec.ts
 */
const REAL_API = process.env.REAL_API;
const OWNER_EMAIL = process.env.REAL_OWNER_EMAIL;
const OWNER_PASSWORD = process.env.REAL_OWNER_PASSWORD;

const APPS = [
  { name: 'dash', url: process.env.REAL_DASH_URL ?? 'http://localhost:5175' },
  { name: 'crm', url: process.env.REAL_CRM_URL ?? 'http://localhost:5181' },
  { name: 'flow', url: process.env.REAL_FLOW_URL ?? 'http://localhost:5183' },
  { name: 'hr', url: process.env.REAL_HR_URL ?? 'http://localhost:5180' },
  { name: 'erp', url: process.env.REAL_ERP_URL ?? 'http://localhost:5179' },
  { name: 'finance', url: process.env.REAL_FINANCE_URL ?? 'http://localhost:5182' },
  { name: 'menu', url: process.env.REAL_MENU_URL ?? 'http://localhost:5174' },
  { name: 'recruit', url: process.env.REAL_RECRUIT_URL ?? 'http://localhost:5184' },
  { name: 'insights', url: process.env.REAL_INSIGHTS_URL ?? 'http://localhost:5187' },
];
const enabled = Boolean(REAL_API && OWNER_EMAIL && OWNER_PASSWORD);

let token: string;
let grantedApps: string[] = [];

test.describe('Every app shell opens for a real tenant session', () => {
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
    const me = await request.get(`${REAL_API}/hub/user/v1/users/me`, {
      headers: { Authorization: `Bearer ${token}` },
    });
    grantedApps = (await me.json()).apps ?? [];
  });

  for (const app of APPS) {
    test(`${app.name} renders its shell for a signed-in owner`, async ({ page }) => {
      // Skip (do not fail) when the fixture set does not grant this app, so a
      // narrower seed does not produce a false red.
      test.skip(
        !grantedApps.includes(app.name),
        `owner is not granted ${app.name} in this fixture set`
      );

      const target = app.url;
      const { hostname } = new URL(target);
      await page.context().addCookies([
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
