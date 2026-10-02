import { flowAuthFields } from './support/flow-auth';
import { test, expect, type Route } from './fixtures';
import { captureResponsivePair } from './support/screenshot';

/**
 * Screenshots of My tasks, at desktop and at the 390px floor.
 *
 * Tagged `@screenshots` so it stays out of the normal run: `pnpm test:screenshots`.
 *
 * `docs/pr-screenshots.md` says the capture lives on the public menu because
 * "every authenticated app's spec currently redirects to the auth app before
 * making a single API call, so a capture placed on one of those screens would
 * have shipped exercising nothing". That is no longer true, and this spec is the
 * demonstration: `flow-task-detail.spec.ts` and `flow-inbox-views.spec.ts` both
 * assert real content on authenticated screens, using the same primed cookie
 * used here. So an authenticated capture is possible, and the note in that doc
 * is corrected alongside this.
 *
 * It matters because a screenshot is part of the definition of done for a change
 * a user can see, and the two screens changed in rtylr#1077 and rtylr#1215 are
 * both behind the guard.
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

/** Relative to the run, so the sections are the ones being demonstrated. */
function dayOffset(days: number): string {
  const d = new Date();
  d.setDate(d.getDate() + days);
  return [
    d.getFullYear(),
    String(d.getMonth() + 1).padStart(2, '0'),
    String(d.getDate()).padStart(2, '0'),
  ].join('-');
}

/** One task per section, so every heading in the urgency order appears. */
const MY_TASKS = [
  { id: '201', title: 'Chase the grease trap inspection', due: dayOffset(-15), priority: 'urgent' },
  { id: '202', title: 'Re: invoice 4471, follow up again', due: dayOffset(-2), priority: 'high' },
  { id: '203', title: 'Approve the Ras Beirut rota', due: dayOffset(0), priority: 'high' },
  {
    id: '204',
    title: 'Send the supplier the revised order',
    due: dayOffset(3),
    priority: 'medium',
  },
  { id: '205', title: 'Quarterly stock count', due: dayOffset(45), priority: 'low' },
  { id: '206', title: 'Draft the staff handbook update', due: '', priority: 'low' },
];

/**
 * snake_case, because the v2 wire is snake_case and the generated client drops
 * what it does not recognise. A camelCase payload arrives as six tasks with no
 * due dates, so every one lands in "No date", which is the bucket that starts
 * collapsed: the capture then shows a single closed section and looks plausible.
 * That is why the test asserts content before capturing.
 */
const wireTask = (t: (typeof MY_TASKS)[number]) => ({
  id: t.id,
  project_id: '1',
  owner_id: '1001',
  title: t.title,
  status: 'todo',
  priority: t.priority,
  assignee_id: '1001',
  assignee_name: 'Test Owner',
  assignee_initials: 'TO',
  due_date: t.due,
  archived: false,
  labels: [],
  identifier: `AP-${t.id}`,
  project_name: 'Ras Beirut',
  created_at: new Date().toISOString(),
  updated_at: new Date().toISOString(),
});

async function primeAuthCookie(context: import('@playwright/test').BrowserContext): Promise<void> {
  await context.addCookies([
    {
      name: 'token',
      value: 'e2e-test-token',
      domain: 'localhost',
      path: '/',
      httpOnly: false,
      sameSite: 'Lax',
      secure: false,
    },
  ]);
}

test.describe('Flow My tasks screenshots @screenshots', () => {
  test.beforeEach(async ({ context, page }) => {
    await primeAuthCookie(context);

    // Catch-all first, so an endpoint this screen happens to touch cannot leave
    // a request in flight and stall the capture.
    await page.route(/localhost:9999\//, (route: Route) =>
      route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify(route.request().method() === 'GET' ? {} : { success: true }),
      })
    );

    await page.route(`${API_HOST}/hub/user/v1/users/me`, (route: Route) =>
      route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify(FAKE_USER),
      })
    );

    await page.route(/localhost:9999\/flow\/task\/v2\/tasks\/my(\?.*)?$/, (route: Route) =>
      route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({ tasks: MY_TASKS.map(wireTask) }),
      })
    );
  });

  test('captures My tasks at desktop and at the 390px floor', async ({ page }) => {
    await page.goto(`${FLOW_URL}/my-issues`);
    await page.waitForTimeout(1500);

    // The section that rtylr#1077 added. Asserted before capturing, so a blank
    // or redirected page cannot be mistaken for a successful shot.
    await expect(page.getByRole('button', { expanded: true }).first()).toBeVisible();
    await expect(page.getByText('Chase the grease trap inspection')).toBeVisible();

    const shots = await captureResponsivePair(page, 'flow-my-tasks');

    // Measured, not eyeballed off the image: a screenshot cannot show you that
    // the page scrolls sideways at 390px.
    expect(shots.overflowsAt390).toBe(false);
  });
});
