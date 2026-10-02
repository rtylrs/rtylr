import { Link, useRouterState } from '@tanstack/react-router';
import {
  Users,
  Folder,
  Shield,
  History,
  Plug2,
  User,
  Building2,
  Globe,
  Mail,
  CreditCard,
  Puzzle,
  Briefcase,
  Sparkles,
  QrCode,
  Palette,
  KeyRound,
  Webhook,
  Settings,
} from 'lucide-react';
import { HomeIcon, MenuIcon, FlowIcon, InsightsIcon } from '@voxire/ui';
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
  { key: 'integrations', path: '/integrations', icon: Plug2, section: 'business' },
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
  { key: 'addons', path: '/settings/addons', icon: Puzzle, section: 'business' },
  { key: 'integrations', path: '/settings/integrations', icon: Plug2, section: 'business' },
  { key: 'apiKeys', path: '/api-keys', icon: KeyRound, section: 'business' },
  { key: 'webhooks', path: '/webhooks', icon: Webhook, section: 'business' },
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
    path === '/' || path === '/insights' || path === '/settings'
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
