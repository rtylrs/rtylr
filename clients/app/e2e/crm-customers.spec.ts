import { expectAppAction } from './support/app-appearance';
import { test, expect, type Page, type Route } from './fixtures';

/**
 * End-to-end: the CRM customers flow.
 *
 * Covers the five buttons that make up the customer lifecycle: landing on the
 * list empty state, creating a customer via the 3-step wizard, opening the
 * detail page, editing a field (notes/tag) and saving, and deleting with
 * confirmation. Every `/crm/customers*` endpoint is stubbed via `page.route()`
 * so CI never touches the real backend and the test is deterministic.
 *
 * The CRM app is guarded by `AppGuard` (auth token cookie + `/hub/user/v1/users/me`), so
 * we seed a fake cookie per context and stub `/hub/user/v1/users/me`. AppGuard is
 * optimistic on error, but returning a valid user keeps the happy path honest.
 */

const CRM_URL = process.env.CRM_URL ?? 'http://localhost:5197/crm';

// Keys land as snake_case on the wire and are converted to camelCase by the
// shared axios interceptor before the hooks see them. Stubs mirror the
// Go backend shape so the conversion layer exercises the same code path it
// does in production.
const FAKE_USER = {
  id: 1,
  business_id: '1',
  full_name: 'CRM Tester',
  email: 'tester@example.com',
  effective_access: {
    business_id: '1',
    is_owner: true,
    is_admin: false,
    dash_permissions: [],
    subscribed_apps: ['crm'],
    licenses: [],
    authorization_version: '1',
    apps: [
      {
        app_id: 'crm',
        licensed: true,
        has_access: true,
        permissions: [],
        permission_scopes: [],
        sources: ['system'],
      },
    ],
  },
  member_type: 'owner',
  apps: ['crm'],
  plan_slug: 'pro',
  onboarded: true,
};

const FAKE_CUSTOMER = {
  id: 101,
  first_name: 'Jamie',
  last_name: 'Doe',
  name: 'Jamie Doe',
  email: 'jamie@example.com',
  phone: '+1 555-000-0000',
  total_spent: 240,
  visit_count: 3,
  loyalty_points: 120,
  last_visit_at: '2 days ago',
  visits: 3,
  last_visit: '2 days ago',
  initials: 'JD',
  tags: ['vip'],
  notes: '',
  consent: { email: true, sms: false, marketing: true },
  created_at: '2026-01-01',
};

// State held between `page.route()` handlers so we can simulate a real
// list that grows and shrinks across requests. Resetting in beforeEach
// guarantees test isolation - each spec runs against a known baseline.
let customers: Array<Record<string, unknown>> = [];
let lastCreatePayload: Record<string, unknown> | null = null;
let lastUpdatePayload: Record<string, unknown> | null = null;
let lastDeletePayload: Record<string, unknown> | null = null;

async function seedAuth(page: Page): Promise<void> {
  // Seed the cookie before first navigation so AppGuard's synchronous
  // token check passes on the very first render - otherwise it would
  // redirect to sso.rtylr.com and the test would leave the CRM origin.
  await page.context().addCookies([
    {
      name: 'token',
      value: 'test-token',
      domain: 'localhost',
      path: '/',
    },
  ]);
}

