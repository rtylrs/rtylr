import { test, expect } from './fixtures';

/**
 * End-to-end: invoice list → create → detail → payment → delete.
 *
 * Drives the Finance app's primary invoice flow through a real browser.
 * Every `/finance/*` call is stubbed via `page.route()` so the spec runs
 * without a backend and without polluting prod data.
 *
 * The Finance app is guarded by `AppGuard`, which needs a `token` cookie
 * plus a `/hub/user/v1/users/me` response before it will render. We satisfy both in
 * the `beforeEach` so specs focus on the invoice UI itself.
 */

const FINANCE_URL = process.env.FINANCE_URL ?? 'http://localhost:5182';
const API_URL = 'https://api.voxire.com';

type InvoiceRow = {
  id: string;
  number?: string;
  client: string;
  client_email?: string;
  amount: number;
  status: 'Draft' | 'Sent' | 'Paid' | 'Overdue';
  currency?: string;
  issued_date: string;
  due_date: string;
  paid?: number;
  items?: unknown[];
  history?: { label: string; at: string }[];
};

const API_STATUS: Record<InvoiceRow['status'], string> = {
  Draft: 'draft',
  Sent: 'sent',
  Paid: 'paid',
  Overdue: 'overdue',
};

function toSummary(row: InvoiceRow): Record<string, unknown> {
  const paid = row.paid ?? 0;
  return {
    id: row.id,
    invoice_number: row.number ?? row.id,
    client_name: row.client,
    status: API_STATUS[row.status],
    issue_date: row.issued_date,
    due_date: row.due_date,
    currency: row.currency ?? 'USD',
    total: row.amount,
    amount_paid: paid,
    amount_due: row.amount - paid,
  };
}

function toDetail(row: InvoiceRow): Record<string, unknown> {
  return {
    ...toSummary(row),
    client_email: row.client_email ?? 'billing@example.com',
    subtotal: row.amount,
    discount_type: 'none',
    discount_value: 0,
    tax_amount: 0,
    notes: '',
    items: row.items ?? [],
  };
}

async function stubFinanceAPI(
  page: import('@playwright/test').Page,
  invoices: InvoiceRow[]
): Promise<{ get: () => InvoiceRow[] }> {
  // Seed a store the route handlers mutate so create/delete appear in the list.
  const store: InvoiceRow[] = [...invoices];

  // The finance app now uses generated OneKit clients against the API host.
  // Register the fallback first so the focused endpoint handlers below win.
  await page.route(`${API_URL}/finance/**`, (route) =>
    route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: '{}',
    })
  );

  // AppGuard fetches /hub/user/v1/users/me on mount. Return a minimal owner user so the
  // guard admits us without bouncing to the auth app.
  await page.route(`${API_URL}/hub/user/v1/users/me`, (route) =>
    route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({
        id: 1,
        email: 'qa@example.com',
        full_name: 'QA',
        business_id: '1',
        role: 'owner',
        apps: ['finance'],
      }),
    })
  );

  await page.route(`${API_URL}/hub/subscription/v1/my`, (route) =>
    route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({ subscription: null }),
    })
  );

  await page.route(`${API_URL}/finance/dashboard`, (route) =>
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

  // GET/POST /finance/invoices/v1/invoices - list and create.
  await page.route(`${API_URL}/finance/invoices/v1/invoices**`, async (route) => {
    const method = route.request().method();
    const url = new URL(route.request().url());
    if (method === 'GET' && url.pathname.endsWith('/invoices')) {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          invoices: store.map(toSummary),
          page: 1,
          page_size: 50,
          has_more: false,
        }),
      });
      return;
    }

    if (method !== 'POST' || !url.pathname.endsWith('/invoices')) {
      await route.continue();
      return;
    }

    const body = route.request().postDataJSON() ?? {};

    // Create: shape the request as the generated API expects, then return its
    // response envelope so the app can invalidate and re-fetch the list.
    const created: InvoiceRow = {
      id: body.id ?? String(2000 + store.length + 1),
      number: body.invoice_number ?? body.number,
      client: body.client_name ?? 'Unknown',
      client_email: body.client_email,
      amount:
        body.items?.reduce(
          (sum: number, item: { total?: number; unit_price?: number; quantity?: number }) =>
            sum + (item.total ?? (item.unit_price ?? 0) * (item.quantity ?? 0)),
          0
        ) ?? 0,
      status: 'Draft',
      issued_date: body.issue_date ?? '2026-04-01',
      due_date: body.due_date ?? '2026-05-01',
      currency: body.currency ?? 'USD',
    };
    store.push(created);
    await route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({ invoice: toDetail(created) }),
    });
  });

  // GET/DELETE /finance/invoices/v1/invoices/{id} - detail and void.
  await page.route(`${API_URL}/finance/invoices/v1/invoices/*`, async (route) => {
    const url = new URL(route.request().url());
    const id = url.pathname.split('/').pop() ?? '';
    const row = store.find((r) => r.id === id);
    if (route.request().method() === 'GET') {
      await route.fulfill({
        status: row ? 200 : 404,
        contentType: 'application/json',
        body: JSON.stringify(row ? { invoice: toDetail(row), payments: [] } : {}),
      });
      return;
    }

    if (route.request().method() === 'DELETE') {
      const idx = store.findIndex((r) => r.id === id);
      if (idx >= 0) store.splice(idx, 1);
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({ message: 'deleted' }),
      });
      return;
    }

    await route.continue();
  });

  // PATCH /finance/invoices/v1/invoices/{id}/status.
  await page.route(`${API_URL}/finance/invoices/v1/invoices/*/status`, async (route) => {
    const id = new URL(route.request().url()).pathname.split('/').at(-2) ?? '';
    const body = route.request().postDataJSON() ?? {};
    const idx = store.findIndex((r) => r.id === id);
    if (idx >= 0) {
      const status = Object.entries(API_STATUS).find(([, value]) => value === body.status)?.[0];
      if (status) store[idx] = { ...store[idx], status: status as InvoiceRow['status'] };
    }
    await route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({ message: 'updated' }),
    });
  });

  // POST /finance/invoices/payments/v1/invoices/{id}/payment.
  await page.route(`${API_URL}/finance/invoices/payments/v1/invoices/*/payment`, async (route) => {
    const url = new URL(route.request().url());
    const invoiceId = url.pathname.split('/').at(-2) ?? '';
    const body = route.request().postDataJSON() ?? {};
    const idx = store.findIndex((r) => r.id === invoiceId);
    if (idx >= 0) {
      store[idx] = { ...store[idx], paid: body.amount, status: 'Paid' };
    }
    await route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({
        payment: {
          id: 'payment-1',
          invoice_id: invoiceId,
          amount: body.amount,
          currency: body.currency,
          payment_date: body.payment_date,
          method: body.method,
          reference: body.reference,
        },
      }),
    });
  });

  // POST /finance/invoices/v1/invoices/{id}/send.
  await page.route(`${API_URL}/finance/invoices/v1/invoices/*/send`, async (route) => {
    await route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({ message: 'sent' }),
    });
  });

  return { get: () => store };
}

