import { test, expect, type Page, type Route } from './fixtures';

/**
 * Throwaway verification for fix/hr-payslip-typed-search.
 *
 * Confirms the typed `validateSearch` refactor still resolves the run the user
 * clicked. The risk being checked is a silent regression of #44: if the search
 * schema turned a valid `runId` into `undefined`, `useGetMyPayrollRun` would
 * skip the fetch and the detail view would render permanently empty.
 */

const HR_URL = process.env.HR_URL ?? 'http://localhost:5180';
const API = 'http://localhost:9999';

const FAKE_USER = {
  id: 1,
  email: 'hr-admin@acme.test',
  full_name: 'HR Admin',
  business_id: '1',
  role: 'admin',
  apps: ['hr'],
  plan_slug: 'pro',
};

const RUNS = [
  {
    id: 41,
    period_start: '2026-06-01',
    period_end: '2026-06-15',
    status: 'Approved',
    total_net: 1111,
  },
  {
    id: 42,
    period_start: '2026-07-01',
    period_end: '2026-07-15',
    status: 'Approved',
    total_net: 2222,
  },
  {
    id: 43,
    period_start: '2026-08-01',
    period_end: '2026-08-15',
    status: 'Approved',
    total_net: 3333,
  },
];

/** Per-run entry, so each detail page renders a distinguishable net pay. */
function entryFor(id: number): Record<string, unknown> {
  const run = RUNS.find((r) => r.id === id);
  return {
    id: id * 10,
    employee_id: 4101,
    base_pay: (run?.total_net ?? 0) - 100,
    overtime_pay: 100,
    bonuses: 0,
    deductions: 0,
    net_pay: run?.total_net ?? 0,
  };
}

/**
 * Records the raw `<id>` segment of every `/hr/payroll/my-runs/<id>` request.
 *
 * Deliberately captures the raw string rather than a number: the whole point of
 * the out-of-range cases is that the old code issued requests for ids like
 * `-3`, and a `\d+` pattern would silently miss exactly those, letting the
 * request fall through to the catch-all and never be recorded.
 */
async function installStubs(page: Page, requested: string[]): Promise<void> {
  await page.route(`${API}/hr/**`, (route: Route) =>
    route.fulfill({ status: 200, contentType: 'application/json', body: '{}' })
  );
  await page.route(`${API}/subscriptions/**`, (route: Route) =>
    route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({ subscription: null, billing: null }),
    })
  );
  await page.route(`${API}/hub/user/v1/users/me`, (route: Route) =>
    route.fulfill({ status: 200, contentType: 'application/json', body: JSON.stringify(FAKE_USER) })
  );
  await page.route(`${API}/hr/dashboard/v1/dashboard`, (route: Route) =>
    route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({
        totalStaff: 1,
        onShiftToday: 0,
        openPositions: 0,
        monthlyPayroll: 0,
        recentActivity: [],
        todayAtGlance: { shiftsScheduled: 0, newApplications: 0, pendingApprovals: 0 },
        config: {
          workWeekStart: 'Monday',
          overtimeAfterHours: 40,
          ptoPolicyDays: 15,
          currency: 'USD',
          weeklyHours: 40,
          industry: 'restaurant',
          payrollConfigured: true,
          payrollEnabled: true,
        },
      }),
    })
  );

  // HRConfigRoute gates every feature route on this; without a
  // `setup_completed_at` it redirects to /setup and never renders payroll.
  await page.route(`${API}/hr/config`, (route: Route) =>
    route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({
        config: {
          id: 1,
          setup_completed_at: '1767225600',
          work_week_start: 'Monday',
          currency: 'USD',
          weekly_hours: 40,
          industry: 'restaurant',
          payroll_enabled: true,
          payroll_configured: true,
        },
      }),
    })
  );

  // List. Registered before the detail handler so the more specific one wins.
  await page.route(/localhost:9999\/hr\/payroll\/my-runs(\?.*)?$/, (route: Route) =>
    route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({ runs: RUNS, total: RUNS.length }),
    })
  );

  // Detail. Records the id so the test can assert which run was resolved.
  await page.route(/localhost:9999\/hr\/payroll\/my-runs\/([^/?]+)$/, (route: Route) => {
    const raw = route.request().url().split('/').pop() ?? '';
    requested.push(raw);
    const run = RUNS.find((r) => String(r.id) === raw);
    if (!run) {
      route.fulfill({ status: 404, contentType: 'application/json', body: '{}' });
      return;
    }
    route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({ run: { ...run, entries: [entryFor(run.id)] } }),
    });
  });
}

