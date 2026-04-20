import { test, expect } from '@playwright/test';

/**
 * Landing-site e2e. Strict-mode selectors (role + name) so assertions
 * don't go stale when we add another "Pricing" or "Free" somewhere on
 * the page — which has happened before and burned this spec.
 */

const LANDING_URL = process.env.LANDING_URL ?? 'http://localhost:4200';

test.describe('Landing page', () => {
  test('homepage loads with h1 hero', async ({ page }) => {
    await page.goto(LANDING_URL);
    // MainHero renders an <h1>; if this breaks we've either dropped
    // the heading (SEO regression) or moved it out of the viewport.
    await expect(page.getByRole('heading', { level: 1 })).toBeVisible();
  });

  test('header has the Pricing link', async ({ page }) => {
    await page.goto(LANDING_URL);
    // `.first()` avoids the strict-mode violation when "Pricing"
    // appears in the header, pricing section, and footer.
    await expect(page.getByRole('link', { name: /^Pricing$/ }).first()).toBeVisible();
  });

  test('pricing page renders the plan heading', async ({ page }) => {
    await page.goto(`${LANDING_URL}/pricing`);
    // Heading is unique per page so it's a stable anchor — beats
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
    const jsonLd = page.locator('script[type="application/ld+json"]').first();
    await expect(jsonLd).toBeAttached();
    const content = await jsonLd.textContent();
    const data = JSON.parse(content!);
    expect(data['@type']).toBe('SoftwareApplication');
    expect(data.name).toBe('rtylr');
  });

  test('robots.txt advertises the sitemap + AI-bot allowances', async ({ request }) => {
    // Pure HTTP GET — no browser needed.
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
