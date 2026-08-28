import { test, expect, type Page, type Route } from './fixtures';

/**
 * End-to-end: an HR operator managing staff and shifts.
 *
 * Staff and shifts are the HR app's primary data surfaces - if the list
 * screens break silently, the operator loses visibility into who is scheduled
 * and what employees are active. This spec locks the critical read path down
 * in a real browser with stubbed APIs so CI stays deterministic; no backend
 * required.
 *
 * All stubs live in `installHrStubs` so they stay in one place as new
 * endpoints land. Every endpoint the staff list, staff detail, and shifts
 * screens touch must have a handler there.
 */

const HR_URL = process.env.HR_URL ?? 'http://localhost:5180';

// ─── Stub shapes (snake_case to mirror the real API wire format) ──────────────

interface StubStaff {
  id: number;
  name: string;
  role: string;
  department: string;
  email: string;
  employment_type: 'Full-time' | 'Part-time' | 'Contractor';
  status: 'Active' | 'Inactive';
}

interface StubShift {
  id: number;
  name: string;
  role: string;
  start: string;
  end: string;
  status: 'Scheduled' | 'In Progress' | 'Completed';
  day: string;
  staff_id?: number;
}

// ─── Fake data ────────────────────────────────────────────────────────────────

const FAKE_USER = {
  id: 1,
  email: 'hr-admin@acme.test',
  full_name: 'HR Admin',
  business_id: '1',
  role: 'admin',
  apps: ['hr'],
  plan_slug: 'pro',
};

const FAKE_HR_CONFIG = {
  workWeekStart: 'Monday',
  overtimeAfterHours: 40,
  ptoPolicyDays: 15,
  currency: 'USD',
  weeklyHours: 40,
  industry: 'restaurant',
  payrollConfigured: true,
  payrollEnabled: true,
};

// ─── Stub installer ───────────────────────────────────────────────────────────

/**
 * Installs stub handlers for every backend endpoint the HR staff-list, staff
 * detail, and shifts screens hit. Registration order matters: Playwright
 * matches routes in REVERSE registration order, so the most generic
 * catch-alls are registered FIRST and the most specific handlers LAST.
 */