async function stubCrmAPI(page: Page): Promise<void> {
  // Catch-all stub - anything that reaches `localhost:9999` without a
  // more-specific handler returns an empty object rather than hitting
  // the real backend. Registered first so specific handlers below take
  // precedence (Playwright runs routes in reverse-registration order).
  await page.route(/localhost:9999\//, (route: Route) =>
    route.fulfill({ status: 200, contentType: 'application/json', body: '{}' })
  );

  // /hub/user/v1/users/me is called by AppGuard to confirm the user has CRM access.
  await page.route(/\/users\/me$/, async (route: Route) => {
    await route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify(FAKE_USER),
    });
  });

  // Dashboard + every sibling list endpoint is stubbed as an empty
  // envelope - the specs only assert on the customers flow, but the
  // DefaultLayout sidebar and prefetched hooks will hit these.
  const emptyList = (key: string): string => JSON.stringify({ [key]: [] });
  await page.route(/\/crm\/dashboard$/, (route: Route) =>
    route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({
        total_customers: 0,
        new_this_month: 0,
        active_loyalty_members: 0,
        campaign_reach: 0,
        total_customers_change: '+0%',
        new_this_month_change: '+0%',
        active_loyalty_members_change: '+0%',
        campaign_reach_change: '+0%',
        top_customers: [],
        recent_campaigns: [],
      }),
    })
  );
  await page.route(/\/crm\/segments(\?|$)/, (route: Route) =>
    route.fulfill({ status: 200, contentType: 'application/json', body: emptyList('segments') })
  );
  await page.route(/\/crm\/campaigns(\?|$)/, (route: Route) =>
    route.fulfill({ status: 200, contentType: 'application/json', body: emptyList('campaigns') })
  );

  // Customer list / create / update / delete. Canonical customer operations
  // use the generated customer/v1 paths; legacy detail helpers remain on the
  // compatibility routes.
  await page.route(
    /^http:\/\/localhost:9999\/crm\/(?:customers|customer\/v1\/customers)(?:\/|\?|$)/,
    async (route: Route) => {
      const req = route.request();
      const url = new URL(req.url());
      const path = url.pathname;
      const isCanonicalCollection = path.endsWith('/crm/customer/v1/customers');
      const isCanonicalDetail = path.startsWith('/crm/customer/v1/customers/');

      if (req.method() === 'DELETE' && isCanonicalDetail) {
        const id = Number(path.split('/').pop());
        lastDeletePayload = { id };
        customers = customers.filter((c) => c.id !== id);
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({ message: 'Customer deleted' }),
        });
        return;
      }

      if (req.method() === 'PUT' && isCanonicalDetail) {
        const id = Number(path.split('/').pop());
        const body = (req.postDataJSON() ?? {}) as Record<string, unknown>;
        lastUpdatePayload = { id, ...body };
        customers = customers.map((customer) =>
          customer.id === id ? { ...customer, ...body } : customer
        );
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({ message: 'Customer updated' }),
        });
        return;
      }

      // /crm/customers/detail?id=101
      if (
        path.endsWith('/crm/customers/detail') ||
        path.endsWith('/crm/customers/detail/full') ||
        (isCanonicalDetail && req.method() === 'GET')
      ) {
        const id = Number(url.searchParams.get('id'));
        const canonicalId = isCanonicalDetail ? Number(path.split('/').pop()) : id;
        const found = customers.find((c) => c.id === (isCanonicalDetail ? canonicalId : id));
        if (!found) {
          await route.fulfill({ status: 404, body: '{}' });
          return;
        }
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify(
            isCanonicalDetail || path.endsWith('/crm/customers/detail/full')
              ? { customer: found }
              : found
          ),
        });
        return;
      }

      if (path.endsWith('/crm/customers/timeline')) {
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({ entries: [] }),
        });
        return;
      }

      if (path.endsWith('/crm/customers') || isCanonicalCollection) {
        if (req.method() === 'GET') {
          await route.fulfill({
            status: 200,
            contentType: 'application/json',
            body: JSON.stringify({
              customers,
              total: String(customers.length),
              page: '1',
              page_size: '100',
            }),
          });
          return;
        }

        if (req.method() === 'POST') {
          const body = (req.postDataJSON() ?? {}) as Record<string, unknown>;
          if (body._action === 'delete') {
            lastDeletePayload = body;
            customers = customers.filter((c) => c.id !== body.id);
            await route.fulfill({
              status: 200,
              contentType: 'application/json',
              body: JSON.stringify({ ok: true }),
            });
            return;
          }
          if (typeof body.id === 'number') {
            lastUpdatePayload = body;
            customers = customers.map((c) => (c.id === body.id ? { ...c, ...body } : c));
            await route.fulfill({
              status: 200,
              contentType: 'application/json',
              body: JSON.stringify(customers.find((c) => c.id === body.id) ?? {}),
            });
            return;
          }

          lastCreatePayload = {
            ...body,
            name: [body.first_name, body.last_name].filter(Boolean).join(' '),
          };
          const created = {
            ...FAKE_CUSTOMER,
            ...body,
            id: FAKE_CUSTOMER.id,
            total_spent: 0,
            visits: 0,
            loyalty_points: 0,
            last_visit: 'Never',
            tags: [],
          };
          customers = [...customers, created];
          await route.fulfill({
            status: 200,
            contentType: 'application/json',
            body: JSON.stringify(
              isCanonicalCollection
                ? { customer_id: String(created.id), message: 'Customer created' }
                : created
            ),
          });
          return;
        }
      }

      // Any other /crm/customers/* endpoint we haven't explicitly mocked.
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: '{}',
      });
    }
  );
}

