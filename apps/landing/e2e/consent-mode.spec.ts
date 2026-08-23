import { test, expect } from './fixtures';

/**
 * GA4 and Google Consent Mode, checked as behaviour in a real browser.
 *
 * The reason this file exists rather than a "the tag is in the head"
 * assertion: the tag WAS in the head, on production, for as long as GA4
 * has been installed, while the site's own consent banner sat beside it
 * being ignored. Presence proved nothing. What matters is the order and
 * content of the commands in `window.dataLayer`, and whether accepting
 * the banner changes them.
 *
 * gtag.js itself is blocked here. The library is not under test and
 * letting it through would send hits to the live property from CI.
 * dataLayer is a plain array that the page fills whether or not the
 * library ever loads, so blocking it changes nothing being asserted.
 */

const LANDING_URL = process.env.LANDING_URL ?? 'http://127.0.0.1:4200';
const GA_ID = 'G-6XKFL1Z8HE';

type Command = [string, ...unknown[]];

/** Read dataLayer as plain arrays. Entries are `arguments` objects. */
async function dataLayer(page: import('@playwright/test').Page): Promise<Command[]> {
  return page.evaluate(
    () =>
      ((window as unknown as { dataLayer?: IArguments[] }).dataLayer ?? []).map(
        (args) => Array.from(args) as Command
      ) as Command[]
  );
}

const find = (calls: Command[], kind: string, subtype?: string): Command | undefined =>
  calls.find((c) => c[0] === kind && (subtype === undefined || c[1] === subtype));

/**
 * The last matching command, which is the one in force. Consent updates
 * accumulate in dataLayer, so reading the first one tells you what was
 * true at page load and nothing about what the visitor just clicked.
 */
const findLast = (calls: Command[], kind: string, subtype?: string): Command | undefined =>
  calls.filter((c) => c[0] === kind && (subtype === undefined || c[1] === subtype)).at(-1);

const indexOf = (calls: Command[], kind: string, subtype?: string): number =>
  calls.findIndex((c) => c[0] === kind && (subtype === undefined || c[1] === subtype));

test.beforeEach(async ({ page }) => {
  await page.route(/googletagmanager\.com/, (route) => route.fulfill({ status: 204, body: '' }));
});

test.describe('Google Consent Mode', () => {
  test('a first-time visitor is measured with storage denied', async ({ page }) => {
    await page.goto(LANDING_URL);

    const calls = await dataLayer(page);
    const dflt = find(calls, 'consent', 'default');

    expect(dflt, 'no consent default was queued: GA4 is running unconsented').toBeDefined();
    expect(dflt?.[2]).toMatchObject({
      ad_storage: 'denied',
      ad_user_data: 'denied',
      ad_personalization: 'denied',
      analytics_storage: 'denied',
    });

    // Google ignores a default that arrives after config, silently.
    expect(indexOf(calls, 'consent', 'default')).toBeLessThan(indexOf(calls, 'config'));
    expect(find(calls, 'config')?.[1]).toBe(GA_ID);

    // And nothing may push an update yet, not even a denial. The
    // bootstrap's wait_for_update is holding the opening hits so that a
    // visitor who accepts a moment from now is measured properly, and
    // any update ends that wait.
    expect(find(calls, 'consent', 'update')).toBeUndefined();
  });

  test('accepting the banner grants analytics storage without a reload', async ({ page }) => {
    await page.goto(LANDING_URL);

    await page.getByRole('button', { name: /^accept$/i }).click();

    // The grant is pushed from a React handler, so give the effect a tick.
    await expect
      .poll(async () => findLast(await dataLayer(page), 'consent', 'update')?.[2])
      .toMatchObject({ analytics_storage: 'granted' });
  });

  test('declining the banner leaves analytics storage denied', async ({ page }) => {
    await page.goto(LANDING_URL);

    await page.getByRole('button', { name: /^decline$/i }).click();

    await expect
      .poll(async () => findLast(await dataLayer(page), 'consent', 'update')?.[2] ?? null)
      .toMatchObject({ analytics_storage: 'denied' });
  });

  test('a returning visitor who accepted is granted before the first hit', async ({
    page,
    context,
  }) => {
    // This is the case the inline head script exists for. Without it the
    // grant waits for React and every returning visitor's first pageview
    // is either delayed or sent cookieless.
    await context.addCookies([
      { name: 'rtylr-consent', value: 'granted', url: LANDING_URL },
    ]);
    await page.goto(LANDING_URL);

    const calls = await dataLayer(page);
    const updateIdx = indexOf(calls, 'consent', 'update');

    expect(updateIdx, 'returning consented visitor was not granted').toBeGreaterThan(-1);
    expect(calls[updateIdx][2]).toMatchObject({ analytics_storage: 'granted' });
    expect(updateIdx).toBeLessThan(indexOf(calls, 'config'));

    // And the banner should not be asking again.
    await expect(page.getByRole('button', { name: /^accept$/i })).toHaveCount(0);
  });

  test('a returning visitor who declined is not granted', async ({ page, context }) => {
    await context.addCookies([{ name: 'rtylr-consent', value: 'denied', url: LANDING_URL }]);
    await page.goto(LANDING_URL);

    const calls = await dataLayer(page);
    for (const call of calls.filter((c) => c[0] === 'consent' && c[1] === 'update')) {
      expect(call[2]).toMatchObject({ analytics_storage: 'denied' });
    }
  });

  test('the consent default is on every page, not just the home page', async ({ page }) => {
    // It lives in _document, so it should be. Asserted because the
    // per-page schema had exactly the opposite bug in reverse.
    for (const path of ['/pricing', '/pos', '/legal/privacy-policy', '/blog']) {
      await page.goto(`${LANDING_URL}${path}`);
      const calls = await dataLayer(page);
      expect(find(calls, 'consent', 'default'), `no consent default on ${path}`).toBeDefined();
    }
  });
});
