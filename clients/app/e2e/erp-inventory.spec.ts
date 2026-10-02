import { test, expect, type Page, type Route } from './fixtures';

/**
 * End-to-end: an operator managing the ERP product catalog.
 *
 * Products is the ERP app's primary data surface - if the New Product /
 * Edit / Delete flow breaks silently, the operator quietly loses the
 * ability to keep the catalog in sync with reality, and every downstream
 * surface (inventory, low-stock, purchasing) drifts. This spec locks the
 * critical path down in a real browser with a stubbed API so CI stays
 * deterministic, no backend required.
 *
 * Stubs live in `installProductStubs` - keep them in one place so the
 * spec doesn't drift as new endpoints land. Every endpoint the Products
 * screen + detail screen touches should have a handler there.
 */

const ERP_URL = process.env.ERP_URL ?? 'http://localhost:5197/erp';

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
  id: '1',
  email: 'owner@acme.test',
  full_name: 'Acme Owner',
  business_id: '1',
  member_type: 'owner',
  effective_access: {
    business_id: '1',
    is_owner: true,
    is_admin: false,
    dash_permissions: [],
    subscribed_apps: ['erp'],
    licenses: [],
    apps: [
      {
        app_id: 'erp',
        licensed: true,
        has_access: true,
        permissions: [],
        permission_scopes: [],
        sources: ['system'],
      },
    ],
    authorization_version: '1',
  },
  apps: ['erp'],
  plan_slug: 'pro',
};

const FAKE_CATEGORIES = [
  { id: '1', name: 'Beverages', description: 'Drinks' },
  { id: '2', name: 'Snacks', description: 'Bites' },
];

const FAKE_BRANDS = [{ id: '1', name: 'Acme', description: 'House brand' }];

const FAKE_SUPPLIERS = [
  {
    id: '1',
    name: 'Global Grains Co.',
    email: 'alice@globalgrains.test',
    phone: '+1 555 0001',
    city: 'Chicago',
    country: 'US',
    is_active: true,
  },
  {
    id: '2',
    name: 'Dairy Direct Ltd.',
    email: 'bob@dairydirect.test',
    phone: '+1 555 0002',
    city: 'Boston',
    country: 'US',
    is_active: true,
  },
  {
    id: '3',
    name: 'FreshFarms',
    email: 'freshfarms@example.test',
    phone: '+1 555 0003',
    city: 'Austin',
    country: 'US',
    is_active: true,
  },
];

function categoryIdForName(name: string): string {
  return FAKE_CATEGORIES.find((category) => category.name === name)?.id ?? '0';
}

