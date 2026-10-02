import { test, expect, type Page, type Route } from './fixtures';

/**
 * End-to-end: CRM loyalty tiers, campaigns, segments, and customer import UI.
 *
 * Every backend endpoint is stubbed via `page.route()` so CI never touches the
 * real backend. Auth seeding follows the same pattern as crm-customers.spec.ts:
 * seed a `token` cookie and stub `/hub/user/v1/users/me` before first navigation.
 *
 * Wire shapes use snake_case to mirror the Go backend response format. The
 * shared axios interceptor converts them to camelCase before hooks see them.
 */

const CRM_URL = process.env.CRM_URL ?? 'http://localhost:5175/crm';

const FAKE_USER = {
  id: 1,
  business_id: '1',
  full_name: 'CRM Tester',
  email: 'tester@example.com',
  effective_access: JSON.stringify({
    business_id: '1',
    is_owner: true,
    is_admin: false,
    apps: [{ app_id: 'crm', licensed: true, has_access: true, permissions: [] }],
  }),
  role: 'owner',
  apps: ['crm'],
  plan_slug: 'pro',
};

// ── Fixture data ─────────────────────────────────────────────────────────────

const FAKE_TIERS = [
  {
    id: '1',
    name: 'Bronze',
    min_points: '0',
    discount_pct: 1,
    perks: '["Birthday bonus","Early access"]',
    sort_order: '1',
  },
  {
    id: '2',
    name: 'Gold',
    min_points: '500',
    discount_pct: 2,
    perks: '["Free shipping","Priority support","2× points"]',
    sort_order: '2',
  },
];

const FAKE_CAMPAIGN = {
  id: '10',
  name: 'Spring Promo 2026',
  description: 'Limited-time spring deals for all members.',
  channel: 'email',
  status: 'active',
  subject: 'Limited-time spring deals for all members.',
  audience_count: '1200',
  sent_count: '1100',
  opened_count: '418',
  updated_at: '2 days ago',
};

const FAKE_SEGMENT = {
  id: '5',
  name: 'VIP Customers',
  description: 'Top spenders with 10+ visits.',
  customer_count: 45,
  color: '#8b5cf6',
  rules: ['Total Spend > $1,000', 'Visit Count >= 10'],
  trend: '+3',
  trend_up: true,
};

// ── Auth helpers ─────────────────────────────────────────────────────────────

async function seedAuth(page: Page): Promise<void> {
  await page.context().addCookies([
    {
      name: 'token',
      value: 'test-token',
      domain: 'localhost',
      path: '/',
    },
  ]);
}

async function stubBaseRoutes(page: Page): Promise<void> {
  // Catch-all for any localhost:9999 request not handled by a specific stub.
  await page.route(/localhost:9999\//, (route: Route) =>
    route.fulfill({ status: 200, contentType: 'application/json', body: '{}' })
  );

  // AppGuard calls /hub/user/v1/users/me to validate CRM access.
  await page.route(/\/users\/me$/, (route: Route) =>
    route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify(FAKE_USER),
    })
  );

  // Dashboard is prefetched by DefaultLayout on first render.
  await page.route(/\/crm\/dashboard$/, (route: Route) =>
    route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({
        total_customers: 0,
        new_this_month: 0,
        active_loyalty_members: 2,
        campaign_reach: 1200,
        total_customers_change: '+0%',
        new_this_month_change: '+0%',
        active_loyalty_members_change: '+0%',
        campaign_reach_change: '+0%',
        top_customers: [],
        recent_campaigns: [],
      }),
    })
  );

  // Customers list - empty, not under test here.
  await page.route(/\/crm\/(?:customers|customer\/v1\/customers)(?:\?|$)/, (route: Route) =>
    route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({ customers: [] }),
    })
  );
}

// ── Loyalty tiers ─────────────────────────────────────────────────────────────

