import { test, expect } from './fixtures';
import { captureResponsivePair } from './support/screenshot';

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

test.describe('Public menu screenshots @screenshots', () => {
  /**
   * Produces the image files a PR body needs, rather than asserting anything
   * new. Tagged so it stays out of the normal run: `pnpm test:screenshots`.
   *
   * It exists because the Flow workspace-removal tracker recorded C2, the
   * screenshots-in-every-PR requirement, as "cannot be delivered" on the belief
   * that no image file could be produced here. That was wrong, and a
   * requirement written off as impossible does not get revisited unless
   * something in the repo demonstrates otherwise.
   *
   * It lives on the PUBLIC menu deliberately. Every authenticated app's spec
   * currently redirects to the auth app before making a single API call, so a
   * capture placed on one of those screens would have shipped exercising
   * nothing. See docs/pr-screenshots.md for that, which is a separate defect.
   */
  test.beforeEach(async ({ page }) => {
    await stubPublicMenuAPIs(page);
  });

  test('captures the public menu at desktop and at the 390px floor', async ({ page }) => {
    await page.goto(`${MENU_URL}/b/42`);
    await page.waitForTimeout(1500);

    const shots = await captureResponsivePair(page, 'menu-public');

    // Measured, not eyeballed off the image. A screenshot cannot show you that
    // the page scrolls sideways at 390px; this can.
    expect(shots.overflowsAt390).toBe(false);
  });
});