function toWireProduct(product: StubProduct): Record<string, unknown> {
  return {
    id: String(product.id),
    name: product.name,
    sku: product.sku,
    category_id: categoryIdForName(product.category),
    supplier_id: '1',
    unit: product.unit,
    cost_price: product.costPrice,
    sale_price: product.salePrice,
    stock_quantity: product.stock,
    low_stock_threshold: product.lowStock,
    reorder_quantity: 0,
    is_active: true,
    track_inventory: true,
    description: 'Test product',
    barcode: `${product.sku}-BARCODE`,
    brand_id: '0',
    currency: 'USD',
  };
}

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

  // All stubs are scoped to the backend host (localhost:9999) - matching
  // against `**/erp/**` would also swallow Vite's own module requests
  // (the dev server hosts source under paths that include `/erp/`), so
  // we anchor the glob to the API origin to keep the two worlds separate.
  //
  // Registration order matters: Playwright matches routes in REVERSE
  // registration order, so the most generic catch-alls must be registered
  // FIRST and the most specific handlers LAST.
  const apiHost = 'http://localhost:9999';

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
  // would trigger an auth redirect that tears the page down - neutral
  // empty response keeps the surface quiet.
  await page.route(`${apiHost}/subscriptions/**`, (route: Route) =>
    route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({ subscription: null, billing: null }),
    })
  );

  // (3) AppGuard expects /hub/user/v1/users/me to resolve with a user who has ERP access.
  await page.route(`${apiHost}/hub/user/v1/users/me`, (route: Route) =>
    route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify(FAKE_USER),
    })
  );

  // (4) Categories + brands feed the filter drawer + quick-add dropdown.
  await page.route(`${apiHost}/erp/category/v1/categories**`, (route: Route) =>
    route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({ categories: FAKE_CATEGORIES }),
    })
  );
  await page.route(`${apiHost}/erp/supplier/v1/suppliers**`, (route: Route) =>
    route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({ suppliers: FAKE_SUPPLIERS, has_more: false }),
    })
  );
  await page.route(`${apiHost}/erp/brand/v1/brands**`, (route: Route) =>
    route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({ brands: FAKE_BRANDS }),
    })
  );

  // (5) Products use the generated OneKit paths and snake_case wire payloads.
  await page.route(`${apiHost}/erp/product/v1/products**`, async (route: Route) => {
    const req = route.request();
    if (req.method() === 'GET') {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          products: state.products.map(toWireProduct),
          total: state.products.length,
          has_more: false,
        }),
      });
      return;
    }
    if (req.method() === 'POST') {
      const body = (req.postDataJSON() ?? {}) as Record<string, unknown>;
      captured.lastPost = body;
      const created: StubProduct = {
        id: state.nextId++,
        name: String(body.name ?? ''),
        sku: String(body.sku ?? ''),
        category:
          FAKE_CATEGORIES.find((category) => category.id === String(body.category_id))?.name ?? '',
        stock: Number(body.stock_quantity ?? 0),
        lowStock: Number(body.low_stock_threshold ?? 10),
        costPrice: Number(body.cost_price ?? 0),
        salePrice: Number(body.sale_price ?? 0),
        status: 'In Stock',
        brand: null,
        unit: String(body.unit ?? 'pcs'),
      };
      state.products.push(created);
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({ product_id: String(created.id), message: 'created' }),
      });
      return;
    }
    await route.fulfill({
      status: 500,
      contentType: 'application/json',
      body: JSON.stringify({ message: `unhandled ${req.method()}` }),
    });
  });

  // (6) Barcode lookup - registered BEFORE the by-id handler so the
  // /products/barcode/* path doesn't get mis-parsed as /products/<id>.
  await page.route(`${apiHost}/erp/product/v1/barcodes/**`, (route: Route) =>
    route.fulfill({ status: 404, contentType: 'application/json', body: '{}' })
  );

  // (7) Most-specific: product-by-id and full-product handlers. Registered
  // LAST so they win the match against the list handler above.
  await page.route(
    /localhost:9999\/erp\/product\/v1\/products\/(\d+)(\/full)?(?:\?.*)?$/,
    async (route: Route) => {
      const match = route
        .request()
        .url()
        .match(/\/erp\/product\/v1\/products\/(\d+)(\/full)?(?:\?.*)?$/);
      const id = Number(match?.[1] ?? 0);
      const product = state.products.find((item) => item.id === id);
      if (!product) {
        await route.fulfill({
          status: 404,
          contentType: 'application/json',
          body: JSON.stringify({ message: 'not found' }),
        });
        return;
      }

      if (route.request().method() === 'GET') {
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({ product: toWireProduct(product) }),
        });
        return;
      }

      if (route.request().method() === 'PUT') {
        const body = (route.request().postDataJSON() ?? {}) as Record<string, unknown>;
        captured.lastPost = { id, ...body };
        const categoryName = FAKE_CATEGORIES.find(
          (category) => category.id === String(body.category_id)
        )?.name;
        state.products = state.products.map((item) =>
          item.id === id
            ? {
                ...item,
                name: typeof body.name === 'string' ? body.name : item.name,
                sku: typeof body.sku === 'string' ? body.sku : item.sku,
                category: categoryName ?? item.category,
                stock: typeof body.stock_quantity === 'number' ? body.stock_quantity : item.stock,
                lowStock:
                  typeof body.low_stock_threshold === 'number'
                    ? body.low_stock_threshold
                    : item.lowStock,
                costPrice: typeof body.cost_price === 'number' ? body.cost_price : item.costPrice,
                salePrice: typeof body.sale_price === 'number' ? body.sale_price : item.salePrice,
                unit: typeof body.unit === 'string' ? body.unit : item.unit,
              }
            : item
        );
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({ message: 'updated' }),
        });
        return;
      }

      if (route.request().method() === 'DELETE') {
        state.products = state.products.filter((item) => item.id !== id);
        captured.lastPost = { id, _action: 'delete' };
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({ message: 'deleted' }),
        });
        return;
      }

      await route.fulfill({ status: 405, body: '' });
    }
  );

  return { captured };
}

