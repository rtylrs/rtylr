import { test, expect } from './fixtures';

const GA_ID = 'G-6XKFL1Z8HE';
const LANDING_URL = process.env.LANDING_URL ?? 'http://127.0.0.1:4200';

test('no Google requests before acceptance; withdrawal disables collection', async ({ page }) => {
  const requests: string[] = [];
  page.on('request', (request) => {
    if (/googletagmanager|google-analytics|news\.google/.test(request.url()))
      requests.push(request.url());
  });
  await page.goto(LANDING_URL);
  await expect(page.getByRole('button', { name: /^accept$/i })).toBeVisible();
  expect(requests).toEqual([]);
  await page.getByRole('button', { name: /^decline$/i }).click();
  expect(requests).toEqual([]);
  await page.getByRole('link', { name: /cookie preferences/i }).click();
  await page.getByRole('button', { name: /^accept$/i }).click();
  await expect
    .poll(() => requests.filter((url) => url.includes('googletagmanager')).length)
    .toBe(1);
  await page.getByRole('link', { name: /cookie preferences/i }).click();
  expect(
    await page.evaluate(
      (id) => (window as unknown as Record<string, unknown>)[`ga-disable-${id}`],
      GA_ID
    )
  ).toBe(true);
  await expect(page.getByRole('button', { name: /^accept$/i })).toBeVisible();
});

for (const state of ['denied', 'granted']) {
  test(`returning ${state} choice controls loading on legal pages`, async ({ page, context }) => {
    await context.addCookies([{ name: 'rtylr-consent', value: state, url: LANDING_URL }]);
    await page.goto(`${LANDING_URL}/legal/privacy-policy/`);
    await expect(page.locator('h1')).toHaveText('Privacy Policy');
    await expect(page.locator('script[src*="googletagmanager"]')).toHaveCount(
      state === 'granted' ? 1 : 0
    );
    await expect(page.getByRole('button', { name: /^accept$/i })).toHaveCount(0);
  });
}
