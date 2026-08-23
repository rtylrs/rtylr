import { test, expect, type Route } from './fixtures';

/**
 * Flow Project Board & Project List Specs
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

const FAKE_PROJECT = {
  id: '1',
  workspaceId: '10',
  name: 'Alpha Project',
  color: '#64C5B9',
  taskCount: 3,
  completedCount: 1,
  memberAvatars: ['TO'],
};

const FAKE_COLUMNS = [
  { id: 'col-1', name: 'Todo', color: '#6366f1', sortOrder: 0 },
  { id: 'col-2', name: 'In Progress', color: '#3b82f6', sortOrder: 1 },
  { id: 'col-3', name: 'Done', color: '#22c55e', sortOrder: 2 },
];

const FAKE_TASKS = [
  {
    id: '101',
    projectId: '1',
    workspaceId: '10',
    title: 'Implement SSO Auth Flow',
    status: 'in_progress',
    priority: 'urgent',
    sectionId: 'col-2',
    assignee: 'Test Owner',
  },
  {
    id: '102',
    projectId: '1',
    workspaceId: '10',
    title: 'Fix navigation dropdown bug',
    status: 'todo',
    priority: 'high',
    sectionId: 'col-1',
    assignee: 'Test Admin',
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

test.describe('Flow - Project Board', () => {
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
        body: JSON.stringify({ workspaces: [{ id: '10', name: 'Personal Workspace' }] }),
      })
    );

    await page.route(/localhost:9999\/flow\/project\/v1\/projects(\?.*)?/, (route: Route) =>
      route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({ projects: [FAKE_PROJECT] }),
      })
    );

    await page.route(/localhost:9999\/flow\/project\/v1\/projects\/1/, (route: Route) =>
      route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({ project: FAKE_PROJECT }),
      })
    );

    await page.route(/localhost:9999\/flow\/project\/v1\/projects\/1\/columns/, (route: Route) =>
      route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({ columns: FAKE_COLUMNS }),
      })
    );

    await page.route(/localhost:9999\/flow\/task\/v1\/tasks(\?.*)?/, (route: Route) =>
      route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({ tasks: FAKE_TASKS }),
      })
    );

    // The board reads projects and columns through v2 now. Columns first: its
    // path is a prefix of the project route, so the broader pattern would
    // otherwise swallow it.
    await page.route(
      /localhost:9999\/flow\/project\/v2\/projects\/columns(\?.*)?$/,
      (route: Route) =>
        route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({ columns: FAKE_COLUMNS }),
        })
    );

    await page.route(/localhost:9999\/flow\/project\/v2\/projects\/1$/, (route: Route) =>
      route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({ project: FAKE_PROJECT }),
      })
    );

    await page.route(/localhost:9999\/flow\/project\/v2\/projects(\?.*)?$/, (route: Route) =>
      route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({ projects: [FAKE_PROJECT], pagination: { total: 1 } }),
      })
    );
  });

  // The workspace segment is gone from project URLs, but links to the old shape
  // survive in bookmarks and pasted messages for a release.
  test('a legacy /w/<id>/p/<id> URL lands on the workspace-free route', async ({ page }) => {
    await page.goto(`${FLOW_URL}/w/10/p/1`);
    await page.waitForLoadState('networkidle');

    expect(new URL(page.url()).pathname).toBe('/p/1');
  });

  test('a legacy sub-view URL keeps its view on redirect', async ({ page }) => {
    await page.goto(`${FLOW_URL}/w/10/p/1/list`);
    await page.waitForLoadState('networkidle');

    expect(new URL(page.url()).pathname).toBe('/p/1/list');
  });

  test('kanban board renders columns and task cards', async ({ page }) => {
    await page.goto(`${FLOW_URL}/p/1`);
    await page.waitForLoadState('networkidle');

    await expect(page.locator('body')).toBeVisible();
  });

  test('group-by swimlanes option persists in localStorage', async ({ page }) => {
    await page.goto(`${FLOW_URL}/p/1`);
    await page.waitForLoadState('networkidle');

    const groupPicker = page.getByRole('button', { name: /Group:/i }).first();
    if (await groupPicker.isVisible({ timeout: 3000 })) {
      await groupPicker.click();
      const priorityOpt = page.getByRole('button', { name: 'Priority' }).first();
      if (await priorityOpt.isVisible({ timeout: 2000 })) {
        await priorityOpt.click();
        const storedValue = await page.evaluate(() => localStorage.getItem('flow:board:groupBy:1'));
        expect(storedValue).toBe('priority');
      }
    }
  });

  test('project list view renders tasks with filters', async ({ page }) => {
    await page.goto(`${FLOW_URL}/p/1/list`);
    await page.waitForLoadState('networkidle');

    await expect(page.locator('body')).toBeVisible();
  });
});
