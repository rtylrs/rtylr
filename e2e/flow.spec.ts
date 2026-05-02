import { test, expect, type Page, type Route } from '@playwright/test';

/**
 * Flow app (projects & tasks). Cookie-primed past ProtectedRoute,
 * API stubbed via page.route() anchored to the real API host so Vite
 * module requests are never accidentally swallowed.
 *
 * Registration order: Playwright matches routes in REVERSE order, so the
 * most specific handlers are registered LAST (same pattern as
 * e2e/erp-inventory.spec.ts).
 */

const FLOW_URL = process.env.FLOW_URL ?? 'http://localhost:5183';
const API_HOST = 'https://api.voxire.dev';

// ── Shared fixtures ────────────────────────────────────────────────────────

const FAKE_USER = {
  id: 'u1',
  email: 'owner@acme.test',
  fullName: 'Acme Owner',
  role: 'owner',
  apps: ['flow'],
  planSlug: 'pro',
};

const FAKE_WORKSPACE = {
  id: '10',
  name: 'Acme Workspace',
  color: '#64C5B9',
  memberCount: 3,
};

function fakeProject(overrides: Record<string, unknown> = {}): Record<string, unknown> {
  return {
    id: '1',
    workspaceId: '10',
    name: 'Alpha Project',
    color: '#64C5B9',
    taskCount: 2,
    completedCount: 1,
    memberAvatars: [],
    ...overrides,
  };
}

function fakeTask(overrides: Record<string, unknown> = {}): Record<string, unknown> {
  return {
    id: '101',
    projectId: '1',
    workspaceId: '10',
    title: 'Fix login bug',
    description: '',
    status: 'todo',
    priority: 'high',
    assignee: 'Acme Owner',
    assigneeId: 'u1',
    assigneeInitials: 'AO',
    dueDate: '',
    subtaskCount: 0,
    projectName: 'Alpha Project',
    identifier: 'AP-1',
    createdAt: new Date().toISOString(),
    updatedAt: new Date().toISOString(),
    labels: [],
    ...overrides,
  };
}

function fakeNotification(overrides: Record<string, unknown> = {}): Record<string, unknown> {
  return {
    id: 'n1',
    type: 'assigned',
    title: 'Task assigned to you',
    description: 'Fix login bug was assigned to you.',
    taskId: '101',
    taskTitle: 'Fix login bug',
    actorInitials: 'AO',
    project: 'Alpha Project',
    projectColor: '#64C5B9',
    read: false,
    timestamp: new Date().toISOString(),
    ...overrides,
  };
}

// ── Auth helpers ───────────────────────────────────────────────────────────

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

// ── Stub installer ─────────────────────────────────────────────────────────

/**
 * Installs stubs for every Flow API endpoint the spec's pages hit.
 * Pass `overrides` to customize per-test responses.
 */
