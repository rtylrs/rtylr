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
    await context.route(/localhost:9999/, (route) => {
      const url = route.request().url();
      route.fulfill({
        status: 500,
        contentType: 'application/json',
        body: JSON.stringify({
          error: 'TEST_LEAK',
          message: `Unstubbed request reached production API: ${url}`,
        }),
      });
    });
    await use(context);
  },
});

export { expect };
