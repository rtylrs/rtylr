import { flowAuthFields } from './support/flow-auth';
import { test, expect, type Route } from './fixtures';
import { captureResponsivePair } from './support/screenshot';

/**
 * Flow Task Detail E2E Specs
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

function wireTaskDetail(): Record<string, unknown> {
  const task = FAKE_TASK_DETAIL.task;
  return {
    task: {
      id: task.id,
      project_id: task.projectId,
      owner_id: task.ownerId,
      title: task.title,
      description: task.description,
      status: task.status,
      priority: task.priority,
      assignee_id: task.assigneeId,
      assignee_name: task.assignee,
      assignee_initials: task.assigneeInitials,
      reporter_id: task.reporterId,
      due_date: task.dueDate,
      start_date: task.startDate,
      estimated_hours: task.estimatedHours,
      actual_hours: task.actualHours,
      archived: task.archived,
      labels: task.labels.map(String),
      identifier: task.identifier,
      created_at: task.createdAt,
      updated_at: task.updatedAt,
    },
    comments: FAKE_TASK_DETAIL.comments.map((comment) => ({
      id: comment.id,
      task_id: comment.taskId,
      user_id: comment.userId,
      content: comment.content,
      is_edited: comment.isEdited,
      created_at: comment.createdAt,
    })),
    subtasks: FAKE_TASK_DETAIL.subtasks.map((subtask) => ({
      id: subtask.id,
      project_id: '1',
      title: subtask.title,
      status: subtask.status,
    })),
    activity: [],
  };
}

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

    await page.route(/localhost:9999\/flow\/task\/v2\/tasks\/101/, (route: Route) =>
      route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify(FAKE_TASK_DETAIL),
      })
    );

    await page.route(/localhost:9999\/flow\/task\/v2\/tasks\/101$/, (route: Route) =>
      route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify(wireTaskDetail()),
      })
    );

    await page.route(/localhost:9999\/flow\/task\/v2\/tasks\/101\/activity$/, (route: Route) =>
      route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({ activity: [] }),
      })
    );

    await page.route(/localhost:9999\/flow\/task\/v2\/tasks\/101\/time-logs$/, (route: Route) =>
      route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({ time_logs: [] }),
      })
    );

    await page.route(/localhost:9999\/flow\/task\/v2\/tasks\/101\/attachments$/, (route: Route) =>
      route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({ attachments: [] }),
      })
    );

    await page.route(/localhost:9999\/flow\/task\/v2\/tasks\/101\/activity/, (route: Route) =>
      route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({ activity: [] }),
      })
    );

    await page.route(/localhost:9999\/flow\/task\/v2\/tasks\/101\/time-logs/, (route: Route) =>
      route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({ timeLogs: [] }),
      })
    );

    await page.route(/localhost:9999\/flow\/task\/v2\/tasks\/101\/attachments/, (route: Route) =>
      route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({ attachments: [] }),
      })
    );

    await page.route(/localhost:9999\/flow\/label\/v2\/labels/, (route: Route) =>
      route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({ labels: [] }),
      })
    );

    await page.route(/localhost:9999\/flow\/project\/v2\/projects/, (route: Route) =>
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

  /**
   * Captures task detail for a PR body, and asserts what a screenshot cannot.
   *
   * In this file rather than its own so it reuses the beforeEach above: task
   * detail needs the primed cookie, `users/me`, and both the v1 and v2 task
   * endpoints.
   */
  test('captures task detail at desktop and at the 390px floor @screenshots', async ({ page }) => {
    await page.goto(`${FLOW_URL}/task/101`);
    await page.waitForLoadState('networkidle');
    await page.waitForTimeout(800);

    // Asserted before capturing, so a blank or redirected page cannot pass as a
    // successful shot.
    // The title is an editable input, so its value is not text content: the
    // existing test above matches on the attribute for the same reason.
    await expect(page.locator('input[value="Implement SSO Auth Flow"]')).toBeVisible();

    // One STATUS row, not two. These were adjacent rows reading `STATUS` and
    // `STATUSES`, one concept as two, with labels a letter apart (rtylr#1079).
    const statusLabels = page.getByText(/^STATUS(ES)?$/i);
    expect(await statusLabels.count()).toBe(1);

    // The time total. This read `20100226:30 of 40:00` because durationMinutes
    // arrives as a string and the sum concatenated it (rtylr#1226).
    const body = await page.locator('body').innerText();
    expect(body).not.toMatch(/\d{7,}:\d\d/);

    const shots = await captureResponsivePair(page, 'flow-task-detail');
    expect(shots.overflowsAt390).toBe(false);
  });

  test('task detail renders task title, comments, subtasks', async ({ page }) => {
    await page.goto(`${FLOW_URL}/task/101`);
    await page.waitForLoadState('networkidle');

    await expect(page.locator('input[value="Implement SSO Auth Flow"]')).toBeVisible();
    await expect(page.getByText('Initial comment on SSO implementation.')).toBeVisible();
    await expect(page.getByText('Configure OAuth callback')).toBeVisible();
  });
});