// ─── Stub factories for additional flows ──────────────────────────────────────

interface StubSupplier {
  id: number;
  name: string;
  contact_name: string;
  email: string;
  phone: string;
  product_count: number;
  categories: string[];
  status: 'Active' | 'Inactive';
}

interface StubPurchaseOrder {
  id: number;
  supplier: string;
  supplier_id: number;
  date: string;
  expected_date: string;
  items: number;
  total: string;
  status: 'Draft' | 'Sent' | 'Received' | 'Pending Approval' | 'Approved';
}

interface StubCategory {
  id: number;
  name: string;
  description?: string;
  product_count: number;
  sort_order: number;
}

/**
 * Installs stubs for the Suppliers screen: GET /erp/suppliers + POST /erp/suppliers.
 * Also installs the common auth + subscription stubs required by AppGuard / layout.
 */
async function installSupplierStubs(
  page: Page,
  initial: StubSupplier[] = []
): Promise<{ captured: { lastPost: Record<string, unknown> | null } }> {
  const state: { suppliers: StubSupplier[]; nextId: number } = {
    suppliers: [...initial],
    nextId: initial.length > 0 ? Math.max(...initial.map((s) => s.id)) + 1 : 1,
  };
  const captured: { lastPost: Record<string, unknown> | null } = { lastPost: null };
  const apiHost = 'http://localhost:9999';

  // (1) Widest ERP catch-all - registered first so specific handlers win.
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

  // (2) Subscription / billing polled by the layout shell.
  await page.route(`${apiHost}/subscriptions/**`, (route: Route) =>
    route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({ subscription: null, billing: null }),
    })
  );

  // (3) AppGuard user identity.
  await page.route(`${apiHost}/hub/user/v1/users/me`, (route: Route) =>
    route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify(FAKE_USER),
    })
  );

  // (4) Suppliers list + create (POST).
  await page.route(`${apiHost}/erp/supplier/v1/suppliers**`, async (route: Route) => {
    const req = route.request();
    if (req.method() === 'GET') {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          suppliers: state.suppliers.map((supplier) => ({
            id: String(supplier.id),
            name: supplier.name,
            email: supplier.email,
            phone: supplier.phone,
            city: 'Chicago',
            country: 'US',
            is_active: supplier.status === 'Active',
          })),
          has_more: false,
        }),
      });
      return;
    }
    if (req.method() === 'POST') {
      const body = (req.postDataJSON() ?? {}) as Record<string, unknown>;
      captured.lastPost = body;
      const created: StubSupplier = {
        id: state.nextId++,
        name: String(body.name ?? ''),
        contact_name: '',
        email: String(body.email ?? ''),
        phone: String(body.phone ?? ''),
        product_count: 0,
        categories: [],
        status: 'Active',
      };
      state.suppliers.push(created);
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({ supplier_id: String(created.id), message: 'created' }),
      });
    }
  });

  return { captured };
}

/**
 * Installs stubs for the Purchase Orders screen.
 */
