import {
  Users,
  Gift,
  Megaphone,
  Filter,
  FileText,
  MessageSquare,
  Kanban,
  Contact,
  FileSignature,
  LayoutDashboard,
  ListTodo,
} from 'lucide-react';
import { canEffectiveAccess, hasEffectiveAppAccess } from '@voxire/permissions';
import type { StoredUser } from '../../shared/stores/userStore';

/** Sidebar groups, labelled from crmNav.sections. */
export type CrmNavGroup = 'sales' | 'customers' | 'marketing' | 'retention';

export const crmNavItems = [
  { key: 'dashboard', path: '/crm', icon: LayoutDashboard },
  { key: 'customers', path: '/crm/customers', icon: Users, group: 'customers' },
  { key: 'campaigns', path: '/crm/campaigns', icon: Megaphone, group: 'marketing' },
  { key: 'segments', path: '/crm/segments', icon: Filter, group: 'marketing' },
  { key: 'templates', path: '/crm/templates', icon: FileText, group: 'marketing' },
  { key: 'loyalty', path: '/crm/loyalty', icon: Gift, group: 'retention' },
  { key: 'feedback', path: '/crm/feedback', icon: MessageSquare, group: 'retention' },
] as const;

export const crmSalesNavItems = [
  { key: 'leads', path: '/crm/sales/leads', icon: Contact, group: 'sales' },
  { key: 'deals', path: '/crm/sales', icon: Kanban, group: 'sales' },
  { key: 'tasks', path: '/crm/sales/tasks', icon: ListTodo, group: 'sales' },
  { key: 'contracts', path: '/crm/sales/contracts', icon: FileSignature, group: 'sales' },
] as const;

/** The CRM sidebar in order: overview, then sales ahead of the other groups. */
export const crmSidebarItems = [crmNavItems[0], ...crmSalesNavItems, ...crmNavItems.slice(1)];

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
    ['/crm/sales/leads/import', 'crm:customers:create'],
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
    crmNavItems.slice(1).find((item) => canAccessCrmPath(user, item.path))?.path ??
    (canAccessCrmPath(user, '/crm/automations') ? '/automations' : '/settings/crm')
  );
}
