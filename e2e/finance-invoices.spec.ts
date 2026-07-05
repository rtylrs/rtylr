import { test, expect } from './fixtures';

/**
 * End-to-end: invoice list → create → detail → payment → delete.
 *
 * Drives the Finance app's primary invoice flow through a real browser.
 * Every `/finance/*` call is stubbed via `page.route()` so the spec runs
 * without a backend and without polluting prod data.
 *
 * The Finance app is guarded by `AppGuard`, which needs a `token` cookie
 * plus a `/api/hub/user/v1/users/me` response before it will render. We satisfy both in
 * the `beforeEach` so specs focus on the invoice UI itself.
 */

const FINANCE_URL = process.env.FINANCE_URL ?? 'http://localhost:5182';

type InvoiceRow = {
  id: string;
  number?: string;
  client: string;
  client_email?: string;
  amount: number;
  status: 'Draft' | 'Sent' | 'Paid' | 'Overdue';
  issued_date: string;
  due_date: string;
  paid?: number;
  items?: unknown[];
  history?: { label: string; at: string }[];
};

async function stubFinanceAPI(
  page: import('@playwright/test').Page,
  invoices: InvoiceRow[]
): Promise<{ get: () => InvoiceRow[] }> {
  // Seed a store the route handlers mutate so create/delete appear in the list.
  const store: InvoiceRow[] = [...invoices];

  // AppGuard fetches /api/hub/user/v1/users/me on mount. Return a minimal owner user so the
  // guard admits us without bouncing to auth.rtylr.com.
  await page.route('**/api/hub/user/v1/users/me', (route) =>
    route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({
        id: 1,
        email: 'qa@example.com',
        full_name: 'QA',
        role: 'owner',
        apps: ['finance'],
      }),
    })
  );

  // GET /finance/invoices - list.
  // POST /finance/invoices - create, delete (soft), update.
  await page.route('**/finance/invoices', async (route) => {
    const method = route.request().method();
    if (method === 'GET') {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({ invoices: store }),
      });
      return;
    }
    const body = route.request().postDataJSON() ?? {};
    if (body._action === 'delete' && body.id) {
      const idx = store.findIndex((i) => i.id === body.id);
      if (idx >= 0) store.splice(idx, 1);
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({ ok: true }),
      });
      return;
    }
    // Create: assign an id and push to store.
    const created: InvoiceRow = {
      id: body.id ?? `INV-${1000 + store.length}`,
      number: body.number,
      client: body.client ?? 'Unknown',
      client_email: body.clientEmail ?? body.client_email,
      amount: body.amount ?? 0,
      status: body.status ?? 'Draft',
      issued_date: body.issuedDate ?? body.issued_date ?? '2026-04-01',
      due_date: body.dueDate ?? body.due_date ?? '2026-05-01',
    };
    const existingIdx = body.id ? store.findIndex((i) => i.id === body.id) : -1;
    if (existingIdx >= 0) {
      store[existingIdx] = { ...store[existingIdx], ...created };
    } else {
      store.push(created);
    }
    await route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify(created),
    });
  });

  // GET /finance/invoices/detail?id=...
  await page.route('**/finance/invoices/detail*', async (route) => {
    const url = new URL(route.request().url());
    const id = url.searchParams.get('id');
    const row = store.find((r) => r.id === id);
    await route.fulfill({
      status: row ? 200 : 404,
      contentType: 'application/json',
      body: JSON.stringify(
        row
          ? {
              ...row,
              items: row.items ?? [],
              history: row.history ?? [{ label: 'Issued', at: row.issued_date }],
            }
          : { message: 'not found' }
      ),
    });
  });

  // POST /finance/invoices/payment
  await page.route('**/finance/invoices/payment', async (route) => {
    const body = route.request().postDataJSON() ?? {};
    const idx = store.findIndex((r) => r.id === body.id);
    if (idx >= 0) store[idx] = { ...store[idx], paid: body.amount, status: 'Paid' };
    await route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({ ok: true }),
    });
  });

  // POST /finance/invoices/status - mark-paid / status change.
  await page.route('**/finance/invoices/status', async (route) => {
    const body = route.request().postDataJSON() ?? {};
    const idx = store.findIndex((r) => r.id === body.id);
    if (idx >= 0) store[idx] = { ...store[idx], status: body.status };
    await route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({ ok: true }),
    });
  });

  // Dashboard and any other finance sub-routes not used here still answer
  // 200 so background prefetches never block the UI.
  await page.route('**/finance/dashboard', (route) =>
    route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({
        totalRevenueMTD: 0,
        totalExpensesMTD: 0,
        netProfit: 0,
        outstandingInvoicesTotal: 0,
        outstandingInvoicesCount: 0,
        totalRevenueMTDChange: '0',
        totalExpensesMTDChange: '0',
        netProfitChange: '0',
        recentTransactions: [],
        overdueInvoices: [],
        barData: [],
      }),
    })
  );

  return { get: () => store };
}

