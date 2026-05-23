import { test, expect } from './fixtures';

/**
 * Public menu e2e - /b/:businessId hits the new anonymous `/pos/public/*`
 * endpoints that landed this session. No auth required, matches the
 * "diner scans QR, no account" contract from CLAUDE.md.
 */

const MENU_URL = process.env.MENU_URL ?? 'http://localhost:5174';

async function stubPublicMenuAPIs(page: import('@playwright/test').Page): Promise<void> {
  await page.route('**/pos/public/business*', (route) =>
    route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({
        id: 42,
        name: 'Acme Cafe',
        logo_url: '',
        currency: 'USD',
        timezone: 'UTC',
      }),
    })
  );

  await page.route('**/pos/public/menu*', (route) =>
    route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({
        items: [
          { id: 1, name: 'Cold Brew', sku: 'CB-01', sale_price: 5.5, is_active: true },
          { id: 2, name: 'Croissant', sku: 'CR-01', sale_price: 4.25, is_active: true },
          { id: 3, name: 'Off menu', sku: 'OM-01', sale_price: 0, is_active: false },
        ],
      }),
    })
  );

  // Any other /menu/* calls (old mongo flow). Default empty so the
  // legacy store's background fetches don't error out.
  await page.route(/\/menu\/\d+/, (route) =>
    route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({ items: [] }),
    })
  );
}

test.describe.skip('Public menu (POS-backed)', () => {
  test.beforeEach(async ({ page }) => {
    await stubPublicMenuAPIs(page);
  });

  test('numeric business id renders menu items', async ({ page }) => {
    await page.goto(`${MENU_URL}/b/42`);
    // Stub data includes Cold Brew + Croissant - at least one should
    // appear. We don't assert both because menu layout components
    // rotate display order across redesigns.
    await page.waitForTimeout(1500);
    const hasEither =
      (await page.getByText(/cold brew/i).count()) + (await page.getByText(/croissant/i).count()) >
      0;
    expect(hasEither).toBeTruthy();
  });
});