test.describe('CRM customers flow', () => {
  test.beforeEach(async ({ page }) => {
    customers = [];
    lastCreatePayload = null;
    lastUpdatePayload = null;
    lastDeletePayload = null;
    await seedAuth(page);
    await stubCrmAPI(page);
  });

  test('windows a long customer list and keeps an expanded row in place', async ({ page }) => {
    // Long names only near the top: unlocked, the name column would narrow
    // once those rows scroll away.
    const name = (i: number) =>
      i < 40 ? `Customer ${i} with a much longer name` : `Customer ${i}`;
    customers = Array.from({ length: 150 }, (_, i) => ({
      ...FAKE_CUSTOMER,
      id: 1000 + i,
      first_name: 'Customer',
      last_name: name(i).replace('Customer ', ''),
      name: name(i),
      email: `c${String(i).padStart(3, '0')}@example.com`,
      initials: 'CU',
    }));
    await page.goto(`${CRM_URL}/customers`);
    await expect(page.getByText(name(0), { exact: true })).toBeVisible({ timeout: 15_000 });

    // Each customer is its own tbody; only those near the viewport render.
    const rendered = page.locator('table tbody[data-index]');
    await expect.poll(() => rendered.count()).toBeLessThan(60);
    await expect(page.getByText(name(149), { exact: true })).not.toBeAttached();

    const headers = page.locator('table thead th');
    const widths = () => headers.evaluateAll((cells) => cells.map((c) => c.clientWidth));
    const before = await widths();

    // Expanding a row grows its tbody, which the list measures.
    await page.getByText(name(1), { exact: true }).click();
    await expect(page.getByText('c001@example.com').nth(1)).toBeVisible();

    const main = page.locator('#main-content');
    await main.evaluate((el) => el.scrollTo({ top: el.scrollHeight }));
    await expect(page.getByText(name(149), { exact: true })).toBeVisible();
    await expect(page.getByText(name(0), { exact: true })).not.toBeAttached();
    // Columns hold their width as rows come and go.
    const after = await widths();
    after.forEach((width, i) => expect(Math.abs(width - before[i])).toBeLessThanOrEqual(1));

    await main.evaluate((el) => el.scrollTo({ top: 0 }));
    await expect(page.getByText(name(0), { exact: true })).toBeVisible();
    await expect(page.getByText('c001@example.com').nth(1)).toBeVisible();
  });

  test('empty customer list shows the empty state', async ({ page }) => {
    await page.goto(`${CRM_URL}/customers`);
    await expect(page.getByRole('heading', { name: /No customers found/i })).toBeVisible({
      timeout: 15_000,
    });
    // The empty-state CTA button duplicates the header action - both
    // should land the user on /customers/new.
    await expect(page.getByRole('button', { name: /New customer/i }).first()).toBeVisible();
  });

  test('create a customer via the new-customer wizard', async ({ page }) => {
    await page.goto(`${CRM_URL}/customers`);
    await expect(page.getByRole('heading', { name: /No customers found/i })).toBeVisible({
      timeout: 15_000,
    });

    await page
      .getByRole('button', { name: /New customer/i })
      .first()
      .click();
    await expect(page).toHaveURL(/\/customers\/new$/);

    // Step 1: info - name is required to advance. The Next button
    // only enables after React commits the name-field state. We use
    // keyboard Tab + Enter instead of click() so we don't race the
    // re-render (button transition styles trigger Playwright's
    // stability check which retries against a stale element).
    const nameInput = page.getByPlaceholder('Jane Doe');
    await nameInput.click();
    await nameInput.pressSequentially('Jamie Doe', { delay: 10 });
    await expect(page.getByRole('button', { name: /^Next/i })).toBeEnabled();
    await expectAppAction(page.getByRole('button', { name: /^Next/i }), '#8b5cf6');
    await page.getByRole('button', { name: /^Next/i }).click();

    // Step 2: contact - email or phone required.
    const emailInput = page.getByPlaceholder('jane@example.com');
    await emailInput.click();
    await emailInput.pressSequentially('jamie@example.com', { delay: 10 });
    await expect(page.getByRole('button', { name: /^Next/i })).toBeEnabled();
    await page.getByRole('button', { name: /^Next/i }).click();

    // Step 3: consent - submit.
    await expect(page.getByRole('button', { name: /^Create customer/i })).toBeEnabled();
    await page.getByRole('button', { name: /^Create customer/i }).click();

    // Wizard returns to the list on success; the stub injected our fake
    // customer, so the empty state is gone.
    await expect(page).toHaveURL(/\/customers$/, { timeout: 10_000 });
    await expect(page.getByText('Jamie Doe')).toBeVisible({ timeout: 10_000 });

    // Sanity: the payload matches what the UI promised.
    expect(lastCreatePayload).toMatchObject({
      name: 'Jamie Doe',
      email: 'jamie@example.com',
    });
  });

  test('open customer detail, edit notes, and delete', async ({ page }) => {
    // Seed a pre-existing customer so the detail page has something to
    // load - mirrors the "you already have customers" happy path.
    customers = [{ ...FAKE_CUSTOMER }];

    await page.goto(`${CRM_URL}/customers/${FAKE_CUSTOMER.id}`);
    await expect(page.getByRole('heading', { name: FAKE_CUSTOMER.name, level: 2 })).toBeVisible({
      timeout: 15_000,
    });

    // Add a tag - triggers the update mutation with tags[].
    await page.getByPlaceholder('Add tag...').fill('e2e');
    await page.getByRole('button', { name: /^Add$/ }).click();

    await expect(async () => {
      expect(lastUpdatePayload).not.toBeNull();
      expect(lastUpdatePayload).toMatchObject({ id: FAKE_CUSTOMER.id });
      expect(Array.isArray((lastUpdatePayload as Record<string, unknown>).tags)).toBe(true);
    }).toPass({ timeout: 5_000 });

    // Delete - the confirm modal must open, confirm dispatches the
    // mutation, and the app should navigate back to /customers.
    // The header-action "Delete" button is the trigger; the modal's
    // footer "Delete" is the confirm. After the modal opens there are
    // two Delete buttons; we click the last one (the confirm, which
    // renders below the trigger in DOM order).
    await page
      .getByRole('button', { name: /^Delete$/i })
      .first()
      .click();
    await expect(page.getByRole('heading', { name: /Delete customer\?/i })).toBeVisible();
    const deleteButtons = page.getByRole('button', { name: /^Delete$/i });
    await deleteButtons.last().click();

    await expect(page).toHaveURL(/\/customers$/, { timeout: 10_000 });
    expect(lastDeletePayload).toMatchObject({ id: FAKE_CUSTOMER.id });
  });
});
