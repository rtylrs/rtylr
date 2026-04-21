import { test, expect } from '@playwright/test';

/**
 * Auth login spec. Rate-limited backend (5/min/IP) plus strict-mode
 * role selectors so the assertions survive UI tweaks. API is stubbed
 * via `page.route()` — never hits api.voxire.dev.
 */

const AUTH_URL = process.env.AUTH_URL ?? 'http://localhost:5176';

async function stubAuth(page: import('@playwright/test').Page): Promise<void> {
  await page.route('**/auth/login', async (route) => {
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

  // Google auth stub — never hit Firebase/Google in CI.
  await page.route('**/auth/register/google', (route) =>
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

  // Register endpoint — keeps the /register spec path self-contained.
  await page.route('**/auth/register', (route) =>
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
    // Password is usually not a role=textbox in Playwright — select by input type.
    await expect(page.locator('input[type="password"]').first()).toBeVisible();
  });

  test('wrong password surfaces an error, right one navigates away', async ({ page }) => {
    await page.goto(AUTH_URL);

    await page.getByRole('textbox', { name: /email/i }).first().fill('user@example.com');
    await page.locator('input[type="password"]').first().fill('wrong');
    await page
      .getByRole('button', { name: /sign ?in|log ?in|continue/i })
      .first()
      .click();

    // Error message appears somewhere on the page — we don't pin
    // exact copy because the Auth UI rotates error text, but we do
    // insist that SOMETHING visible says "invalid" in some form.
    await expect(page.getByText(/invalid|incorrect|wrong/i)).toBeVisible({ timeout: 5_000 });

    // Retry with the right password; the mock returns 200 and the
    // app persists the token (cookie) before redirecting.
    await page.locator('input[type="password"]').first().fill('correct');
    await page
      .getByRole('button', { name: /sign ?in|log ?in|continue/i })
      .first()
      .click();

    // After a successful login we expect the page to either redirect
    // externally (window.location change) or show a logged-in state.
    // Assert on the cookie being set as the most stable signal.
    await page.waitForTimeout(500);
    const cookies = await page.context().cookies();
    const tokenCookie = cookies.find((c) => c.name === 'token');
    expect(tokenCookie?.value).toBe('fake-jwt');
  });
});
