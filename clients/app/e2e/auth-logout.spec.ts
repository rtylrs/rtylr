import { test, expect } from './fixtures';
import { PORTAL_URL, PREVIEW_TOKEN, stubPortalSession } from './portal-session';

const AUTH_URL = process.env.AUTH_URL ?? 'http://localhost:5176';

// The shared cookie unit tests cover parent-domain deletion; localhost exercises the visible sign-out flow.
test.describe('Sign out clears the auth cookie', () => {
  test('a member can sign out from Portal without retaining the token', async ({
    page,
    context,
  }) => {
    await stubPortalSession(page, context);
    await page.goto(PORTAL_URL);
    const userMenu = page.getByRole('button', { name: /user menu/i });
    await expect(userMenu).toBeVisible();
    expect((await context.cookies()).find((cookie) => cookie.name === 'token')?.value).toBe(
      PREVIEW_TOKEN
    );
    await userMenu.click();
    await page.getByRole('button', { name: /sign out/i }).click();
    await page.waitForURL((url) => url.toString().startsWith(AUTH_URL));
    expect((await context.cookies()).find((cookie) => cookie.name === 'token')).toBeUndefined();
    expect(new URL(page.url()).searchParams.get('redirect')).toBe('dash');
  });

  test('a protected deep link without a token returns through SSO', async ({ page, context }) => {
    await context.clearCookies();
    await page.goto(`${PORTAL_URL}/insights/financial`);
    await page.waitForURL((url) => url.toString().startsWith(AUTH_URL));
    const destination = new URL(page.url());
    expect(destination.searchParams.get('redirect')).toBe('dash');
    expect(destination.searchParams.get('returnTo')).toBe('/insights/financial');
    expect((await context.cookies()).find((cookie) => cookie.name === 'token')).toBeUndefined();
  });
});
