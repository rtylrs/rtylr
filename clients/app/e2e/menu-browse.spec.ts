import { test, expect } from './fixtures';

/**
 * Public menu browsing e2e:
 * Covers the unauthenticated customer browsing half under the Portal origin:
 * - /:slug/menu (SlugLanding: business + location picker)
 * - /:slug/menu/:location (LocationMenu: location menu items)
 * - Error and not-found states
 * - Language persistence and RTL for Arabic
 * - Scrolling behavior within the Portal scroll container (.public-menu-root)
 * - Auth boundaries and gate verification
 */

const PORTAL_URL = process.env.PORTAL_URL ?? 'http://localhost:5197';

const BROWSE_STUB_DATA = {
  business: {
    id: 42,
    name: 'Acme Cafe',
    slug: 'acme',
    logo_url: '',
    currency: 'USD',
    timezone: 'UTC',
  },
  locations: [
    {
      id: 101,
      name: 'Downtown',
      slug: 'downtown',
      address: '123 Main St',
      timezone: 'UTC',
      is_active: true,
    },
    {
      id: 102,
      name: 'Uptown',
      slug: 'uptown',
      address: '456 North Ave',
      timezone: 'UTC',
      is_active: true,
    },
  ],
  items: [
    { id: 1, name: 'Cold Brew', sku: 'CB-01', sale_price: 5.5, is_active: true },
    { id: 2, name: 'Croissant', sku: 'CR-01', sale_price: 4.25, is_active: true },
    { id: 3, name: 'Matcha Latte', sku: 'ML-01', sale_price: 6.0, is_active: true },
    { id: 4, name: 'Avocado Toast', sku: 'AT-01', sale_price: 9.5, is_active: true },
    { id: 5, name: 'Blueberry Muffin', sku: 'BM-01', sale_price: 3.75, is_active: true },
    { id: 6, name: 'Espresso', sku: 'ES-01', sale_price: 3.0, is_active: true },
    { id: 7, name: 'Iced Tea', sku: 'IT-01', sale_price: 4.0, is_active: true },
    { id: 8, name: 'Bagel with Cream Cheese', sku: 'BG-01', sale_price: 4.5, is_active: true },
    { id: 9, name: 'Scone', sku: 'SC-01', sale_price: 3.5, is_active: true },
    { id: 10, name: 'Chai Latte', sku: 'CL-01', sale_price: 5.0, is_active: true },
    { id: 11, name: 'Hot Chocolate', sku: 'HC-01', sale_price: 4.5, is_active: true },
    { id: 12, name: 'Seasonal Tart', sku: 'ST-01', sale_price: 6.5, is_active: false },
  ],
};

async function stubMenuBrowseAPIs(page: import('@playwright/test').Page): Promise<void> {
  await page.route('**/hub/business/v1/businesses/by-slug*', (route) => {
    const url = new URL(route.request().url());
    const slug = url.searchParams.get('slug');
    if (slug === 'acme') {
      return route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          business: BROWSE_STUB_DATA.business,
          locations: BROWSE_STUB_DATA.locations,
        }),
      });
    }
    return route.fulfill({
      status: 404,
      contentType: 'application/json',
      body: JSON.stringify({ error: 'Business not found' }),
    });
  });

  await page.route('**/hub/business/v1/businesses/public*', (route) =>
    route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({
        business: BROWSE_STUB_DATA.business,
      }),
    })
  );

  await page.route('**/pos/menu-item/v1/menu-items/public*', (route) =>
    route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({
        items: BROWSE_STUB_DATA.items,
      }),
    })
  );
}

