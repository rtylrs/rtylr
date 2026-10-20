import { stubPortalSession } from './portal-session';
import { test, expect } from './fixtures';

/**
 * End-to-end: Finance - Bank Feeds and Reconcile when the bank account list
 * fails. Production answered it with a 500 for days, and both pages read as
 * empty instead of broken, so each one has to name the failure.
 */

const FINANCE_URL = process.env.FINANCE_URL ?? 'http://localhost:5197/finance';
const API_URL = 'http://localhost:9999';

test.describe('Finance - bank accounts fail to load', () => {
  test.beforeEach(async ({ context, page }) => {
    await context.addCookies([{ name: 'token', value: 'e2e-stub-token', url: FINANCE_URL }]);
    await page.route(`${API_URL}/finance/**`, (route) =>
      route.fulfill({ status: 200, contentType: 'application/json', body: '{}' })
    );
    await stubPortalSession(page, page.context(), { isOwner: true, appIds: ['finance'] });
    await page.route(`${API_URL}/hub/subscription/v1/my`, (route) =>
      route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({ subscription: null }),
      })
    );
    await page.route(`${API_URL}/finance/bank-account/v1/bank-accounts**`, (route) =>
      route.fulfill({
        status: 500,
        contentType: 'application/json',
        body: JSON.stringify({ message: 'Failed to list bank accounts' }),
      })
    );
  });

  test('Bank Feeds says the accounts did not load', async ({ page }) => {
    await page.goto(`${FINANCE_URL}/bank`);
    const alert = page.getByRole('alert');
    await expect(alert.getByRole('heading', { name: "Bank accounts didn't load" })).toBeVisible();
    await expect(alert.getByRole('button', { name: 'Refresh' })).toBeVisible();
    await expect(page.getByText(/connected account/)).toHaveCount(0);
  });

  test('Reconcile says what did not load instead of showing only its heading', async ({ page }) => {
    await page.goto(`${FINANCE_URL}/reconcile`);
    const alert = page.getByRole('alert');
    await expect(alert.getByRole('heading', { name: "Bank accounts didn't load" })).toBeVisible();
    await expect(alert).toContainText('nothing here can be matched yet');
    await expect(page.getByText('All transactions reconciled')).toHaveCount(0);
  });
});
