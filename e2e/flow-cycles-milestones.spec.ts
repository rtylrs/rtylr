import { test, expect, type Route } from './fixtures';

/**
 * Flow Cycles & Milestones E2E Specs
 */

const FLOW_URL = process.env.FLOW_URL ?? 'http://localhost:5183';
const API_HOST = 'http://localhost:9999';

const FAKE_USER = {
  id: '1001',
  email: 'owner@test.local',
  fullName: 'Test Owner',
  role: 'owner',
  apps: ['flow'],
};

const FAKE_CYCLES = [
  {
    id: '1',
    name: 'Sprint 1 - Foundations',
    description: 'Initial sprint cycle',
    startDate: '2026-08-01',
    endDate: '2026-08-15',
    status: 'active',
    workspaceId: '10',
  },
];

const FAKE_MILESTONES = [
  {
    id: '1',
    name: 'v1.0 Beta Release',
    description: 'Public beta milestone',
    targetDate: '2026-08-30',
    dueDate: '2026-08-30',
    status: 'active',
    workspaceId: '10',
  },
];

async function primeAuthCookie(context: import('@playwright/test').BrowserContext): Promise<void> {
  const fakeJwt = 'eyJhbGciOiJub25lIn0.eyJ1c2VyX2lkIjoiMTAwMSJ9.fake_sig';
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

test.describe('Flow - Cycles & Milestones', () => {
  test.beforeEach(async ({ context, page }) => {
    await primeAuthCookie(context);

    await page.route(`${API_HOST}/hub/user/v1/users/me`, (route: Route) =>
      route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify(FAKE_USER),
      })
    );

    await page.route(/localhost:9999\/flow\/cycle\/v1\/cycles/, (route: Route) =>
      route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({ cycles: FAKE_CYCLES }),
      })
    );

    await page.route(/localhost:9999\/flow\/milestone\/v1\/milestones/, (route: Route) =>
      route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({ milestones: FAKE_MILESTONES }),
      })
    );

    await page.route(/localhost:9999\/flow\/workspace\/v1\/workspaces/, (route: Route) =>
      route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({ workspaces: [{ id: '10', name: 'Personal Workspace' }] }),
      })
    );
  });

  test('cycles view renders cycle list', async ({ page }) => {
    await page.goto(`${FLOW_URL}/cycles`);
    await page.waitForLoadState('networkidle');

    await expect(page.getByText('Sprint 1 - Foundations').first()).toBeVisible();
  });

  test('milestones view renders milestone list', async ({ page }) => {
    await page.goto(`${FLOW_URL}/milestones`);
    await page.waitForLoadState('networkidle');

    await expect(page.getByText('v1.0 Beta Release').first()).toBeVisible();
  });
});
