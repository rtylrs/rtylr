import { test, expect, type Page } from './fixtures';

/**
 * Regional prices on /pricing/. The page is static and carries no price; a
 * visitor whose Cloudflare country has an admin-set regional price sees it
 * after load. Both lookups are stubbed: the country (/cdn-cgi/trace, served
 * by Cloudflare, not by the local static server) and onion's plans.
 */

const LANDING_URL = process.env.LANDING_URL ?? 'http://127.0.0.1:4200';

const CAD_PLANS = {
  plans: [
    { id: '1', slug: 'free', name: 'Free', currency: 'USD' },
    {
      id: '2',
      slug: 'pro',
      name: 'Pro',
      price_per_seat_cents: 899,
      annual_price_per_seat_cents: 649,
      currency: 'CAD',
      region: 'CA',
    },
    // A plan with no override comes back at its default USD price and no
    // region. It must stay a no-price card.
    { id: '3', slug: 'scale', name: 'Scale', price_per_seat_cents: 499, currency: 'USD' },
  ],
};

async function stubGeo(page: Page, country: string, plans: unknown = CAD_PLANS): Promise<string[]> {
  const asked: string[] = [];
  await page.route('**/cdn-cgi/trace', (route) =>
    route.fulfill({ contentType: 'text/plain', body: `fl=1\ncolo=YUL\nloc=${country}\n` })
  );
  await page.route('**/hub/subscription/v1/plans*', (route) => {
    asked.push(new URL(route.request().url()).searchParams.get('region') ?? '');
    return route.fulfill({
      contentType: 'application/json',
      headers: { 'access-control-allow-origin': '*' },
      body: JSON.stringify(plans),
    });
  });
  return asked;
}

test.describe('Regional pricing', () => {
  test('a Canadian visitor sees the Canadian price, and plans without one stay quoted', async ({ page }) => {
    const asked = await stubGeo(page, 'CA');
    await page.goto(`${LANDING_URL}/pricing`);

    const pro = page.getByTestId('price-pro');
    await expect(pro).toBeVisible();
    await expect(pro).toContainText('$8.99');
    await expect(pro).toContainText(/per seat \/ month/i);
    await expect(pro).toContainText('$6.49');
    await expect(pro).toContainText(/billed annually/i);

    // Every card carries the block, so the lists line up. Free is free, and
    // a plan with no regional price is still quoted, not priced at its USD
    // default.
    await expect(page.getByTestId('price-free')).toContainText('Free');
    const scale = page.getByTestId('price-scale');
    await expect(scale).toContainText(/quoted/i);
    await expect(scale).not.toContainText(/\d/);
    expect(asked).toEqual(['CA']);
    await expect(page.getByTestId('regional-pricing-note')).toContainText('Canada');
  });

  test('the visitor can switch to the standard page and back', async ({ page }) => {
    await stubGeo(page, 'CA');
    await page.goto(`${LANDING_URL}/pricing`);
    await expect(page.getByTestId('price-pro')).toBeVisible();

    await page.getByRole('button', { name: /see standard pricing/i }).click();
    await expect(page.getByTestId('price-pro')).toHaveCount(0);
    const body = (await page.locator('body').innerText()).replace(/\s+/g, ' ');
    expect(body).not.toMatch(/[$€£]\s?\d/);

    await page.getByRole('button', { name: /show prices for canada/i }).click();
    await expect(page.getByTestId('price-pro')).toBeVisible();
  });

  test('a visitor from a region with no prices gets the no-price page', async ({ page }) => {
    // Onion answers with default prices and no region for an unpriced region.
    await stubGeo(page, 'US', {
      plans: [{ id: '2', slug: 'pro', price_per_seat_cents: 699, currency: 'USD' }],
    });
    await page.goto(`${LANDING_URL}/pricing`);
    await expect(page.getByRole('heading', { level: 1, name: /one rtylr suite/i })).toBeVisible();
    await expect(page.getByTestId('regional-pricing-note')).toHaveCount(0);
    const body = (await page.locator('body').innerText()).replace(/\s+/g, ' ');
    expect(body).not.toMatch(/[$€£]\s?\d/);
  });

  test('?region=CA shows Canadian prices without a Canadian IP', async ({ page }) => {
    const asked = await stubGeo(page, 'LB');
    await page.goto(`${LANDING_URL}/pricing?region=CA`);
    await expect(page.getByTestId('price-pro')).toContainText('$8.99');
    expect(asked).toEqual(['CA']);
  });

  test('?region=default keeps a Canadian visitor on the standard page', async ({ page }) => {
    const asked = await stubGeo(page, 'CA');
    await page.goto(`${LANDING_URL}/pricing?region=default`);
    await expect(page.getByRole('heading', { level: 1, name: /one rtylr suite/i })).toBeVisible();
    await expect(page.getByTestId('price-pro')).toHaveCount(0);
    expect(asked).toEqual([]);
  });

  test('if the country cannot be told, the page stays as it is', async ({ page }) => {
    const asked = await stubGeo(page, 'XX');
    await page.goto(`${LANDING_URL}/pricing`);
    await expect(page.getByRole('heading', { level: 1, name: /one rtylr suite/i })).toBeVisible();
    await expect(page.getByTestId('price-pro')).toHaveCount(0);
    expect(asked).toEqual([]);
  });
});
