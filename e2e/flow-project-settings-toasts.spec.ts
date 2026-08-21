import { test, expect, type Page, type Route } from './fixtures';

/**
 * Regression coverage for the double-toast bug in ProjectSettings.tsx:
 * useMutation's hook-level onError and the call site's own onError both
 * fired on failure, showing the same error toast twice. Each test here
 * forces the relevant endpoint to 500 and asserts exactly one
 * [data-sonner-toast] renders - proving the call-site's redundant
 * toast.error was removed while the hook's own toast still fires.
 *
 * transferOwnership is not covered here: useFlowTeam's org-member list
 * comes from a Connect-RPC client, not a plain REST endpoint, so it isn't
 * a simple page.route() stub. That call site is covered by a unit test
 * in ProjectSettings.test.tsx instead.
 */

const FLOW_URL = process.env.FLOW_URL ?? 'http://localhost:5183';
const API_HOST = 'http://localhost:9999';

const FAKE_USER = {
  id: 'u1',
  email: 'owner@acme.test',
  fullName: 'Acme Owner',
  role: 'owner',
  apps: ['flow'],
  planSlug: 'pro',
};

const FAKE_PROJECT = {
  id: '1',
  workspaceId: '10',
  name: 'Alpha Project',
  description: 'Existing description',
  color: '#64C5B9',
  status: 'active',
  taskCount: 0,
  completedCount: 0,
  memberAvatars: [],
};

const FAKE_MEMBER = {
  id: 'm1',
  projectId: '1',
  userId: '5',
  name: 'Ada Lovelace',
  email: 'ada@example.com',
  role: 'member',
  addedAt: '2026-01-01T00:00:00Z',
};

async function primeAuthCookie(context: import('@playwright/test').BrowserContext): Promise<void> {
  const fakeJwt = 'eyJhbGciOiJub25lIn0.eyJ1c2VyX2lkIjoiMSJ9.fake_sig';
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

async function installBaseStubs(
  page: Page,
  opts: { workspaces?: Record<string, unknown>[] } = {}
): Promise<void> {
  const workspaces = opts.workspaces ?? [{ id: '10', name: 'Acme Workspace', color: '#64C5B9' }];

  // Broadest catch-all so unstubbed calls (Caret, realtime, etc.) don't stall networkidle.
  await page.route(/localhost:9999\//, (route: Route) => {
    const method = route.request().method();
    route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify(method === 'GET' ? {} : { success: true }),
    });
  });

  await page.route(`${API_HOST}/flow/**`, (route: Route) => {
    const method = route.request().method();
    route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify(
        method === 'GET'
          ? { projects: [], tasks: [], workspaces: [], members: [], notifications: [] }
          : { success: true }
      ),
    });
  });

  await page.route(`${API_HOST}/hub/user/v1/users/me`, (route: Route) =>
    route.fulfill({ status: 200, contentType: 'application/json', body: JSON.stringify(FAKE_USER) })
  );

  await page.route(
    /localhost:9999\/(get-configs|subscriptions|business\/me|hub\/kpis|team|activity|notifications)/,
    (route: Route) =>
      route.fulfill({ status: 200, contentType: 'application/json', body: JSON.stringify({}) })
  );

  await page.route(/localhost:9999\/flow\/workspaces(\?.*)?$/, (route: Route) =>
    route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({ workspaces }),
    })
  );

  await page.route(/localhost:9999\/flow\/projects(\?.*)?$/, (route: Route) =>
    route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({ projects: [FAKE_PROJECT] }),
    })
  );

  // Project members list - GET only. Mutations on this exact path/sub-paths
  // are overridden per-test below to return 500.
  await page.route(
    /localhost:9999\/flow\/project\/v1\/projects\/members(\?.*)?$/,
    (route: Route) => {
      if (route.request().method() === 'GET') {
        route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({ members: [FAKE_MEMBER] }),
        });
        return;
      }
      route.fulfill({ status: 200, contentType: 'application/json', body: '{"success":true}' });
    }
  );
}

async function goToSettings(page: Page): Promise<void> {
  await page.goto(`${FLOW_URL}/w/10/p/1/settings`);
  await page.waitForLoadState('networkidle', { timeout: 30_000 });
  await expect(page.getByRole('heading', { name: 'Settings' })).toBeVisible({ timeout: 15_000 });
}

