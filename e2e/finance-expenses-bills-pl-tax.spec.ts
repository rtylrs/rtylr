import { test, expect } from './fixtures';

/**
 * End-to-end: Finance - Expenses, Bills, Profit & Loss, Tax screens.
 *
 * Every `/finance/*` call is stubbed via `page.route()` so the spec runs
 * without a backend. The AppGuard is satisfied the same way as the invoice
 * spec: a `token` cookie plus a stubbed `/api/hub/user/v1/users/me` response.
 */

const FINANCE_URL = process.env.FINANCE_URL ?? 'http://localhost:5182';

// ── Shared auth + fallback stub ───────────────────────────────────────────

async function seedAuth(page: import('@playwright/test').Page): Promise<void> {
  // AppGuard reads /api/hub/user/v1/users/me on mount.
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

  // Silence the dashboard prefetch that most screens trigger.
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

  // Catch-all for any un-stubbed finance sub-routes so rogue prefetches
  // never cause a 401 redirect.
  await page.route(/localhost:9999/, (route) =>
    route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: '{}',
    })
  );
}

// ── beforeEach helper: place auth cookie ─────────────────────────────────

test.describe.skip('Finance - Expenses list', () => {
  test.beforeEach(async ({ context, page }) => {
    await context.addCookies([{ name: 'token', value: 'e2e-stub-token', url: FINANCE_URL }]);
    await seedAuth(page);
  });

  test('renders two expense rows from stub data', async ({ page }) => {
    // Stub GET /finance/expenses - envelope format used by the Go backend.
    await page.route('**/finance/expenses', async (route) => {
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
              date: '2026-04-10',
              description: 'Office Supplies Purchase',
              category: 'Other',
              amount: 250,
              receipt: false,
              approval: 'pending',
            },
            {
              id: 'exp-1002',
              date: '2026-04-15',
              description: 'Cloud Hosting Invoice',
              category: 'Utilities',
              amount: 1200,
              receipt: true,
              approval: 'approved',
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
    await page.route('**/finance/expenses', async (route) => {
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
              date: '2026-03-01',
              description: 'Staff Lunch',
              category: 'Staff',
              amount: 80,
              receipt: true,
              approval: 'approved',
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

test.describe.skip('Finance - Bills list', () => {
  test.beforeEach(async ({ context, page }) => {
    await context.addCookies([{ name: 'token', value: 'e2e-stub-token', url: FINANCE_URL }]);
    await seedAuth(page);
  });

  test('renders a draft bill with vendor name', async ({ page }) => {
    await page.route('**/finance/bills', async (route) => {
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
              vendor: 'Acme Suppliers',
              supplier: 'Acme Suppliers',
              category: 'Inventory',
              amount: 4500,
              status: 'unpaid',
              due_date: '2026-05-20',
              issued_date: '2026-04-20',
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
    await page.route('**/finance/bills', async (route) => {
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

test.describe.skip('Finance - Profit & Loss report', () => {
  test.beforeEach(async ({ context, page }) => {
    await context.addCookies([{ name: 'token', value: 'e2e-stub-token', url: FINANCE_URL }]);
    await seedAuth(page);
  });

  test('renders revenue and net profit from stub data', async ({ page }) => {
    // The PL screen calls GET /api/finance/report/v1/pl?period=month by default.
    await page.route('**/api/finance/report/v1/pl*', async (route) => {
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
    await page.route('**/api/finance/report/v1/pl*', async (route) => {
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

test.describe.skip('Finance - Tax rates screen', () => {
  test.beforeEach(async ({ context, page }) => {
    await context.addCookies([{ name: 'token', value: 'e2e-stub-token', url: FINANCE_URL }]);
    await seedAuth(page);
  });

  test('renders Standard VAT rate name and 15% after opening config panel', async ({ page }) => {
    await page.route('**/finance/tax/config', async (route) => {
      if (route.request().method() !== 'GET') {
        await route.continue();
        return;
      }
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify([{ id: 1, name: 'Standard VAT', rate: 15 }]),
      });
    });

    // Stub the tax report so the page mounts cleanly.
    await page.route('**/finance/tax/report', async (route) => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          vat_summary: { collected: 0, owed: 0, paid: 0 },
          periods: [],
          configs: [{ id: 1, name: 'Standard VAT', rate: 15 }],
        }),
      });
    });

    await page.goto(`${FINANCE_URL}/tax`);

    // Open the "Tax Rates" config panel.
    await page.getByRole('button', { name: /Tax Rates/i }).click();

    // The panel should now show the name and the rate value.
    await expect(page.getByDisplayValue('Standard VAT')).toBeVisible({ timeout: 15_000 });
    // Rate input should show 15.
    await expect(page.getByDisplayValue('15')).toBeVisible({ timeout: 10_000 });
  });

  test('Tax page heading is visible without config panel', async ({ page }) => {
    await page.route('**/finance/tax/config', async (route) => {
      if (route.request().method() !== 'GET') {
        await route.continue();
        return;
      }
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify([{ id: 1, name: 'Standard VAT', rate: 15 }]),
      });
    });

    await page.route('**/finance/tax/report', async (route) => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          vat_summary: { collected: 0, owed: 0, paid: 0 },
          periods: [],
          configs: [],
        }),
      });
    });

    await page.goto(`${FINANCE_URL}/tax`);

    await expect(page.getByRole('heading', { name: /^Tax$/i })).toBeVisible({ timeout: 15_000 });
  });
});
