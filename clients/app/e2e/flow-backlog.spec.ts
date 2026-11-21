import { flowAuthFields } from './support/flow-auth';
import { test, expect, type Route } from './fixtures';
import { captureForPR } from './support/screenshot';
import { defaultStatuses, stubStatuses } from './support/flow-statuses';

/**
 * Flow project Backlog (T-570): the active cycle, upcoming cycles and
 * unplanned work on one screen, with drag between sections.
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

const PROJECT = {
  id: '1',
  name: 'Alpha Project',
  description: '',
  icon: 'folder',
  color: '#64C5B9',
  status: 'active',
  visibility: 'private',
  task_count: 4,
  completed_count: 1,
};

const CYCLES = [
  {
    id: '50',
    project_id: '1',
    name: 'Sprint 4',
    status: 'active',
    start_date: '2026-11-23',
    end_date: '2026-12-04',
    goal: 'Ship the planning screen',
    scope: 2,
    completed: 1,
    estimate_minutes: 540,
    carried_over: [{ from_cycle_id: '49', from_cycle_name: 'Sprint 3', task_ids: ['101'] }],
  },
  {
    id: '51',
    project_id: '1',
    name: 'Sprint 5',
    status: 'upcoming',
    start_date: '2026-12-07',
    end_date: '2026-12-18',
    scope: 1,
  },
];

const wireTask = (id: string, title: string, extra: Record<string, unknown> = {}) => ({
  id,
  project_id: '1',
  owner_id: '1001',
  title,
  description: '',
  status: 'todo',
  priority: 'medium',
  assignee_id: '1001',
  assignee_name: 'Test Owner',
  project_name: 'Alpha Project',
  identifier: `ALP-${id}`,
  archived: false,
  created_at: `2026-11-0${id.slice(-1)}T00:00:00Z`,
  ...extra,
});

const TASKS = [
  wireTask('101', 'Wire the move endpoint', {
    cycle_id: '50',
    plan_rank: '1024',
    effort_minutes: 480,
  }),
  wireTask('102', 'Draft the release note', {
    cycle_id: '50',
    plan_rank: '2048',
    status: 'done',
    effort_minutes: 60,
  }),
  wireTask('103', 'Capacity view polish', { cycle_id: '51', plan_rank: '1024' }),
  wireTask('104', 'Arabic copy review'),
];

test.describe('Flow - Project Backlog', () => {
  test.beforeEach(async ({ context, page }) => {
    await context.addCookies([
      { name: 'token', value: 'e2e-test-token', domain: 'localhost', path: '/', sameSite: 'Lax' },
      { name: 'rtylr-consent', value: 'denied', domain: 'localhost', path: '/', sameSite: 'Lax' },
    ]);
    await page.route(`${API_HOST}/hub/user/v1/users/me`, (route: Route) =>
      route.fulfill({ json: FAKE_USER })
    );
    await page.route(/localhost:9999\/flow\/workspace\/v1\/workspaces/, (route: Route) =>
      route.fulfill({ json: { workspaces: [] } })
    );
    await stubStatuses(page, () => defaultStatuses('1'));
    await page.route(/localhost:9999\/flow\/project\/v2\/projects\/1$/, (route: Route) =>
      route.fulfill({ json: { project: PROJECT } })
    );
    await page.route(/localhost:9999\/flow\/project\/v2\/projects(\?.*)?$/, (route: Route) =>
      route.fulfill({ json: { projects: [PROJECT], pagination: { total: 1 } } })
    );
    await page.route(/localhost:9999\/flow\/project\/v2\/projects\/members(\?.*)?$/, (route) =>
      route.fulfill({
        json: {
          members: [{ id: '1', project_id: '1', user_id: '1001', name: 'Test Owner', role: 'admin' }],
          pagination: { total: 1 },
        },
      })
    );
    await page.route(/localhost:9999\/flow\/task\/v2\/tasks(\?.*)?$/, (route: Route) =>
      route.fulfill({ json: { tasks: TASKS, pagination: { total: TASKS.length } } })
    );
    await page.route(/localhost:9999\/flow\/cycle\/v2\/cycles(\?.*)?$/, (route: Route) =>
      route.fulfill({ json: { cycles: CYCLES, pagination: { total: CYCLES.length } } })
    );
    await page.route(/localhost:9999\/flow\/cycle\/v2\/cycles\/\d+\/capacity$/, (route: Route) =>
      route.fulfill({
        json: {
          leave_applied: true,
          minutes_per_day: 480,
          members: [
            {
              user_id: '1001',
              name: 'Test Owner',
              working_days: 10,
              leave_days: 2,
              available_minutes: 3840,
              committed_minutes: 3600,
              task_count: 1,
            },
          ],
        },
      })
    );
  });

  test('lists the active cycle, the upcoming cycle and unplanned work @screenshots', async ({
    page,
  }) => {
    await page.setViewportSize({ width: 1280, height: 900 });
    await page.goto(`${FLOW_URL}/p/1/backlog`);
    const sections = page.locator('[data-backlog-section]');
    await expect(sections).toHaveCount(3, { timeout: 15_000 });
    await expect(sections.nth(0).getByRole('heading', { name: 'Sprint 4' })).toBeVisible();
    await expect(sections.nth(0).getByText('Ship the planning screen')).toBeVisible();
    await expect(sections.nth(0).getByText(/carried over from Sprint 3/)).toBeVisible();
    await expect(sections.nth(1).getByRole('heading', { name: 'Sprint 5' })).toBeVisible();
    await expect(sections.nth(2).getByText('Arabic copy review')).toBeVisible();
    // A second cycle cannot be started while Sprint 4 runs.
    await expect(sections.nth(1).getByRole('button', { name: 'Start cycle' })).toBeDisabled();

    await sections.nth(0).getByRole('button', { name: 'Capacity' }).click();
    await expect(sections.nth(0).getByText('2 days of approved leave')).toBeVisible();
    await captureForPR(page, 'flow-backlog-desktop');

    await page.setViewportSize({ width: 820, height: 1180 });
    await captureForPR(page, 'flow-backlog-tablet');
    const overflows = await page.evaluate(
      () => document.documentElement.scrollWidth > document.documentElement.clientWidth
    );
    expect(overflows).toBe(false);
  });

  test('dragging a task into another cycle sends one move', async ({ page }) => {
    let moveBody: string | null = null;
    await page.route(/localhost:9999\/flow\/cycle\/v2\/cycles\/tasks\/order$/, (route: Route) => {
      moveBody = route.request().postData();
      return route.fulfill({ json: { cycle_id: '51', plan_rank: '0' } });
    });
    await page.setViewportSize({ width: 1280, height: 900 });
    await page.goto(`${FLOW_URL}/p/1/backlog`);
    const row = page.locator('[data-backlog-row="104"]');
    const target = page.locator('[data-backlog-row="103"]');
    await expect(row).toBeVisible({ timeout: 15_000 });

    const from = await row.boundingBox();
    const to = await target.boundingBox();
    if (!from || !to) throw new Error('rows are not laid out');
    await page.mouse.move(from.x + 40, from.y + from.height / 2);
    await page.mouse.down();
    await page.mouse.move(from.x + 40, from.y - 20, { steps: 4 });
    await page.mouse.move(to.x + 40, to.y + 4, { steps: 8 });
    await page.mouse.up();

    await expect.poll(() => moveBody, { timeout: 10_000 }).not.toBeNull();
    expect(JSON.parse(moveBody ?? '{}')).toEqual({
      task_id: '104',
      cycle_id: '51',
      before_task_id: '103',
    });
  });
});