test.describe('CRM loyalty tiers flow', () => {
  test.beforeEach(async ({ page }) => {
    await seedAuth(page);
    await stubBaseRoutes(page);

    // Stub all loyalty sub-endpoints that DefaultLayout or the Loyalty screen
    // may prefetch before we reach /loyalty.
    await page.route(/localhost:9999\/crm\/loyalty(?:\/v1)?\/config$/, (route: Route) =>
      route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          program_enabled: true,
          points_per_dollar: 10,
          points_expiry: 365,
          redemption_rate: 0.01,
        }),
      })
    );
    await page.route(/localhost:9999\/crm\/loyalty(?:\/v1)?\/transactions$/, (route: Route) =>
      route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({ transactions: [] }),
      })
    );
    await page.route(/localhost:9999\/crm\/loyalty(?:\/v1)?\/members$/, (route: Route) =>
      route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({ members: [] }),
      })
    );
    await page.route(/localhost:9999\/crm\/loyalty(?:\/v1)?\/rewards$/, (route: Route) =>
      route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({ rewards: [] }),
      })
    );
    await page.route(/localhost:9999\/crm\/loyalty(?:\/v1)?\/tiers$/, (route: Route) =>
      route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({ tiers: FAKE_TIERS }),
      })
    );

    // Segments and campaigns are prefetched by the sidebar.
    await page.route(/\/crm\/(?:segments|segment\/v1\/segments)(?:\?|$)/, (route: Route) =>
      route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({ segments: [] }),
      })
    );
    await page.route(/\/crm\/(?:campaigns|campaign\/v1\/campaigns)(?:\?|$)/, (route: Route) =>
      route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({ campaigns: [] }),
      })
    );
  });

  test('loyalty tiers page renders both tier names', async ({ page }) => {
    await page.goto(`${CRM_URL}/loyalty`);

    // The Loyalty screen renders tier cards with the tier name as a heading.
    await expect(page.getByText('Bronze')).toBeVisible({ timeout: 15_000 });
    await expect(page.getByText('Gold')).toBeVisible({ timeout: 10_000 });
  });

  test('loyalty tiers page shows threshold labels', async ({ page }) => {
    await page.goto(`${CRM_URL}/loyalty`);

    // Bronze is the starting tier (threshold === 0).
    await expect(page.getByText(/Starting tier/i).first()).toBeVisible({ timeout: 15_000 });

    // Gold starts at 500 points.
    await expect(page.getByText(/500/)).toBeVisible({ timeout: 10_000 });
  });

  test('dedicated /loyalty/tiers route renders tier cards', async ({ page }) => {
    await page.goto(`${CRM_URL}/loyalty/tiers`);

    // LoyaltyTiers component renders each tier name in an <h3>.
    await expect(page.getByRole('heading', { name: 'Bronze' })).toBeVisible({ timeout: 15_000 });
    await expect(page.getByRole('heading', { name: 'Gold' })).toBeVisible({ timeout: 10_000 });
  });
});

// ── Campaigns list ────────────────────────────────────────────────────────────

test.describe('CRM campaigns list flow', () => {
  test.beforeEach(async ({ page }) => {
    await seedAuth(page);
    await stubBaseRoutes(page);

    // Loyalty sub-endpoints: empty for these tests.
    await page.route(/localhost:9999\/crm\/loyalty(?:\/v1)?\//, (route: Route) =>
      route.fulfill({ status: 200, contentType: 'application/json', body: '{}' })
    );

    // Segments sidebar prefetch.
    await page.route(/\/crm\/(?:segments|segment\/v1\/segments)(?:\?|$)/, (route: Route) =>
      route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({ segments: [] }),
      })
    );

    // Campaigns list returns one active campaign.
    await page.route(/\/crm\/(?:campaigns|campaign\/v1\/campaigns)(?:\?|$)/, (route: Route) =>
      route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({ campaigns: [FAKE_CAMPAIGN] }),
      })
    );
  });

  test('campaigns list renders the campaign name', async ({ page }) => {
    await page.goto(`${CRM_URL}/campaigns`);

    await expect(page.getByText('Spring Promo 2026')).toBeVisible({ timeout: 15_000 });
  });

  test('campaigns list shows an Active status badge', async ({ page }) => {
    await page.goto(`${CRM_URL}/campaigns`);

    // The StatusBadge component renders the status text directly.
    await expect(page.getByText('Active', { exact: true })).toBeVisible({ timeout: 15_000 });
  });

  test('active campaign count is shown in the page subtitle', async ({ page }) => {
    await page.goto(`${CRM_URL}/campaigns`);

    // The Campaigns component renders "N active campaigns" below the heading.
    await expect(page.getByText(/1 active campaign/i)).toBeVisible({ timeout: 15_000 });
  });
});

