import { test, expect, type Page } from './fixtures';

/**
 * Public menu ordering e2e - the cart, checkout and order-status half of the
 * menu that moved onto the Portal's root route.
 *
 * Like `menu-public.spec.ts`, this targets the Portal origin with no auth
 * cookie and stubs every API call with `page.route()`. The request and response
 * shapes come from the generated `pos/orders/v1` contract: create is a POST to
 * `/pos/orders/v1/orders/public` and status is a GET to
 * `/pos/orders/v1/orders/{order_id}/public-status`.
 */

const PORTAL_URL = process.env.PORTAL_URL ?? 'http://localhost:5197';

/**
 * Pre-set the consent cookie so the Portal's cookie banner never renders. At
 * the 390px viewport the banner overlaps the cart bar and intercepts clicks.
 */
async function suppressConsentBanner(
  context: import('@playwright/test').BrowserContext
): Promise<void> {
  await context.addCookies([
    { name: 'rtylr-consent', value: 'denied', url: PORTAL_URL, sameSite: 'Lax' },
  ]);
}

const BUSINESSES: Record<string, { id: number; name: string }> = {
  acme: { id: 1, name: 'Acme Diner' },
  beta: { id: 2, name: 'Beta Kitchen' },
};

const BUSINESS_NAMES_BY_ID: Record<number, string> = {
  1: 'Acme Diner',
  2: 'Beta Kitchen',
  42: 'Acme Cafe',
};

const ITEMS: Record<number, { id: number; name: string; salePrice: number }[]> = {
  1: [
    { id: 10, name: 'Falafel wrap', salePrice: 5.5 },
    { id: 11, name: 'Mint tea', salePrice: 2 },
  ],
  2: [{ id: 20, name: 'Beta Burger', salePrice: 8 }],
  42: [{ id: 7, name: 'Cold Brew', salePrice: 4.5 }],
};

/** Stubs the anonymous menu reads: slug lookup, business branding and items. */
async function stubMenuData(page: Page): Promise<void> {
  await page.route('**/hub/business/v1/businesses/by-slug*', (route) => {
    const slug = new URL(route.request().url()).searchParams.get('slug') ?? '';
    const business = BUSINESSES[slug] ?? BUSINESSES.acme;
    return route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({
        business: {
          id: business.id,
          name: business.name,
          slug,
          logo_url: '',
          currency: 'USD',
          timezone: 'UTC',
        },
        locations: [
          {
            id: business.id * 10,
            name: 'Downtown',
            slug: 'downtown',
            address: '1 Main St',
            timezone: 'UTC',
            is_active: true,
          },
        ],
      }),
    });
  });

  await page.route('**/hub/business/v1/businesses/public*', (route) => {
    const id = Number(new URL(route.request().url()).searchParams.get('business_id'));
    return route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({
        business: {
          id,
          name: BUSINESS_NAMES_BY_ID[id] ?? 'Acme',
          logo_url: '',
          currency: 'USD',
          timezone: 'UTC',
        },
      }),
    });
  });

  await page.route('**/pos/menu-item/v1/menu-items/public*', (route) => {
    const id = Number(new URL(route.request().url()).searchParams.get('business_id'));
    const items = (ITEMS[id] ?? []).map((item) => ({
      id: item.id,
      owner_id: id,
      category_id: 1,
      name: item.name,
      sku: `SKU-${item.id}`,
      sale_price: item.salePrice,
      is_active: true,
      created_at: '',
      updated_at: '',
    }));
    return route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({ items }),
    });
  });
}

interface OrderStubOptions {
  /** Status returned by the create call. >= 400 simulates a failed submit. */
  submitStatus?: number;
  /** Successive statuses returned by the poll endpoint, last one repeats. */
  statusSequence?: string[];
}