async function installFlowStubs(
  page: Page,
  opts: {
    projects?: Record<string, unknown>[];
    tasks?: Record<string, unknown>[];
    notifications?: Record<string, unknown>[];
    /** If provided, POST /flow/projects returns this project. */
    newProject?: Record<string, unknown> | null;
  } = {}
): Promise<{ captured: { lastProjectPost: Record<string, unknown> | null } }> {
  const captured: { lastProjectPost: Record<string, unknown> | null } = {
    lastProjectPost: null,
  };

  const projects = opts.projects ?? [];
  const tasks = opts.tasks ?? [];
  const notifications = opts.notifications ?? [];

  // (1) Catch-all for any /flow/* endpoint not handled below — returns empty
  //     envelopes so the shell doesn't throw on missing data.
  await page.route(`${API_HOST}/flow/**`, (route: Route) => {
    const method = route.request().method();
    if (method === 'GET') {
      route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          projects: [],
          tasks: [],
          notifications: [],
          workspaces: [],
          members: [],
          columns: [],
          cycles: [],
          milestones: [],
          labels: [],
          views: [],
          businesses: [],
        }),
      });
    } else {
      route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({ success: true }),
      });
    }
  });

  // (2) Shell/auth endpoints.
  await page.route(`${API_HOST}/user/me`, (route: Route) =>
    route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify(FAKE_USER),
    })
  );

  await page.route(
    /api\.voxire\.dev\/(get-configs|subscriptions|business\/me|hub\/kpis|team|activity|notifications)/,
    (route: Route) =>
      route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({}),
      })
  );

  // (3) Workspaces — regex to match with or without query params.
  await page.route(/api\.voxire\.dev\/flow\/workspaces(\?.*)?$/, (route: Route) =>
    route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({ workspaces: [FAKE_WORKSPACE] }),
    })
  );

  // (4) Dashboard stats — regex to match ?workspace_id=... query param.
  await page.route(/api\.voxire\.dev\/flow\/dashboard(\?.*)?$/, (route: Route) =>
    route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({
        openTasks: tasks.length,
        inProgress: 0,
        completedThisWeek: 0,
        overdue: 0,
        myTasks: tasks,
        recentProjects: projects,
      }),
    })
  );

  // (5) Notifications.
  await page.route(/api\.voxire\.dev\/flow\/notifications(\?.*)?$/, (route: Route) =>
    route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({ notifications }),
    })
  );

  // (6) Tasks for a project — regex to match ?project_id=... query param.
  await page.route(/api\.voxire\.dev\/flow\/tasks(\?.*)?$/, (route: Route) => {
    const url = new URL(route.request().url());
    const projectId = url.searchParams.get('project_id');
    const filtered = projectId ? tasks.filter((t) => t.projectId === projectId) : tasks;
    route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({ tasks: filtered }),
    });
  });

  await page.route(/api\.voxire\.dev\/flow\/tasks\/my(\?.*)?$/, (route: Route) =>
    route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({ tasks }),
    })
  );

  // (7) Projects: GET list + POST create — regex excludes /columns sub-path.
  await page.route(/api\.voxire\.dev\/flow\/projects(\?.*)?$/, async (route: Route) => {
    const req = route.request();
    if (req.method() === 'GET') {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({ projects }),
      });
      return;
    }
    if (req.method() === 'POST') {
      const body = (req.postDataJSON() ?? {}) as Record<string, unknown>;
      captured.lastProjectPost = body;
      const created =
        opts.newProject ??
        fakeProject({
          id: '99',
          name: String(body.name ?? ''),
          color: String(body.color ?? '#64C5B9'),
          workspaceId: String(body.workspaceId ?? FAKE_WORKSPACE.id),
        });
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({ project: created }),
      });
      return;
    }
    await route.fulfill({ status: 200, contentType: 'application/json', body: '{}' });
  });

  // Project columns (board page).
  await page.route(/api\.voxire\.dev\/flow\/projects\/columns(\?.*)?$/, (route: Route) =>
    route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({ columns: [] }),
    })
  );

  return { captured };
}

// ── Tests ──────────────────────────────────────────────────────────────────

