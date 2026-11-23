import { test, expect, type Page } from './fixtures';
import { stubPortalSession } from './portal-session';

const RENNILI_URL = process.env.RENNILI_URL ?? 'http://localhost:5198';

/**
 * Rennili at the sizes it is used at: a customer's phone, and a tablet at the
 * counter. Layout is only real in a browser, so these measure the rendered page.
 */

const LANDSCAPE_TABLET = { width: 1024, height: 768 };

async function stubBoard(page: Page): Promise<void> {
  const now = new Date().toISOString();
  await page.route('**/hub/config/v1/config', (route) =>
    route.fulfill({
      json: { business: { id: '1', name: 'Northwind', slug: 'northwind' }, maintenance: false },
    })
  );
  await page.route('**/pager/v1/tickets', (route) =>
    route.fulfill({
      json: {
        tickets: [
          { id: '1', code: '427', status: 'waiting', created_at: now },
          { id: '2', label: '1234567890AB', label_kind: 'number', status: 'waiting' },
          { id: '3', label: 'Sara', label_kind: 'name', status: 'ready', ready_at: now },
        ],
        server_time: now,
      },
    })
  );
  await page.route('**/pager/v1/settings', (route) =>
    route.fulfill({ json: { enabled: true, label_mode: 'either' } })
  );
}

const box = async (page: Page, name: string) => {
  const found = await page.getByRole('heading', { name }).locator('xpath=..').boundingBox();
  if (!found) throw new Error(`no column ${name}`);
  return found;
};

