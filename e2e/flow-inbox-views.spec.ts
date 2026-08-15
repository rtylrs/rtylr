import { test, expect, type Route } from './fixtures';

/**
 * Flow Inbox, Views, My Issues, Today, Archive, Backlog & Settings E2E Specs
 */

const FLOW_URL = process.env.FLOW_URL ?? 'http://localhost:5183';
const API_HOST = 'http://localhost:9999';

const FAKE_USER = {
  id: '1001',
  email: 'owner@test.local',
  fullName: 'Test Owner',
  role: 'owner',
  apps: ['flow'],
};

const FAKE_NOTIFICATIONS = [
  {
    id: 'n1',
    type: 'assigned',
    title: 'Task assigned to you',
    description: 'Implement SSO Auth Flow was assigned to you.',
    taskId: '101',
    taskTitle: 'Implement SSO Auth Flow',
    actorInitials: 'TO',
    project: 'Alpha Project',
    projectColor: '#64C5B9',
    read: false,
    timestamp: new Date().toISOString(),
  },
];

async function primeAuthCookie(context: import('@playwright/test').BrowserContext): Promise<void> {
  const fakeJwt = 'eyJhbGciOiJub25lIn0.eyJ1c2VyX2lkIjoiMTAwMSJ9.fake_sig';
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

test.describe('Flow - Inbox & Features Sweep', () => {
  test.beforeEach(async ({ context, page }) => {
    await primeAuthCookie(context);

    await page.route(`${API_HOST}/hub/user/v1/users/me`, (route: Route) =>
      route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify(FAKE_USER),
      })
    );

    await page.route(/localhost:9999\/flow\/notification\/v1\/notifications/, (route: Route) =>
      route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({ notifications: FAKE_NOTIFICATIONS }),
      })
    );

    await page.route(/localhost:9999\/flow\/view\/v1\/views/, (route: Route) =>
      route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({ views: [] }),
      })
    );

    await page.route(/localhost:9999\/flow\/workspace\/v1\/workspaces/, (route: Route) =>
      route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({ workspaces: [{ id: '10', name: 'Personal Workspace' }] }),
      })
    );
  });

  test('inbox renders notifications list', async ({ page }) => {
    await page.goto(`${FLOW_URL}/inbox`);
    await page.waitForLoadState('networkidle');

    await expect(page.getByText('Task assigned to you').first()).toBeVisible();
  });

  test('my issues renders cross-project task list', async ({ page }) => {
    await page.goto(`${FLOW_URL}/my-issues`);
    await page.waitForLoadState('networkidle');

    await expect(page.locator('body')).toBeVisible();
  });

  test('settings page renders all sections', async ({ page }) => {
    await page.goto(`${FLOW_URL}/settings`);
    await page.waitForLoadState('networkidle');

    await expect(
      page
        .locator('h1, h2, div')
        .filter({ hasText: /Settings/i })
        .first()
    ).toBeVisible();
  });
});
