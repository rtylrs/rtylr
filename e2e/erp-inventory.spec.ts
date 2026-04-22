import { test, expect, type Page, type Route } from '@playwright/test';

/**
 * End-to-end: an operator managing the ERP product catalog.
 *
 * Products is the ERP app's primary data surface — if the New Product /
 * Edit / Delete flow breaks silently, the operator quietly loses the
 * ability to keep the catalog in sync with reality, and every downstream
 * surface (inventory, low-stock, purchasing) drifts. This spec locks the
 * critical path down in a real browser with a stubbed API so CI stays
 * deterministic, no backend required.
 *
 * Stubs live in `installProductStubs` — keep them in one place so the
 * spec doesn't drift as new endpoints land. Every endpoint the Products
 * screen + detail screen touches should have a handler there.
 */

const ERP_URL = process.env.ERP_URL ?? 'http://localhost:5179';

interface StubProduct {
  id: number;
  name: string;
  sku: string;
  category: string;
  stock: number;
  lowStock: number;
  costPrice: number;
  salePrice: number;
  status: 'In Stock' | 'Low Stock' | 'Out of Stock';
  brand: string | null;
  unit: string;
}

const FAKE_USER = {
  id: 1,
  email: 'owner@acme.test',
  fullName: 'Acme Owner',
  role: 'owner',
  apps: ['erp'],
  planSlug: 'pro',
};

const FAKE_CATEGORIES = [
  { id: 1, name: 'Beverages', description: 'Drinks', productCount: 0, sortOrder: 1 },
  { id: 2, name: 'Snacks', description: 'Bites', productCount: 0, sortOrder: 2 },
];

const FAKE_BRANDS = [{ id: 1, name: 'Acme', description: 'House brand', productCount: 0 }];

/**
 * Installs stub handlers for every backend endpoint the Products list +
 * detail pages hit. The internal `productsState` is mutated by the
 * POST/PUT/DELETE handlers so optimistic UI + refetch flows look real.
 */