async function installPurchaseOrderStubs(
  page: Page,
  initial: StubPurchaseOrder[] = []
): Promise<{ captured: { lastPost: Record<string, unknown> | null } }> {
  const state: { orders: StubPurchaseOrder[]; nextId: number } = {
    orders: [...initial],
    nextId: initial.length > 0 ? Math.max(...initial.map((o) => o.id)) + 1 : 1,
  };
  const captured: { lastPost: Record<string, unknown> | null } = { lastPost: null };
  const apiHost = 'http://localhost:9999';

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

  await page.route(`${apiHost}/subscriptions/**`, (route: Route) =>
    route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({ subscription: null, billing: null }),
    })
  );

  await page.route(`${apiHost}/hub/user/v1/users/me`, (route: Route) =>
    route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify(FAKE_USER),
    })
  );

  await page.route(`${apiHost}/erp/supplier/v1/suppliers**`, (route: Route) =>
    route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({ suppliers: FAKE_SUPPLIERS, has_more: false }),
    })
  );

  await page.route(`${apiHost}/erp/purchase-order/v1/purchase-orders**`, async (route: Route) => {
    const req = route.request();
    if (req.method() === 'GET') {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          purchase_orders: state.orders.map((order) => ({
            id: String(order.id),
            supplier_id: String(order.supplier_id),
            po_number: `PO-${order.id}`,
            status: order.status === 'Pending Approval' ? 'sent' : order.status.toLowerCase(),
            subtotal: Number(order.total.replace(/[^0-9.]/g, '')),
            tax_amount: 0,
            total: Number(order.total.replace(/[^0-9.]/g, '')),
            expected_delivery: order.expected_date,
            items: Array.from({ length: order.items }, (_, index) => ({
              id: String(index + 1),
              product_id: '1',
              quantity_ordered: 1,
              quantity_received: 0,
              unit_cost: 0,
              total_cost: 0,
            })),
          })),
          has_more: false,
        }),
      });
      return;
    }
    if (req.method() === 'POST') {
      const body = (req.postDataJSON() ?? {}) as Record<string, unknown>;
      captured.lastPost = body;
      const created: StubPurchaseOrder = {
        id: state.nextId++,
        supplier: '',
        supplier_id: Number(body.supplier_id ?? 0),
        date: new Date().toISOString().slice(0, 10),
        expected_date: String(body.expected_delivery ?? new Date().toISOString().slice(0, 10)),
        items: Array.isArray(body.items) ? body.items.length : 0,
        total: '$0.00',
        status: 'Draft',
      };
      state.orders.push(created);
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({ purchase_order_id: String(created.id), message: 'created' }),
      });
    }
  });

  // Approve endpoint.
  await page.route(
    /localhost:9999\/erp\/purchase-order\/v1\/purchase-orders\/(\d+)\/approve/,
    async (route: Route) => {
      const url = new URL(route.request().url());
      const parts = url.pathname.split('/');
      const id = Number(parts[parts.indexOf('purchase-orders') + 1]);
      const order = state.orders.find((o) => o.id === id);
      if (order) {
        order.status = 'Approved';
      }
      captured.lastPost = { id, action: 'approve' };
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({ message: 'approved' }),
      });
    }
  );

  return { captured };
}

/**
 * Installs stubs for the Categories screen.
 */
async function installCategoryStubs(
  page: Page,
  initial: StubCategory[] = []
): Promise<{ captured: { lastPost: Record<string, unknown> | null } }> {
  const state: { categories: StubCategory[]; nextId: number } = {
    categories: [...initial],
    nextId: initial.length > 0 ? Math.max(...initial.map((c) => c.id)) + 1 : 1,
  };
  const captured: { lastPost: Record<string, unknown> | null } = { lastPost: null };
  const apiHost = 'http://localhost:9999';

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

  await page.route(`${apiHost}/subscriptions/**`, (route: Route) =>
    route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({ subscription: null, billing: null }),
    })
  );

  await page.route(`${apiHost}/hub/user/v1/users/me`, (route: Route) =>
    route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify(FAKE_USER),
    })
  );

  await page.route(`${apiHost}/erp/category/v1/categories**`, async (route: Route) => {
    const req = route.request();
    if (req.method() === 'GET') {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          categories: state.categories.map((category) => ({
            id: String(category.id),
            name: category.name,
            description: category.description ?? '',
          })),
          has_more: false,
          total: state.categories.length,
        }),
      });
      return;
    }
    if (req.method() === 'POST') {
      const body = (req.postDataJSON() ?? {}) as Record<string, unknown>;
      captured.lastPost = body;
      const created: StubCategory = {
        id: state.nextId++,
        name: String(body.name ?? ''),
        product_count: 0,
        sort_order: state.categories.length + 1,
      };
      state.categories.push(created);
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({ category_id: String(created.id), message: 'created' }),
      });
      return;
    }

    const id = Number(req.url().match(/\/categories\/(\d+)(?:\?.*)?$/)?.[1] ?? 0);
    if (req.method() === 'PUT') {
      const body = (req.postDataJSON() ?? {}) as Record<string, unknown>;
      captured.lastPost = { id, ...body, _action: 'update' };
      state.categories = state.categories.map((category) =>
        category.id === id
          ? {
              ...category,
              name: typeof body.name === 'string' ? body.name : category.name,
              description:
                typeof body.description === 'string' ? body.description : category.description,
            }
          : category
      );
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({ message: 'updated' }),
      });
      return;
    }

    if (req.method() === 'DELETE') {
      state.categories = state.categories.filter((category) => category.id !== id);
      captured.lastPost = { id, _action: 'delete' };
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({ message: 'deleted' }),
      });
    }
  });

  return { captured };
}

