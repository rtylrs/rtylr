import { stubPortalSession } from './portal-session';
import { test, expect, type Page, type Route } from './fixtures';

const HR_URL = process.env.HR_URL ?? 'http://localhost:5197/hr';
const API = 'http://localhost:9999';

const USER = {
  id: 1,
  email: 'hr-admin@acme.test',
  full_name: 'HR Admin',
  business_id: '1',
  role: 'admin',
  apps: ['hr'],
  plan_slug: 'pro',
};

async function installStubs(page: Page): Promise<void> {
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
  await stubPortalSession(page, page.context(), { isOwner: true, appIds: ['hr'] });
  await page.route(`${API}/hr/config`, (route: Route) =>
    route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({
        config: {
          id: 1,
          setup_completed_at: '1767225600',
          currency: 'USD',
          weekly_hours: 40,
          payroll_enabled: true,
          payroll_configured: true,
          payroll_period: 'monthly',
        },
      }),
    })
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
        config: { currency: 'USD', weeklyHours: 40, payrollConfigured: true, payrollEnabled: true },
      }),
    })
  );
  await page.route(/localhost:9999\/hr\/employees\/me(\?.*)?$/, (route: Route) =>
    route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({ employee: { id: '42' } }),
    })
  );
  await page.route(/localhost:9999\/hr\/payroll\/runs(\?.*)?$/, (route: Route) =>
    route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({
        runs: [
          {
            id: 7,
            period_start: '2026-08-01',
            period_end: '2026-08-31',
            status: 'approved',
            total_gross: 12000,
            total_net: 9600,
            entries: [{ id: 1 }],
          },
        ],
        total: 1,
      }),
    })
  );
  await page.route(/localhost:9999\/hr\/leave\/v1\/balances(\?.*)?$/, (route: Route) =>
    route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({
        balances: [
          {
            leave_type_id: 3,
            entitled_days: 20,
            used_days: 5,
            pending_days: 2,
            remaining_days: 13,
          },
        ],
      }),
    })
  );
  await page.route(`${API}/hr/time-entry/v1/timesheets**`, (route: Route) =>
    route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({
        timesheets: [
          {
            id: 12,
            employee_id: 42,
            period_start: '2026-08-03',
            period_end: '2026-08-09',
            total_minutes: 2400,
            entries: [{ id: 1 }],
            status: 'submitted',
          },
        ],
        total: 1,
      }),
    })
  );
  await page.route(/localhost:9999\/hr\/timesheets(\?.*)?$/, (route: Route) =>
    route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({
        timesheets: [
          {
            id: 12,
            employee_id: 42,
            period_start: '2026-08-03',
            period_end: '2026-08-09',
            total_minutes: 2400,
            entries: [{ id: 1 }],
            status: 'submitted',
          },
        ],
        total: 1,
      }),
    })
  );
}

test.describe('HR critical staff flows', () => {
  test.beforeEach(async ({ context }) => {
    await context.addCookies([{ name: 'token', value: 'test-token', url: HR_URL }]);
  });

  test('payroll shows an approved run and net total', async ({ page }) => {
    await installStubs(page);
    await page.goto(`${HR_URL}/payroll`);
    await expect(page.getByText('2026-08-01, 2026-08-31')).toBeVisible();
    await expect(page.getByText('$9,600.00', { exact: true }).first()).toBeVisible();
  });

  test('leave balances show remaining and used days', async ({ page }) => {
    await installStubs(page);
    await page.goto(`${HR_URL}/leave`);
    await expect(page.getByText('13')).toBeVisible();
    await expect(page.getByText('5')).toBeVisible();
  });

  test('approval chain lists submitted timesheet with approve action', async ({ page }) => {
    await installStubs(page);
    await page.goto(`${HR_URL}/timesheets/approvals`);
    await expect(page.getByText(/Employee #?42/i)).toBeVisible();
    await expect(page.getByRole('button', { name: /approve/i })).toBeVisible();
  });
});
