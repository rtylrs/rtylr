import type { BrowserContext, Page } from './fixtures';
import { encodeListMyBusinessesResponse } from '@voxire/voxie/hub/business/v1/types';
import { encodeGetMeResponse } from '@voxire/voxie/hub/user/v1/types';

export const PORTAL_URL = process.env.PORTAL_URL ?? 'http://localhost:5197';
export const PREVIEW_TOKEN = 'fake-jwt.test.signature';

interface PortalSessionOptions {
  isOwner?: boolean;
  appIds?: string[];
  permissions?: string[];
}

export async function stubPortalSession(
  page: Page,
  context: BrowserContext,
  {
    isOwner = false,
    appIds = ['insights'],
    permissions = ['insights:dashboard:view'],
  }: PortalSessionOptions = {}
): Promise<void> {
  await context.addCookies([
    { name: 'token', value: PREVIEW_TOKEN, url: PORTAL_URL, sameSite: 'Lax' },
    { name: 'rtylr-consent', value: 'denied', url: PORTAL_URL, sameSite: 'Lax' },
  ]);
  await page.route('**/hub/user/v1/users/me', (route) =>
    route.fulfill({
      json: encodeGetMeResponse({
        id: '1',
        email: 'sam@example.test',
        fullName: 'Sam Test',
        businessId: '1',
        businessName: 'Northwind',
        onboarded: true,
        effectiveAccess: {
          businessId: '1',
          dashPermissions: [],
          subscribedApps: appIds,
          licenses: appIds.map((appId) => ({
            appId,
            status: 'active',
            assignmentSource: 'explicit',
            grantedAt: '2026-04-01T00:00:00Z',
          })),
          apps: appIds.map((appId) => ({
            appId,
            licensed: true,
            hasAccess: true,
            permissions: permissions.filter((permission) => permission.startsWith(`${appId}:`)),
            permissionScopes: permissions
              .filter((permission) => permission.startsWith(`${appId}:`))
              .map((permissionKey) => ({ permissionKey, scopes: [] })),
            sources: ['explicit'],
          })),
          isOwner,
          isAdmin: false,
          authorizationVersion: '1',
        },
      }),
    })
  );
  await page.route('**/hub/config/v1/config', (route) =>
    route.fulfill({ json: { business: { id: '1', name: 'Northwind' }, maintenance: false } })
  );
  await page.route('**/hub/business/v1/businesses/mine', (route) =>
    route.fulfill({
      json: encodeListMyBusinessesResponse({
        businesses: [
          { id: '1', name: 'Northwind', isOwned: isOwner, isActive: true, parentId: '0' },
        ],
      }),
    })
  );
  await page.route('**/hub/notification/v1/notifications**', (route) =>
    route.fulfill({ json: { notifications: [], next_cursor: '' } })
  );
  await page.route('**/hub/app/v1/addons', (route) => route.fulfill({ json: { addons: [] } }));
}