test.describe('Finance invoices flow', () => {
  test.beforeEach(async ({ context, page }) => {
    // AppGuard reads this cookie synchronously before rendering.
    await context.addCookies([
      {
        name: 'token',
        value: 'e2e-stub-token',
        url: FINANCE_URL,
      },
    ]);
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
      await fetch('https://api.voxire.com/finance/invoices/v1/invoices', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          client_name: 'Acme Ltd',
          client_email: 'billing@acme.example',
          issue_date: '2026-04-10',
          due_date: '2026-05-10',
          currency: 'USD',
          discount_type: 'none',
          items: [{ description: 'Acme Ltd', quantity: 1, unit_price: 2500, total: 2500 }],
        }),
      });
    });
    await page.reload();
    await expect(page.getByText('Acme Ltd')).toBeVisible({ timeout: 15_000 });
    await expect(page.getByText('2001').first()).toBeVisible();
  });

  test('opening detail and recording a payment hits the payment endpoint', async ({ page }) => {
    let paymentPostBody: Record<string, unknown> | null = null;
    await stubFinanceAPI(page, [
      {
        id: '3001',
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
    await page.route(
      `${API_URL}/finance/invoices/payments/v1/invoices/*/payment`,
      async (route) => {
        paymentPostBody = route.request().postDataJSON();
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({ payment: { id: 'payment-1', invoice_id: '3001' } }),
        });
      }
    );

    await page.goto(`${FINANCE_URL}/invoices/3001`);
    await expect(page.getByRole('heading', { name: /Invoice 3001/i })).toBeVisible({
      timeout: 15_000,
    });

    // motion/react animates on mount - settle before the first click so
    // playwright's stability check doesn't retry against a moving element.
    await page.waitForTimeout(700);

    await page
      .getByRole('button', { name: /record payment/i })
      .first()
      .click({ force: true });
    await page.getByPlaceholder(/Transaction ID or receipt number/i).fill('e2e-payment');
    // Modal "Record Payment" button confirms the payment. Use the last
    // match to disambiguate from the trigger button which shares the label.
    await page
      .getByRole('button', { name: /record payment|recording/i })
      .last()
      .click({ force: true });

    await expect
      .poll(() => paymentPostBody, { timeout: 10_000 })
      .toMatchObject({
        invoice_id: '3001',
      });
  });

  test('voiding an invoice removes it from the list', async ({ page }) => {
    await stubFinanceAPI(page, [
      {
        id: '4001',
        number: 'INV-4001',
        client: 'Initech',
        amount: 500,
        status: 'Sent',
        issued_date: '2026-04-01',
        due_date: '2026-05-01',
      },
    ]);

    await page.goto(`${FINANCE_URL}/invoices/4001`);
    await expect(page.getByRole('heading', { name: /Invoice 4001/i })).toBeVisible({
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
