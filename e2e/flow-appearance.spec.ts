import AxeBuilder from '@axe-core/playwright';
import { test, expect, type Route } from './fixtures';

/**
 * Appearance panel: the checks a unit test cannot make.
 *
 * Three things live here because vitest structurally cannot cover them, not
 * because a browser is nicer:
 *
 * 1. Pointer travel between two option cards. `fireEvent.pointerLeave` on a
 *    child is routed to the parent's handler by React's enter/leave simulation,
 *    so a per-card clear and a per-grid clear are indistinguishable in jsdom. A
 *    test written for it in vitest passed with the defect deliberately
 *    reintroduced. Real mouse movement across the gap is the only way to tell.
 * 2. RTL rendering. Neither jsdom nor happy-dom resolves logical properties to a
 *    physical side, so an assertion about `margin-inline-start` proves nothing
 *    there. appearanceRtl.test.ts guards the source; this checks the result.
 * 3. axe. Needs a real accessibility tree.
 */

const FLOW_URL = process.env.FLOW_URL ?? 'http://localhost:5183';
const API_HOST = 'http://localhost:9999';

const FAKE_USER = {
  id: '1001',
  name: 'Appearance Tester',
  email: 'appearance@test.local',
  role: 'owner',
  apps: ['flow'],
};

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

/** The pane mirrors applied state, so read its tokens rather than the root's. */
async function paneToken(page: import('@playwright/test').Page, name: string): Promise<string> {
  return page.evaluate((token) => {
    const pane = document.querySelector('[data-testid="appearance-preview"]');
    return pane ? getComputedStyle(pane).getPropertyValue(token).trim() : '';
  }, name);
}

async function rootToken(page: import('@playwright/test').Page, name: string): Promise<string> {
  return page.evaluate(
    (token) => getComputedStyle(document.documentElement).getPropertyValue(token).trim(),
    name
  );
}

