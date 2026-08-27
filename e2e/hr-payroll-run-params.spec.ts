import { test, expect, type Page, type Route } from './fixtures';

/**
 * Verification for the typed `/payroll/$id` path param (issue #46).
 *
 * Mirrors `hr-payslip-search.spec.ts`, which covers the search-param half of
 * the same defect class. Here the risk is the admin payroll run detail: the
 * old code read the param with `useParams({ strict: false })` and forwarded
 * `Number(id)`, so `Number('-3')` was truthy and `useGetPayrollRun` issued
 * GET /hr/payroll/runs/-3 - a request for an id that cannot exist.
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
    status: 'approved',
    total_net: 1111,
  },
  {
    id: 42,
    period_start: '2026-07-01',
    period_end: '2026-07-15',
    status: 'approved',
    total_net: 2222,
  },
  {
    id: 43,
    period_start: '2026-08-01',
    period_end: '2026-08-15',
    status: 'approved',
    total_net: 3333,
  },
];

/** Per-run entry, so each detail page renders a distinguishable net pay. */
function entryFor(id: number): Record<string, unknown> {
  const run = RUNS.find((r) => r.id === id);
  return {
    id: id * 10,
    employee_id: 4101,
    employee_name: 'Test Employee',
    hours_worked: 80,
    base_pay: (run?.total_net ?? 0) - 100,
    overtime_pay: 100,
    bonuses: 0,
    deductions: 0,
    net_pay: run?.total_net ?? 0,
  };
}

/**
 * Records the raw `<id>` segment of every `/hr/payroll/runs/<id>` request.
 *
 * Deliberately captures the raw string rather than a number: the whole point
 * of the out-of-range cases is that the old code issued requests for ids like
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
  // `payroll_period` must be present too: `Payroll.tsx` renders its
  // "configure payroll first" empty state without it.
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
          payroll_period: 'monthly',
        },
      }),
    })
  );

  // List. Registered before the detail handler so the more specific one wins.
  await page.route(/localhost:9999\/hr\/payroll\/runs(\?.*)?$/, (route: Route) =>
    route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({ runs: RUNS, total: RUNS.length }),
    })
  );

  // Detail. Records the id so the test can assert which run was resolved.
  await page.route(/localhost:9999\/hr\/payroll\/runs\/([^/?]+)$/, (route: Route) => {
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

test.describe('payroll run typed path param', () => {
  test.beforeEach(async ({ context }) => {
    await context.addCookies([{ name: 'token', value: 'test-token', url: HR_URL }]);
  });

  test('clicking a run opens that run', async ({ page }) => {
    const requested: string[] = [];
    await installStubs(page, requested);

    await page.goto(`${HR_URL}/payroll`);
    await expect(page.getByText('2026-07-01, 2026-07-15')).toBeVisible({ timeout: 15_000 });

    // Click the MIDDLE run: a bug that defaults to the first row would still
    // look correct if we clicked row one.
    await page.getByText('2026-07-01, 2026-07-15').click();

    await expect(page).toHaveURL(/\/payroll\/42$/, { timeout: 10_000 });
    // `PayrollRunDetail`'s formatter uses `minimumFractionDigits: 0`, so the
    // net renders as `$2,222`, not `$2,222.00`.
    await expect(page.getByText('$2,222').first()).toBeVisible({ timeout: 10_000 });
    expect(requested).toContain('42');
    expect(requested).not.toContain('41');
  });

  /**
   * The cases that distinguish the typed param from the manual
   * `Number(useParams({ strict: false }).id)` parsing it replaced.
   *
   * Run ids are backend serial primary keys, so a negative or fractional id
   * is always malformed input, never a real run. The route's `params.parse`
   * schema resolves these to `undefined` before the query is ever enabled,
   * and the component renders its own "not found" state - not the router
   * error boundary, and with no backend request at all.
   */
  for (const [label, value] of [
    ['negative', '-3'],
    ['fractional', '12.5'],
  ]) {
    test(`does not call the backend for a ${label} run id`, async ({ page }) => {
      const requested: string[] = [];
      await installStubs(page, requested);

      await page.goto(`${HR_URL}/payroll/${value}`);

      // The route must still render its own component - not the error
      // boundary. The "back to payroll runs" link only exists in the
      // feature's not-found state.
      await expect(page.getByRole('button', { name: /back to payroll runs/i })).toBeVisible({
        timeout: 15_000,
      });
      expect(requested).toHaveLength(0);
    });
  }
});