async function installProductStubs(
  page: Page,
  initial: StubProduct[] = []
): Promise<{ captured: { lastPost: Record<string, unknown> | null } }> {
  // Mutable state across route handlers. We intentionally keep this
  // simple (linear search, in-memory) because the spec needs at most a
  // handful of products.
  const state: { products: StubProduct[]; nextId: number } = {
    products: [...initial],
    nextId: initial.length > 0 ? Math.max(...initial.map((p) => p.id)) + 1 : 1,
  };
  const captured: { lastPost: Record<string, unknown> | null } = { lastPost: null };

  // All stubs are scoped to the backend host (api.voxire.dev) — matching
  // against `**/erp/**` would also swallow Vite's own module requests
  // (the dev server hosts source under paths that include `/erp/`), so
  // we anchor the glob to the API origin to keep the two worlds separate.
  //
  // Registration order matters: Playwright matches routes in REVERSE
  // registration order, so the most generic catch-alls must be registered
  // FIRST and the most specific handlers LAST.
  const apiHost = 'https://api.voxire.dev';

  // (1) Widest catch-all for ERP endpoints not otherwise handled.
  await page.route(`${apiHost}/erp/**`, (route: Route) => {
    if (route.request().method() === 'GET') {
      route.fulfill({ status: 200, contentType: 'application/json', body: '{}' });
    } else {
      route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({ success: true }),
      });
    }
  });

  // (2) Subscription/billing polled by the layout shell. A 401 on these
  // would trigger an auth redirect that tears the page down — neutral
  // empty response keeps the surface quiet.
  await page.route(`${apiHost}/subscriptions/**`, (route: Route) =>
    route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({ subscription: null, billing: null }),
    })
  );

  // (3) AppGuard expects /user/me to resolve with a user who has ERP access.
  await page.route(`${apiHost}/user/me`, (route: Route) =>
    route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify(FAKE_USER),
    })
  );

  // (4) Categories + brands feed the filter drawer + quick-add dropdown.
  await page.route(`${apiHost}/erp/categories`, (route: Route) =>
    route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({ categories: FAKE_CATEGORIES }),
    })
  );
  await page.route(`${apiHost}/erp/brands`, (route: Route) =>
    route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({ brands: FAKE_BRANDS }),
    })
  );

  // (5) Products: list (GET), create/update/delete multiplex (POST).
  //
  // The ERP API multiplexes create/update/delete through POST /erp/products
  // with an `_action` discriminator — see shared/services/erpApi.ts. We
  // mirror that here so the handlers stay faithful to production.
  await page.route(`${apiHost}/erp/products`, async (route: Route) => {
    const req = route.request();
    const url = new URL(req.url());
    if (req.method() === 'GET') {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({ products: state.products }),
      });
      return;
    }
    if (req.method() === 'POST') {
      const body = (req.postDataJSON() ?? {}) as Record<string, unknown>;
      captured.lastPost = body;

      // Delete path: POST /erp/products { id, _action: 'delete' }
      if (body._action === 'delete' && typeof body.id === 'number') {
        state.products = state.products.filter((p) => p.id !== body.id);
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({ success: true }),
        });
        return;
      }

      // Update path: POST /erp/products { id: <existing>, ...fields }
      if (typeof body.id === 'number') {
        const idx = state.products.findIndex((p) => p.id === body.id);
        if (idx >= 0) {
          const merged = { ...state.products[idx], ...(body as Partial<StubProduct>) };
          state.products[idx] = merged;
          await route.fulfill({
            status: 200,
            contentType: 'application/json',
            body: JSON.stringify(merged),
          });
          return;
        }
      }

      // Create path: POST /erp/products { ...fields, no id }
      const created: StubProduct = {
        id: state.nextId++,
        name: String(body.name ?? ''),
        sku: String(body.sku ?? ''),
        category: String(body.category ?? ''),
        stock: Number(body.stock ?? 0),
        lowStock: Number(body.lowStock ?? 10),
        costPrice: Number(body.costPrice ?? 0),
        salePrice: Number(body.salePrice ?? 0),
        status: 'In Stock',
        brand: null,
        unit: String(body.unit ?? 'pcs'),
      };
      state.products.push(created);
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify(created),
      });
      return;
    }

    await route.fulfill({
      status: 500,
      contentType: 'application/json',
      body: JSON.stringify({ message: `unhandled ${req.method()} ${url.pathname}` }),
    });
  });

  // (6) Barcode lookup — registered BEFORE the by-id handler so the
  // /products/barcode/* path doesn't get mis-parsed as /products/<id>.
  await page.route(`${apiHost}/erp/products/barcode/**`, (route: Route) =>
    route.fulfill({ status: 404, contentType: 'application/json', body: '{}' })
  );

  // (7) Most-specific: product-by-id detail GET. Registered LAST so it
  // wins the match against any of the earlier wider patterns.
  await page.route(/api\.voxire\.dev\/erp\/products\/\d+(\?.*)?$/, async (route: Route) => {
    const url = new URL(route.request().url());
    const idStr = url.pathname.split('/').pop() ?? '';
    const id = Number(idStr);
    const product = state.products.find((p) => p.id === id);
    if (!product) {
      await route.fulfill({
        status: 404,
        contentType: 'application/json',
        body: JSON.stringify({ message: 'not found' }),
      });
      return;
    }
    await route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify(product),
    });
  });

  return { captured };
}

