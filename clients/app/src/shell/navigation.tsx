import { Link, useRouterState } from '@tanstack/react-router';
import {
  Users,
  Folder,
  Shield,
  History,
  User,
  Building2,
  Globe,
  Mail,
  CreditCard,
  ToyBrick,
  Briefcase,
  Sparkles,
  QrCode,
  Palette,
  KeyRound,
  Settings,
} from 'lucide-react';
import {
  HomeIcon,
  MenuIcon,
  FlowIcon,
  InsightsIcon,
  CrmIcon,
  ErpIcon,
  FinanceIcon,
  HrIcon,
  RecruitIcon,
} from '@voxire/ui';
import type { ComponentType } from 'react';
import { routeKeys } from './routeTitles';
export { routeKeys };

export interface PortalNavItem {
  key?: string;
  path: string;
  icon: ComponentType<{ size?: number; className?: string }>;
  section?: 'account' | 'business' | 'apps';
}
export const rawNavItems: PortalNavItem[] = [
  { key: 'dashboard', path: '/', icon: HomeIcon },
  { key: 'team', path: '/team', icon: Users },
  { key: 'groups', path: '/team/groups', icon: Folder },
  { key: 'roles', path: '/roles', icon: Shield },
  { key: 'auditLog', path: '/audit', icon: History },
  { key: 'extensions', path: '/extensions', icon: ToyBrick },
];
export const rawMenuSubItems: PortalNavItem[] = [
  { key: 'menus', path: '/menus', icon: MenuIcon },
  { key: 'qrCodes', path: '/qr-codes', icon: QrCode },
  { key: 'design', path: '/design', icon: Palette },
];
export const settingsNavItems: PortalNavItem[] = [
  { key: 'settings', path: '/settings', icon: Settings },
  { key: 'account', path: '/settings/account', icon: User, section: 'account' },
  { key: 'security', path: '/security', icon: Shield, section: 'account' },
  { key: 'businesses', path: '/settings/businesses', icon: Building2, section: 'account' },
  { key: 'general', path: '/settings/general', icon: Building2, section: 'business' },
  { key: 'industry', path: '/settings/industry', icon: Briefcase, section: 'business' },
  { key: 'aiUsage', path: '/settings/ai', icon: Sparkles, section: 'business' },
  { key: 'billing', path: '/settings/billing', icon: CreditCard, section: 'business' },
  { key: 'apiKeys', path: '/api-keys', icon: KeyRound, section: 'business' },
  { key: 'recruitSettings', path: '/settings/recruit', icon: RecruitIcon, section: 'apps' },
  {
    key: 'hrPayrollSettings',
    path: '/settings/hr/payroll-config',
    icon: CreditCard,
    section: 'apps',
  },
  { key: 'hrLeaveSettings', path: '/settings/hr/leave-policy', icon: Users, section: 'apps' },
  { key: 'hrSettings', path: '/settings/hr', icon: HrIcon, section: 'apps' },
  { key: 'financeSettings', path: '/settings/finance', icon: FinanceIcon, section: 'apps' },
  { key: 'erpSettings', path: '/settings/erp', icon: ErpIcon, section: 'apps' },
  { key: 'crmSettings', path: '/settings/crm', icon: CrmIcon, section: 'apps' },
  { key: 'flowSettings', path: '/settings/flow', icon: FlowIcon, section: 'apps' },
  { key: 'insightsSettings', path: '/settings/insights', icon: InsightsIcon, section: 'apps' },
];
export const addonSettingsItems = [
  {
    key: 'customDomain',
    path: '/settings/domain',
    icon: Globe,
    section: 'business' as const,
    addonId: 'custom_domain',
  },
  {
    key: 'emailSettings',
    path: '/settings/email',
    icon: Mail,
    section: 'business' as const,
    addonId: 'email_marketing',
  },
];

export function PortalNavLink({
  path,
  label,
  icon: Icon,
  tourId,
  onNavigate,
}: PortalNavItem & { label: string; tourId?: string; onNavigate?: () => void }) {
  const pathname = useRouterState({ select: (state) => state.location.pathname });
  const active =
    path === '/' ||
    path === '/insights' ||
    path === '/crm' ||
    path === '/erp' ||
    path === '/finance' ||
    path === '/hr' ||
    path === '/crm/sales' ||
    path === '/settings' ||
    path === '/automations'
      ? pathname === path
      : path === '/team'
        ? pathname === '/team' ||
          (pathname.startsWith('/team/') && !pathname.startsWith('/team/groups'))
        : pathname === path || pathname.startsWith(`${path}/`);
  return (
    <Link
      to={path}
      data-tour={tourId}
      aria-current={active ? 'page' : undefined}
      className={`portal-nav-link${active ? ' is-active' : ''}`}
      onClick={onNavigate}
    >
      <Icon size={17} />
      <span>{label}</span>
    </Link>
  );
}
