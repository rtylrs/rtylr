import { test, expect } from './fixtures';

/**
 * End-to-end: Finance - Expenses, Bills, Profit & Loss, Tax screens.
 *
 * Every `/finance/*` call is stubbed via `page.route()` so the spec runs
 * without a backend. The AppGuard is satisfied the same way as the invoice
 * spec: a `token` cookie plus a stubbed `/hub/user/v1/users/me` response.
 */

const FINANCE_URL = process.env.FINANCE_URL ?? 'http://localhost:5182';
const API_URL = 'https://api.voxire.com';

// ── Shared auth + fallback stub ───────────────────────────────────────────

async function seedAuth(page: import('@playwright/test').Page): Promise<void> {
  // The finance app now uses generated OneKit clients against the API host,
  // rather than the old localhost:9999 REST stub.
  await page.route(`${API_URL}/finance/**`, (route) =>
    route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: '{}',
    })
  );

  // AppGuard reads /hub/user/v1/users/me on mount.
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

  // The current layout fetches the subscription before rendering the page.
  await page.route(`${API_URL}/hub/subscription/v1/my`, (route) =>
    route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({ subscription: null }),
    })
  );

  // Silence the dashboard prefetch that most screens trigger.
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
}

// ── beforeEach helper: place auth cookie ─────────────────────────────────

test.describe('Finance - Expenses list', () => {
  test.beforeEach(async ({ context, page }) => {
    await context.addCookies([{ name: 'token', value: 'e2e-stub-token', url: FINANCE_URL }]);
    await seedAuth(page);
  });

  test('renders two expense rows from stub data', async ({ page }) => {
    // Stub the generated finance.expenses.v1 list method.
    await page.route(`${API_URL}/finance/expenses/v1/expenses`, async (route) => {
      if (route.request().method() !== 'GET') {
        await route.continue();
        return;
      }
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          expenses: [
            {
              id: 'exp-1001',
              expense_date: '2026-04-10',
              description: 'Office Supplies Purchase',
              category_id: '1',
              amount: 250,
              currency: 'USD',
              receipt_url: '',
              status: 'pending',
            },
            {
              id: 'exp-1002',
              expense_date: '2026-04-15',
              description: 'Cloud Hosting Invoice',
              category_id: '2',
              amount: 1200,
              currency: 'USD',
              receipt_url: 'https://example.com/receipts/exp-1002',
              status: 'approved',
            },
          ],
        }),
      });
    });

    await page.goto(`${FINANCE_URL}/expenses`);

    // Both descriptions should be visible in the table.
    await expect(page.getByText('Office Supplies Purchase')).toBeVisible({ timeout: 15_000 });
    await expect(page.getByText('Cloud Hosting Invoice')).toBeVisible({ timeout: 15_000 });

    // Amounts rendered with currency formatting (the component uses toLocaleString USD).
    await expect(page.getByText('$250').first()).toBeVisible({ timeout: 10_000 });
    await expect(page.getByText('$1,200').first()).toBeVisible({ timeout: 10_000 });
  });

  test('shows expense descriptions even when receipt flag differs', async ({ page }) => {
    await page.route(`${API_URL}/finance/expenses/v1/expenses`, async (route) => {
      if (route.request().method() !== 'GET') {
        await route.continue();
        return;
      }
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          expenses: [
            {
              id: 'exp-2001',
              expense_date: '2026-03-01',
              description: 'Staff Lunch',
              category_id: '3',
              amount: 80,
              currency: 'USD',
              receipt_url: 'https://example.com/receipts/exp-2001',
              status: 'approved',
            },
          ],
        }),
      });
    });

    await page.goto(`${FINANCE_URL}/expenses`);
    await expect(page.getByText('Staff Lunch')).toBeVisible({ timeout: 15_000 });
    // Receipt present → "Yes" indicator.
    await expect(page.getByText('Yes').first()).toBeVisible({ timeout: 10_000 });
  });
});

// ── Bills list ────────────────────────────────────────────────────────────

test.describe('Finance - Bills list', () => {
  test.beforeEach(async ({ context, page }) => {
    await context.addCookies([{ name: 'token', value: 'e2e-stub-token', url: FINANCE_URL }]);
    await seedAuth(page);
  });

  test('renders a draft bill with vendor name', async ({ page }) => {
    await page.route(`${API_URL}/finance/bill/v1/bills`, async (route) => {
      if (route.request().method() !== 'GET') {
        await route.continue();
        return;
      }
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          bills: [
            {
              id: 'BILL-3001',
              supplier_id: '1',
              supplier_name: 'Acme Suppliers',
              amount: 4500,
              currency: 'USD',
              status: 'unpaid',
              due_date: '2026-05-20',
              description: 'May office supplies',
            },
          ],
        }),
      });
    });

    await page.goto(`${FINANCE_URL}/bills`);

    await expect(page.getByText('Acme Suppliers')).toBeVisible({ timeout: 15_000 });
    await expect(page.getByText('BILL-3001').first()).toBeVisible({ timeout: 10_000 });
  });

  test('shows empty state when no bills exist', async ({ page }) => {
    await page.route(`${API_URL}/finance/bill/v1/bills`, async (route) => {
      if (route.request().method() !== 'GET') {
        await route.continue();
        return;
      }
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({ bills: [] }),
      });
    });

    await page.goto(`${FINANCE_URL}/bills`);
    await expect(page.getByText('No bills')).toBeVisible({ timeout: 15_000 });
  });
});

