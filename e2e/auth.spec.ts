import { test, expect } from '@playwright/test';

const AUTH_URL = process.env.AUTH_URL ?? 'http://localhost:5176';

test.describe('Auth flows', () => {
  test('login page loads', async ({ page }) => {
    await page.goto(AUTH_URL);
    await expect(page.getByRole('heading', { level: 1 })).toBeVisible();
    await expect(page.getByPlaceholder(/email/i)).toBeVisible();
    await expect(page.getByPlaceholder(/password/i)).toBeVisible();
  });

  test('shows error for invalid credentials', async ({ page }) => {
    await page.goto(AUTH_URL);
    await page.getByPlaceholder(/email/i).fill('bad@example.com');
    await page.getByPlaceholder(/password/i).fill('wrongpass');
    await page.getByRole('button', { name: /sign in/i }).click();

    // Should show an error message
    await expect(page.getByText(/invalid|failed|error|incorrect/i)).toBeVisible({
      timeout: 5000,
    });
  });

  test('navigate to register page', async ({ page }) => {
    await page.goto(AUTH_URL);
    await page.getByRole('button', { name: /sign up/i }).click();
    await expect(page).toHaveURL(/register/);
    await expect(page.getByPlaceholder(/name/i)).toBeVisible();
  });

  test('navigate to reset password', async ({ page }) => {
    await page.goto(AUTH_URL);
    await page.getByText(/forgot/i).click();
    await expect(page).toHaveURL(/password-reset/);
  });

  test('register form has required fields', async ({ page }) => {
    await page.goto(`${AUTH_URL}/register`);
    await expect(page.getByPlaceholder(/name/i)).toBeVisible();
    await expect(page.getByPlaceholder(/email/i)).toBeVisible();
    await expect(page.getByPlaceholder(/password/i)).toBeVisible();
  });
});