// ─── Test suites ──────────────────────────────────────────────────────────────

test.describe('ERP inventory flow', () => {
  // AppGuard redirects to auth when there's no token cookie. A dummy
  // value is enough because we stub /hub/user/v1/users/me - the guard only checks
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
    await expect(page.getByRole('button', { name: 'Create Product' })).toBeVisible();
  });

  test('create a product via the New Product button', async ({ page }) => {
    const { captured } = await installProductStubs(page, []);
    await page.goto(`${ERP_URL}/products`);

    // Wait for empty state to confirm initial list loaded.
    await expect(page.getByText(/No products found/i)).toBeVisible({ timeout: 10_000 });

    // Open the Quick Add modal.
    await page
      .getByRole('dialog', { name: 'Cookie consent' })
      .getByRole('button', { name: 'Decline' })
      .click();
    await page.getByTestId('products-new-button').click();
    await expect(page.getByTestId('quick-add-product-modal')).toBeVisible();

    await page.getByTestId('qap-name').fill('Arabica Coffee');
    await page.getByTestId('qap-sku').fill('BEV-001');
    await page.getByTestId('qap-category').selectOption({ label: 'Beverages' });
    await page.getByLabel('Supplier').selectOption({ label: 'Global Grains Co.' });
    await page.getByLabel(/barcode/i).fill('BEV-001-BARCODE');
    await page.getByLabel(/description/i).fill('Arabica coffee beans');

    await page.getByTestId('qap-submit').click();

    // Row appears after refetch.
    await expect(page.getByText('Arabica Coffee')).toBeVisible({ timeout: 5_000 });

    // Backend received the expected payload.
    expect(captured.lastPost).toMatchObject({
      name: 'Arabica Coffee',
      sku: 'BEV-001',
      category_id: '1',
      barcode: 'BEV-001-BARCODE',
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
    await page
      .getByRole('dialog', { name: 'Cookie consent' })
      .getByRole('button', { name: 'Decline' })
      .click();
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

    // Confirm dialog - target the modal's confirm button by testid so
    // we don't collide with the detail screen's own Delete button.
    await expect(page.getByTestId('confirm-modal')).toBeVisible();
    await page.getByTestId('confirm-modal-confirm').click();

    // Redirects back to /products and shows the empty state because the
    // single product has been removed from the stub state.
    await expect(page).toHaveURL(/\/inventory$/, { timeout: 5_000 });
    await expect(page.getByText(/No products found/i)).toBeVisible({ timeout: 5_000 });

    // Backend saw the delete payload.
    expect(captured.lastPost).toMatchObject({ id: 99, _action: 'delete' });
  });
});

// ─── Suppliers flow ───────────────────────────────────────────────────────────

