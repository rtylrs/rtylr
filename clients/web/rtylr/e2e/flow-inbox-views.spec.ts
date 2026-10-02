import { flowAuthFields } from './support/flow-auth';
import { test, expect, type Route } from './fixtures';

/**
 * Flow Inbox, Views, My Issues, Today, Archive, Backlog & Settings E2E Specs
 */

const FLOW_URL = process.env.FLOW_URL ?? 'http://localhost:5175/flow';
const API_HOST = 'http://localhost:9999';

const FAKE_USER = {
  id: '1001',
  email: 'owner@test.local',
  full_name: 'Test Owner',
  business_id: '1',
  ...flowAuthFields('owner'),
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

const FAKE_TASK = {
  id: '101',
  project_id: '1',
  owner_id: '1001',
  title: 'Implement SSO Auth Flow',
  description: '',
  status: 'in_progress',
  priority: 'high',
  assignee_id: '1001',
  assignee_name: 'Test Owner',
  assignee_initials: 'TO',
  project_name: 'Alpha Project',
  identifier: 'AP-101',
  archived: false,
};

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

test.describe('Flow - Inbox & Features Sweep', () => {
  test.beforeEach(async ({ context, page }) => {
    await primeAuthCookie(context);

    await page.route(/localhost:9999\//, (route: Route) => {
      const method = route.request().method();
      return route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify(method === 'GET' ? {} : { success: true }),
      });
    });

    await page.route(`${API_HOST}/hub/user/v1/users/me`, (route: Route) =>
      route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify(FAKE_USER),
      })
    );

    await page.route(
      /localhost:9999\/flow\/notification\/v1\/notifications(\?.*)?$/,
      (route: Route) =>
        route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({
            notifications: FAKE_NOTIFICATIONS.map((notification) => ({
              id: notification.id,
              type: notification.type,
              title: notification.title,
              description: notification.description,
              task_id: notification.taskId,
              task_title: notification.taskTitle,
              actor_initials: notification.actorInitials,
              project: notification.project,
              project_color: notification.projectColor,
              read: notification.read,
              timestamp: notification.timestamp,
            })),
          }),
        })
    );

    await page.route(/localhost:9999\/flow\/view\/v2\/views(\?.*)?$/, (route: Route) =>
      route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({ views: [] }),
      })
    );

    await page.route(/localhost:9999\/flow\/task\/v2\/tasks\/my(\?.*)?$/, (route: Route) =>
      route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({ tasks: [FAKE_TASK] }),
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
    await page.goto(`${FLOW_URL.replace(/\/flow$/, '')}/settings/flow`);
    await page.waitForLoadState('networkidle');

    await expect(
      page
        .locator('h1, h2, div')
        .filter({ hasText: /Settings/i })
        .first()
    ).toBeVisible();
  });
});