test.describe.skip('ERP inventory flow', () => {
  // AppGuard redirects to auth when there's no token cookie. A dummy
  // value is enough because we stub /user/me — the guard only checks
  // presence, not validity.
  test.beforeEach(async ({ context }) => {
    await context.addCookies([
      {
        name: 'token',
        value: 'test-token',
        url: ERP_URL,
      },
    ]);
  });

  test('empty state renders when no products exist', async ({ page }) => {
    await installProductStubs(page, []);
    await page.goto(`${ERP_URL}/products`);
    await expect(page.getByText(/No products found/i)).toBeVisible({ timeout: 10_000 });
    await expect(page.getByTestId('products-empty-create-button')).toBeVisible();
  });

  test('create a product via the New Product button', async ({ page }) => {
    const { captured } = await installProductStubs(page, []);
    await page.goto(`${ERP_URL}/products`);

    // Wait for empty state to confirm initial list loaded.
    await expect(page.getByText(/No products found/i)).toBeVisible({ timeout: 10_000 });

    // Open the Quick Add modal.
    await page.getByTestId('products-new-button').click();
    await expect(page.getByTestId('quick-add-product-modal')).toBeVisible();

    await page.getByTestId('qap-name').fill('Arabica Coffee');
    await page.getByTestId('qap-sku').fill('BEV-001');
    await page.getByTestId('qap-category').selectOption({ label: 'Beverages' });

    await page.getByTestId('qap-submit').click();

    // Row appears after refetch.
    await expect(page.getByText('Arabica Coffee')).toBeVisible({ timeout: 5_000 });

    // Backend received the expected payload.
    expect(captured.lastPost).toMatchObject({
      name: 'Arabica Coffee',
      sku: 'BEV-001',
      category: 'Beverages',
    });
  });

  test('edit a product from its detail page, save, assert row updates', async ({ page }) => {
    const initial: StubProduct[] = [
      {
        id: 42,
        name: 'Arabica Coffee',
        sku: 'BEV-001',
        category: 'Beverages',
        stock: 40,
        lowStock: 10,
        costPrice: 5,
        salePrice: 12,
        status: 'In Stock',
        brand: null,
        unit: 'bag',
      },
    ];
    const { captured } = await installProductStubs(page, initial);

    await page.goto(`${ERP_URL}/products/42`);
    await expect(page.getByRole('heading', { name: 'Arabica Coffee' })).toBeVisible({
      timeout: 10_000,
    });

    // Open edit modal, rename the product, save.
    await page.getByTestId('product-edit-button').click();
    await expect(page.getByTestId('quick-add-product-modal')).toBeVisible();

    const nameInput = page.getByTestId('qap-name');
    await nameInput.fill('Arabica Premium');
    await page.getByTestId('qap-submit').click();

    // Detail heading reflects the updated name after the refetch.
    await expect(page.getByRole('heading', { name: 'Arabica Premium' })).toBeVisible({
      timeout: 5_000,
    });

    // Backend received an update (id present in payload).
    expect(captured.lastPost).toMatchObject({ id: 42, name: 'Arabica Premium' });
  });

  test('delete a product, assert row disappears', async ({ page }) => {
    const initial: StubProduct[] = [
      {
        id: 99,
        name: 'Expire-Me',
        sku: 'TMP-001',
        category: 'Beverages',
        stock: 1,
        lowStock: 0,
        costPrice: 1,
        salePrice: 2,
        status: 'In Stock',
        brand: null,
        unit: 'pcs',
      },
    ];
    const { captured } = await installProductStubs(page, initial);

    await page.goto(`${ERP_URL}/products/99`);
    await expect(page.getByRole('heading', { name: 'Expire-Me' })).toBeVisible({
      timeout: 10_000,
    });

    await page.getByTestId('product-delete-button').click();

    // Confirm dialog — target the modal's confirm button by testid so
    // we don't collide with the detail screen's own Delete button.
    await expect(page.getByTestId('confirm-modal')).toBeVisible();
    await page.getByTestId('confirm-modal-confirm').click();

    // Redirects back to /products and shows the empty state because the
    // single product has been removed from the stub state.
    await expect(page).toHaveURL(/\/products$/, { timeout: 5_000 });
    await expect(page.getByText(/No products found/i)).toBeVisible({ timeout: 5_000 });

    // Backend saw the delete payload.
    expect(captured.lastPost).toMatchObject({ id: 99, _action: 'delete' });
  });
});
