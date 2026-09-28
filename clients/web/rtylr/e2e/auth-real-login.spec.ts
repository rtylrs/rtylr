import {
  test,
  expect,
  type Route,
  type APIRequestContext,
  type APIResponse,
} from '@playwright/test';

/**
 * Real-backend auth + tenant-isolation check, opt-in via REAL_API plus the
 * REAL_OWNER_EMAIL, REAL_OWNER_PASSWORD, REAL_OTHER_EMAIL, REAL_OTHER_PASSWORD
 * and optional REAL_VIEWER_EMAIL, REAL_VIEWER_PASSWORD credentials. It skips
 * without them, so a default `test:e2e` stays green and no secret is committed.
 *
 * Imports `test` from `@playwright/test` rather than `./fixtures` on purpose:
 * the shared fixture routes anything reaching `localhost:9999` to `TEST_LEAK`,
 * which would abort the very request under test. Isolation is re-established
 * per test by aborting non-local hosts, so this can never contact production.
 *
 * The login route allows 5 requests/minute per IP, so every credential login
 * happens in `beforeAll` and a 429 is retried once after `retry_after`.
 */
const REAL_API = process.env.REAL_API;
const AUTH_URL = process.env.AUTH_URL ?? 'http://localhost:5176';
const OWNER_EMAIL = process.env.REAL_OWNER_EMAIL;
const OWNER_PASSWORD = process.env.REAL_OWNER_PASSWORD;
const OTHER_EMAIL = process.env.REAL_OTHER_EMAIL;
const OTHER_PASSWORD = process.env.REAL_OTHER_PASSWORD;
const VIEWER_EMAIL = process.env.REAL_VIEWER_EMAIL;
const VIEWER_PASSWORD = process.env.REAL_VIEWER_PASSWORD;

const enabled = Boolean(REAL_API && OWNER_EMAIL && OWNER_PASSWORD && OTHER_EMAIL && OTHER_PASSWORD);

interface Tenant {
  token: string;
  me: {
    id: number;
    business_id: string;
    business_name: string;
    email: string;
    is_owner: boolean;
    apps?: string[];
  };
}

let owner: Tenant;
let other: Tenant;
let viewer: Tenant | undefined;
let grantedApps: string[] = [];

// The login route is limited to 5 requests/minute per IP
// (PerRouteRateLimit(5, time.Minute)). This file needs several logins in a
// single run, so a 429 is an expected environmental condition rather than a
// broken fixture. Retrying once after the advertised retry_after keeps the
// suite reliable; every other status is returned immediately, so bad
// credentials or a backend regression still fail loudly instead of skipping.
const LOGIN_RATE_LIMIT_WAIT_MS = 61_000;

async function postLogin(
  request: APIRequestContext,
  email: string,
  password: string
): Promise<APIResponse> {
  let res = await request.post(`${REAL_API}/common/auth/v1/login`, {
    data: { email, password },
  });
  if (res.status() === 429) {
    await new Promise((resolve) => setTimeout(resolve, LOGIN_RATE_LIMIT_WAIT_MS));
    res = await request.post(`${REAL_API}/common/auth/v1/login`, {
      data: { email, password },
    });
  }
  return res;
}

async function loginAs(
  request: APIRequestContext,
  email: string,
  password: string
): Promise<Tenant> {
  const res = await postLogin(request, email, password);
  if (res.status() !== 200) {
    throw new Error(`login failed for ${email}: ${res.status()} ${await res.text()}`);
  }
  // A 200 without a usable token would otherwise become `Bearer undefined` and
  // the protected assertions would run against an anonymous path.
  const { token } = (await res.json()) as { token?: unknown };
  if (typeof token !== 'string' || token.length === 0) {
    throw new Error(`login for ${email} returned no usable token`);
  }
  const meRes = await request.get(`${REAL_API}/hub/user/v1/users/me`, {
    headers: { Authorization: `Bearer ${token}` },
  });
  if (meRes.status() !== 200) {
    throw new Error(`users/me failed for ${email}: ${meRes.status()} ${await meRes.text()}`);
  }
  return { token, me: await meRes.json() };
}