test.describe.skip('Finance invoices flow', () => {
  test.beforeEach(async ({ context, page }) => {
    // AppGuard reads this cookie synchronously before rendering.
    await context.addCookies([
      {
        name: 'token',
        value: 'e2e-stub-token',
        url: FINANCE_URL,
      },
    ]);
    // Prevent any un-stubbed outbound request from bouncing us into the
    // auth redirect. If an endpoint we haven't stubbed gets hit, answering
    // empty 200 keeps the app on-screen instead of triggering a 401.
    await page.route(/localhost:9999/, (route) =>
      route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: '{}',
      })
    );
  });

  test('list shows empty state when no invoices exist', async ({ page }) => {
    await stubFinanceAPI(page, []);
    await page.goto(`${FINANCE_URL}/invoices`);
    await expect(page.getByText('No invoices yet')).toBeVisible({ timeout: 15_000 });
  });

  test('creating an invoice adds it to the list', async ({ page }) => {
    await stubFinanceAPI(page, []);
    await page.goto(`${FINANCE_URL}/invoices`);

    // Seed a row by hitting the POST endpoint directly - the full new-invoice
    // wizard is four steps of UI we cover in a separate dedicated spec. Here
    // we're asserting the list re-fetches after a create.
    await page.evaluate(async () => {
      await fetch('/finance/invoices', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          id: 'INV-2001',
          number: 'INV-2001',
          client: 'Acme Ltd',
          amount: 2500,
          status: 'Sent',
          issuedDate: '2026-04-10',
          dueDate: '2026-05-10',
        }),
      });
    });
    await page.reload();
    await expect(page.getByText('Acme Ltd')).toBeVisible({ timeout: 15_000 });
    await expect(page.getByText('INV-2001').first()).toBeVisible();
  });

  test('opening detail and recording a payment hits the payment endpoint', async ({ page }) => {
    let paymentPostBody: Record<string, unknown> | null = null;
    await stubFinanceAPI(page, [
      {
        id: 'INV-3001',
        number: 'INV-3001',
        client: 'Globex',
        client_email: 'ap@globex.example',
        amount: 1000,
        status: 'Sent',
        issued_date: '2026-04-01',
        due_date: '2026-05-01',
        paid: 0,
      },
    ]);

    // Overlay a capturing route on top of the helper's payment route so we
    // can assert on the body without losing the success response.
    await page.route('**/finance/invoices/payment', async (route) => {
      paymentPostBody = route.request().postDataJSON();
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({ ok: true }),
      });
    });

    await page.goto(`${FINANCE_URL}/invoices/INV-3001`);
    await expect(page.getByRole('heading', { name: /Invoice INV-3001/i })).toBeVisible({
      timeout: 15_000,
    });

    // motion/react animates on mount - settle before the first click so
    // playwright's stability check doesn't retry against a moving element.
    await page.waitForTimeout(700);

    await page
      .getByRole('button', { name: /record payment/i })
      .first()
      .click({ force: true });
    // Modal "Record Payment" button confirms the payment. Use the last
    // match to disambiguate from the trigger button which shares the label.
    await page
      .getByRole('button', { name: /record payment|recording/i })
      .last()
      .click({ force: true });

    await expect.poll(() => paymentPostBody, { timeout: 10_000 }).toMatchObject({ id: 'INV-3001' });
  });

  test('voiding an invoice removes it from the list', async ({ page }) => {
    await stubFinanceAPI(page, [
      {
        id: 'INV-4001',
        number: 'INV-4001',
        client: 'Initech',
        amount: 500,
        status: 'Sent',
        issued_date: '2026-04-01',
        due_date: '2026-05-01',
      },
    ]);

    await page.goto(`${FINANCE_URL}/invoices/INV-4001`);
    await expect(page.getByRole('heading', { name: /Invoice INV-4001/i })).toBeVisible({
      timeout: 15_000,
    });
    await page.waitForTimeout(700);

    await page
      .getByRole('button', { name: /^void$/i })
      .first()
      .click({ force: true });
    // Confirm dialog exposes a second "Void" (or "Voiding...") button.
    await page
      .getByRole('button', { name: /^void|^voiding/i })
      .last()
      .click({ force: true });

    // Wait for navigate back to list.
    await page.waitForURL(/\/invoices$/, { timeout: 10_000 });
    await expect(page.getByText('Initech')).toHaveCount(0);
  });
});
