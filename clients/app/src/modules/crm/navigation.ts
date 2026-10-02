import {
  Users,
  Gift,
  Megaphone,
  Filter,
  Zap,
  FileText,
  MessageSquare,
  Kanban,
  Contact,
  FileSignature,
  LayoutDashboard,
} from 'lucide-react';
import { canEffectiveAccess, hasEffectiveAppAccess } from '@voxire/permissions';
import type { StoredUser } from '../../shared/stores/userStore';

export const crmNavItems = [
  { key: 'dashboard', path: '/crm', icon: LayoutDashboard },
  { key: 'customers', path: '/crm/customers', icon: Users },
  { key: 'loyalty', path: '/crm/loyalty', icon: Gift },
  { key: 'campaigns', path: '/crm/campaigns', icon: Megaphone },
  { key: 'segments', path: '/crm/segments', icon: Filter },
  { key: 'automations', path: '/crm/automations', icon: Zap },
  { key: 'templates', path: '/crm/templates', icon: FileText },
  { key: 'feedback', path: '/crm/feedback', icon: MessageSquare },
] as const;

export const crmSalesNavItems = [
  { key: 'pipeline', path: '/crm/sales', icon: Kanban },
  { key: 'leads', path: '/crm/sales/leads', icon: Contact },
  { key: 'contracts', path: '/crm/sales/contracts', icon: FileSignature },
] as const;

export function canAccessCrmPath(user: StoredUser | null, pathname: string): boolean {
  const access = user?.effectiveAccess;
  if (!hasEffectiveAppAccess(access, 'crm')) return false;
  if (access?.isOwner || access?.isAdmin) return true;
  if (pathname === '/crm' || pathname === '/crm/settings' || pathname === '/settings/crm')
    return true;
  if (['/crm/campaigns/email', '/crm/campaigns/sms'].includes(pathname))
    return ['crm:campaigns:create', 'crm:campaigns:manage'].some((permission) =>
      canEffectiveAccess(access, permission, 'crm')
    );
  const routes: Array<[string, string]> = [
    ['/crm/customers/new', 'crm:customers:create'],
    ['/crm/customers/import', 'crm:customers:create'],
    ['/crm/customers/merge', 'crm:customers:delete'],
    ['/crm/campaigns/email', 'crm:campaigns:create'],
    ['/crm/campaigns/sms', 'crm:campaigns:create'],
    ['/crm/campaigns', 'crm:campaigns:view'],
    ['/crm/segments/new', 'crm:segments:manage'],
    ['/crm/automations/new', 'crm:automations:manage'],
    ['/crm/customers', 'crm:customers:view'],
    ['/crm/loyalty', 'crm:loyalty:manage'],
    ['/crm/segments', 'crm:customers:view'],
    ['/crm/automations', 'crm:automations:manage'],
    ['/crm/templates', 'crm:campaigns:view'],
    ['/crm/feedback', 'crm:customers:view'],
    ['/crm/sales', 'crm:customers:view'],
  ];
  const match = routes.find(([path]) => pathname === path || pathname.startsWith(`${path}/`));
  return !!match && canEffectiveAccess(access, match[1], 'crm');
}

export function getCrmHomePath(user: StoredUser | null): string {
  if (canEffectiveAccess(user?.effectiveAccess, 'crm:customers:view', 'crm')) return '/crm';
  return (
    crmNavItems.slice(1).find((item) => canAccessCrmPath(user, item.path))?.path ?? '/settings/crm'
  );
}
