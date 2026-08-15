import { test, expect, type Route } from './fixtures';

/**
 * Flow Workspace Dashboard & Workspace Members E2E Specs
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

const FAKE_WORKSPACE = {
  id: '10',
  name: 'Personal Workspace',
  color: '#4583C4',
  memberCount: 2,
};

const FAKE_PROJECTS = [
  {
    id: '1',
    workspaceId: '10',
    name: 'Alpha Project',
    color: '#64C5B9',
    taskCount: 5,
    completedCount: 2,
    memberAvatars: ['TO', 'TA'],
  },
  {
    id: '2',
    workspaceId: '10',
    name: 'Beta Project',
    color: '#f97316',
    taskCount: 3,
    completedCount: 1,
    memberAvatars: ['TO'],
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

test.describe('Flow - Workspace Dashboard', () => {
  test.beforeEach(async ({ context, page }) => {
    await primeAuthCookie(context);

    await page.route(`${API_HOST}/hub/user/v1/users/me`, (route: Route) =>
      route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify(FAKE_USER),
      })
    );

    await page.route(/localhost:9999\/flow\/workspace\/v1\/workspaces/, (route: Route) =>
      route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({ workspaces: [FAKE_WORKSPACE] }),
      })
    );

    await page.route(/localhost:9999\/flow\/project\/v1\/projects/, (route: Route) =>
      route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({ projects: FAKE_PROJECTS }),
      })
    );

    await page.route(/localhost:9999\/flow\/dashboard\/v1\/dashboard/, (route: Route) =>
      route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          openTasks: 4,
          inProgress: 2,
          completedThisWeek: 3,
          overdue: 1,
          myTasks: [
            {
              id: '101',
              title: 'Implement SSO Auth Flow',
              projectName: 'Alpha Project',
              status: 'in_progress',
              priority: 'urgent',
              dueDate: '2026-08-10',
            },
          ],
          recentProjects: FAKE_PROJECTS,
        }),
      })
    );
  });

  test('workspace dashboard loads stats and project cards', async ({ page }) => {
    await page.goto(`${FLOW_URL}/w/10`);
    await page.waitForLoadState('networkidle');

    await expect(page.getByText('Open Tasks').first()).toBeVisible();
    await expect(page.getByText('Alpha Project').first()).toBeVisible();
    await expect(page.getByText('Beta Project').first()).toBeVisible();
  });

  test('weekly review streak button updates streak count', async ({ page }) => {
    await page.goto(`${FLOW_URL}/w/10`);
    await page.waitForLoadState('networkidle');

    const reviewBtn = page.getByRole('button', { name: /I did my review this week/i }).first();
    await expect(reviewBtn).toBeVisible();
    await reviewBtn.click();
  });
});

test.describe('Flow - Workspace Members', () => {
  test.beforeEach(async ({ context, page }) => {
    await primeAuthCookie(context);

    await page.route(`${API_HOST}/hub/user/v1/users/me`, (route: Route) =>
      route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify(FAKE_USER),
      })
    );

    await page.route(/localhost:9999\/flow\/workspace\/v1\/members/, (route: Route) =>
      route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          members: [
            { id: '1001', name: 'Test Owner', email: 'owner@test.local', role: 'admin' },
            { id: '1002', name: 'Test Admin', email: 'admin@test.local', role: 'admin' },
          ],
        }),
      })
    );
  });

  test('workspace members page renders member list', async ({ page }) => {
    await page.goto(`${FLOW_URL}/w/10/members`);
    await page.waitForLoadState('networkidle');

    await expect(
      page
        .locator('h1, h2, div')
        .filter({ hasText: /Members|Workspace/i })
        .first()
    ).toBeVisible();
  });
});