test.describe('Public Menu Browsing', () => {
  test.beforeEach(async ({ page }) => {
    await stubMenuBrowseAPIs(page);
  });

  test('/acme/menu renders business and locations', async ({ page }) => {
    await page.goto(`${PORTAL_URL}/acme/menu`);

    await expect(page.getByText('Acme Cafe')).toBeVisible();
    await expect(page.getByText('Downtown')).toBeVisible();
    await expect(page.getByText('Uptown')).toBeVisible();
    await expect(page.getByText('123 Main St')).toBeVisible();
  });

  test('clicking a location navigates to /acme/menu/<location> and renders menu sections/items', async ({
    page,
  }) => {
    await page.goto(`${PORTAL_URL}/acme/menu`);

    await expect(page.getByText('Downtown')).toBeVisible();
    await page.getByRole('link', { name: /Downtown/i }).click();

    await page.waitForURL(`**/acme/menu/downtown`);
    expect(page.url()).toContain('/acme/menu/downtown');

    // Menu header displays business and location
    await expect(page.getByRole('heading', { name: /Acme Cafe - Downtown/i })).toBeVisible();

    // Menu category and items render
    await expect(page.getByText(/Cold Brew/i)).toBeVisible();
    await expect(page.getByText(/Croissant/i)).toBeVisible();
    await expect(page.getByText(/Matcha Latte/i)).toBeVisible();
  });

  test('unknown slug shows not-found state without redirecting to login', async ({ page }) => {
    await page.goto(`${PORTAL_URL}/unknown-slug/menu`);

    // Must show business not found empty state
    await expect(page.getByText(/Business not found/i)).toBeVisible();
    await expect(page.getByText(/unknown-slug/i)).toBeVisible();

    // Must not redirect to auth host or login
    expect(page.url()).toContain('/unknown-slug/menu');
    expect(page.url()).not.toContain('5176');
    expect(page.url()).not.toContain('/login');
  });

  test('deep link to /acme/menu/<location> directly loads menu', async ({ page }) => {
    await page.goto(`${PORTAL_URL}/acme/menu/downtown`);

    await expect(page.getByRole('heading', { name: /Acme Cafe - Downtown/i })).toBeVisible();
    await expect(page.getByText(/Cold Brew/i)).toBeVisible();
    await expect(page.getByText(/Croissant/i)).toBeVisible();

    expect(page.url()).toContain('/acme/menu/downtown');
    expect(page.url()).not.toContain('5176');
  });

  test('language switch to Arabic sets dir=rtl and translated strings', async ({
    page,
    context,
  }) => {
    await context.addCookies([{ name: 'language', value: 'ar', url: PORTAL_URL }]);

    await page.goto(`${PORTAL_URL}/acme/menu`);

    // HTML root has RTL direction
    await expect(page.locator('html')).toHaveAttribute('dir', 'rtl');

    // Arabic translated string for pick location is displayed
    await expect(page.getByText('اختر موقعًا لعرض قائمته')).toBeVisible();
    await expect(page.getByText('Acme Cafe')).toBeVisible();
  });

  test('the page scrolls within .public-menu-root container', async ({ page }) => {
    await page.goto(`${PORTAL_URL}/acme/menu/downtown`);

    // Wait until items are rendered
    await expect(page.getByText(/Cold Brew/i)).toBeVisible();
    await expect(page.getByText(/Seasonal Tart/i)).toBeVisible();

    const root = page.locator('.public-menu-root');
    await expect(root).toBeVisible();

    // Verify container has scrollable overflow
    const scrollHeight = await root.evaluate((el) => el.scrollHeight);
    const clientHeight = await root.evaluate((el) => el.clientHeight);
    expect(scrollHeight).toBeGreaterThan(clientHeight);

    // Scroll down the container
    await root.evaluate((el) => {
      el.scrollTop = 250;
    });

    const scrollTop = await root.evaluate((el) => el.scrollTop);
    expect(scrollTop).toBeGreaterThan(0);
  });

  test('unauthenticated visit to public menu URLs never redirects to auth host', async ({
    page,
  }) => {
    const publicUrls = [
      `${PORTAL_URL}/acme/menu`,
      `${PORTAL_URL}/acme/menu/downtown`,
      `${PORTAL_URL}/unknown-slug/menu`,
    ];

    for (const url of publicUrls) {
      await page.goto(url);
      await page.waitForTimeout(500);
      expect(page.url()).not.toContain('5176');
      expect(page.url()).not.toContain('/login');
    }
  });

  test('Portal route like /crm still redirects unauthenticated users to auth host', async ({
    page,
  }) => {
    await page.goto(`${PORTAL_URL}/crm`);

    // BusinessRoute detects no token and redirects to auth host
    await page.waitForURL((url) => url.port === '5176' || url.pathname.includes('/login'), {
      timeout: 10_000,
    });
    expect(page.url()).toContain('5176');
    expect(page.url()).toContain('redirect=dash');
  });

  test('reserved-looking path /menu still goes through auth gate', async ({ page }) => {
    await page.goto(`${PORTAL_URL}/menu`);

    // Admin /menu is guarded by BusinessRoute and must redirect to auth host
    await page.waitForURL((url) => url.port === '5176' || url.pathname.includes('/login'), {
      timeout: 10_000,
    });
    expect(page.url()).toContain('5176');
    expect(page.url()).toContain('redirect=dash');
  });
});
