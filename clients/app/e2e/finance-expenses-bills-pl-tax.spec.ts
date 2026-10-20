import { stubPortalSession } from './portal-session';
import { test, expect } from './fixtures';

/**
 * End-to-end: Finance - Expenses, Bills, Profit & Loss, Tax screens.
 *
 * Every `/finance/*` call is stubbed via `page.route()` so the spec runs
 * without a backend. The AppGuard is satisfied the same way as the invoice
 * spec: a `token` cookie plus a stubbed `/hub/user/v1/users/me` response.
 */

const FINANCE_URL = process.env.FINANCE_URL ?? 'http://localhost:5197/finance';
const API_URL = 'http://localhost:9999';

/** A day in the month the browser is standing in; the expenses page opens on it. */
function thisMonthDay(day: number): string {
  const now = new Date();
  const month = String(now.getMonth() + 1).padStart(2, '0');
  return `${now.getFullYear()}-${month}-${String(day).padStart(2, '0')}`;
}

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
  await stubPortalSession(page, page.context(), { isOwner: true, appIds: ['finance'] });

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
              expense_date: thisMonthDay(1),
              description: 'Office Supplies Purchase',
              category_id: '1',
              amount: 250,
              currency: 'USD',
              receipt_url: '',
              status: 'pending',
            },
            {
              id: 'exp-1002',
              expense_date: thisMonthDay(2),
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

    // Amounts are plain currency with cents; an expense list carries no minus sign.
    await expect(page.getByText('$250.00').first()).toBeVisible({ timeout: 10_000 });
    await expect(page.getByText('$1,200.00').first()).toBeVisible({ timeout: 10_000 });
    await expect(page.getByText(/-\$/)).toHaveCount(0);
    // Each row says where it stands in review.
    // The stub session can approve, so a pending row offers the review buttons.
    await expect(page.getByRole('row', { name: /Office Supplies Purchase/ })).toContainText(
      'Approve'
    );
    await expect(page.getByRole('row', { name: /Cloud Hosting Invoice/ })).toContainText(
      'Approved'
    );
  });

  // Regression: the contract required a vendor and a receipt URL, which this
  // form never collects, so the generated client refused the request before
  // sending it and Submit appeared to do nothing.
  test('submits an expense with no vendor, notes or receipt', async ({ page }) => {
    await page.route(`${API_URL}/finance/expenses/v1/categories`, (route) =>
      route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          categories: [
            { id: '11', name: 'Communications', color: '#14b8a6' },
            { id: '12', name: 'Communications', color: '#14b8a6' },
          ],
        }),
      })
    );
    let posted: Record<string, unknown> | null = null;
    await page.route(`${API_URL}/finance/expenses/v1/expenses`, async (route) => {
      if (route.request().method() === 'POST') {
        posted = route.request().postDataJSON() as Record<string, unknown>;
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({ expense_id: '501', message: 'Expense created' }),
        });
        return;
      }
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({ expenses: [] }),
      });
    });

    await page.goto(`${FINANCE_URL}/expenses/new`);
    await page.getByPlaceholder('Client dinner, Uber ride...').fill('Phone number');
    await page.getByPlaceholder('0.00').fill('35.00');
    await page.locator('input[type="date"]').fill('2026-10-19');
    await page.getByLabel('Category').selectOption('11');
    await page.getByRole('button', { name: 'Submit' }).click();

    await expect(page).toHaveURL(/\/finance\/expenses$/, { timeout: 15_000 });
    expect(posted).toMatchObject({
      category_id: '11',
      amount: 35,
      description: 'Phone number',
      expense_date: '2026-10-19',
    });
  });

  test('says why when the server refuses a new expense', async ({ page }) => {
    await page.route(`${API_URL}/finance/expenses/v1/categories`, (route) =>
      route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({ categories: [{ id: '11', name: 'Communications' }] }),
      })
    );
    await page.route(`${API_URL}/finance/expenses/v1/expenses`, (route) =>
      route.fulfill({
        status: 400,
        contentType: 'application/json',
        body: JSON.stringify({ message: 'Expense category not found' }),
      })
    );

    await page.goto(`${FINANCE_URL}/expenses/new`);
    await page.getByPlaceholder('Client dinner, Uber ride...').fill('Phone number');
    await page.getByPlaceholder('0.00').fill('35.00');
    await page.getByLabel('Category').selectOption('11');
    await page.getByRole('button', { name: 'Submit' }).click();

    await expect(page.getByText('Expense category not found')).toBeVisible({ timeout: 10_000 });
    await expect(page.locator('[data-sonner-toast]')).toHaveCount(1);
    await expect(page).toHaveURL(/\/finance\/expenses\/new$/);
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
              expense_date: thisMonthDay(1),
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
    // A receipt shows as a paperclip beside the description.
    await expect(page.locator('[aria-label="Receipt attached"]')).toHaveCount(1);
  });
});