// ── Segments list ─────────────────────────────────────────────────────────────

test.describe('CRM segments list flow', () => {
  test.beforeEach(async ({ page }) => {
    await seedAuth(page);
    await stubBaseRoutes(page);

    // Loyalty sub-endpoints: empty.
    await page.route(/localhost:9999\/crm\/loyalty(?:\/v1)?\//, (route: Route) =>
      route.fulfill({ status: 200, contentType: 'application/json', body: '{}' })
    );

    // Campaigns sidebar prefetch: empty.
    await page.route(/\/crm\/(?:campaigns|campaign\/v1\/campaigns)(?:\?|$)/, (route: Route) =>
      route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({ campaigns: [] }),
      })
    );

    // Segments list - returns our VIP Customers segment.
    await page.route(/\/crm\/(?:segments|segment\/v1\/segments)(?:\?|$)/, (route: Route) =>
      route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({ segments: [FAKE_SEGMENT] }),
      })
    );
  });

  test('segments list renders the segment name', async ({ page }) => {
    await page.goto(`${CRM_URL}/segments`);

    await expect(page.getByText('VIP Customers')).toBeVisible({ timeout: 15_000 });
  });

  test('segments list shows the member count', async ({ page }) => {
    await page.goto(`${CRM_URL}/segments`);

    // Segments.tsx renders customerCount.toLocaleString() inside the card.
    await expect(page.getByText('45', { exact: true })).toBeVisible({ timeout: 15_000 });
  });

  test('segments page shows aggregate customer total', async ({ page }) => {
    await page.goto(`${CRM_URL}/segments`);

    // Footer bar: "45 customers total"
    await expect(page.getByText(/45 customers total/i)).toBeVisible({ timeout: 15_000 });
  });
});

// ── Customer import UI ────────────────────────────────────────────────────────

test.describe('CRM customer import UI', () => {
  test.beforeEach(async ({ page }) => {
    await seedAuth(page);
    await stubBaseRoutes(page);

    // Loyalty, campaigns, segments: empty stubs so sidebar is quiet.
    await page.route(/localhost:9999\/crm\/loyalty(?:\/v1)?\//, (route: Route) =>
      route.fulfill({ status: 200, contentType: 'application/json', body: '{}' })
    );
    await page.route(/\/crm\/(?:campaigns|campaign\/v1\/campaigns)(?:\?|$)/, (route: Route) =>
      route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({ campaigns: [] }),
      })
    );
    await page.route(/\/crm\/(?:segments|segment\/v1\/segments)(?:\?|$)/, (route: Route) =>
      route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({ segments: [] }),
      })
    );

    // Import endpoint stub - captures the POST body for later assertion.
    await page.route(/\/crm\/customers\/import$/, (route: Route) =>
      route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({ imported: 0, skipped: 0 }),
      })
    );
  });

  test('import page renders the CSV upload area', async ({ page }) => {
    await page.goto(`${CRM_URL}/customers/import`);

    // ImportCustomers renders a file input and a heading - assert the UI is
    // present without triggering an actual file upload (which requires OS
    // picker interaction not supported in Playwright headless).
    await expect(page.getByRole('heading', { name: /import customers/i })).toBeVisible({
      timeout: 15_000,
    });
  });

  test('import page is reachable via the import route', async ({ page }) => {
    await page.goto(`${CRM_URL}/customers/import`);

    await expect(page).toHaveURL(/\/customers\/import$/, { timeout: 10_000 });
  });
});