test.describe('Rennili on phones and tablets', () => {
  /** A scanned phone whose ticket shows the match number 427. */
  async function stubScannedPhone(page: Page, venueName: string): Promise<void> {
    await page.route('**/pager/v1/public/push-key', (route) =>
      route.fulfill({ json: { public_key: '' } })
    );
    await page.route('**/pager/v1/public/venues/acme/scan', (route) =>
      route.fulfill({ json: { public_token: 'tok-rr', code: '427', status: 'waiting' } })
    );
    await page.route('**/pager/v1/public/tickets/tok-rr', (route) =>
      route.fulfill({
        json: {
          code: '427',
          status: 'waiting',
          venue_name: venueName,
          server_time: new Date().toISOString(),
        },
      })
    );
  }

  async function chooseLanguage(page: Page, language: 'en' | 'ar'): Promise<void> {
    if (language === 'ar') await page.getByRole('button', { name: 'العربية' }).click();
    await expect(page.locator('.pager-code')).toHaveText('427');
  }

  for (const language of ['en', 'ar'] as const) {
    test(`the customer page fits a small phone with the one tap in view (${language})`, async ({
      page,
    }) => {
      await page.setViewportSize({ width: 360, height: 740 });
      await stubScannedPhone(page, 'Al Mokhtar Bakery & Roastery Hamra Street');

      await page.goto(`${RENNILI_URL}/acme`);
      await chooseLanguage(page, language);

      const notify = page.getByRole('button', {
        name: language === 'ar' ? 'نبّهني عندما يجهز' : "Notify me when it's ready",
      });
      await expect(notify).toBeInViewport();
      const overflow = await page.evaluate(() => {
        const root = document.querySelector('.pager-root');
        return root ? root.scrollWidth - root.clientWidth : -1;
      });
      expect(overflow).toBe(0);
    });

    test(`the code reads left to right on one line, centred in its card, on a tablet (${language})`, async ({
      page,
    }) => {
      await page.setViewportSize({ width: 820, height: 1180 });
      await stubScannedPhone(page, 'Acme Cafe');

      await page.goto(`${RENNILI_URL}/acme`);
      await chooseLanguage(page, language);

      const code = page.locator('.pager-code');
      await expect(code).toHaveAttribute('dir', 'ltr');
      const layout = await code.evaluate((element) => {
        const range = document.createRange();
        range.selectNodeContents(element);
        const lines = new Set(Array.from(range.getClientRects()).map((r) => Math.round(r.top)));
        const card = element.closest('.pager-code-card')?.getBoundingClientRect();
        const text = range.getBoundingClientRect();
        return {
          lines: lines.size,
          offCentre: card
            ? Math.abs(text.left + text.width / 2 - (card.left + card.width / 2))
            : 99,
          inside: card ? text.left >= card.left && text.right <= card.right : false,
        };
      });
      expect(layout.lines).toBe(1);
      expect(layout.inside).toBe(true);
      expect(layout.offCentre).toBeLessThan(12);
    });
  }

  // A name is as wide as its letters, not its count: "Ahmad" once split as "Ahma / d".
  for (const [label, kind] of [
    ['Ahmad', 'name'],
    ['WWMMW', 'name'],
    ['Mohammad', 'name'],
    ['1234567890AB', 'number'],
  ] as const) {
    test(`the order label "${label}" stays on one line on a small phone`, async ({ page }) => {
      await page.setViewportSize({ width: 360, height: 740 });
      await page.route('**/pager/v1/public/push-key', (route) =>
        route.fulfill({ json: { public_key: '' } })
      );
      await page.route('**/pager/v1/public/tickets/tok-label', (route) =>
        route.fulfill({
          json: {
            label,
            label_kind: kind,
            status: 'waiting',
            venue_name: 'Acme Cafe',
            server_time: new Date().toISOString(),
          },
        })
      );

      await page.goto(`${RENNILI_URL}/t/tok-label`);
      const shown = page.locator('.pager-order-number');
      await expect(shown).toHaveText(label);
      const layout = await shown.evaluate((element) => {
        const range = document.createRange();
        range.selectNodeContents(element);
        const lines = new Set(Array.from(range.getClientRects()).map((r) => Math.round(r.top)));
        const text = range.getBoundingClientRect();
        return { lines: lines.size, inside: text.left >= 0 && text.right <= window.innerWidth };
      });
      expect(layout.lines).toBe(1);
      expect(layout.inside).toBe(true);
    });
  }

  test('the board puts its three columns side by side on a landscape tablet', async ({
    page,
    context,
  }) => {
    await page.setViewportSize(LANDSCAPE_TABLET);
    await stubPortalSession(page, context);
    await stubBoard(page);

    await page.goto(`${RENNILI_URL}/board`);

    const fresh = await box(page, 'New phones (1)');
    const waiting = await box(page, 'Waiting (1)');
    const ready = await box(page, 'Ready (1)');
    expect(Math.abs(fresh.y - waiting.y)).toBeLessThan(2);
    expect(Math.abs(fresh.y - ready.y)).toBeLessThan(2);

    const label = page.getByRole('textbox', { name: /for code/ });
    expect((await label.boundingBox())?.width).toBeGreaterThan(120);
    for (const button of [
      page.getByRole('button', { name: 'Save', exact: true }),
      page.getByRole('button', { name: 'Ready: 1234567890AB' }),
      page.getByRole('button', { name: 'Collected: Sara' }),
    ]) {
      expect((await button.boundingBox())?.height).toBeGreaterThanOrEqual(44);
    }
  });

  test('the demo board beside the phone keeps a usable label field', async ({ page }) => {
    await page.setViewportSize(LANDSCAPE_TABLET);

    await page.goto(`${RENNILI_URL}/try`);

    const label = page.getByRole('textbox', { name: /for code/ });
    await expect(label).toBeVisible();
    expect((await label.boundingBox())?.width).toBeGreaterThan(120);
  });

  test('the counter display fills a landscape tablet with its QR', async ({ page, context }) => {
    await page.setViewportSize(LANDSCAPE_TABLET);
    await stubPortalSession(page, context);
    await stubBoard(page);

    await page.goto(`${RENNILI_URL}/display`);

    const qr = await page.locator('.pager-display canvas').boundingBox();
    expect(qr?.width).toBeGreaterThan(LANDSCAPE_TABLET.height * 0.55);
  });
});