test.describe('Auth real-backend login + tenant isolation', () => {
  test.skip(!enabled, 'set REAL_API + REAL_* credentials to run against a live API');

  test.beforeAll(async ({ request }, testInfo) => {
    // A single-instance local API is rate limited to 5 logins/min per IP. Every
    // credential login therefore happens here, back to back, so the count is
    // fixed and known rather than depending on test order. Running it under the
    // 390px project too would double the logins and 429 that limit, so it runs
    // once, on the desktop project. Viewport is not what is under test here.
    // A rate-limited retry can wait past the 60s default hook timeout.
    test.setTimeout(180_000);
    test.skip(
      testInfo.project.name !== 'chromium',
      'desktop project only; single-instance local API'
    );
    owner = await loginAs(request, OWNER_EMAIL!, OWNER_PASSWORD!);
    other = await loginAs(request, OTHER_EMAIL!, OTHER_PASSWORD!);
    if (VIEWER_EMAIL && VIEWER_PASSWORD) {
      viewer = await loginAs(request, VIEWER_EMAIL, VIEWER_PASSWORD);
    }
    grantedApps = owner.me.apps ?? [];
  });

  test.beforeEach(async ({ page }) => {
    // Block any production host so a live run can never touch prod. Local API
    // and same-origin app assets pass through untouched.
    await page.route(
      (url) => !url.hostname.startsWith('127.0.0.1') && !url.hostname.startsWith('localhost'),
      (route: Route) => route.abort()
    );
  });

  test('a seeded user signs in through the real form and the session carries the tenant', async ({
    page,
    request,
  }) => {
    const failures: string[] = [];
    page.on('console', (m) => {
      if (m.type() === 'error') failures.push(`console: ${m.text()}`);
    });
    page.on('response', (r) => {
      if (r.url().includes('/common/auth/v1/login')) failures.push(`login ${r.status()}`);
    });

    await page.goto(AUTH_URL);
    const email = page.getByRole('textbox', { name: /email/i }).first();
    await expect(email).toBeVisible({ timeout: 15_000 });
    await email.fill(OWNER_EMAIL!);
    await page.locator('input[type="password"]').first().fill(OWNER_PASSWORD!);
    await page
      .getByRole('button', { name: /sign ?in|log ?in|continue/i })
      .first()
      .click();

    // The token cookie is the stable signal the API accepted a real login.
    await expect
      .poll(async () => (await page.context().cookies()).find((c) => c.name === 'token')?.value, {
        timeout: 15_000,
        message: `no token cookie. observed: ${failures.join(' | ') || '(nothing)'}`,
      })
      .toBeTruthy();

    // Bind the assertion to the session this form just created. Checking the
    // cookie exists is not enough: an arbitrary or stale value would satisfy
    // the poll above, and `owner.me` came from a separate API login in
    // beforeAll, so neither proves the form's cookie authenticates. Exchange the
    // browser's own cookie for an identity and assert it is the owner tenant.
    const browserToken = (await page.context().cookies()).find((c) => c.name === 'token')?.value;
    const browserMe = await request.get(`${REAL_API}/hub/user/v1/users/me`, {
      headers: { Authorization: `Bearer ${browserToken}` },
    });
    expect(
      browserMe.status(),
      `the cookie written by the sign-in form did not authenticate: ${await browserMe.text()}`
    ).toBe(200);
    const browserIdentity = (await browserMe.json()) as {
      email: string;
      business_id: string;
    };
    expect(browserIdentity.email, 'the browser session is not the owner').toBe(OWNER_EMAIL);
    expect(
      browserIdentity.business_id,
      'the browser session did not resolve to the owner tenant'
    ).toBe(owner.me.business_id);
    expect(owner.me.is_owner).toBe(true);
    expect(Number(owner.me.business_id)).toBeGreaterThan(0);
  });

  test('a wrong password is rejected by the real API', async ({ request }) => {
    const res = await postLogin(request, OWNER_EMAIL!, 'definitely-not-the-password');
    expect(res.status()).toBe(401);
  });

  test('tenants are isolated: neither user can read the other business', async ({ request }) => {
    // Distinct tenants.
    expect(other.me.business_id).not.toBe(owner.me.business_id);

    // Each can read their own business.
    const ownRes = await request.get(
      `${REAL_API}/hub/business/v1/businesses/${owner.me.business_id}`,
      {
        headers: { Authorization: `Bearer ${owner.token}` },
      }
    );
    expect(ownRes.status()).toBe(200);

    // The other tenant cannot. Assert the exact authorization response: a bare
    // `>= 400` would accept a 500/502 panic or outage as proof of isolation.
    const crossRes = await request.get(
      `${REAL_API}/hub/business/v1/businesses/${owner.me.business_id}`,
      { headers: { Authorization: `Bearer ${other.token}` } }
    );
    expect(crossRes.status(), 'cross-tenant business read was not refused with 403').toBe(403);
    const crossBody = await crossRes.text();
    expect(crossBody).not.toContain(owner.me.business_name);
  });

  test('tenant data plane is isolated (CRM customers)', async ({ request }) => {
    const listOwn = await request.get(`${REAL_API}/crm/customers?page_size=50`, {
      headers: { Authorization: `Bearer ${owner.token}` },
    });
    test.skip(listOwn.status() !== 200, 'owner lacks a CRM license in this fixture set');
    const own = await listOwn.json();

    // The other tenant's CRM list must not contain any of the owner's customers.
    const listOther = await request.get(`${REAL_API}/crm/customers?page_size=50`, {
      headers: { Authorization: `Bearer ${other.token}` },
    });
    test.skip(listOther.status() !== 200, 'second tenant lacks a CRM license in this fixture set');
    const otherList = await listOther.json();

    const ownIds = new Set((own.customers ?? []).map((c: { id: string }) => String(c.id)));
    const leaked = (otherList.customers ?? []).filter((c: { id: string }) =>
      ownIds.has(String(c.id))
    );
    expect(leaked, 'a customer owned by one tenant appeared in another tenant list').toEqual([]);
  });

  test('every licensed app opens its data plane for the owner', async ({ request }) => {
    // One real read route per app, taken from the generated contracts. A 403
    // here means the app is licensed in the manifest but the resolver still
    // denies it — the entitlement boundary failing, which is the bug this
    // guards against. Anything that is not 403 proves the app's HTTP boundary
    // let a genuine owner through (a 200 with data, or a 404/400 for an empty
    // or under-filled route, both of which mean the request reached the app).
    const routes: Array<[string, string]> = [
      ['pos', '/pos/loyalty/v1/balance'],
      ['erp', '/erp/warehouse/v1/warehouses'],
      ['hr', '/hr/work-schedule/v1/work-schedules'],
      ['finance', '/finance/tax/v1/tax'],
      ['insights', '/insights/dashboard/v1/summary'],
      ['recruit', '/recruit/career/v1/submissions'],
      ['menu', '/hub/menu/v1/menu'],
      ['crm', '/crm/customers?page_size=1'],
      ['flow', '/flow/task/v2/tasks/my'],
    ];
    const denied: string[] = [];
    for (const [app, path] of routes) {
      if (!grantedApps.includes(app)) continue;
      const res = await request.get(`${REAL_API}${path}`, {
        headers: { Authorization: `Bearer ${owner.token}` },
      });
      if (res.status() === 403) denied.push(`${app} (${path})`);
    }
    expect(denied, 'licensed apps still denied by the app-access boundary').toEqual([]);
  });

  test('a limited (viewer) user is denied writes and unlicensed apps', async ({ request }) => {
    // A viewer exercises the deny paths an owner never can. This test needs a
    // third seeded identity; skip (not fail) when it is absent so a narrower
    // fixture set does not go red. When REAL_VIEWER_* *is* set the login already
    // happened in beforeAll and loginAs throws on failure, so a broken fixture
    // or a backend regression fails the suite instead of being skipped.
    test.skip(!VIEWER_EMAIL || !VIEWER_PASSWORD, 'set REAL_VIEWER_* for a limited-role user');
    if (!viewer) {
      throw new Error('viewer token missing: beforeAll did not authenticate the configured viewer');
    }
    const auth = { Authorization: `Bearer ${viewer.token}` };

    // Reads the viewer is entitled to must succeed: crm.viewer carries
    // crm:customers:view. This is the regression guard for the entitlement
    // source, which must be the member's role assignments rather than the
    // legacy user_app_permissions.apps column.
    const read = await request.get(`${REAL_API}/crm/customers?page_size=5`, {
      headers: auth,
    });
    expect(read.status(), 'a viewer was denied a read it is entitled to').toBe(200);

    // Writes must be refused: the viewer role has no :create/:edit/:delete.
    const write = await request.post(`${REAL_API}/crm/customers`, {
      headers: auth,
      data: { first_name: 'Nope', last_name: 'Nope', email: 'nope@viewer.test' },
    });
    expect(write.status(), 'a viewer was allowed to create a customer').toBe(403);

    // An app the viewer is not licensed for must be refused.
    const unlicensed = await request.get(`${REAL_API}/erp/warehouse/v1/warehouses`, {
      headers: auth,
    });
    expect(unlicensed.status(), 'a viewer reached an unlicensed app').toBe(403);

    // And the viewer must not be an owner of the tenant. The field is omitted
    // (undefined) rather than false for a non-owner, so assert falsy.
    const me = await request.get(`${REAL_API}/hub/user/v1/users/me`, { headers: auth });
    const body = await me.json();
    expect(Boolean(body.is_owner), 'a viewer was reported as the tenant owner').toBe(false);
  });
});