test.describe('pay slip typed search', () => {
  test.beforeEach(async ({ context }) => {
    await context.addCookies([{ name: 'token', value: 'test-token', url: HR_URL }]);
  });

  test('clicking a run opens that run, and back returns to the list', async ({ page }) => {
    const requested: string[] = [];
    await installStubs(page, requested);

    await page.goto(`${HR_URL}/payroll/my-payslips`);
    await expect(page.getByText('2026-07-01, 2026-07-15')).toBeVisible({ timeout: 15_000 });

    // Click the MIDDLE run: a bug that defaults to the first row would still
    // look correct if we clicked row one.
    await page.getByText('2026-07-01, 2026-07-15').click();

    await expect(page).toHaveURL(/runId=42/, { timeout: 10_000 });
    // The amount renders twice (gross row and net pay row), hence `.first()`.
    await expect(page.getByText('$2,222.00').first()).toBeVisible({ timeout: 10_000 });
    await expect(page.getByText('Pay period: 2026-07-01, 2026-07-15')).toBeVisible();
    expect(requested).toContain('42');
    expect(requested).not.toContain('41');

    // Back returns to the list, still populated.
    await page.goBack();
    await expect(page.getByText('2026-08-01, 2026-08-15')).toBeVisible({ timeout: 10_000 });

    // A different run resolves to its own data, not a cached first run.
    await page.getByText('2026-08-01, 2026-08-15').click();
    await expect(page).toHaveURL(/runId=43/, { timeout: 10_000 });
    await expect(page.getByText('$3,333.00').first()).toBeVisible({ timeout: 10_000 });
    await expect(page.getByText('Pay period: 2026-08-01, 2026-08-15')).toBeVisible();
    expect(requested).toContain('43');
  });

  test('a malformed runId shows the empty state instead of an error boundary', async ({ page }) => {
    const requested: string[] = [];
    await installStubs(page, requested);

    await page.goto(`${HR_URL}/payroll/my-payslips/view?runId=abc`);

    // The route must still render its own component - not the error boundary.
    await expect(page.getByRole('button', { name: /back/i })).toBeVisible({ timeout: 15_000 });
    expect(requested).toHaveLength(0);
  });

  test('a missing runId shows the empty state instead of an error boundary', async ({ page }) => {
    const requested: string[] = [];
    await installStubs(page, requested);

    await page.goto(`${HR_URL}/payroll/my-payslips/view`);

    await expect(page.getByRole('button', { name: /back/i })).toBeVisible({ timeout: 15_000 });
    expect(requested).toHaveLength(0);
  });

  /**
   * This is the case that distinguishes the typed search from the manual
   * `Number(sp.get('runId'))` parsing it replaced.
   *
   * `Number('-3')` is a perfectly good number, so the old code passed it to
   * `useGetMyPayrollRun`, whose `enabled: !!id` saw a truthy value and issued
   * GET /hr/payroll/my-runs/-3 - a request for an id that cannot exist. The
   * schema's `.positive()` now rejects it before the query is ever enabled.
   *
   * Run ids are backend serial primary keys, so a negative or fractional id is
   * always malformed input, never a real run.
   */
  for (const [label, value] of [
    ['negative', '-3'],
    ['fractional', '12.5'],
  ]) {
    test(`does not call the backend for a ${label} runId`, async ({ page }) => {
      const requested: string[] = [];
      await installStubs(page, requested);

      await page.goto(`${HR_URL}/payroll/my-payslips/view?runId=${value}`);

      await expect(page.getByRole('button', { name: /back/i })).toBeVisible({ timeout: 15_000 });
      expect(requested).toHaveLength(0);
    });
  }
});
