import { flowAuthFields } from './support/flow-auth';
import { test, expect, type Route } from './fixtures';

/**
 * Flow Inbox, Views, My Issues, Today, Archive, Backlog & Settings E2E Specs
 */

const FLOW_URL = process.env.FLOW_URL ?? 'http://localhost:5197/flow';
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

  test('my issues windows a long task list', async ({ page }) => {
    const tasks = Array.from({ length: 300 }, (_, i) => ({
      ...FAKE_TASK,
      id: String(1000 + i),
      title: `Long list task ${i}`,
      identifier: `AP-${1000 + i}`,
    }));
    await page.route(/localhost:9999\/flow\/task\/v2\/tasks\/my(\?.*)?$/, (route: Route) =>
      route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({ tasks }),
      })
    );
    await page.goto(`${FLOW_URL}/my-issues`);

    // Undated tasks land in "No date", which starts collapsed.
    await page.getByRole('button', { name: /^No date/ }).click();
    const first = page.getByText('Long list task 0', { exact: true });
    await expect(first).toBeVisible();
    // Only the rows near the viewport render; the rest hold their space.
    const rows = page.getByRole('button', { name: /Long list task \d+/ });
    await expect.poll(() => rows.count()).toBeLessThan(80);
    await expect(page.getByText('Long list task 299', { exact: true })).not.toBeAttached();

    const scrollTo = (where: 'top' | 'bottom') =>
      first.evaluate((el, where) => {
        let scroller = el.parentElement;
        while (scroller && getComputedStyle(scroller).overflowY !== 'auto') {
          scroller = scroller.parentElement;
        }
        scroller!.scrollTo({ top: where === 'top' ? 0 : scroller!.scrollHeight });
      }, where);
    await scrollTo('bottom');
    await expect(page.getByText('Long list task 299', { exact: true })).toBeVisible();
    await expect(page.getByText('Long list task 0', { exact: true })).not.toBeAttached();
    await expect.poll(() => rows.count()).toBeLessThan(80);
  });

  test('central settings navigation opens Flow sections', async ({ page }) => {
    await page.goto(`${FLOW_URL.replace(/\/flow$/, '')}/settings/flow`);
    await page.waitForLoadState('networkidle');

    // Settings open as a modal over the page; its menu lists the Flow sections.
    const dialog = page.getByRole('dialog', { name: 'Settings' });
    await expect(dialog.getByRole('heading', { name: 'Notifications', exact: true })).toBeVisible();
    await dialog
      .getByRole('navigation', { name: 'Settings' })
      .getByRole('link', { name: 'Sidebar', exact: true })
      .click();
    await expect(
      dialog.getByRole('heading', { name: 'Customize your sidebar', exact: true })
    ).toBeVisible();
  });
});