test.describe('Flow - ProjectSettings error toasts', () => {
  test.beforeEach(async ({ context }) => {
    await primeAuthCookie(context);
  });

  test('saving a project update failure shows exactly one toast', async ({ page }) => {
    await installBaseStubs(page);
    await page.route(/localhost:9999\/flow\/project\/v1\/projects\/1(\?.*)?$/, (route: Route) => {
      if (route.request().method() === 'PUT') {
        route.fulfill({ status: 500, contentType: 'application/json', body: '{"error":"boom"}' });
        return;
      }
      route.fallback();
    });

    await goToSettings(page);
    await page.getByRole('button', { name: 'Save changes' }).click();

    await expect(page.locator('[data-sonner-toast]')).toHaveCount(1, { timeout: 10_000 });
  });

  test('removing a member failure shows exactly one toast', async ({ page }) => {
    await installBaseStubs(page);
    await page.route(
      /localhost:9999\/flow\/project\/v1\/projects\/members(\?.*)?$/,
      (route: Route) => {
        if (route.request().method() === 'DELETE') {
          route.fulfill({ status: 500, contentType: 'application/json', body: '{"error":"boom"}' });
          return;
        }
        route.fallback();
      }
    );

    await goToSettings(page);
    await page.getByRole('button', { name: 'Members' }).click();
    await page.getByTitle('Remove').click();

    await expect(page.locator('[data-sonner-toast]')).toHaveCount(1, { timeout: 10_000 });
  });

  test('updating a member role failure shows exactly one toast', async ({ page }) => {
    await installBaseStubs(page);
    await page.route(
      /localhost:9999\/flow\/project\/v1\/projects\/members\/role(\?.*)?$/,
      (route: Route) => {
        route.fulfill({ status: 500, contentType: 'application/json', body: '{"error":"boom"}' });
      }
    );

    await goToSettings(page);
    await page.getByRole('button', { name: 'Members' }).click();
    await page.locator('select').filter({ hasText: 'Viewer' }).selectOption('admin');

    await expect(page.locator('[data-sonner-toast]')).toHaveCount(1, { timeout: 10_000 });
  });

  test('moving a project failure shows exactly one toast and resets the confirm state', async ({
    page,
  }) => {
    await installBaseStubs(page, {
      workspaces: [
        { id: '10', name: 'Acme Workspace', color: '#64C5B9' },
        { id: '11', name: 'Other Workspace', color: '#111111' },
      ],
    });
    await page.route(
      /localhost:9999\/flow\/project\/v1\/projects\/move(\?.*)?$/,
      (route: Route) => {
        route.fulfill({ status: 500, contentType: 'application/json', body: '{"error":"boom"}' });
      }
    );

    await goToSettings(page);
    await page.getByRole('button', { name: 'Members' }).click();
    await page.getByText('Move to workspace').waitFor();
    await page.locator('select').filter({ hasText: 'Other Workspace' }).selectOption('11');
    await page.getByRole('button', { name: 'Move' }).click();
    await page.getByRole('button', { name: 'Confirm' }).click();

    await expect(page.locator('[data-sonner-toast]')).toHaveCount(1, { timeout: 10_000 });
    await expect(page.getByRole('button', { name: 'Move' })).toBeVisible();
  });

  test('deleting a project failure shows exactly one toast', async ({ page }) => {
    await installBaseStubs(page);
    await page.route(/localhost:9999\/flow\/project\/v1\/projects\/1(\?.*)?$/, (route: Route) => {
      if (route.request().method() === 'DELETE') {
        route.fulfill({ status: 500, contentType: 'application/json', body: '{"error":"boom"}' });
        return;
      }
      route.fallback();
    });

    await goToSettings(page);
    await page.getByRole('button', { name: 'Danger Zone' }).click();
    await page.getByRole('button', { name: 'Delete' }).click();
    await page.getByPlaceholder('Alpha Project').fill('Alpha Project');
    await page.getByRole('button', { name: 'Delete project' }).click();

    await expect(page.locator('[data-sonner-toast]')).toHaveCount(1, { timeout: 10_000 });
  });
});
