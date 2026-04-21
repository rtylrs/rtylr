import { test, expect } from '@playwright/test';

/**
 * Flow app (projects & tasks). Cookie-primed past ProtectedRoute,
 * API stubbed. We assert the app mounts and that the inbox / focus
 * surface renders without console errors — deeper drill-through
 * belongs in a focused spec once backend fixtures land.
 */

const FLOW_URL = process.env.FLOW_URL ?? 'http://localhost:5183';

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

async function stubFlowAPIs(page: import('@playwright/test').Page): Promise<void> {
  await page.route(/\/flow\/.*/, (route) =>
    route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({ projects: [], tasks: [], items: [], notifications: [] }),
    })
  );

  await page.route(
    /\/(user\/me|get-configs|business\/me|hub\/kpis|notifications|team|activity)/,
    (route) =>
      route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({}),
      })
  );
}

test.describe('Flow', () => {
  test.beforeEach(async ({ page, context }) => {
    await primeAuthCookie(context);
    await stubFlowAPIs(page);
  });

  test('inbox loads past auth guard', async ({ page }) => {
    await page.goto(`${FLOW_URL}/inbox`);
    // Any rendered landmark signals we're past the protected shell.
    await expect(
      page.getByRole('heading').first().or(page.getByRole('banner').first())
    ).toBeVisible({ timeout: 10_000 });
  });

  test('renders without console errors', async ({ page }) => {
    const errors: string[] = [];
    page.on('console', (msg) => {
      if (msg.type() === 'error') errors.push(msg.text());
    });
    await page.goto(FLOW_URL);
    await page.waitForTimeout(2_000);
    const meaningful = errors.filter(
      (e) => !/posthog|sentry|Failed to load resource|WebSocket/i.test(e)
    );
    expect(meaningful, meaningful.join('\n')).toHaveLength(0);
  });
});