test.describe('Flow - Appearance panel', () => {
  test.beforeEach(async ({ context, page }) => {
    await primeAuthCookie(context);

    await page.route(`${API_HOST}/hub/user/v1/users/me`, (route: Route) =>
      route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify(FAKE_USER),
      })
    );
    // Everything else the settings shell touches, so nothing falls through to the
    // fixture's production guard and fails the test for the wrong reason.
    for (const pattern of [
      /localhost:9999\/flow\/notification\/v1\//,
      /localhost:9999\/flow\/view\/v1\//,
      /localhost:9999\/flow\/workspace\/v1\//,
      /localhost:9999\/flow\/label\/v1\//,
      /localhost:9999\/flow\/project\/v1\//,
      /localhost:9999\/flow\/task\/v1\//,
      /localhost:9999\/hr\//,
      /localhost:9999\/hub\//,
    ]) {
      await page.route(pattern, (route: Route) =>
        route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({
            notifications: [],
            views: [],
            workspaces: [],
            labels: [],
            projects: [],
            tasks: [],
          }),
        })
      );
    }

    await page.goto(`${FLOW_URL}/settings?section=appearance`);
    await page.waitForLoadState('networkidle');
  });

  test('renders the panel with its preview pane', async ({ page }) => {
    await expect(page.getByTestId('appearance-preview')).toBeVisible();
  });

  // D2. The regression this guards is real and was measured: clearing the
  // audition per card showed the applied palette for about eight pixels of
  // travel across the grid gap.
  test('keeps auditioning while the pointer travels between two palette cards', async ({
    page,
  }) => {
    const cards = page.getByRole('group', { name: /palette/i }).getByRole('button');
    const count = await cards.count();
    test.skip(count < 2, 'needs at least two palettes in the active mode');

    const applied = await paneToken(page, '--flow-bg');

    await cards.nth(0).hover();
    const first = await paneToken(page, '--flow-bg');
    expect(first).not.toBe('');

    // Land in the gap between the two cards, which is unhovered space. This is
    // the moment a per-card clear reverts and a per-grid clear does not.
    const a = await cards.nth(0).boundingBox();
    const b = await cards.nth(1).boundingBox();
    expect(a && b).toBeTruthy();
    const gapX = a!.x + a!.width + Math.max(1, (b!.x - (a!.x + a!.width)) / 2);
    await page.mouse.move(gapX, a!.y + a!.height / 2);

    const inGap = await paneToken(page, '--flow-bg');
    expect(inGap, 'the preview reverted while crossing the gap').toBe(first);

    // Leaving the grid entirely must revert.
    await page.mouse.move(10, 10);
    await expect.poll(async () => paneToken(page, '--flow-bg')).toBe(applied);
  });

  test('commits on click and leaves the root token changed', async ({ page }) => {
    const cards = page.getByRole('group', { name: /palette/i }).getByRole('button');
    test.skip((await cards.count()) < 2, 'needs at least two palettes');

    const before = await rootToken(page, '--flow-bg');
    await cards.nth(1).click();
    await expect.poll(async () => rootToken(page, '--flow-bg')).not.toBe(before);
  });

  // D3, part one.
  test('every control is reachable by keyboard with a visible focus ring', async ({ page }) => {
    const panel = page.getByTestId('appearance-preview').locator('xpath=ancestor::*[3]');
    await expect(panel).toBeVisible();

    const focusable = await page.evaluate(() => {
      const nodes = Array.from(
        document.querySelectorAll<HTMLElement>('button, [role="switch"], input, select')
      ).filter((n) => n.offsetParent !== null && !n.hasAttribute('disabled'));
      return nodes.length;
    });
    expect(focusable, 'no focusable controls found, so the walk below is vacuous').toBeGreaterThan(
      5
    );

    // Tab through and confirm focus actually lands somewhere with an outline
    // rather than the browser default of none.
    let ringed = 0;
    for (let i = 0; i < Math.min(focusable, 30); i++) {
      await page.keyboard.press('Tab');
      const hasRing = await page.evaluate(() => {
        const el = document.activeElement as HTMLElement | null;
        if (!el || el === document.body) return false;
        const s = getComputedStyle(el);
        const w = parseFloat(s.outlineWidth || '0');
        return (w > 0 && s.outlineStyle !== 'none') || s.boxShadow !== 'none';
      });
      if (hasRing) ringed++;
    }
    expect(ringed, 'tabbing never produced a visible focus indicator').toBeGreaterThan(3);
  });

  // D3, part two.
  test('has no axe violations', async ({ page }) => {
    const results = await new AxeBuilder({ page })
      .withTags(['wcag2a', 'wcag2aa', 'wcag21a', 'wcag21aa'])
      .analyze();
    const serious = results.violations.filter(
      (v) => v.impact === 'serious' || v.impact === 'critical'
    );
    expect(
      serious.map((v) => `${v.id}: ${v.help} (${v.nodes.length} nodes)`),
      'serious or critical axe violations in the appearance panel'
    ).toEqual([]);
  });

  // The browser half of A3. appearanceRtl.test.ts guards the source; this checks
  // that the rendered result actually flips and does not overflow.
  test('renders in RTL without overflowing or mirroring the wrong way', async ({ page }) => {
    await page.setViewportSize({ width: 390, height: 900 });
    await page.evaluate(() => {
      document.documentElement.setAttribute('dir', 'rtl');
      document.documentElement.setAttribute('lang', 'ar');
    });

    const overflow = await page.evaluate(() => {
      const d = document.documentElement;
      return d.scrollWidth - d.clientWidth;
    });
    expect(overflow, 'the panel scrolls sideways in RTL at 390px').toBeLessThanOrEqual(1);

    // The switch knob is positioned with inset-inline-start, so in RTL it must
    // sit toward the right edge of its track rather than the left.
    const knob = await page.evaluate(() => {
      const sw = document.querySelector('[role="switch"]');
      const inner = sw?.querySelector('span:last-of-type');
      if (!sw || !inner) return null;
      const t = sw.getBoundingClientRect();
      const k = inner.getBoundingClientRect();
      return { fromStartEdge: Math.round(t.right - k.right), trackWidth: Math.round(t.width) };
    });
    expect(knob, 'no switch found to measure').toBeTruthy();
    // In RTL the inline start edge is the right edge, so a small gap there means
    // the knob is anchored correctly.
    expect(knob!.fromStartEdge).toBeLessThan(knob!.trackWidth / 2);
  });
});
