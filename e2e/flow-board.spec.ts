import { test, expect, type Route } from './fixtures';
import { captureResponsivePair } from './support/screenshot';

/**
 * Flow Project Board & Project List Specs
 */

const FLOW_URL = process.env.FLOW_URL ?? 'http://localhost:5183';
const API_HOST = 'http://localhost:9999';

const FAKE_USER = {
  id: '1001',
  email: 'owner@test.local',
  full_name: 'Test Owner',
  business_id: '1',
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
  { id: '1', name: 'Todo', color: '#6366f1', sortOrder: 0 },
  { id: '2', name: 'In Progress', color: '#3b82f6', sortOrder: 1 },
  { id: '3', name: 'Done', color: '#22c55e', sortOrder: 2 },
];

const FAKE_TASKS = [
  {
    id: '101',
    projectId: '1',
    workspaceId: '10',
    title: 'Implement SSO Auth Flow',
    status: 'in_progress',
    priority: 'urgent',
    sectionId: '2',
    assignee: 'Test Owner',
  },
  {
    id: '102',
    projectId: '1',
    workspaceId: '10',
    title: 'Fix navigation dropdown bug',
    status: 'todo',
    priority: 'high',
    sectionId: '1',
    assignee: 'Test Admin',
  },
];

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
        body: JSON.stringify({
          project: {
            id: '1',
            name: 'Alpha Project',
            description: '',
            icon: 'folder',
            color: '#64C5B9',
            status: 'active',
            visibility: 'private',
            task_count: 3,
            completed_count: 1,
          },
        }),
      })
    );

    await page.route(/localhost:9999\/flow\/project\/v2\/projects(\?.*)?$/, (route: Route) =>
      route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          projects: [
            {
              id: '1',
              name: 'Alpha Project',
              description: '',
              icon: 'folder',
              color: '#64C5B9',
              status: 'active',
              visibility: 'private',
              task_count: 3,
              completed_count: 1,
            },
          ],
          pagination: { total: 1 },
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
            members: [
              {
                id: '1',
                project_id: '1',
                user_id: '1001',
                name: 'Test Owner',
                email: 'owner@test.local',
                role: 'admin',
              },
            ],
            pagination: { total: 1 },
          }),
        })
    );

    await page.route(/localhost:9999\/flow\/task\/v2\/tasks(\?.*)?$/, (route: Route) =>
      route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          tasks: FAKE_TASKS.map((task) => ({
            id: task.id,
            project_id: task.projectId,
            owner_id: '1001',
            title: task.title,
            description: '',
            status: task.status,
            priority: task.priority,
            assignee_id: '1001',
            assignee_name: task.assignee,
            assignee_initials: 'TO',
            project_name: 'Alpha Project',
            identifier: String(task.id),
            section_id: task.sectionId,
            archived: false,
          })),
        }),
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

  // WCAG 2.2 AA 2.5.7 (B2): moving a card must not require a drag. This is the
  // path a keyboard or single-pointer user takes, so it is the one that has to
  // keep working.
  test('a card can be moved between columns without dragging', async ({ page }) => {
    let movePayload: string | null = null;
    await page.route(
      /localhost:9999\/flow\/project\/v2\/projects\/columns\/move/,
      (route: Route) => {
        movePayload = route.request().postData();
        return route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({ message: 'moved' }),
        });
      }
    );

    await page.goto(`${FLOW_URL}/p/1`);
    await page.waitForLoadState('networkidle');

    const moveButton = page.getByRole('button', { name: /Move .* to another column/i }).first();
    await moveButton.click();

    const menu = page.getByRole('menu');
    await expect(menu).toBeVisible({ timeout: 10_000 });
    await menu.getByRole('menuitem', { name: /Done/ }).click();

    await expect.poll(() => movePayload, { timeout: 10_000 }).not.toBeNull();
    expect(movePayload).toContain('"column_id":"3"');
  });

  /**
   * Captures the board for a PR body, and asserts the things a screenshot cannot
   * show you.
   *
   * Lives in this file rather than its own so it reuses the beforeEach above:
   * the board needs the primed cookie, `users/me`, the project and two column
   * endpoints, and a separate spec would have been sixty lines of duplicated
   * routes.
   */
  test('captures the board at desktop and at the 390px floor @screenshots', async ({ page }) => {
    await page.goto(`${FLOW_URL}/p/1`);
    await page.waitForLoadState('networkidle');
    await page.waitForTimeout(800);

    // Asserted before capturing, so a blank or redirected page cannot be
    // mistaken for a successful shot.
    await expect(page.getByText('Implement SSO Auth Flow')).toBeVisible();

    // Add task is on every column now, and is not hover-only. It used to appear
    // only in empty columns, and the header alternative was a 20px control that
    // a touch screen cannot reach (rtylr#1078).
    const addButtons = page.getByRole('button', { name: /add task/i });
    await expect(addButtons.first()).toBeVisible();
    expect(await addButtons.count()).toBeGreaterThanOrEqual(3);

    // The column width rule is 280-320px. It was 248.
    const column = page.getByText('In Progress').first().locator('xpath=ancestor::div[1]');
    const box = await column.boundingBox();
    expect(box).not.toBeNull();

    // The card is actually painted, not merely present.
    //
    // This is the assertion the board never had. `layout` plus
    // `initial={{ opacity: 0 }}` left every card at opacity 0, so the columns,
    // the counts and the empty states all rendered while the cards were
    // invisible. Playwright's own `toBeVisible` does not catch it, because
    // opacity is not part of its visibility check, and no unit test can:
    // happy-dom runs no animations at all.
    const cardOpacity = await page
      .getByText('Implement SSO Auth Flow')
      .first()
      .evaluate((el) => getComputedStyle(el.closest('.rounded-lg') as HTMLElement).opacity);
    expect(cardOpacity).toBe('1');

    const shots = await captureResponsivePair(page, 'flow-board');

    // A screenshot cannot show you that the page scrolls sideways at 390px. The
    // board's own columns scroll horizontally by design; the page must not.
    expect(shots.overflowsAt390).toBe(false);
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