test.describe('ERP suppliers flow', () => {
  test.beforeEach(async ({ context }) => {
    await context.addCookies([{ name: 'token', value: 'test-token', url: ERP_URL }]);
  });

  test('supplier names render from GET /erp/supplier/v1/suppliers', async ({ page }) => {
    const fakeSuppliers: StubSupplier[] = [
      {
        id: 1,
        name: 'Global Grains Co.',
        contact_name: 'Alice',
        email: 'alice@globalgrains.test',
        phone: '+1 555 0001',
        product_count: 4,
        categories: ['Grains'],
        status: 'Active',
      },
      {
        id: 2,
        name: 'Dairy Direct Ltd.',
        contact_name: 'Bob',
        email: 'bob@dairydirect.test',
        phone: '+1 555 0002',
        product_count: 7,
        categories: ['Dairy'],
        status: 'Active',
      },
    ];

    await installSupplierStubs(page, fakeSuppliers);
    await page.goto(`${ERP_URL}/suppliers`);

    await expect(page.getByText('Global Grains Co.')).toBeVisible({ timeout: 10_000 });
    await expect(page.getByText('Dairy Direct Ltd.')).toBeVisible({ timeout: 5_000 });
  });

  test('creating a supplier via the Add button makes it appear in the list', async ({ page }) => {
    const { captured } = await installSupplierStubs(page, []);
    await page.goto(`${ERP_URL}/suppliers`);

    // Open the "Add Supplier" modal by clicking the button that has
    // data-testid="suppliers-add-button".
    await page.getByTestId('suppliers-add-button').click();

    // Fill in the Company name field (first input rendered in the modal).
    const inputs = page.locator('input[placeholder="Company name"]');
    await inputs.first().fill('Fresh Farms Inc.');
    await page.locator('input[placeholder="Email"]').fill('freshfarms@example.test');
    await page.locator('input[placeholder="City"]').fill('Austin');
    await page.locator('input[placeholder="Country"]').fill('US');

    // Submit - the modal's "Create" button is the last enabled button.
    await page.getByRole('button', { name: /^Create$/i }).click();

    // The new card should be visible after the optimistic refetch.
    await expect(page.getByText('Fresh Farms Inc.')).toBeVisible({ timeout: 5_000 });

    // Backend received the name.
    expect(captured.lastPost).toMatchObject({ name: 'Fresh Farms Inc.' });
  });
});

// ─── Purchase orders flow ──────────────────────────────────────────────────────

test.describe('ERP purchase orders flow', () => {
  test.beforeEach(async ({ context }) => {
    await context.addCookies([{ name: 'token', value: 'test-token', url: ERP_URL }]);
  });

  test('PO number and supplier name render from GET /erp/purchase-order/v1/purchase-orders', async ({
    page,
  }) => {
    const fakePOs: StubPurchaseOrder[] = [
      {
        id: 101,
        supplier: 'Global Grains Co.',
        supplier_id: 1,
        date: '2026-04-01',
        expected_date: '2026-04-15',
        items: 3,
        total: '$450.00',
        status: 'Draft',
      },
    ];

    await installPurchaseOrderStubs(page, fakePOs);
    await page.goto(`${ERP_URL}/purchasing`);

    // The table renders "PO-<id>" in a monospace span.
    await expect(page.getByText('PO-101')).toBeVisible({ timeout: 10_000 });
    await expect(page.getByText('Global Grains Co.')).toBeVisible({ timeout: 5_000 });
  });

  test('draft PO shows a "Send" action button', async ({ page }) => {
    const fakePOs: StubPurchaseOrder[] = [
      {
        id: 202,
        supplier: 'Dairy Direct Ltd.',
        supplier_id: 2,
        date: '2026-04-10',
        expected_date: '2026-04-20',
        items: 1,
        total: '$120.00',
        status: 'Draft',
      },
    ];

    await installPurchaseOrderStubs(page, fakePOs);
    await page.goto(`${ERP_URL}/purchasing`);

    await expect(page.getByText('PO-202')).toBeVisible({ timeout: 10_000 });
    // Draft status rows show a "Send" button (not "Approve").
    await expect(page.getByRole('button', { name: /^Send$/i })).toBeVisible({ timeout: 5_000 });
  });

  test('Pending Approval PO shows an "Approve" button', async ({ page }) => {
    const fakePOs: StubPurchaseOrder[] = [
      {
        id: 303,
        supplier: 'FreshFarms',
        supplier_id: 3,
        date: '2026-04-12',
        expected_date: '2026-04-22',
        items: 5,
        total: '$900.00',
        status: 'Pending Approval',
      },
    ];

    await installPurchaseOrderStubs(page, fakePOs);
    await page.goto(`${ERP_URL}/purchasing`);

    await expect(page.getByText('PO-303')).toBeVisible({ timeout: 10_000 });
    await expect(page.getByRole('button', { name: /^Approve$/i })).toBeVisible({ timeout: 5_000 });
  });
});

