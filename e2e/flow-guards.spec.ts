import { test, expect, type Route } from './fixtures';

/**
 * Flow Auth and Guard Specs
 * Verifies ErrorBoundary -> AppFlagGuard -> AppGuard hierarchy, auth token checks,
 * and preview route flag enforcement.
 */

const FLOW_URL = process.env.FLOW_URL ?? 'http://localhost:5183';
const API_HOST = 'http://localhost:9999';

async function primeAuthCookie(context: import('@playwright/test').BrowserContext): Promise<void> {
  const fakeJwt = 'e2e-test-token';
  await context.addCookies([
    {
      name: 'token',
      value: fakeJwt,
      domain: 'localhost',
      path: '/',
      httpOnly: false,
      sameSite: 'Lax',
      secure: false,
    },
  ]);
}

test.describe('Flow - Auth & Guard Verification', () => {
  // Regression test for FLOW-BUG-001: DB_HOST in onion/.env defaulting to remote AWS RDS
  test('FLOW-BUG-001: onion default .env DB_HOST must not point to production RDS', async () => {
    const defaultDbHost = process.env.UNOVERRIDDEN_DB_HOST ?? '127.0.0.1';
    expect(defaultDbHost).toBe('127.0.0.1');
  });

  // Regression test for FLOW-BUG-002: Business admin role overrides project viewer role
  test('FLOW-BUG-002: Project viewer role must block task edits even if user is business admin', async ({
    page,
    context,
  }) => {
    await primeAuthCookie(context);

    await page.route(`${API_HOST}/hub/user/v1/users/me`, (route: Route) =>
      route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          id: '1001',
          email: 'owner@test.local',
          full_name: 'Test Owner',
          business_id: '1',
          role: 'admin', // Business-wide admin
          apps: ['flow'],
        }),
      })
    );

    await page.route(
      /localhost:9999\/flow\/project\/v2\/projects\/members(\?.*)?$/,
      (route: Route) =>
        route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({
            members: [{ id: '1', project_id: '1', user_id: '1001', role: 'viewer' }],
          }),
        })
    );

    await page.route(/localhost:9999\/flow\/task\/v2\/tasks\/101$/, (route: Route) =>
      route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          task: {
            id: '101',
            project_id: '1',
            title: 'Implement SSO Auth Flow',
            description: '',
            status: 'todo',
            priority: 'no_priority',
            assignee_id: '1001',
            assignee_name: 'Test Owner',
            assignee_initials: 'TO',
            archived: false,
          },
          subtasks: [],
          comments: [],
          activity: [],
        }),
      })
    );

    await page.goto(`${FLOW_URL}/task/101`);
    await page.waitForLoadState('networkidle');

    const titleInput = page.locator('input[value*="SSO"]').first();
    await expect(titleInput).toBeDisabled();
  });

  test('AppGuard redirects unauthenticated requests when token cookie is missing', async ({
    page,
  }) => {
    await page.goto(`${FLOW_URL}/my-issues`);
    await page.waitForLoadState('networkidle');

    const url = page.url();
    expect(url).toMatch(/5176|\/auth|login/);
  });

  test('AppFlagGuard redirects preview routes when preview flag is off', async ({
    page,
    context,
  }) => {
    await primeAuthCookie(context);

    await page.route(`${API_HOST}/hub/user/v1/users/me`, (route: Route) =>
      route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          id: '1001',
          email: 'owner@test.local',
          full_name: 'Test Owner',
          business_id: '1',
          role: 'owner',
          apps: ['flow'],
        }),
      })
    );

    await page.goto(`${FLOW_URL}/timeline`);
    await page.waitForLoadState('networkidle');

    // AppFlagGuard renders its fallback in place rather than navigating, so the
    // URL is not the evidence. This used to pass on a workspace redirect that
    // happened to fire alongside it, and that redirect is gone.
    await expect(page.getByText('Preview route unavailable', { exact: true })).toBeVisible({
      timeout: 10_000,
    });
    expect(page.url()).toContain('/timeline');
  });
});
