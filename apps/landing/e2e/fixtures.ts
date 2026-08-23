import { test as base, expect } from '@playwright/test';
export type { BrowserContext, Page, Route } from '@playwright/test';

/**
 * Extended test fixture that keeps every test on the local origin.
 *
 * Two context-level routes:
 *
 *   1. **localhost:9999** returns a loud 500 with a `TEST_LEAK` body, so an
 *      unstubbed API call fails visibly in the trace instead of silently
 *      reaching a real service.
 *   2. **Every other origin** returns an empty 204.
 *
 * The second one was added after the Preferred Sources button landed. That
 * button loads a script from news.google.com which then injects two
 * cross-origin iframes, and `page.goto(..., waitUntil: 'load')` stopped
 * completing: the suite went from 13 seconds to 4 minutes and six tests
 * timed out at 60s. Nothing was wrong with the site. The tests were
 * waiting on Google.
 *
 * 204 rather than abort() on purpose: an aborted subresource can leave the
 * load event pending, which is the failure being fixed. An empty success
 * lets the document finish loading.
 *
 * This also makes the fixture do what its docstring always claimed. It
 * only ever blocked one port, so "no test can accidentally reach
 * production" was aspirational.
 *
 * Per-test `page.route()` stubs still win, because Playwright evaluates
 * page routes before context routes. A test that genuinely needs a real
 * third-party response has to say so explicitly, which is the right way
 * round.
 */
const LOCAL = /^https?:\/\/(127\.0\.0\.1|localhost|\[::1\])(:\d+)?(\/|$)/;

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

    await context.route('**/*', async (route) => {
      const url = route.request().url();
      if (LOCAL.test(url) || url.startsWith('data:') || url.startsWith('blob:')) {
        return route.fallback();
      }
      return route.fulfill({ status: 204, body: '' });
    });

    await use(context);
  },
});

export { expect };
