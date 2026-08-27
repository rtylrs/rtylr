import { test, expect } from './fixtures';

/**
 * Logout regression spec. The bug from commit c81034fc shipped a
 * silent failure: clicking Sign out cleared a host-only cookie but
 * left the `.rtylr.com` parent-domain cookie alive, so auth.rtylr.com
 * saw the token on the next navigation and bounced the user right
 * back in. This spec asserts the user-visible behavior - cookie is
 * gone after click, the URL navigates away from dash, and reloading
 * the protected page does not auto-rebound into the app.
 *
 * On localhost we can't reproduce the Domain= attribute mismatch (all
 * cookies are host-only), so the unit test in
 * `packages/auth/src/cookies.test.ts` covers the static contract that
 * removeCookie emits the Domain attr. This file covers the dynamic
 * "the user can sign out" behavior end-to-end.
 *
 * Test target: dash StaffView. Why staff and not owner? StaffView's
 * sign-out button is the one fixed in commit c81034fc and it's the
 * simplest path through dash that doesn't require menu/QR/widget data.
 */

const DASH_URL = process.env.DASH_URL ?? 'http://localhost:5175';
const AUTH_URL = process.env.AUTH_URL ?? 'http://localhost:5176';

interface StubUserOptions {
  isOwner?: boolean;
  apps?: string[];
}

async function stubBackend(
  page: import('@playwright/test').Page,
  options: StubUserOptions = {}
): Promise<void> {
  await page.route('**/hub/user/v1/users/me', (route) =>
    route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({
        user_id: 1,
        business_id: '1',
        email: 'staff@example.com',
        full_name: 'Staff User',
        business_name: 'Example Co',
        is_owner: options.isOwner ?? false,
        apps: options.apps ?? [],
        role: 'staff',
        onboarded: true,
      }),
    })
  );

  // BusinessRoute also pulls business config; stub a minimal payload
  // so the dash shell doesn't error before StaffView mounts.
  await page.route('**/get-configs', (route) =>
    route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({
        config: {
          business_name: 'Example Co',
          maintanance: false,
          theme: { text: '#111', bg: '#fff', border: '#e5e5e5' },
        },
      }),
    })
  );
}

test.describe('Sign out clears the auth cookie', () => {
  test('staff user can sign out from dash StaffView and the token cookie is gone', async ({
    page,
    context,
  }) => {
    await stubBackend(page, { isOwner: false, apps: [] });

    // Pre-set the token cookie the same way the auth app would after
    // a successful login. Localhost cookies are host-only by default;
    // that's fine - the regression we're catching is "click does
    // nothing", not the Domain= attribute drift (which is unit-tested).
    await context.addCookies([
      {
        name: 'token',
        value: 'fake-jwt',
        url: DASH_URL,
        httpOnly: false,
        secure: false,
        sameSite: 'Lax',
      },
      // Strict-opt-in cookie banner from @voxire/consent reads this
      // cookie to decide whether to render. Setting it to "rejected"
      // skips the overlay so the test isn't blocked by it.
      {
        name: 'rtylr-consent',
        value: 'rejected',
        url: DASH_URL,
        httpOnly: false,
        secure: false,
        sameSite: 'Lax',
      },
    ]);

    await page.goto(DASH_URL);

    // StaffView renders a "User menu" button. Wait for it.
    const userMenu = page.getByRole('button', { name: /user menu/i });
    await expect(userMenu).toBeVisible({ timeout: 15_000 });

    // Token cookie must be present right before logout.
    const before = await context.cookies();
    expect(before.find((c) => c.name === 'token')?.value).toBe('fake-jwt');

    await userMenu.click();

    const signOut = page.getByRole('button', { name: /sign out/i });
    await expect(signOut).toBeVisible();
    await signOut.click();

    // The handler navigates via window.location.href to AUTH_URL.
    // Wait for the navigation rather than racing it - dash redirects
    // to auth.rtylr.com (or AUTH_URL in dev), so URL pattern is
    // "anywhere except dash".
    await page.waitForURL((url) => !url.toString().startsWith(DASH_URL), { timeout: 15_000 });

    // Cookie must be cleared.
    const after = await context.cookies();
    expect(after.find((c) => c.name === 'token')).toBeUndefined();

    // The redirect target is auth, with redirect=dash query string.
    const finalUrl = page.url();
    expect(finalUrl).toContain(AUTH_URL);
    expect(finalUrl).toContain('redirect=dash');
  });

  test('reloading dash without the token cookie redirects to auth, no auto-login', async ({
    page,
    context,
  }) => {
    await stubBackend(page);

    // Start without a cookie - this simulates a user landing on dash
    // post-logout. The expected behavior is a clean redirect, not a
    // momentary flash of the dashboard.
    await context.clearCookies();

    await page.goto(DASH_URL);

    // BusinessRoute redirects via window.location.href when token
    // cookie is missing.
    await page.waitForURL((url) => url.toString().includes(AUTH_URL), { timeout: 15_000 });
    expect(page.url()).toContain('redirect=dash');

    const cookies = await context.cookies();
    expect(cookies.find((c) => c.name === 'token')).toBeUndefined();
  });
});
