import { test, expect } from './fixtures';
import { stubPortalSession, PORTAL_URL } from './portal-session';

for (const isOwner of [true, false]) {
  test(`Portal POS ${isOwner ? 'owner' : 'member'} requires terminal and PIN, then preserves the cart on navigation`, async ({
    page,
    context,
  }) => {
    await stubPortalSession(page, context, {
      isOwner,
      appIds: ['pos'],
      permissions: ['pos:orders:create', 'pos:orders:view'],
    });
    await page.route('http://localhost:9999/pos/settings/v1/settings', (r) =>
      r.fulfill({
        json: {
          settings: {
            business_profile: 'retail',
            features: { tables: false, kitchen_display: false, gift_cards: false },
          },
        },
      })
    );
    await page.route('http://localhost:9999/common/oauth/v1/me', (r) =>
      r.fulfill({ json: { user_id: '1' } })
    );
    await page.route('http://localhost:9999/pos/access/v1/my-locations', (r) =>
      r.fulfill({
        json: { locations: [{ location_id: '10', name: 'Main location', is_active: true }] },
      })
    );
    await page.route('http://localhost:9999/pos/location/v1/locations', (r) =>
      r.fulfill({ json: { locations: [{ id: '10', name: 'Main location', is_active: true }] } })
    );
    await page.route('http://localhost:9999/pos/terminal/v1/terminals', (r) =>
      r.fulfill({
        json: {
          terminals: [
            { id: '20', branch_id: '10', name: 'Register A', type: 'web', is_active: true },
          ],
        },
      })
    );
    await page.route('http://localhost:9999/pos/access/v1/location-roster**', (r) =>
      r.fulfill({
        json: {
          roster: [
            {
              user_id: '1',
              name: 'Sam Cashier',
              has_pin: true,
              is_manager: false,
              pos_role: 'cashier',
            },
          ],
        },
      })
    );
    let authorized = false;
    await page.route('http://localhost:9999/pos/user/v1/users/verify-pin', (r) => {
      const request = r.request().postDataJSON();
      expect(request).toMatchObject({ location_id: '10', user_id: '1' });
      if (request.pin !== '2468')
        return r.fulfill({ status: 401, json: { message: 'Incorrect PIN' } });
      authorized = true;
      return r.fulfill({
        json: {
          user_id: '1',
          pos_user_id: '11',
          display_name: 'Sam Cashier',
          location_id: '10',
          role: 'cashier',
        },
      });
    });
    await page.route('http://localhost:9999/pos/menu-item/v1/catalog', (r) =>
      r.fulfill({
        json: {
          items: [
            {
              id: '30',
              name: 'Migration Coffee',
              sale_price: 5,
              category: 'Drinks',
              is_active: true,
            },
          ],
        },
      })
    );
    await page.route('http://localhost:9999/pos/terminal/v1/terminals/heartbeat', (r) =>
      r.fulfill({ json: {} })
    );
    await page.route('http://localhost:9999/pos/stock/v1/low-stock**', (r) =>
      r.fulfill({ json: { items: [] } })
    );
    await page.goto(`${PORTAL_URL}/pos`);
    await page.getByRole('button', { name: /Register A/ }).click();
    await page.getByRole('button', { name: /Sam Cashier/ }).click();
    expect(authorized).toBe(false);
    await page.keyboard.type('2468');
    await page.getByRole('button', { name: 'Sign in', exact: true }).click();
    await expect(page.getByRole('button', { name: /Migration Coffee/ })).toBeVisible();
    expect(authorized).toBe(true);
    await page.getByRole('button', { name: /Migration Coffee/ }).click();
    await expect(page.getByRole('button', { name: /Charge/ })).toBeEnabled();
    const apps = page.getByRole('navigation', { name: 'Apps', exact: true });
    await apps.getByRole('link', { name: 'POS', exact: true }).click();
    await expect(page.getByRole('button', { name: /Charge/ })).toBeEnabled();
    await page.reload();
    await expect(page.getByRole('button', { name: /Charge/ })).toBeEnabled();
    await expect(
      page
        .getByRole('navigation', { name: 'Home navigation' })
        .getByRole('link', { name: 'Tables' })
    ).toHaveCount(0);
  });
}

test('direct POS settings load the saved configuration before saving', async ({
  page,
  context,
}) => {
  await stubPortalSession(page, context, { isOwner: true, appIds: ['pos'] });
  await page.route('http://localhost:9999/pos/settings/v1/settings', (route) => {
    if (route.request().method() === 'PUT') {
      expect(route.request().postDataJSON()).toMatchObject({
        business_profile: 'restaurant',
        receipt_footer: 'Updated receipt',
        features: { tables: true },
      });
    }
    return route.fulfill({
      json: {
        settings: {
          business_profile: 'restaurant',
          receipt_footer: route.request().method() === 'PUT' ? 'Updated receipt' : 'Saved receipt',
          features: { tables: true },
        },
      },
    });
  });
  await page.goto(`${PORTAL_URL}/settings/pos`);
  await expect(page.locator('textarea')).toHaveValue('Saved receipt');
  await page.locator('textarea').fill('Updated receipt');
  const save = page.waitForRequest(
    (r) => r.method() === 'PUT' && r.url().endsWith('/pos/settings/v1/settings')
  );
  await page.getByRole('button', { name: 'Save Business Settings' }).click();
  await save;
  await expect(page.getByText('POS settings saved', { exact: true })).toBeVisible();
});

test('POS settings failures show an error and retry without exposing defaults', async ({
  page,
  context,
}) => {
  await stubPortalSession(page, context, { isOwner: true, appIds: ['pos'] });
  let fail = true;
  await page.route('http://localhost:9999/pos/settings/v1/settings', (r) =>
    fail
      ? r.fulfill({ status: 503, json: { message: 'Settings unavailable' } })
      : r.fulfill({
          json: { settings: { business_profile: 'retail', receipt_footer: 'Recovered receipt' } },
        })
  );
  await page.goto(`${PORTAL_URL}/settings/pos`);
  await expect(page.getByText('Something went wrong', { exact: true })).toBeVisible();
  await expect(page.getByRole('button', { name: 'Save Business Settings' })).toHaveCount(0);
  fail = false;
  await page.getByRole('button', { name: 'Try again', exact: true }).click();
  await expect(page.locator('textarea')).toHaveValue('Recovered receipt');
});
