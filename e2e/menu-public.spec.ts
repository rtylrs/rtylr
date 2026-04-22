import { test, expect } from '@playwright/test';

/**
 * Public menu e2e — /b/:businessId hits the new anonymous `/pos/public/*`
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

test.describe('Public menu (POS-backed)', () => {
  test.beforeEach(async ({ page }) => {
    await stubPublicMenuAPIs(page);
  });

  test('numeric business id renders items', async ({ page }) => {
    await page.goto(`${MENU_URL}/b/42`);

    // Active items appear, inactive is filtered out.
    await expect(page.getByText('Cold Brew')).toBeVisible({ timeout: 10_000 });
    await expect(page.getByText('Croissant')).toBeVisible();
    await expect(page.getByText('Off menu')).toHaveCount(0);
  });

  test('invalid business id path shows the "invalid menu" state', async ({ page }) => {
    // The :businessId route only matches numeric ids. A non-numeric
    // segment like "not-a-number" falls through to PosMenu which
    // refuses to request and renders the error panel.
    await page.goto(`${MENU_URL}/b/not-a-number`);
    await expect(page.getByText(/invalid menu/i)).toBeVisible({ timeout: 5_000 });
  });
});