// ── Bills list ────────────────────────────────────────────────────────────

test.describe('Finance - Bills list', () => {
  test.beforeEach(async ({ context, page }) => {
    await context.addCookies([{ name: 'token', value: 'e2e-stub-token', url: FINANCE_URL }]);
    await seedAuth(page);
  });

  test('renders a draft bill with vendor name', async ({ page }) => {
    await page.route(`${API_URL}/finance/bill/v1/bills**`, async (route) => {
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
    await page.route(`${API_URL}/finance/bill/v1/bills**`, async (route) => {
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
          pending_expenses: 35,
          pending_expense_count: '1',
          expenses_by_category: [
            { category_id: '11', name: 'Communications', color: '#14b8a6', amount: 15000 },
          ],
        }),
      });
    });

    await page.goto(`${FINANCE_URL}/pl`);

    // Pending expenses are named rather than silently left out.
    await expect(page.getByRole('status')).toContainText(
      '1 expense ($35.00) in this period is waiting for approval',
      { timeout: 15_000 }
    );
    await expect(page.getByText('Communications')).toBeVisible();

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

    await expect(page.getByRole('tab', { name: 'This month' })).toBeVisible({
      timeout: 15_000,
    });
    await expect(page.getByRole('tab', { name: 'This year' })).toBeVisible({ timeout: 10_000 });
    // An empty period reads $0.00, never "-$0", and the dead fiscal-year button is gone.
    await expect(page.getByText(/-\$0/)).toHaveCount(0);
    await expect(page.getByRole('button', { name: /Close Fiscal Year/i })).toHaveCount(0);
  });
});

// ── Tax rates screen ──────────────────────────────────────────────────────

test.describe('Finance - Tax rates screen', () => {
  test.beforeEach(async ({ context, page }) => {
    await context.addCookies([{ name: 'token', value: 'e2e-stub-token', url: FINANCE_URL }]);
    await seedAuth(page);
  });

  test('lists the Standard VAT rate at 15%', async ({ page }) => {
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

    const row = page.getByRole('row', { name: /Standard VAT/ });
    await expect(row).toBeVisible({ timeout: 15_000 });
    await expect(row).toContainText('15%');
  });

  test('Tax rates page is titled for what it does', async ({ page }) => {
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

    await expect(page.getByRole('heading', { name: 'Tax rates', exact: true })).toBeVisible({
      timeout: 15_000,
    });
    await expect(page.getByText('No tax rates configured yet')).toBeVisible();
  });
});

// ── Dashboard ─────────────────────────────────────────────────────────────

test.describe('Finance - Dashboard', () => {
  test.beforeEach(async ({ context, page }) => {
    await context.addCookies([{ name: 'token', value: 'e2e-stub-token', url: FINANCE_URL }]);
    await seedAuth(page);
  });

  test('names pending expenses and explains its empty sections', async ({ page }) => {
    await page.route(`${API_URL}/finance/dashboard/v1/dashboard`, (route) =>
      route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({ pending_expenses: 35, pending_expense_count: '1' }),
      })
    );

    await page.goto(FINANCE_URL);

    await expect(page.getByRole('status')).toContainText(
      '1 expense ($35.00) is waiting for approval',
      { timeout: 15_000 }
    );
    await expect(page.getByText(/Nothing to chart yet/)).toBeVisible();
    await expect(page.getByText(/No activity yet/)).toBeVisible();
    await expect(page.getByText(/vs last month/)).toHaveCount(0);
  });
});