/** Stubs order creation and status polling. Later calls override earlier ones. */
async function stubOrders(page: Page, options: OrderStubOptions = {}): Promise<void> {
  const statusSequence = options.statusSequence ?? ['pending'];
  let statusCalls = 0;

  await page.route('**/pos/orders/v1/orders/public', (route) => {
    if (route.request().method() !== 'POST') return route.fallback();
    const submitStatus = options.submitStatus ?? 200;
    if (submitStatus >= 400) {
      return route.fulfill({
        status: submitStatus,
        contentType: 'application/json',
        body: JSON.stringify({ error: 'ORDER_FAILED', message: 'Could not place the order' }),
      });
    }
    return route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({
        order_id: '9001',
        order_number: 'ORD-9001',
        status: 'pending',
        public_token: 'tok-9001',
      }),
    });
  });

  await page.route('**/pos/orders/v1/orders/9001/public-status*', (route) => {
    const status = statusSequence[Math.min(statusCalls, statusSequence.length - 1)];
    statusCalls += 1;
    return route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({ order_id: '9001', order_number: 'ORD-9001', status }),
    });
  });
}

async function addItemToCart(page: Page, itemName: string): Promise<void> {
  await page.getByRole('button', { name: `Add ${itemName} to cart` }).click();
}

/** Opens the cart drawer and hands off to checkout. */
async function openCartAndCheckout(page: Page): Promise<void> {
  await page.getByRole('button', { name: /View Order/ }).click();
  const dialog = page.getByRole('dialog', { name: 'Your Order' });
  await expect(dialog).toBeVisible();
  await dialog.getByRole('button', { name: /Place Order/ }).click();
  await expect(page).toHaveURL(/\/menu\/checkout/);
  // The first lazy load of the checkout chunk can make Vite re-optimise deps
  // and reload the page. Wait for the menu route to be gone so the item name
  // resolves to the checkout summary alone.
  await expect(page.getByRole('heading', { name: 'Order Summary' })).toBeVisible();
  await expect(page.getByLabel('Your Order')).toHaveCount(0);
}

async function fillCustomerDetails(page: Page): Promise<void> {
  await page.getByLabel('Name *').fill('Sam Tester');
  await page.getByLabel('Phone Number *').fill('+1 555 0199');
}