// ── Profit & Loss ─────────────────────────────────────────────────────────

test.describe('Finance - Profit & Loss report', () => {
  test.beforeEach(async ({ context, page }) => {
    await context.addCookies([{ name: 'token', value: 'e2e-stub-token', url: FINANCE_URL }]);
    await seedAuth(page);
  });

  test('renders revenue and net profit from stub data', async ({ page }) => {
    // The PL screen calls GET /finance/report/v1/pl?period=month by default.
    await page.route(`${API_URL}/finance/report/v1/pl**`, async (route) => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          period: 'month',
          revenue: 50000,
          cogs: 15000,
          operating_expenses: 15000,
          gross_profit: 35000,
          net_profit: 20000,
        }),
      });
    });

    await page.goto(`${FINANCE_URL}/pl`);

    // The PL component renders summary cards for "Revenue" and "Net Profit".
    // formatMoney(50000) → "$50,000"
    await expect(page.getByText(/\$50,000/).first()).toBeVisible({ timeout: 15_000 });
    // formatMoney(20000) → "$20,000"
    await expect(page.getByText(/\$20,000/).first()).toBeVisible({ timeout: 15_000 });

    // Heading is always present.
    await expect(page.getByRole('heading', { name: /Profit & Loss/i }).first()).toBeVisible({
      timeout: 10_000,
    });
  });

  test('period selector buttons render', async ({ page }) => {
    await page.route(`${API_URL}/finance/report/v1/pl**`, async (route) => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          period: 'month',
          revenue: 0,
          cogs: 0,
          operating_expenses: 0,
          gross_profit: 0,
          net_profit: 0,
        }),
      });
    });

    await page.goto(`${FINANCE_URL}/pl`);

    await expect(page.getByRole('button', { name: 'This Month' })).toBeVisible({
      timeout: 15_000,
    });
    await expect(page.getByRole('button', { name: 'This Year' })).toBeVisible({ timeout: 10_000 });
  });
});

// ── Tax rates screen ──────────────────────────────────────────────────────

test.describe('Finance - Tax rates screen', () => {
  test.beforeEach(async ({ context, page }) => {
    await context.addCookies([{ name: 'token', value: 'e2e-stub-token', url: FINANCE_URL }]);
    await seedAuth(page);
  });

  test('renders Standard VAT rate name and 15% after opening config panel', async ({ page }) => {
    await page.route(`${API_URL}/finance/report/v1/tax-config`, async (route) => {
      if (route.request().method() !== 'GET') {
        await route.continue();
        return;
      }
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          configs: [{ id: '1', name: 'Standard VAT', rate: 15, type: 'sales' }],
        }),
      });
    });

    // Stub the tax report so the page mounts cleanly.
    await page.route(`${API_URL}/finance/report/v1/tax`, async (route) => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({ period: 'month', tax_collected: 0, tax_owed: 0 }),
      });
    });

    await page.goto(`${FINANCE_URL}/tax`);

    // Open the "Tax Rates" config panel.
    await page.getByRole('button', { name: /Tax Rates/i }).click();

    // The panel should now show the name and the rate value.
    await expect(page.locator('input[value="Standard VAT"]')).toBeVisible({ timeout: 15_000 });
    // Rate input should show 15.
    await expect(page.locator('input[type="number"][value="15"]')).toBeVisible({
      timeout: 10_000,
    });
  });

  test('Tax page heading is visible without config panel', async ({ page }) => {
    await page.route(`${API_URL}/finance/report/v1/tax-config`, async (route) => {
      if (route.request().method() !== 'GET') {
        await route.continue();
        return;
      }
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({ configs: [] }),
      });
    });

    await page.route(`${API_URL}/finance/report/v1/tax`, async (route) => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({ period: 'month', tax_collected: 0, tax_owed: 0 }),
      });
    });

    await page.goto(`${FINANCE_URL}/tax`);

    await expect(page.getByRole('heading', { name: /^Tax$/i })).toBeVisible({ timeout: 15_000 });
  });
});
