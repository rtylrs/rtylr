import type { StoredUser } from '../stores/userStore';
import { canEffectiveAccess, hasEffectiveAppAccess } from '@voxire/permissions';

const DASH_ROUTE_PERMISSIONS: Array<[string, string]> = [
  ['/team/invite', 'hub:team:invite'],
  ['/team/groups', 'hub:groups:manage'],
  ['/team', 'hub:team:view'],
  ['/roles', 'hub:roles:view'],
  ['/settings/billing', 'hub:billing:view'],
  ['/settings', 'hub:business:settings:view'],
  ['/menus', 'menu:menus:view'],
  ['/menu', 'menu:menus:view'],
  ['/orders', 'menu:orders:view'],
  ['/qr-codes', 'menu:menus:view'],
  ['/design', 'menu:branding:edit'],
  ['/customers', 'crm:customers:view'],
  ['/products', 'erp:products:view'],
];

export function canAccessPortalPath(user: StoredUser | null, pathname: string): boolean {
  if (!user?.effectiveAccess) return false;
  if (user.effectiveAccess.isOwner || user.effectiveAccess.isAdmin) return true;
  if (pathname === '/insights' || pathname.startsWith('/insights/')) {
    if (!hasEffectiveAppAccess(user.effectiveAccess, 'insights')) return false;
    if (pathname === '/insights/reports') {
      return ['insights:reports:create', 'insights:reports:export', 'insights:reports:share'].some(
        (permission) => canEffectiveAccess(user.effectiveAccess, permission)
      );
    }
    const permission =
      pathname === '/insights/financial'
        ? 'insights:financial:view'
        : pathname === '/insights/customers'
          ? 'insights:customers:view'
          : 'insights:dashboard:view';
    return canEffectiveAccess(user.effectiveAccess, permission);
  }
  if (
    pathname === '/' ||
    pathname === '/settings/account' ||
    pathname === '/security' ||
    pathname === '/notifications'
  ) {
    return true;
  }
  const match = DASH_ROUTE_PERMISSIONS.find(([prefix]) => pathname.startsWith(prefix));
  return match ? canEffectiveAccess(user.effectiveAccess, match[1]) : false;
}
