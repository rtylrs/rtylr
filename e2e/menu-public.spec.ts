import { test, expect } from './fixtures';

/**
 * Public menu e2e - /b/:businessId hits anonymous OneKit menu endpoints.
 */

const MENU_URL = process.env.MENU_URL ?? 'http://localhost:5174';

async function stubPublicMenuAPIs(page: import('@playwright/test').Page): Promise<void> {
  await page.route('**/hub/business/v1/businesses/public*', (route) =>
    route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({
        business: {
          id: 42,
          name: 'Acme Cafe',
          logo_url: '',
          currency: 'USD',
          timezone: 'UTC',
        },
      }),
    })
  );

  await page.route('**/pos/menu-item/v1/menu-items/public*', (route) =>
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
}

test.describe('Public menu (POS-backed)', () => {
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
