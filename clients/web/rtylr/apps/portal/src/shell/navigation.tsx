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
} from 'lucide-react';
import { HomeIcon, MenuIcon } from '@voxire/ui';
import type { ComponentType } from 'react';
import { routeKeys } from './routeTitles';
export { routeKeys };

export interface PortalNavItem {
  key?: string;
  path: string;
  icon: ComponentType<{ size?: number; className?: string }>;
}
export const rawNavItems: PortalNavItem[] = [
  { key: 'dashboard', path: '/', icon: HomeIcon },
  { key: 'team', path: '/team', icon: Users },
  { key: 'groups', path: '/team/groups', icon: Folder },
  { key: 'roles', path: '/roles', icon: Shield },
  { key: 'auditLog', path: '/audit', icon: History },
  { key: 'integrations', path: '/integrations', icon: Plug2 },
];
export const rawMenuSubItems: PortalNavItem[] = [
  { key: 'menus', path: '/menus', icon: MenuIcon },
  { key: 'qrCodes', path: '/qr-codes', icon: QrCode },
  { key: 'design', path: '/design', icon: Palette },
];
export const settingsNavItems: PortalNavItem[] = [
  { key: 'account', path: '/settings/account', icon: User },
  { key: 'security', path: '/security', icon: Shield },
  { key: 'businesses', path: '/settings/businesses', icon: Building2 },
  { key: 'general', path: '/settings/general', icon: Building2 },
  { key: 'industry', path: '/settings/industry', icon: Briefcase },
  { key: 'aiUsage', path: '/settings/ai', icon: Sparkles },
  { key: 'billing', path: '/settings/billing', icon: CreditCard },
  { key: 'addons', path: '/settings/addons', icon: Puzzle },
  { key: 'integrations', path: '/settings/integrations', icon: Plug2 },
];
export const addonSettingsItems = [
  { key: 'customDomain', path: '/settings/domain', icon: Globe, addonId: 'custom_domain' },
  { key: 'emailSettings', path: '/settings/email', icon: Mail, addonId: 'email_marketing' },
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
    path === '/' || path === '/insights'
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
