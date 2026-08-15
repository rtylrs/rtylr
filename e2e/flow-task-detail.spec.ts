import { test, expect, type Route } from './fixtures';

/**
 * Flow Task Detail E2E Specs
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

const FAKE_TASK_DETAIL = {
  task: {
    id: '101',
    projectId: '1',
    workspaceId: '10',
    ownerId: '1001',
    businessId: 1,
    title: 'Implement SSO Auth Flow',
    description: 'Build SSO login screen and session handling.',
    status: 'in_progress',
    priority: 'urgent',
    assigneeId: '1001',
    assignee: 'Test Owner',
    assigneeInitials: 'TO',
    reporterId: '1001',
    dueDate: '2026-08-10',
    startDate: '2026-08-01',
    estimatedHours: 8,
    actualHours: 4,
    archived: false,
    labels: [1, 2],
    identifier: 'AP-101',
    createdAt: new Date().toISOString(),
    updatedAt: new Date().toISOString(),
  },
  comments: [
    {
      id: 'c1',
      taskId: '101',
      userId: '1001',
      content: 'Initial comment on SSO implementation.',
      isEdited: false,
      createdAt: new Date().toISOString(),
    },
  ],
  subtasks: [
    {
      id: 's1',
      taskId: '101',
      title: 'Configure OAuth callback',
      status: 'done',
    },
  ],
};

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

test.describe('Flow - Task Detail', () => {
  test.beforeEach(async ({ context, page }) => {
    await primeAuthCookie(context);

    await page.route(`${API_HOST}/hub/user/v1/users/me`, (route: Route) =>
      route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify(FAKE_USER),
      })
    );

    await page.route(/localhost:9999\/flow\/task\/v1\/tasks\/101/, (route: Route) =>
      route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify(FAKE_TASK_DETAIL),
      })
    );

    await page.route(/localhost:9999\/flow\/task\/v1\/tasks\/101\/activity/, (route: Route) =>
      route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({ activity: [] }),
      })
    );

    await page.route(/localhost:9999\/flow\/task\/v1\/tasks\/101\/time-logs/, (route: Route) =>
      route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({ timeLogs: [] }),
      })
    );

    await page.route(/localhost:9999\/flow\/task\/v1\/tasks\/101\/attachments/, (route: Route) =>
      route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({ attachments: [] }),
      })
    );

    await page.route(/localhost:9999\/flow\/label\/v1\/labels/, (route: Route) =>
      route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({ labels: [] }),
      })
    );

    await page.route(/localhost:9999\/flow\/project\/v1\/projects/, (route: Route) =>
      route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({ projects: [] }),
      })
    );

    await page.route(/localhost:9999\/flow\/workspace\/v1\/members/, (route: Route) =>
      route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({ members: [{ id: '1001', name: 'Test Owner', initials: 'TO' }] }),
      })
    );
  });

  test('task detail renders task title, comments, subtasks', async ({ page }) => {
    await page.goto(`${FLOW_URL}/task/101`);
    await page.waitForLoadState('networkidle');

    await expect(page.locator('input[value="Implement SSO Auth Flow"]')).toBeVisible();
    await expect(page.getByText('Initial comment on SSO implementation.')).toBeVisible();
    await expect(page.getByText('Configure OAuth callback')).toBeVisible();
  });
});