// ─── Categories flow ───────────────────────────────────────────────────────────

test.describe('ERP categories flow', () => {
  test.beforeEach(async ({ context }) => {
    await context.addCookies([{ name: 'token', value: 'test-token', url: ERP_URL }]);
  });

  test('category names render from GET /erp/category/v1/categories', async ({ page }) => {
    const fakeCategories: StubCategory[] = [
      { id: 1, name: 'Beverages', description: 'Drinks', product_count: 5, sort_order: 1 },
      { id: 2, name: 'Snacks', description: 'Bites', product_count: 3, sort_order: 2 },
    ];

    await installCategoryStubs(page, fakeCategories);
    await page.goto(`${ERP_URL}/categories`);

    await expect(page.getByText('Beverages')).toBeVisible({ timeout: 10_000 });
    await expect(page.getByText('Snacks')).toBeVisible({ timeout: 5_000 });
  });

  test('editing a category name inline persists the updated name', async ({ page }) => {
    const fakeCategories: StubCategory[] = [
      { id: 10, name: 'Produce', product_count: 2, sort_order: 1 },
    ];

    const { captured } = await installCategoryStubs(page, fakeCategories);
    await page.goto(`${ERP_URL}/categories`);

    // Wait for the category to appear.
    await expect(page.getByText('Produce')).toBeVisible({ timeout: 10_000 });

    // Click the Edit (pencil) button for the first category row - it has
    // aria-label="Edit category".
    await page.getByRole('button', { name: 'Edit category' }).first().click();

    // The inline input should now be visible; clear and type the new name.
    const editInput = page.locator('input[placeholder]').last();
    await editInput.clear();
    await editInput.fill('Fresh Produce');

    // Press Enter to save.
    await editInput.press('Enter');

    // The updated name must appear in the row.
    await expect(page.getByText('Fresh Produce', { exact: true })).toBeVisible({ timeout: 5_000 });

    // Backend received the update action.
    expect(captured.lastPost).toMatchObject({ name: 'Fresh Produce' });
  });
});

// ─── Low-stock indicator ───────────────────────────────────────────────────────

test.describe('ERP inventory low-stock indicator', () => {
  test.beforeEach(async ({ context }) => {
    await context.addCookies([{ name: 'token', value: 'test-token', url: ERP_URL }]);
  });

  test('product with stock below lowStock threshold shows Low Stock badge', async ({ page }) => {
    // The /inventory route uses useProducts (GET /erp/products). We stub
    // both the common endpoints and the products list with a low-stock item.
    const lowStockProduct: StubProduct = {
      id: 77,
      name: 'Oat Milk',
      sku: 'OAT-001',
      category: 'Beverages',
      stock: 2,
      lowStock: 10,
      costPrice: 3,
      salePrice: 6,
      status: 'Low Stock',
      brand: null,
      unit: 'L',
    };

    await installProductStubs(page, [lowStockProduct]);
    await page.goto(`${ERP_URL}/inventory`);

    // Wait for the product row to appear.
    await expect(page.getByText('Oat Milk')).toBeVisible({ timeout: 10_000 });

    // The status badge cell should contain "Low Stock".
    await expect(page.getByText('Low Stock', { exact: true }).last()).toBeVisible({
      timeout: 5_000,
    });

    // The stock count cell is rendered with the amber colour when stock < lowStock.
    // We verify via text content that the value "2" is present in the table.
    await expect(page.getByText('2')).toBeVisible({ timeout: 5_000 });
  });
});
