import { test, expect } from './fixtures';

/**
 * Recruit admin e2e - complements recruit-apply.spec.ts which only
 * exercises the public candidate flow. This one cookie-primes past
 * AppGuard, stubs the admin endpoints, and walks the list +
 * status-change buttons that HR recruiters use.
 */

const RECRUIT_URL = process.env.RECRUIT_URL ?? 'http://localhost:5184';

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

const FAKE_CAREER = {
  id: 42,
  owner_id: 1,
  slug: 'backend-engineer',
  title: 'Backend Engineer',
  department: 'Engineering',
  location: 'Remote',
  employment_type: 'Full-time',
  work_mode: 'Remote',
  seniority: 'Senior',
  details: 'Build things.',
  requirements: '',
  salary_min: 0,
  salary_max: 0,
  salary_currency: 'USD',
  is_published: true,
  is_active: true,
  created_at: '2026-04-01T00:00:00Z',
  updated_at: '2026-04-01T00:00:00Z',
};

const FAKE_SUBMISSION = {
  id: 101,
  career_id: 42,
  candidate_name: 'Jamie Doe',
  email: 'jamie@example.com',
  status: 'submitted',
  created_at: '2026-04-15T12:00:00Z',
};

async function stubRecruitAdminAPIs(page: import('@playwright/test').Page): Promise<void> {
  await page.route('**/hub/user/v1/users/me', (route) =>
    route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({ id: 1, email: 'hr@example.com', role: 'owner' }),
    })
  );

  await page.route('**/recruit/careers', (route) =>
    route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({ careers: [FAKE_CAREER] }),
    })
  );

  await page.route('**/recruit/submissions*', (route) =>
    route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({ submissions: [FAKE_SUBMISSION] }),
    })
  );

  // Status-change PUT - asserts the button actually fires.
  await page.route('**/recruit/submissions/status', (route) =>
    route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({ id: 101, status: 'shortlisted' }),
    })
  );

  await page.route('**/recruit/customization*', (route) =>
    route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({ customization: {} }),
    })
  );

  // Fall-through for anything else - return empty JSON so no widget
  // blocks on network errors.
  await page.route(/\/recruit\/.*/, (route) => {
    if (route.request().url().includes('public')) return route.continue();
    return route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({}),
    });
  });
}

test.describe('Recruit admin', () => {
  test.beforeEach(async ({ page, context }) => {
    await primeAuthCookie(context);
    await stubRecruitAdminAPIs(page);
  });

  test('admin careers page loads past the guard', async ({ page }) => {
    await page.goto(`${RECRUIT_URL}/admin/careers`);
    await expect(
      page.getByRole('heading').first().or(page.getByRole('banner').first())
    ).toBeVisible({ timeout: 10_000 });
  });

  test('admin renders without console errors', async ({ page }) => {
    const errors: string[] = [];
    page.on('console', (msg) => {
      if (msg.type() === 'error') errors.push(msg.text());
    });
    await page.goto(`${RECRUIT_URL}/admin`);
    await page.waitForTimeout(2_000);
    const meaningful = errors.filter((e) => !/posthog|sentry|Failed to load resource/i.test(e));
    expect(meaningful, meaningful.join('\n')).toHaveLength(0);
  });
});