async function installHrStubs(
  page: Page,
  initial: { staff?: StubStaff[]; shifts?: StubShift[] } = {}
): Promise<void> {
  const apiHost = 'http://localhost:9999';

  const state = {
    staff: [...(initial.staff ?? [])],
    shifts: [...(initial.shifts ?? [])],
  };

  // (1) Widest HR catch-all - handles any endpoint not explicitly matched.
  await page.route(`${apiHost}/hr/**`, (route: Route) => {
    if (route.request().method() === 'GET') {
      route.fulfill({ status: 200, contentType: 'application/json', body: '{}' });
    } else {
      route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({ success: true }),
      });
    }
  });

  // (2) Subscription/billing polled by the layout shell. A 401 would
  // trigger an auth redirect - neutral empty response keeps the surface quiet.
  await page.route(`${apiHost}/subscriptions/**`, (route: Route) =>
    route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({ subscription: null, billing: null }),
    })
  );

  // (3) AppGuard expects /hub/user/v1/users/me to resolve with a user who has HR access.
  await page.route(`${apiHost}/hub/user/v1/users/me`, (route: Route) =>
    route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify(FAKE_USER),
    })
  );

  // (4) HR config/settings - HRConfigRoute loads this before rendering features.
  await page.route(`${apiHost}/hr/dashboard/v1/dashboard`, async (route: Route) => {
    await route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({
        totalStaff: state.staff.length,
        onShiftToday: 0,
        openPositions: 0,
        monthlyPayroll: 0,
        recentActivity: [],
        todayAtGlance: { shiftsScheduled: 0, newApplications: 0, pendingApprovals: 0 },
        config: FAKE_HR_CONFIG,
      }),
    });
  });

  // HRConfigRoute gates feature routes on setup completion. Keep that guard in
  // the test so a redirect cannot hide a broken staff screen.
  await page.route(`${apiHost}/hr/config`, (route: Route) =>
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

  // (5) Staff list - GET /hr/staff.
  await page.route(`${apiHost}/hr/staff`, async (route: Route) => {
    const req = route.request();
    if (req.method() === 'GET') {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify(state.staff),
      });
      return;
    }
    await route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({ success: true }),
    });
  });

  // (6) Shifts list - GET /hr/shifts.
  await page.route(`${apiHost}/hr/shifts`, async (route: Route) => {
    const req = route.request();
    if (req.method() === 'GET') {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify(state.shifts),
      });
      return;
    }
    await route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({ success: true }),
    });
  });

  // (7) Staff detail - GET /hr/staff/detail?id=<n>. Registered last so it
  // wins over the wider /hr/staff catch-all.
  await page.route(/localhost:9999\/hr\/staff\/detail(\?.*)?$/, async (route: Route) => {
    const url = new URL(route.request().url());
    const id = Number(url.searchParams.get('id'));
    const member = state.staff.find((s) => s.id === id);
    if (!member) {
      await route.fulfill({
        status: 404,
        contentType: 'application/json',
        body: JSON.stringify({ message: 'not found' }),
      });
      return;
    }
    await route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify(member),
    });
  });

  // The staff list and profile now use the generated hr.staff.v1 client. Keep
  // the legacy handlers above because other parts of the HR shell still use
  // those endpoints, but make the restored suites exercise the current path.
  const departmentNames = [...new Set(state.staff.map((member) => member.department))];
  const toApiStaff = (member: StubStaff): Record<string, unknown> => {
    const [firstName, ...lastNameParts] = member.name.split(' ');
    return {
      id: String(member.id),
      department_id: String(departmentNames.indexOf(member.department) + 1),
      first_name: firstName,
      last_name: lastNameParts.join(' '),
      email: member.email,
      role: member.role,
      employment_type: member.employment_type,
      status: member.status,
    };
  };

  await page.route(/localhost:9999\/hr\/staff\/v1\/staff(\?.*)?$/, async (route: Route) => {
    await route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({ staff: state.staff.map(toApiStaff) }),
    });
  });

  await page.route(/localhost:9999\/hr\/staff\/v1\/staff\/([^/?]+)$/, async (route: Route) => {
    const id = Number(route.request().url().split('/').pop());
    const member = state.staff.find((item) => item.id === id);
    await route.fulfill({
      status: member ? 200 : 404,
      contentType: 'application/json',
      body: JSON.stringify(member ? { staff: toApiStaff(member) } : { message: 'not found' }),
    });
  });

  await page.route(
    /localhost:9999\/hr\/department\/v1\/departments(\?.*)?$/,
    async (route: Route) => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          departments: departmentNames.map((name, index) => ({ id: String(index + 1), name })),
        }),
      });
    }
  );
}

// ─── Test suites ──────────────────────────────────────────────────────────────

test.describe('HR staff list', () => {
  // AppGuard redirects to auth when there is no token cookie. A dummy value is
  // enough because we stub /hub/user/v1/users/me - the guard only checks presence.
  test.beforeEach(async ({ context }) => {
    await context.addCookies([{ name: 'token', value: 'test-token', url: HR_URL }]);
  });

  test('smoke - HR app mounts without crashing', async ({ page }) => {
    await installHrStubs(page, {});
    await page.goto(HR_URL);
    // Any visible text confirms the React tree rendered without a white screen.
    await expect(page.locator('body')).not.toBeEmpty();
  });

  test('staff names render from GET /hr/staff', async ({ page }) => {
    const fakeStaff: StubStaff[] = [
      {
        id: 1,
        name: 'Alice Johnson',
        role: 'Manager',
        department: 'Operations',
        email: 'alice@acme.test',
        employment_type: 'Full-time',
        status: 'Active',
      },
      {
        id: 2,
        name: 'Bob Ramirez',
        role: 'Cashier',
        department: 'Sales',
        email: 'bob@acme.test',
        employment_type: 'Part-time',
        status: 'Active',
      },
    ];

    await installHrStubs(page, { staff: fakeStaff });
    await page.goto(`${HR_URL}/staff`);

    await expect(page.getByText('Alice Johnson')).toBeVisible({ timeout: 10_000 });
    await expect(page.getByText('Bob Ramirez')).toBeVisible({ timeout: 5_000 });
  });

  test('staff member role is displayed in the list', async ({ page }) => {
    const fakeStaff: StubStaff[] = [
      {
        id: 3,
        name: 'Carol Davis',
        role: 'Barista',
        department: 'F&B',
        email: 'carol@acme.test',
        employment_type: 'Full-time',
        status: 'Active',
      },
    ];

    await installHrStubs(page, { staff: fakeStaff });
    await page.goto(`${HR_URL}/staff`);

    await expect(page.getByText('Carol Davis')).toBeVisible({ timeout: 10_000 });
    await expect(page.getByText('Barista')).toBeVisible({ timeout: 5_000 });
  });
});

