import { test, expect } from './fixtures';

/**
 * Landing-site e2e. Strict-mode selectors (role + name) so assertions
 * don't go stale when we add another "Pricing" or "Free" somewhere on
 * the page - which has happened before and burned this spec.
 */

const LANDING_URL = process.env.LANDING_URL ?? 'http://127.0.0.1:4200';

test.describe('Landing page', () => {
  test('homepage loads with h1 hero', async ({ page }) => {
    await page.goto(LANDING_URL);
    // MainHero renders an <h1>; if this breaks we've either dropped
    // the heading (SEO regression) or moved it out of the viewport.
    await expect(page.getByRole('heading', { level: 1 })).toBeVisible();
  });

  test('pricing page renders the plan heading', async ({ page }) => {
    await page.goto(`${LANDING_URL}/pricing`);
    // Heading is unique per page so it's a stable anchor - beats
    // scanning for "Free"/"Pro"/"Scale" which appear 11 times.
    await expect(page.getByRole('heading', { level: 1, name: /one rtylr suite/i })).toBeVisible();
  });

  test('pricing page states no price anywhere', async ({ page }) => {
    await page.goto(`${LANDING_URL}/pricing`);
    // rtylr has not published prices. Placeholder figures were live
    // here, in the plan cards on the home page, and asserted to Google
    // as an AggregateOffer. Nothing may reintroduce a currency amount
    // until there is a real one, so assert the absence rather than the
    // shape of what replaced it.
    const body = (await page.locator('body').innerText()).replace(/\s+/g, ' ');
    expect(body).not.toMatch(/[$€£]\s?\d/);
    expect(body).not.toMatch(/\d+\s?(?:USD|EUR)/i);
    expect(body).not.toMatch(/\/\s?seat\s?\/\s?month/i);
  });

  test('homepage has a canonical link to the root URL', async ({ page }) => {
    await page.goto(LANDING_URL);
    // Per-page canonicals landed in the SEO pass. The homepage may
    // emit either "https://rtylr.com" (global fallback) or
    // "https://rtylr.com/" (per-page). Either is valid; match both.
    const canonical = page.locator('link[rel="canonical"]').first();
    await expect(canonical).toHaveAttribute('href', /^https:\/\/rtylr\.com\/?$/);
  });

  /** Collect every @type on the page, flattening any @graph. */
  async function jsonLdTypes(page: import('@playwright/test').Page): Promise<string[]> {
    const blocks = await page.locator('script[type="application/ld+json"]').all();
    const types: string[] = [];
    for (const el of blocks) {
      const content = await el.textContent();
      if (!content) continue;
      const data = JSON.parse(content);
      for (const node of Array.isArray(data['@graph']) ? data['@graph'] : [data]) {
        if (node['@type']) types.push(node['@type']);
      }
    }
    return types;
  }

  test('homepage emits SoftwareApplication JSON-LD', async ({ page }) => {
    await page.goto(LANDING_URL);
    // This used to pass because _document put SoftwareApplication on
    // every page. It now has to come from the home page's own schema.
    const types = await jsonLdTypes(page);
    expect(types).toContain('SoftwareApplication');
    expect(types).toContain('WebSite');
    expect(types).toContain('Organization');
  });

  test('a legal page claims to be nothing but a page', async ({ page }) => {
    await page.goto(`${LANDING_URL}/legal/privacy-policy`);
    // The counterpart to the test above, and the reason it was moved:
    // the privacy policy spent months telling Google it was a business
    // application with a price of zero.
    const types = await jsonLdTypes(page);
    expect(types).toContain('WebPage');
    expect(types).not.toContain('SoftwareApplication');
    expect(types).not.toContain('Product');
  });

  test('no page asserts a price in its structured data', async ({ page }) => {
    for (const path of ['', '/pricing', '/workspace', '/legal/privacy-policy']) {
      await page.goto(`${LANDING_URL}${path}`);
      const blocks = await page.locator('script[type="application/ld+json"]').all();
      for (const el of blocks) {
        const content = (await el.textContent()) ?? '';
        expect(content, `price asserted on ${path || '/'}`).not.toMatch(
          /"(?:price|lowPrice|highPrice|priceCurrency)"/
        );
      }
    }
  });

  test('robots.txt advertises the sitemap + AI-bot allowances', async ({ request }) => {
    // Pure HTTP GET - no browser needed.
    const res = await request.get(`${LANDING_URL}/robots.txt`);
    expect(res.status()).toBe(200);
    const body = await res.text();
    expect(body).toContain('Sitemap:');
    // AI-bot allowances are a hard contract for this site;
    // accidentally blocking them would kill AI-SEO visibility.
    expect(body).toContain('GPTBot');
    expect(body).toContain('PerplexityBot');
  });
});