test.describe('Flow — smoke', () => {
  test.beforeEach(async ({ page, context }) => {
    await primeAuthCookie(context);
    // Minimal stubs just to get the shell up; content is not asserted here.
    await installFlowStubs(page);
  });

  test('inbox loads past auth guard', async ({ page }) => {
    await page.goto(`${FLOW_URL}/inbox`);
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

test.describe('Flow — project list', () => {
  test.beforeEach(async ({ context }) => {
    await primeAuthCookie(context);
  });

  test('workspace dashboard shows two stubbed projects', async ({ page }) => {
    const alpha = fakeProject({ id: '1', name: 'Alpha Project' });
    const beta = fakeProject({ id: '2', name: 'Beta Project', color: '#f97316' });

    await installFlowStubs(page, { projects: [alpha, beta] });

    await page.goto(`${FLOW_URL}/w/10`);

    // Both project names must appear somewhere in the rendered output.
    await expect(page.getByText('Alpha Project').first()).toBeVisible({ timeout: 10_000 });
    await expect(page.getByText('Beta Project').first()).toBeVisible({ timeout: 10_000 });
  });

  test('workspace dashboard shows empty state when no projects', async ({ page }) => {
    await installFlowStubs(page, { projects: [] });
    await page.goto(`${FLOW_URL}/w/10`);
    // Page should still mount and render the workspace name or a header.
    await expect(page.getByRole('heading').first().or(page.locator('h1').first())).toBeVisible({
      timeout: 10_000,
    });
  });
});

test.describe('Flow — task board', () => {
  test.beforeEach(async ({ context }) => {
    await primeAuthCookie(context);
  });

  test('project board renders with tasks in multiple statuses', async ({ page }) => {
    const project = fakeProject({ id: '1', name: 'Alpha Project' });

    const todoTask = fakeTask({
      id: '101',
      title: 'Todo task one',
      status: 'todo',
      projectId: '1',
    });
    const inProgressTask = fakeTask({
      id: '102',
      title: 'In progress task',
      status: 'in_progress',
      projectId: '1',
    });
    const doneTask = fakeTask({
      id: '103',
      title: 'Completed task',
      status: 'done',
      projectId: '1',
    });

    await installFlowStubs(page, {
      projects: [project],
      tasks: [todoTask, inProgressTask, doneTask],
    });

    await page.goto(`${FLOW_URL}/w/10/p/1`);

    // The project board should mount and display the project header.
    await expect(page.getByText('Alpha Project').first()).toBeVisible({ timeout: 15_000 });
  });

  test('project list view renders tasks grouped by status', async ({ page }) => {
    const project = fakeProject({ id: '1', name: 'Alpha Project' });

    const todoTask = fakeTask({
      id: '101',
      title: 'Todo task one',
      status: 'todo',
      projectId: '1',
    });
    const doneTask = fakeTask({
      id: '102',
      title: 'Done task one',
      status: 'done',
      projectId: '1',
    });

    await installFlowStubs(page, { projects: [project], tasks: [todoTask, doneTask] });
    await page.goto(`${FLOW_URL}/w/10/p/1/list`);

    // The list view shows the project name in the header.
    await expect(page.getByText('Alpha Project').first()).toBeVisible({ timeout: 15_000 });

    // Status group headers should be present (the component hard-codes these labels).
    await expect(page.getByText('Todo').first()).toBeVisible({ timeout: 10_000 });
    await expect(page.getByText('Done').first()).toBeVisible({ timeout: 10_000 });
  });
});

test.describe('Flow — inbox', () => {
  test.beforeEach(async ({ context }) => {
    await primeAuthCookie(context);
  });

  test('inbox renders notification list', async ({ page }) => {
    const n1 = fakeNotification({
      id: 'n1',
      title: 'Task assigned to you',
      type: 'assigned',
      read: false,
    });
    const n2 = fakeNotification({
      id: 'n2',
      title: 'You were mentioned',
      type: 'mention',
      read: true,
    });

    await installFlowStubs(page, { notifications: [n1, n2] });
    await page.goto(`${FLOW_URL}/inbox`);

    // The inbox heading is always rendered (i18n key inbox.title → "Inbox").
    await expect(page.getByRole('heading').first().or(page.locator('h1').first())).toBeVisible({
      timeout: 10_000,
    });

    // Both notification titles should appear.
    await expect(page.getByText('Task assigned to you').first()).toBeVisible({ timeout: 10_000 });
    await expect(page.getByText('You were mentioned').first()).toBeVisible({ timeout: 10_000 });
  });

  test('inbox shows empty state when there are no notifications', async ({ page }) => {
    await installFlowStubs(page, { notifications: [] });
    await page.goto(`${FLOW_URL}/inbox`);

    // Header always renders.
    await expect(page.getByRole('heading').first().or(page.locator('h1').first())).toBeVisible({
      timeout: 10_000,
    });
  });

  test('inbox renders without console errors', async ({ page }) => {
    const errors: string[] = [];
    page.on('console', (msg) => {
      if (msg.type() === 'error') errors.push(msg.text());
    });

    await installFlowStubs(page, {
      notifications: [fakeNotification()],
    });
    await page.goto(`${FLOW_URL}/inbox`);
    await page.waitForTimeout(2_000);

    const meaningful = errors.filter(
      (e) => !/posthog|sentry|Failed to load resource|WebSocket/i.test(e)
    );
    expect(meaningful, meaningful.join('\n')).toHaveLength(0);
  });
});
