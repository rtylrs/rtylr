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
          {
            id: '1',
            word: 'chefskiss',
            word_number: 2,
            color: 3,
            status: 'waiting',
            created_at: now,
          },
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
  test('the customer page fits a small phone with the one tap in view', async ({ page }) => {
    await page.setViewportSize({ width: 360, height: 740 });
    await page.route('**/pager/v1/public/push-key', (route) =>
      route.fulfill({ json: { public_key: '' } })
    );
    await page.route('**/pager/v1/public/venues/acme/scan', (route) =>
      route.fulfill({
        json: {
          public_token: 'tok-rr',
          word: 'chefskiss',
          word_number: 2,
          color: 3,
          status: 'waiting',
        },
      })
    );
    await page.route('**/pager/v1/public/tickets/tok-rr', (route) =>
      route.fulfill({
        json: {
          word: 'chefskiss',
          word_number: 2,
          color: 3,
          status: 'waiting',
          venue_name: 'Al Mokhtar Bakery & Roastery Hamra Street',
          server_time: new Date().toISOString(),
        },
      })
    );

    await page.goto(`${RENNILI_URL}/acme`);

    const notify = page.getByRole('button', { name: "Notify me when it's ready" });
    await expect(notify).toBeInViewport();
    const overflow = await page.evaluate(() => {
      const root = document.querySelector('.pager-root');
      return root ? root.scrollWidth - root.clientWidth : -1;
    });
    expect(overflow).toBe(0);
  });

  test('a ten-letter code word stays on one line on a tablet', async ({ page }) => {
    await page.setViewportSize({ width: 820, height: 1180 });
    await page.route('**/pager/v1/public/push-key', (route) =>
      route.fulfill({ json: { public_key: '' } })
    );
    await page.route('**/pager/v1/public/venues/acme/scan', (route) =>
      route.fulfill({
        json: { public_token: 'tok-rr', word: 'strawberry', color: 3, status: 'waiting' },
      })
    );
    await page.route('**/pager/v1/public/tickets/tok-rr', (route) =>
      route.fulfill({
        json: {
          word: 'strawberry',
          color: 3,
          status: 'waiting',
          venue_name: 'Acme Cafe',
          server_time: new Date().toISOString(),
        },
      })
    );

    await page.goto(`${RENNILI_URL}/acme`);

    await expect(page.locator('.pager-code--word')).toContainText('Strawberry');
    const lines = await page.locator('.pager-code--word').evaluate((badge) => {
      const word = Array.from(badge.childNodes).find(
        (node) => node.nodeType === Node.TEXT_NODE && node.textContent?.includes('Strawberry')
      );
      if (!word) return 0;
      const range = document.createRange();
      range.selectNodeContents(word);
      return new Set(Array.from(range.getClientRects()).map((rect) => Math.round(rect.top))).size;
    });
    expect(lines).toBe(1);
  });

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
      page.getByRole('button', { name: 'Save' }),
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
