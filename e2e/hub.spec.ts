import { test, expect } from '@playwright/test';

const HUB_URL = process.env.HUB_URL ?? 'http://localhost:5175';
const AUTH_URL = process.env.AUTH_URL ?? 'http://localhost:5176';

test.describe('Hub', () => {
  test('redirects to auth when not logged in', async ({ page }) => {
    await page.goto(HUB_URL);
    // Should redirect to auth
    await page.waitForURL(/auth|login/, { timeout: 5000 });
  });

  test.describe('authenticated', () => {
    test.beforeEach(async ({ page }) => {
      // Set a test token cookie
      await page.goto(HUB_URL);
      await page.evaluate(() => {
        document.cookie = 'token=test-token; path=/;';
      });
    });

    test('shows loading state', async ({ page }) => {
      await page.goto(HUB_URL);
      // Should show loading spinner or dashboard
      const spinner = page.locator('[style*="spin"]');
      const dashboard = page.getByText(/dashboard|welcome/i);
      await expect(spinner.or(dashboard)).toBeVisible({ timeout: 5000 });
    });
  });
});
