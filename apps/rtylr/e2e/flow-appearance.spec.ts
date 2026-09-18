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

/**
 * Wait for CSS transitions to finish before measuring colour.
 *
 * The option cards carry `transition-colors`, so for a moment after a mode
 * switch their background is an interpolated value that belongs to no palette.
 * axe measured one of those and reported 4.17:1 against `#f3f3f3`, a colour no
 * palette defines. `deriveAccentFg` guarantees its result against
 * `palette.card`, so any measurement against a value that is not a palette
 * colour is measuring the transition, not the design.
 *
 * This passed before a merge purely on timing luck, which is the worst kind of
 * green.
 */
async function settleTransitions(page: import('@playwright/test').Page): Promise<void> {
  await page.evaluate(async () => {
    const running = document
      .getAnimations()
      .filter((a) => a.playState === 'running')
      .map((a) => a.finished.catch(() => undefined));
    // Capped: a looping animation would never settle and must not hang the test.
    await Promise.race([Promise.all(running), new Promise((resolve) => setTimeout(resolve, 1500))]);
  });
  // One frame for the committed values to be observable.
  await page.evaluate(() => new Promise((r) => requestAnimationFrame(() => r(undefined))));
}

test.describe('Flow - Appearance panel', () => {
  test.beforeEach(async ({ context, page }) => {
    await primeAuthCookie(context);

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

    // Registered AFTER the catch-alls on purpose. Playwright gives precedence to
    // the route registered LAST, so with this first the /hub/ catch-all below
    // shadowed it, AppGuard received an empty envelope instead of a user, and
    // the whole panel never rendered. Every test in this file failed for that one
    // reason, which is what an unexecuted spec hides.
    await page.route(`${API_HOST}/hub/user/v1/users/me`, (route: Route) =>
      route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify(FAKE_USER),
      })
    );

    await page.goto(`${FLOW_URL}/settings?section=appearance`);
    await page.waitForLoadState('networkidle');
  });

  test('renders the panel with its preview pane', async ({ page }) => {
    await expect(page.getByTestId('appearance-preview')).toBeVisible();
  });

  // D2. The regression this guards is real and was measured: clearing the
  // audition per card showed the applied palette for about eight pixels of
  // travel across the grid gap.
  //
  // The first version of this test was VACUOUS. It hovered card 0, which is the
  // card for the palette already applied, so the audition changed nothing and
  // "unchanged while crossing the gap" was trivially true. It passed with the
  // per-card clear deliberately reintroduced, which is the precise defect it
  // exists to catch.
  //
  // Two things fix that and both are load-bearing: audition a card whose palette
  // DIFFERS from the applied one, and assert the audition actually moved the
  // token before asserting the gap preserves it.
  test('keeps auditioning while the pointer travels between two palette cards', async ({
    page,
  }) => {
    const cards = page.getByRole('group', { name: /palette/i }).getByRole('button');
    const count = await cards.count();
    expect(count, 'the active mode must expose at least two palettes').toBeGreaterThanOrEqual(2);

    const applied = await paneToken(page, '--flow-bg');
    expect(applied).not.toBe('');

    // Find a card that genuinely changes the preview. Index is not assumed:
    // which card matches the applied palette depends on the active mode.
    let auditioned = -1;
    let hoveredValue = '';
    for (let i = 0; i < count; i++) {
      await cards.nth(i).hover();
      const v = await paneToken(page, '--flow-bg');
      if (v !== applied && v !== '') {
        auditioned = i;
        hoveredValue = v;
        break;
      }
    }
    expect(
      auditioned,
      'no palette card changed the preview, so this test cannot observe an audition'
    ).toBeGreaterThanOrEqual(0);

    // A neighbour to travel towards, preferring the one to the right when the
    // responsive grid is horizontal and the one below when it is stacked.
    const neighbour = auditioned + 1 < count ? auditioned + 1 : auditioned - 1;
    // Keep both cards and their gap in the viewport. At 390px the palette is
    // stacked, and the first two cards can otherwise straddle the bottom edge.
    await cards.nth(neighbour).scrollIntoViewIfNeeded();
    const a = await cards.nth(auditioned).boundingBox();
    const b = await cards.nth(neighbour).boundingBox();
    expect(a && b).toBeTruthy();

    await cards.nth(auditioned).hover();
    expect(await paneToken(page, '--flow-bg'), 'audition did not take effect').toBe(hoveredValue);

    // Land in unhovered space between the two cards. This is the moment a
    // per-card clear reverts and a per-grid clear does not. The cards are a
    // single column at 390px, so the travel axis must follow their geometry.
    const horizontal = Math.abs(a!.x - b!.x) >= Math.abs(a!.y - b!.y);
    const first = horizontal ? (a!.x <= b!.x ? a! : b!) : a!.y <= b!.y ? a! : b!;
    const second = first === a! ? b! : a!;
    const gapStart = horizontal ? first.x + first.width : first.y + first.height;
    const gapEnd = horizontal ? second.x : second.y;
    const gapMiddle = gapStart + Math.max(1, (gapEnd - gapStart) / 2);
    await page.mouse.move(
      horizontal ? gapMiddle : first.x + first.width / 2,
      horizontal ? first.y + first.height / 2 : gapMiddle
    );

    expect(
      await paneToken(page, '--flow-bg'),
      'the preview reverted while crossing the gap between two cards'
    ).toBe(hoveredValue);

    // Leaving the grid entirely must revert.
    await page.mouse.move(5, 5);
    await expect.poll(async () => paneToken(page, '--flow-bg')).toBe(applied);
  });

  test('commits on click and leaves the root token changed', async ({ page }) => {
    const cards = page.getByRole('group', { name: /palette/i }).getByRole('button');
    expect(
      await cards.count(),
      'the appearance panel must expose at least two palettes'
    ).toBeGreaterThanOrEqual(2);

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
  //
  // Runs in BOTH modes, and that is the whole point. Checked only in the default
  // dark mode this test passed with the rtylr#970 fix reverted, because in dark
  // --accent-on-muted resolves to the same value as --accent-fg and there was
  // never a violation there. The defect existed only in light. A mode-specific
  // bug needs a mode-parameterised test, or the assertion is decorative.
  for (const mode of ['Light', 'Dark'] as const) {
    test(`has no axe violations in ${mode.toLowerCase()} mode`, async ({ page }) => {
      const modeCard = page
        .getByRole('button')
        .filter({ hasText: new RegExp(`^\\s*${mode}`) })
        .first();
      await expect(modeCard).toBeVisible();
      await modeCard.click();
      await expect
        .poll(async () =>
          page.evaluate(() =>
            getComputedStyle(document.documentElement).getPropertyValue('--accent-fg').trim()
          )
        )
        .not.toBe('');
      await settleTransitions(page);

      // Scoped to main, which is the panel. The app shell around it carries two
      // pre-existing serious violations that this PR does not touch and must not
      // silently absorb, so they are named here rather than excluded quietly:
      //
      //   .hover:opacity-90 > .text-start.flex-1   the sidebar search label,
      //                                            #585858 on #1d1d1d, 2.36:1
      //   .font-mono                               the Cmd-K badge,
      //                                            #656564 on #2f2f2f, 2.29:1 at 9px
      //
      // Both are --flow-text-dim, which cannot reach 4.5:1 in either mode and is
      // used as a text colour in 127 places in Flow, and the 9px one also
      // breaches the 10px type floor globals.css documents. Tracked on
      // rtylr#970. They are identical in light and dark, which is the tell that
      // they belong to the shell rather than to the appearance tokens.
      const results = await new AxeBuilder({ page })
        .include('main')
        .withTags(['wcag2a', 'wcag2aa', 'wcag21a', 'wcag21aa'])
        .analyze();
      const serious = results.violations.filter(
        (v) => v.impact === 'serious' || v.impact === 'critical'
      );
      expect(
        serious.flatMap((v) =>
          v.nodes.map((n) => `${v.id} @ ${n.target.join(' ')}: ${n.any[0]?.message ?? ''}`)
        ),
        `serious or critical axe violations in the appearance panel (${mode} mode)`
      ).toEqual([]);
    });
  }

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
