import { test, expect } from './fixtures';

/**
 * Landing-site e2e. Strict-mode selectors (role + name) so assertions
 * don't go stale when we add another "Pricing" or "Free" somewhere on
 * the page - which has happened before and burned this spec.
 */

const LANDING_URL = process.env.LANDING_URL ?? 'http://127.0.0.1:4200';

test.describe('Landing page', () => {
  test.beforeEach(async ({ page }) => {
    await page.route('https://api.voxire.com/features', (route) =>
      route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          variant: 'a',
          features: {
            is_desktop_mode_default: {
              key: 'is_desktop_mode_default',
              enabled: false,
              variant: 'false',
            },
          },
        }),
      })
    );
  });

  test('homepage loads with h1 hero', async ({ page }) => {
    await page.goto(LANDING_URL);
    // MainHero renders an <h1>; if this breaks we've either dropped
    // the heading (SEO regression) or moved it out of the viewport.
    await expect(page.getByRole('heading', { level: 1 })).toBeVisible();
  });

  test('homepage defaults to website mode with desktop opt-in', async ({ page }) => {
    await page.goto(LANDING_URL);
    await expect(page.locator('.rtylr-desktop-shell')).toHaveCount(0);
    await expect(page.getByRole('link', { name: /^Pricing$/i }).first()).toBeVisible();

    await page.getByRole('button', { name: /switch to os mode/i }).click();
    await expect(page.locator('.rtylr-desktop-shell')).toBeVisible();
    await expect(page.getByRole('button', { name: /^Pricing$/i }).first()).toBeVisible();
  });

  test('pricing page renders the plan heading', async ({ page }) => {
    await page.goto(`${LANDING_URL}/pricing`);
    // Heading is unique per page so it's a stable anchor - beats
    // scanning for "Free"/"Pro"/"Scale" which appear 11 times.
    await expect(page.getByRole('heading', { name: /per seat, per day/i })).toBeVisible();
  });

  test('homepage has a canonical link to the root URL', async ({ page }) => {
    await page.goto(LANDING_URL);
    // Per-page canonicals landed in the SEO pass. The homepage may
    // emit either "https://rtylr.com" (global fallback) or
    // "https://rtylr.com/" (per-page). Either is valid; match both.
    const canonical = page.locator('link[rel="canonical"]').first();
    await expect(canonical).toHaveAttribute('href', /^https:\/\/rtylr\.com\/?$/);
  });

  test('homepage emits SoftwareApplication JSON-LD', async ({ page }) => {
    await page.goto(LANDING_URL);
    const allJsonLd = await page.locator('script[type="application/ld+json"]').all();
    expect(allJsonLd.length).toBeGreaterThan(0);
    const types: string[] = [];
    for (const el of allJsonLd) {
      const content = await el.textContent();
      if (!content) continue;
      const data = JSON.parse(content);
      if (data['@type']) types.push(data['@type']);
      if (Array.isArray(data['@graph'])) {
        for (const node of data['@graph']) {
          if (node['@type']) types.push(node['@type']);
        }
      }
    }
    expect(types).toContain('SoftwareApplication');
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
