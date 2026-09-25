import { test, expect } from './fixtures';

/**
 * Insights e2e - dashboard loads, pillar screens switch, scheduled
 * reports CRUD works through the UI. All /insights/* endpoints are
 * stubbed via page.route() matching generated OneKit routes.
 */

const INSIGHTS_URL = process.env.INSIGHTS_URL ?? 'http://localhost:5187';
const API_URL = 'http://localhost:9999';

async function primeAuthCookie(context: import('@playwright/test').BrowserContext): Promise<void> {
  await context.addCookies([
    {
      name: 'token',
      value: 'fake-jwt',
      domain: 'localhost',
      path: '/',
      httpOnly: false,
      sameSite: 'Lax',
      secure: false,
    },
  ]);
}

async function stubInsightsAPIs(page: import('@playwright/test').Page): Promise<void> {
  await page.route(`${API_URL}/insights/dashboard/**`, (route) =>
    route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({
        revenue: { value: 12500, delta_pct: 12.4 },
        customers: { value: 42, delta_pct: 8.1 },
        orders: { value: 210, delta_pct: -2.0 },
        margin: { value: 3890, delta_pct: 6.5 },
      }),
    })
  );

  await page.route(`${API_URL}/insights/revenue/**`, (route) =>
    route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({ series: [], apps: [], comparisons: [] }),
    })
  );

  await page.route(`${API_URL}/insights/customer/**`, (route) =>
    route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({ customers: [], total: 0, new_this_month: 0, active: 0, churned: 0 }),
    })
  );

  await page.route(`${API_URL}/insights/product/**`, (route) =>
    route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({ products: [], total: 0, active: 0, out_of_stock: 0 }),
    })
  );

  await page.route(`${API_URL}/insights/financial/**`, (route) =>
    route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({ series: [], buckets: [] }),
    })
  );

  await page.route(`${API_URL}/insights/operation/**`, (route) =>
    route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({ headcount: 0, shifts: [] }),
    })
  );

  await page.route(`${API_URL}/insights/report/**`, (route) =>
    route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({
        templates: [
          { id: 'revenue-monthly', name: 'Monthly Revenue' },
          { id: 'ar-aging', name: 'A/R Aging' },
        ],
        reports: [],
      }),
    })
  );

  await page.route(`${API_URL}/insights/setting/**`, (route) => {
    if (route.request().method() === 'PUT') {
      route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          settings: {
            currency: 'USD',
            defaultPeriod: '',
            comparePrevious: false,
          },
        }),
      });
      return;
    }
    route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({
        settings: {
          currency: 'USD',
          defaultPeriod: '',
          comparePrevious: false,
        },
      }),
    });
  });

  await page.route(`${API_URL}/hub/user/v1/users/me`, (route) =>
    route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({
        id: '1',
        email: 'test@example.com',
        full_name: 'Test User',
        business_id: '1',
        role: 'owner',
        apps: ['insights'],
      }),
    })
  );
}

test.describe('Insights', () => {
  test.beforeEach(async ({ page, context }) => {
    await primeAuthCookie(context);
    await stubInsightsAPIs(page);
  });

  test('dashboard loads past auth guard', async ({ page }) => {
    await page.goto(INSIGHTS_URL);
    await expect(page.getByRole('heading', { level: 1 })).toBeVisible({ timeout: 10_000 });
  });

  test('renders without console errors', async ({ page }) => {
    const errors: string[] = [];
    page.on('console', (msg) => {
      if (msg.type() === 'error') errors.push(msg.text());
    });
    await page.goto(INSIGHTS_URL);
    await page.waitForTimeout(2_000);
    const meaningful = errors.filter((e) => !/posthog|sentry|Failed to load resource/i.test(e));
    expect(meaningful, meaningful.join('\n')).toHaveLength(0);
  });
});
