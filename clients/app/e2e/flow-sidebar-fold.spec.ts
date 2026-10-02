import { flowAuthFields } from './support/flow-auth';
import { test, expect } from './fixtures';

/**
 * #1076 was filed from a measurement, so it is closed with one.
 *
 * At 1440x900 with three projects, the sidebar's scroll container measured
 * scrollHeight 1067 against clientHeight 614: 453px below the fold, with no
 * scrollbar and no fade to say the region scrolled. P-Alpha sat at y=1165 in a
 * 900px viewport, and "New project" at y=1031.
 *
 * Two defaults caused it: Projects was ordered last, and every group was
 * expanded, so Plan's five items and Time & people's six pushed the project
 * list off the bottom.
 */

const FLOW_URL = process.env.FLOW_URL ?? 'http://localhost:5197/flow';

const PROJECT_NAMES = ['P-Alpha', 'P-Beta', 'P-Gamma'];

/** Snake case, matching the wire shape the other Flow specs stub. */
const PROJECTS = PROJECT_NAMES.map((name, i) => ({
  id: String(i + 1),
  name,
  description: '',
  icon: 'folder',
  color: '#64C5B9',
  status: 'active',
  visibility: 'private',
  task_count: 0,
  completed_count: 0,
  member_count: 0,
  start_date: '',
  due_date: '',
}));

const PROJECTS_BODY = JSON.stringify({
  projects: PROJECTS,
  pagination: { total: PROJECTS.length },
});

test.describe('Flow sidebar puts projects above the fold @screenshots', () => {
  test.use({ viewport: { width: 1440, height: 900 } });

  test.beforeEach(async ({ page, context }) => {
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

    // Broadest first: Playwright matches in reverse registration order, so the
    // specific handlers below win.
    await page.route(/localhost:9999\//, (route) =>
      route.fulfill({ status: 200, contentType: 'application/json', body: '{}' })
    );

    for (const pattern of [
      /localhost:9999\/flow\/projects(\?.*)?$/,
      /localhost:9999\/flow\/project\/v2\/projects(\?.*)?$/,
    ]) {
      await page.route(pattern, (route) =>
        route.fulfill({ status: 200, contentType: 'application/json', body: PROJECTS_BODY })
      );
    }

    // Seed the server-resolved Flow grant used by the Portal shell.
    await page.route(/localhost:9999\/hub\/user\/v1\/users\/me/, (route) =>
      route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          id: '1001',
          email: 'owner@acme.test',
          full_name: 'Acme Owner',
          business_id: '1',
          ...flowAuthFields('owner'),
          plan_slug: 'pro',
        }),
      })
    );
  });

  test('every stubbed project row sits within the viewport', async ({ page }) => {
    await page.goto(`${FLOW_URL}/my-issues`);
    await page.waitForLoadState('networkidle');
    await expect(page.getByText("You don't have access to this app.")).toHaveCount(0);

    const rows = await Promise.all(
      PROJECT_NAMES.map(async (name) => {
        const row = page.getByRole('link', { name: new RegExp(name) }).first();
        const visible = await row.isVisible().catch(() => false);
        return { name, box: visible ? await row.boundingBox() : null };
      })
    );

    const found = rows.filter((r) => r.box);
    // If the stub shape drifts and no project renders, the assertion below
    // would pass by measuring nothing.
    expect(found.length, 'no project rows rendered, so nothing was measured').toBeGreaterThan(0);

    for (const { name, box } of found) {
      expect(box!.y, `${name} is below the 900px fold at y=${Math.round(box!.y)}`).toBeLessThan(
        900
      );
    }
  });

  test('the desktop create control is visible and hittable', async ({ page }) => {
    await page.goto(`${FLOW_URL}/my-issues`);
    await page.waitForLoadState('networkidle');
    await expect(page.getByText("You don't have access to this app.")).toHaveCount(0);

    // The header + is inside an lg:hidden container, so at this width it
    // measured 0x0 and the desktop shell had no create control at all.
    const create = page.getByRole('button', { name: /new task/i });
    const boxes = await Promise.all(
      (await create.all()).map(async (b) => ({ box: await b.boundingBox() }))
    );

    const hittable = boxes.filter((b) => b.box && b.box.width > 0 && b.box.height > 0);
    expect(hittable.length, 'no new-task control has a non-zero box at 1440px').toBeGreaterThan(0);
    expect(hittable[0].box!.y).toBeLessThan(900);
  });
  test('Flow fills the Portal workspace on a wide display', async ({ page }) => {
    await page.setViewportSize({ width: 2560, height: 1440 });
    await page.goto(`${FLOW_URL}/my-issues`);
    await expect(page.getByRole('heading', { name: 'My tasks', exact: true })).toBeVisible();
    const workspace = await page.locator('.portal-page.is-flow').boundingBox();
    const flow = await page.locator('.portal-flow').boundingBox();
    expect(workspace).not.toBeNull();
    expect(flow).not.toBeNull();
    expect(flow!.width).toBeGreaterThan(2000);
    expect(Math.abs(flow!.x - workspace!.x)).toBeLessThan(1);
    expect(Math.abs(flow!.width - workspace!.width)).toBeLessThan(1);
  });
});
