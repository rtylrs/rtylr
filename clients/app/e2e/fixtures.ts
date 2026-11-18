import { test as base, expect } from '@playwright/test';
export type { BrowserContext, Page, Route } from '@playwright/test';

/**
 * Extended test fixture that blocks all requests to localhost:9999 at the
 * context level. Per-test page.route() stubs are registered on the page,
 * which Playwright evaluates before context routes, so they still win.
 *
 * This guarantees no test can accidentally reach production: any unstubbed
 * endpoint returns 500 with a clear "TEST_LEAK" body that shows up in the
 * Playwright trace and fails the test loudly.
 */
export const test = base.extend<object, object>({
  context: async ({ context }, use) => {
    await context.route(
      /localhost:9999|https?:\/\/(?:api\.(?:rtylr|voxire)\.com|[^/]*(?:posthog|sentry)[^/]*)\//,
      (route) => {
        const url = route.request().url();
        route.fulfill({
          status: 500,
          contentType: 'application/json',
          body: JSON.stringify({
            error: 'TEST_LEAK',
            message: `Unstubbed request reached production API: ${url}`,
          }),
        });
      }
    );
    // The one-time avatar welcome opens over whatever page a test is on.
    // Specs that stub /users/me never have an avatar, so mark it as seen for
    // every user and let the tests see the page they are asserting on.
    await context.addInitScript(() => {
      const getItem = Storage.prototype.getItem;
      Storage.prototype.getItem = function (key: string): string | null {
        if (key.startsWith('rtylr:avatar-welcome-seen:')) return '1';
        return getItem.call(this, key);
      };
    });
    await use(context);
  },
});

export { expect };
