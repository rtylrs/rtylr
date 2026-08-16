import { test, expect } from './fixtures';

/**
 * Dash (cross-app command center) spec. Stubs the auth cookie so
 * the app renders past its ProtectedRoute, then stubs each widget's
 * endpoint so every tile has data without a real backend.
 */

const DASH_URL = process.env.DASH_URL ?? 'http://localhost:5175';

async function primeAuthCookie(context: import('@playwright/test').BrowserContext): Promise<void> {
  // The auth flow sets `token` on `.rtylr.com` / `localhost` with path=/.
  // In Playwright the cookie needs an explicit domain that matches the
  // baseURL - `localhost` for local dev.
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

async function stubDashAPIs(page: import('@playwright/test').Page): Promise<void> {
  // Widgets.
  await page.route('**/hub/kpis', (route) =>
    route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({
        revenue_today: 4212,
        revenue_trend_pct: 12.4,
        orders_today: 38,
        orders_trend_pct: -4.1,
        headcount: 24,
        active_staff: 11,
        low_stock_items: 3,
        open_invoices: 7,
        new_customers: 5,
        api_health_pct: 99.9,
        revenue_series: [1, 2, 3, 4, 5, 6, 7],
        orders_series: [1, 2, 3, 4, 5, 6, 7],
      }),
    })
  );

  // Insights-backed widgets added this session. Stub with empty
  // shapes - widgets render their empty state, which is the
  // behaviour we actually want to verify.
  await page.route('**/insights/revenue/by-app', (route) =>
    route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({ by_app: [] }),
    })
  );
  await page.route('**/insights/financial/invoice-aging', (route) =>
    route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({ buckets: [] }),
    })
  );
  await page.route('**/insights/customers/top*', (route) =>
    route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({ customers: [] }),
    })
  );

  // Team, activity, notifications, low-stock, orders, invoices -
  // generic 200 empty response lets every widget render.
  await page.route(
    /\/(team|activity|notifications|erp\/inventory|erp\/products|hr\/time-off|user\/me|get-configs|business\/me)/,
    (route) => {
      const url = route.request().url();
      if (
        url.includes('/src/') ||
        url.endsWith('.tsx') ||
        url.endsWith('.ts') ||
        url.endsWith('.jsx') ||
        url.endsWith('.js')
      ) {
        return route.continue();
      }
      return route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({}),
      });
    }
  );
}

test.describe('Dash', () => {
  test.beforeEach(async ({ page, context }) => {
    await primeAuthCookie(context);
    await stubDashAPIs(page);
  });

  test('dashboard loads past auth guard', async ({ page }) => {
    await page.goto(DASH_URL);
    // Broad smoke: app mounted past Suspense/protected shell when
    // at least one interactive element is on screen. Don't pin to a
    // specific landmark - the dashboard layout evolves.
    await page.waitForTimeout(1500);
    const count = await page.locator('button, a, h1, h2, [role]').count();
    expect(count).toBeGreaterThan(0);
  });

  test('widgets render without runtime errors', async ({ page }) => {
    const consoleErrors: string[] = [];
    page.on('console', (msg) => {
      if (msg.type() === 'error') consoleErrors.push(msg.text());
    });

    await page.goto(DASH_URL);
    // Give the widgets time to mount, fetch, and render their empty
    // states. A widget that throws will show up in the console log.
    await page.waitForTimeout(2_000);

    // We tolerate 3rd-party noise (posthog not configured, sentry
    // not initialised without DSN) by filtering on app-origin
    // substrings. React hydration errors fail the test though.
    const meaningful = consoleErrors.filter(
      (e) => !/posthog|sentry|Failed to load resource/i.test(e)
    );
    expect(meaningful, meaningful.join('\n')).toHaveLength(0);
  });
});
