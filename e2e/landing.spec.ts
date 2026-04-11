import { test, expect } from '@playwright/test';

const LANDING_URL = process.env.LANDING_URL ?? 'http://localhost:4200';

test.describe('Landing page', () => {
  test('homepage loads with hero', async ({ page }) => {
    await page.goto(LANDING_URL);
    await expect(page.getByRole('heading', { level: 1 })).toBeVisible();
  });

  test('has navigation with Features dropdown', async ({ page }) => {
    await page.goto(LANDING_URL);
    await expect(page.getByText('Features')).toBeVisible();
    await expect(page.getByText('Pricing')).toBeVisible();
  });

  test('pricing page loads', async ({ page }) => {
    await page.goto(`${LANDING_URL}/pricing`);
    await expect(page.getByText(/Free|Pro|Scale/)).toBeVisible();
  });

  test('has canonical tag', async ({ page }) => {
    await page.goto(LANDING_URL);
    const canonical = page.locator('link[rel="canonical"]');
    await expect(canonical).toHaveAttribute('href', 'https://rtylr.com');
  });

  test('has structured data', async ({ page }) => {
    await page.goto(LANDING_URL);
    const jsonLd = page.locator('script[type="application/ld+json"]');
    await expect(jsonLd).toBeAttached();
    const content = await jsonLd.textContent();
    const data = JSON.parse(content!);
    expect(data['@type']).toBe('SoftwareApplication');
    expect(data.name).toBe('rtylr');
  });
});
