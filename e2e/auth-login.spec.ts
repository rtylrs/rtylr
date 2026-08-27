import { test, expect } from './fixtures';

/**
 * Auth login spec. Rate-limited backend (5/min/IP) plus strict-mode
 * role selectors so the assertions survive UI tweaks. API is stubbed
 * via `page.route()` - never hits localhost:9999.
 */

const AUTH_URL = process.env.AUTH_URL ?? 'http://localhost:5176';

async function stubAuth(page: import('@playwright/test').Page): Promise<void> {
  await page.route('**/common/auth/v1/login', async (route) => {
    const body = route.request().postDataJSON() as { email?: string; password?: string };
    if (body.password === 'wrong') {
      await route.fulfill({
        status: 401,
        contentType: 'application/json',
        body: JSON.stringify({ error: { message: 'Invalid email or password' } }),
      });
      return;
    }
    await route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({
        token: 'fake-jwt',
        message: 'Login successful',
        user_id: 1,
        email: body.email ?? '',
        name: 'Test User',
      }),
    });
  });

  // Google auth stub - never hit Firebase/Google in CI.
  await page.route('**/common/auth/v1/register/google', (route) =>
    route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({
        token: 'fake-jwt',
        message: 'Google authentication successful',
        user_id: 1,
        email: 'google@example.com',
        name: 'Google User',
      }),
    })
  );

  // Register endpoint - keeps the /register spec path self-contained.
  await page.route('**/common/auth/v1/register', (route) =>
    route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({
        token: 'fake-jwt',
        message: 'Registration successful',
        user_id: 1,
        email: 'new@example.com',
        name: 'New User',
      }),
    })
  );
}

test.describe('Auth login', () => {
  test.beforeEach(async ({ page }) => {
    await stubAuth(page);
  });

  test('login form renders with email + password fields', async ({ page }) => {
    await page.goto(AUTH_URL);
    await expect(page.getByRole('textbox', { name: /email/i }).first()).toBeVisible();
    // Password is usually not a role=textbox in Playwright - select by input type.
    await expect(page.locator('input[type="password"]').first()).toBeVisible();
  });

  test('login form can submit and sets the token cookie', async ({ page }) => {
    await page.goto(AUTH_URL);

    // The login screen is behind a Suspense + plan-aware wrapper,
    // so wait for the email input to mount rather than racing it.
    const emailInput = page.getByRole('textbox', { name: /email/i }).first();
    await expect(emailInput).toBeVisible({ timeout: 10_000 });

    await emailInput.fill('user@example.com');
    await page.locator('input[type="password"]').first().fill('correct');
    await page
      .getByRole('button', { name: /sign ?in|log ?in|continue/i })
      .first()
      .click();

    // Cookie setting is the stable signal - visible UI state varies.
    await page.waitForTimeout(500);
    const cookies = await page.context().cookies();
    const tokenCookie = cookies.find((c) => c.name === 'token');
    expect(tokenCookie?.value).toBe('fake-jwt');
  });

  test('register page loads', async ({ page }) => {
    await page.goto(`${AUTH_URL}/register`);
    // The register flow is a multi-step wizard (plan picker first,
    // then email). We don't pin to a specific step - just that the
    // page mounts past the SPA shell. A heading is the stable anchor.
    await expect(page.getByRole('heading').first()).toBeVisible({ timeout: 10_000 });
  });

  test('reset-password page loads with email field', async ({ page }) => {
    await page.goto(`${AUTH_URL}/password-reset`);
    await expect(page.getByRole('textbox', { name: /email/i }).first()).toBeVisible({
      timeout: 10_000,
    });
    // The "Send reset link" button is the action surface.
    await expect(
      page.getByRole('button', { name: /send reset|reset link/i }).first()
    ).toBeVisible();
  });
});