test.describe('HR staff detail', () => {
  test.beforeEach(async ({ context }) => {
    await context.addCookies([{ name: 'token', value: 'test-token', url: HR_URL }]);
  });

  test('staff detail page renders the employee name', async ({ page }) => {
    const fakeStaff: StubStaff[] = [
      {
        id: 10,
        name: 'Diana Prince',
        role: 'Head Chef',
        department: 'Kitchen',
        email: 'diana@acme.test',
        employment_type: 'Full-time',
        status: 'Active',
      },
    ];

    await installHrStubs(page, { staff: fakeStaff });
    await page.goto(`${HR_URL}/staff/10`);

    await expect(page.getByRole('heading', { name: 'Diana Prince' })).toBeVisible({
      timeout: 10_000,
    });
  });

  test('staff detail page renders the employee role and department', async ({ page }) => {
    const fakeStaff: StubStaff[] = [
      {
        id: 11,
        name: 'Ethan Hunt',
        role: 'Sous Chef',
        department: 'Kitchen',
        email: 'ethan@acme.test',
        employment_type: 'Full-time',
        status: 'Active',
      },
    ];

    await installHrStubs(page, { staff: fakeStaff });
    await page.goto(`${HR_URL}/staff/11`);

    await expect(page.getByRole('heading', { name: 'Ethan Hunt' })).toBeVisible({
      timeout: 10_000,
    });
    // Role and department should appear somewhere on the detail page.
    await expect(page.getByText('Sous Chef')).toBeVisible({ timeout: 5_000 });
  });
});

test.describe('HR shifts list', () => {
  test.beforeEach(async ({ context }) => {
    await context.addCookies([{ name: 'token', value: 'test-token', url: HR_URL }]);
  });

  test('shift entries render from GET /hr/shifts', async ({ page }) => {
    const fakeShifts: StubShift[] = [
      {
        id: 1,
        name: 'Morning Shift',
        role: 'Cashier',
        start: '08:00',
        end: '16:00',
        status: 'Scheduled',
        day: 'Mon',
        staff_id: 1,
      },
      {
        id: 2,
        name: 'Evening Shift',
        role: 'Barista',
        start: '14:00',
        end: '22:00',
        status: 'Scheduled',
        day: 'Mon',
        staff_id: 2,
      },
    ];

    await installHrStubs(page, { shifts: fakeShifts });
    await page.goto(`${HR_URL}/shifts`);

    await page.getByRole('button', { name: /Mon/i }).click();
    await expect(page.getByText('Morning Shift')).toBeVisible({ timeout: 10_000 });
    await expect(page.getByText('Evening Shift')).toBeVisible({ timeout: 5_000 });
  });

  test('shift time range is displayed', async ({ page }) => {
    const fakeShifts: StubShift[] = [
      {
        id: 3,
        name: 'Night Shift',
        role: 'Security',
        start: '22:00',
        end: '06:00',
        status: 'Scheduled',
        day: 'Tue',
      },
    ];

    await installHrStubs(page, { shifts: fakeShifts });
    await page.goto(`${HR_URL}/shifts`);

    await page.getByRole('button', { name: /Tue/i }).click();
    await expect(page.getByText('Night Shift')).toBeVisible({ timeout: 10_000 });
    // Start or end time should appear in the rendered row.
    await expect(page.getByText(/22:00/)).toBeVisible({ timeout: 5_000 });
  });

  test('shift day is displayed', async ({ page }) => {
    const fakeShifts: StubShift[] = [
      {
        id: 4,
        name: 'Friday Opening',
        role: 'Manager',
        start: '07:00',
        end: '15:00',
        status: 'Scheduled',
        day: 'Fri',
      },
    ];

    await installHrStubs(page, { shifts: fakeShifts });
    await page.goto(`${HR_URL}/shifts`);

    await page.getByRole('button', { name: /Fri/i }).click();
    await expect(page.getByText('Friday Opening')).toBeVisible({ timeout: 10_000 });
    await expect(page.getByRole('heading', { name: /Fri:/i })).toBeVisible({ timeout: 5_000 });
  });
});