test.describe('Public menu ordering (slug routes)', () => {
  test.beforeEach(async ({ page, context }) => {
    await suppressConsentBanner(context);
    await stubMenuData(page);
    await stubOrders(page);
  });

  test('adds an item, checks out, and lands on the order status page', async ({ page }) => {
    await page.goto(`${PORTAL_URL}/acme/menu/downtown`);
    await addItemToCart(page, 'Falafel wrap');
    await openCartAndCheckout(page);

    await expect(page.getByText('Falafel wrap')).toBeVisible();
    await expect(page.getByRole('button', { name: 'Place Order · $5.50' })).toBeVisible();

    await fillCustomerDetails(page);
    await page.getByRole('button', { name: /Place Order/ }).click();

    await expect(page).toHaveURL(/\/menu\/order-status\/9001/);
    await expect(page).toHaveURL(/number=ORD-9001/);
    await expect(page).toHaveURL(/slug=acme/);
    await expect(page.getByText('Order received!')).toBeVisible();
    await expect(page.getByText('ORD-9001')).toBeVisible();

    // The slug carried through checkout lets the CTA return to the business
    // menu rather than the Portal root.
    await page.getByRole('button', { name: 'Order Another' }).click();
    await expect(page).toHaveURL(/\/acme\/menu$/);
  });

  test('rejects an empty name, an empty phone and a malformed email', async ({ page }) => {
    await page.goto(`${PORTAL_URL}/acme/menu/downtown`);
    await addItemToCart(page, 'Falafel wrap');
    await openCartAndCheckout(page);

    await page.getByRole('button', { name: /Place Order/ }).click();
    await expect(page.getByRole('alert')).toHaveText('Please enter your name.');

    await page.getByLabel('Name *').fill('Sam');
    await page.getByRole('button', { name: /Place Order/ }).click();
    await expect(page.getByRole('alert')).toHaveText('Please enter your phone number.');

    await page.getByLabel('Phone Number *').fill('+1 555 0199');
    await page.getByLabel('Email (optional)').fill('not-an-email');
    await page.getByRole('button', { name: /Place Order/ }).click();
    await expect(page.getByRole('alert')).toHaveText('Please enter a valid email address.');
  });

  test('shows an error and keeps the cart when the order API fails', async ({ page }) => {
    // Registered after the default stub, so this one wins.
    await stubOrders(page, { submitStatus: 500 });

    await page.goto(`${PORTAL_URL}/acme/menu/downtown`);
    await addItemToCart(page, 'Falafel wrap');
    await openCartAndCheckout(page);
    await fillCustomerDetails(page);
    await page.getByRole('button', { name: /Place Order/ }).click();

    await expect(page.getByRole('alert')).toHaveText('Failed to place order. Please try again.');
    await expect(page).toHaveURL(/\/menu\/checkout/);
    await expect(page.getByText('Falafel wrap')).toBeVisible();
    await expect(page.getByRole('button', { name: /Place Order/ })).toBeEnabled();
  });

  test('updates the displayed status when polling sees a new state', async ({ page }) => {
    await stubOrders(page, { statusSequence: ['pending', 'preparing'] });
    // Timers are faked so the 30s poll interval fires without a real wait.
    await page.clock.install();

    await page.goto(
      `${PORTAL_URL}/menu/order-status/9001?number=ORD-9001&business_id=1&public_token=tok-9001`
    );
    await expect(page.getByText('Order received!')).toBeVisible();

    await page.clock.fastForward(31_000);

    await expect(page.getByText('Being prepared...')).toBeVisible();
  });

  test('keeps the cart across a page reload on checkout', async ({ page }) => {
    await page.goto(`${PORTAL_URL}/acme/menu/downtown`);
    await addItemToCart(page, 'Falafel wrap');
    await openCartAndCheckout(page);
    await expect(page.getByText('Falafel wrap')).toBeVisible();

    await page.reload();

    await expect(page.getByText('Falafel wrap')).toBeVisible();
    await expect(page.getByRole('button', { name: /Place Order/ })).toBeVisible();
  });

  test('does not leak the cart between two businesses', async ({ page }) => {
    await page.goto(`${PORTAL_URL}/acme/menu/downtown`);
    await addItemToCart(page, 'Falafel wrap');
    await openCartAndCheckout(page);
    await expect(page.getByText('Falafel wrap')).toBeVisible();

    // Switching business must drop the abandoned Acme cart rather than submit
    // Acme's items and prices under Beta's id.
    await page.goto(`${PORTAL_URL}/beta/menu/downtown`);
    await addItemToCart(page, 'Beta Burger');
    await openCartAndCheckout(page);

    await expect(page.getByText('Beta Burger')).toBeVisible();
    await expect(page.getByRole('button', { name: 'Place Order · $8.00' })).toBeVisible();
    await expect(page.getByText('Falafel wrap')).not.toBeVisible();
  });

  test('order-another returns to the business menu through the slug', async ({ page }) => {
    await page.goto(
      `${PORTAL_URL}/menu/order-status/9001?number=ORD-9001&business_id=1&slug=acme&public_token=tok-9001`
    );
    await expect(page.getByText('Order received!')).toBeVisible();

    await page.getByRole('button', { name: 'Order Another' }).click();

    await expect(page).toHaveURL(/\/acme\/menu$/);
  });

  test('ignores a return_url on another origin instead of following it', async ({ page }) => {
    await page.goto(
      `${PORTAL_URL}/menu/order-status/9001?number=ORD-9001&business_id=1&slug=acme&public_token=tok-9001` +
        `&return_url=${encodeURIComponent('https://evil.example/steal')}`
    );
    await expect(page.getByText('Order received!')).toBeVisible();

    await page.getByRole('button', { name: 'Order Another' }).click();

    // The foreign origin is dropped and the slug path is used instead.
    await expect(page).toHaveURL(/\/acme\/menu$/);
    expect(page.url().startsWith(PORTAL_URL)).toBe(true);
  });
});

test.describe('Public menu ordering (POS-backed /b route)', () => {
  test.beforeEach(async ({ page, context }) => {
    await suppressConsentBanner(context);
    await stubMenuData(page);
    await stubOrders(page);
  });

  test('adds an item on /b/42/menu and checks out under business 42', async ({ page }) => {
    await page.goto(`${PORTAL_URL}/b/42/menu`);
    await addItemToCart(page, 'Cold Brew');
    await openCartAndCheckout(page);

    await expect(page.getByText('Cold Brew')).toBeVisible();
    await fillCustomerDetails(page);
    await page.getByRole('button', { name: /Place Order/ }).click();

    await expect(page).toHaveURL(/\/menu\/order-status\/9001/);
    await expect(page).toHaveURL(/business_id=42/);
    await expect(page.getByText('Order received!')).toBeVisible();
  });
});
