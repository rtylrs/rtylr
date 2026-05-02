import React, { useRef } from 'react';
import { useLocation, useNavigate, Link } from 'react-router-dom';
import { useForesightLinks } from '@voxire/utils';
import { motion, AnimatePresence } from 'motion/react';
import { useQuery } from '@tanstack/react-query';
import api from '../shared/lib/api';
import { APP_URLS } from '@voxire/config';
import {
  LayoutDashboard,
  UtensilsCrossed,
  ClipboardList,
  QrCode,
  Palette,
  BarChart3,
  Settings,
  Bell,
  LogOut,
  ChevronRight,
  Sun,
  Moon,
  Monitor,
  MonitorSpeaker,
  Package,
  UserCheck,
  CreditCard,
  Layers,
  ExternalLink,
  Users,
  Grid3x3,
  Store,
  MessageSquare,
  Activity as ActivityIcon,
  Search,
  Sparkles,
} from 'lucide-react';
import { CaretPanel, CaretButton } from '@voxire/caret';
import { useUserStore } from '../shared/stores/userStore';
import { useThemeStore } from '../shared/stores/themeStore';
import { cn } from '../shared/lib/utils';
import {
  Avatar,
  DropdownMenu,
  DropdownMenuTrigger,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
} from '../shared/components/ui';
import { SearchCommand } from '../shared/components/dash';
import { useCommandPalette } from '../shared/stores/commandPaletteStore';
import { useKeyboardShortcuts } from '../shared/hooks/useKeyboardShortcuts';
import { useOnboardingTour } from '../shared/hooks/useOnboardingTour';
import { useNotifications } from '../shared/hooks/queries/useNotifications';

interface NavItem {
  label: string;
  path: string;
  icon: React.ComponentType<{ size?: number; className?: string }>;
}

const navItems: NavItem[] = [
  { label: 'Dashboard', path: '/', icon: LayoutDashboard },
  { label: 'Team', path: '/team', icon: Users },
  { label: 'Apps', path: '/apps', icon: Grid3x3 },
  { label: 'Activity', path: '/activity', icon: ActivityIcon },
];

const menuSubItems: NavItem[] = [
  { label: 'Menus', path: '/menus', icon: UtensilsCrossed },
  { label: 'Orders', path: '/orders', icon: ClipboardList },
  { label: 'QR Codes', path: '/qr-codes', icon: QrCode },
  { label: 'Design', path: '/design', icon: Palette },
];

const routeLabels: Record<string, string> = {
  '/': 'Dashboard',
  '/chat': '__caret__',
  '/menus': 'Menus',
  '/orders': 'Orders',
  '/qr-codes': 'QR Codes',
  '/design': 'Design',
  '/analytics': 'Analytics',
  '/menu': 'Menu',
  '/create-item': 'Create Item',
  '/create-banner': 'Create Banner',
  '/team': 'Team',
  '/team/invite': 'Invite member',
  '/permissions': 'Permissions',
  '/roles': 'Roles',
  '/apps': 'Apps',
  '/activity': 'Activity',
  '/audit': 'Audit log',
  '/notifications': 'Notifications',
  '/account': 'Account',
  '/billing': 'Billing',
  '/integrations': 'Integrations',
  '/api-keys': 'API Keys',
  '/webhooks': 'Webhooks',
  '/security': 'Security',
  '/marketplace': 'Marketplace',
  '/changelog': "What's new",
  '/flags': 'Feature flags',
  '/theme': 'Theme',
  '/referrals': 'Referrals',
  '/help': 'Help center',
  '/export': 'Export data',
  '/import': 'Import data',
  '/settings': 'Manage account',
  '/settings/account': 'Account',
  '/settings/general': 'General',
  '/settings/industry': 'Industry',
  '/settings/domain': 'Custom Domain',
  '/settings/email': 'Email Settings',
  '/settings/billing': 'Billing',
  '/settings/addons': 'Addons',
};

function getPageTitle(pathname: string): string {
  if (routeLabels[pathname]) return routeLabels[pathname];
  const match = Object.keys(routeLabels).find((key) => pathname.startsWith(key) && key !== '/');
  return match ? routeLabels[match] : 'Dash';
}

function getInitials(name?: string | null): string {
  if (!name) return '?';
  return name
    .split(' ')
    .map((n) => n[0])
    .join('')
    .toUpperCase()
    .slice(0, 2);
}

const SidebarNavItem: React.FC<NavItem> = ({ label, path, icon: Icon }) => {
  const location = useLocation();
  const isActive = path === '/' ? location.pathname === '/' : location.pathname.startsWith(path);

  return (
    <Link
      to={path}
      data-tour={label.toLowerCase()}
      className={cn(
        'relative flex items-center gap-3 px-3 py-2 rounded-lg text-sm font-medium transition-colors duration-150',
        isActive ? '' : 'hover:bg-[var(--dash-hover)]'
      )}
      style={{ color: isActive ? 'var(--dash-text)' : 'var(--dash-text-muted)' }}
    >
      {isActive && (
        <motion.span
          layoutId="sidebar-active"
          className="absolute inset-0 rounded-lg"
          style={{ background: 'var(--dash-active)' }}
          transition={{ type: 'spring', stiffness: 400, damping: 35 }}
        />
      )}
      <Icon size={16} className="relative shrink-0" />
      <span className="relative">{label}</span>
      {isActive && (
        <span
          className="relative ml-auto h-1 w-1 rounded-full shrink-0"
          style={{ background: '#4583C4' }}
        />
      )}
    </Link>
  );
};

const SidebarSection: React.FC<{
  label: string;
  icon: React.ComponentType<{ size?: number; className?: string }>;
  items: NavItem[];
}> = ({ label, icon: Icon, items }) => {
  const location = useLocation();
  const isActive = items.some((item) => location.pathname.startsWith(item.path));
  const [open, setOpen] = React.useState(isActive);

  return (
    <div className="pt-1">
      <button
        onClick={() => setOpen((v) => !v)}
        className={cn(
          'flex items-center gap-3 px-3 py-2 rounded-lg text-sm w-full transition-colors duration-150',
          isActive ? '' : 'hover:bg-[var(--dash-hover)]'
        )}
        style={{ color: isActive ? 'var(--dash-text)' : 'var(--dash-text-muted)' }}
      >
        <Icon size={16} className="shrink-0" />
        <span className="flex-1 text-left">{label}</span>
        <ChevronRight
          size={12}
          className={cn('shrink-0 transition-transform', open && 'rotate-90')}
          style={{ color: 'var(--dash-text-dim)' }}
        />
      </button>
      {open && (
        <div
          className="ml-5 pl-2 mt-0.5 space-y-0.5"
          style={{ borderLeft: '1px solid var(--dash-border)' }}
        >
          {items.map((item) => (
            <SidebarNavItem key={item.path} {...item} />
          ))}
        </div>
      )}
    </div>
  );
};

interface DefaultLayoutProps {
  children?: React.ReactNode;
}

const DefaultLayout: React.FC<DefaultLayoutProps> = ({ children }) => {
  const location = useLocation();
  const navigate = useNavigate();
  const { user, logout } = useUserStore();
  const { theme, setTheme } = useThemeStore();
  const { setOpen: setCommandPaletteOpen } = useCommandPalette();
  const { data: notifications = [] } = useNotifications();
  const { data: addonsData } = useQuery<Array<{ id: string; isEnabled: boolean }>>({
    queryKey: ['addons'],
    queryFn: async () => {
      const res = await api.get('/addons');
      return res.data?.addons ?? [];
    },
    staleTime: 60_000,
  });
  const menuAddonEnabled = (addonsData ?? []).some((a) => a.id === 'menu' && a.isEnabled);
  const sidebarRef = useRef<HTMLElement | null>(null);

  useForesightLinks(sidebarRef, {
    '/': () => import('../features/dashboard/Dashboard/VendorDashboard'),
    '/team': () => import('../features/team/Team'),
    '/apps': () => import('../features/addons/Addons'),
    '/activity': () => import('../features/activity/Activity'),
    '/menus': () => import('../features/menu/Menus'),
    '/orders': () => import('../features/orders/Orders'),
    '/qr-codes': () => import('../features/qr-codes/QrCodes'),
    '/design': () => import('../features/design/Design'),
    '/analytics': () => import('../features/analytics/Analytics'),
  });

  useKeyboardShortcuts();
  useOnboardingTour();

  const userId = Number(user?.id) || 0;
  const unreadCount = Array.isArray(notifications)
    ? notifications.filter((n) => !n.read).length
    : 0;
  const pageTitle = getPageTitle(location.pathname);
  const userEmail = user?.email || user?.emailAddress || null;
  const userName = user?.name || user?.firstName || userEmail || 'User';
  const businessName = user?.businessName || user?.storeName || 'My Business';

  const handleLogout = () => {
    logout();
    window.location.href = APP_URLS.auth;
  };

  return (
    <>
      <div className="flex h-screen overflow-hidden" style={{ background: 'var(--dash-bg)' }}>
        {/* Sidebar */}
        <aside
          className="w-64 flex flex-col shrink-0 overflow-hidden"
          style={{
            background: 'var(--dash-surface)',
            borderRight: '1px solid var(--dash-border)',
          }}
        >
          {/* Logo */}
          <div
            className="flex items-center gap-2.5 px-4 py-4"
            style={{ borderBottom: '1px solid var(--dash-border)' }}
          >
            <img
              src="/logo.png"
              alt="rtylr"
              className="h-6 w-auto"
              style={{
                filter:
                  theme === 'dark' ||
                  (theme === 'system' && window.matchMedia('(prefers-color-scheme: dark)').matches)
                    ? 'brightness(0) invert(1)'
                    : 'brightness(0)',
              }}
            />
          </div>

          {/* Navigation */}
          <nav className="flex-1 overflow-y-auto py-2 px-2 space-y-0.5">
            {navItems.map((item) => (
              <SidebarNavItem key={item.path} {...item} />
            ))}

            {/* Storefront section - only shown when menu addon is enabled */}
            {menuAddonEnabled && (
              <SidebarSection label="Storefront" icon={Store} items={menuSubItems} />
            )}
          </nav>

          {/* Footer */}
          <div className="p-2" style={{ borderTop: '1px solid var(--dash-border)' }}>
            <Link
              to="/settings/account"
              data-tour="settings"
              className={cn(
                'flex items-center gap-3 px-3 py-2 rounded-lg text-sm transition-colors duration-150 mb-0.5',
                location.pathname.startsWith('/settings')
                  ? 'bg-[var(--dash-active)]'
                  : 'hover:bg-[var(--dash-hover)]'
              )}
              style={{
                color: location.pathname.startsWith('/settings')
                  ? 'var(--dash-text)'
                  : 'var(--dash-text-muted)',
              }}
            >
              <Settings size={16} />
              <span>Manage account</span>
            </Link>

            <DropdownMenu>
              <DropdownMenuTrigger>
                <button className="flex items-center gap-2.5 w-full px-3 py-2 rounded-lg transition-colors duration-150 hover:bg-[var(--dash-hover)]">
                  <Avatar
                    initials={getInitials(userName)}
                    size="sm"
                    className="shrink-0"
                    style={{ background: 'rgba(69,131,196,0.15)', color: '#4583C4' }}
                  />
                  <div className="flex flex-col items-start min-w-0 flex-1">
                    <span
                      className="text-xs font-medium truncate w-full text-left"
                      style={{ color: 'var(--dash-text)' }}
                    >
                      {userName}
                    </span>
                    {userEmail && (
                      <span
                        className="text-[10px] truncate w-full text-left"
                        style={{ color: 'var(--dash-text-muted)' }}
                      >
                        {userEmail}
                      </span>
                    )}
                  </div>
                  <ChevronRight
                    size={12}
                    className="shrink-0"
                    style={{ color: 'var(--dash-text-dim)' }}
                  />
                </button>
              </DropdownMenuTrigger>
              <DropdownMenuContent align="start" side="top" className="w-52">
                {/* Theme selector */}
                <div className="px-2 py-1.5">
                  <p
                    className="text-[10px] font-semibold uppercase tracking-wider mb-1.5"
                    style={{ color: 'var(--dash-text-dim)' }}
                  >
                    Appearance
                  </p>
                  <div className="flex gap-1">
                    {(
                      [
                        { value: 'light', icon: Sun, label: 'Light' },
                        { value: 'dark', icon: Moon, label: 'Dark' },
                        { value: 'system', icon: Monitor, label: 'System' },
                      ] as const
                    ).map(({ value, icon: Icon, label }) => (
                      <button
                        key={value}
                        onClick={() => setTheme(value)}
                        className="flex-1 flex flex-col items-center gap-1 py-2 rounded-lg text-[10px] font-medium transition-colors"
                        style={{
                          background:
                            theme === value ? 'rgba(69,131,196,0.15)' : 'var(--dash-hover)',
                          color: theme === value ? '#4583C4' : 'var(--dash-text-muted)',
                          border:
                            theme === value
                              ? '1px solid rgba(69,131,196,0.3)'
                              : '1px solid transparent',
                        }}
                      >
                        <Icon size={13} />
                        {label}
                      </button>
                    ))}
                  </div>
                </div>
                <DropdownMenuSeparator />
                <DropdownMenuItem destructive onClick={handleLogout}>
                  <LogOut size={14} />
                  Sign out
                </DropdownMenuItem>
              </DropdownMenuContent>
            </DropdownMenu>
          </div>
        </aside>

        {/* Main content */}
        <div className="flex flex-col flex-1 overflow-hidden">
          {/* Topbar */}
          <header
            className="h-11 flex items-center justify-between px-6 shrink-0"
            style={{
              background: 'var(--dash-surface)',
              borderBottom: '1px solid var(--dash-border)',
            }}
          >
            {pageTitle === '__caret__' ? (
              <div className="flex items-center gap-2">
                <div
                  className="h-6 w-6 rounded-lg flex items-center justify-center"
                  style={{ background: 'rgba(139,92,246,0.1)' }}
                >
                  <Sparkles size={13} style={{ color: '#8b5cf6' }} />
                </div>
                <span className="text-sm font-semibold" style={{ color: 'var(--dash-text)' }}>
                  Caret
                </span>
                <span
                  className="text-[9px] font-bold uppercase tracking-wider px-1.5 py-0.5 rounded"
                  style={{ background: 'rgba(139,92,246,0.1)', color: '#8b5cf6' }}
                >
                  AI
                </span>
              </div>
            ) : (
              <h1 className="text-sm font-semibold" style={{ color: 'var(--dash-text)' }}>
                {pageTitle}
              </h1>
            )}
            <div className="flex items-center gap-1">
              <button
                onClick={() => setCommandPaletteOpen(true)}
                className="flex items-center gap-2 px-2.5 h-7 rounded-lg transition-colors duration-150 hover:bg-[var(--dash-hover)]"
                style={{ color: 'var(--dash-text-muted)', border: '1px solid var(--dash-border)' }}
                aria-label="Open command palette"
              >
                <Search size={12} />
                <span className="text-[11px]">Search</span>
                <kbd
                  className="hidden sm:inline-flex items-center px-1 py-0.5 rounded text-[9px] font-mono"
                  style={{ background: 'var(--dash-hover)', color: 'var(--dash-text-dim)' }}
                >
                  ⌘K
                </kbd>
              </button>
              <button
                onClick={() => navigate('/chat')}
                className="h-7 rounded-lg flex items-center gap-1.5 px-3 transition-all duration-150"
                style={{
                  background: 'rgba(139,92,246,0.1)',
                  color: '#8b5cf6',
                  border: '1px solid rgba(139,92,246,0.2)',
                }}
                onMouseEnter={(e) => {
                  (e.currentTarget as HTMLElement).style.background = 'rgba(139,92,246,0.18)';
                }}
                onMouseLeave={(e) => {
                  (e.currentTarget as HTMLElement).style.background = 'rgba(139,92,246,0.1)';
                }}
                aria-label="Chat with Caret"
                data-tour="caret"
              >
                <Sparkles size={13} />
                <span className="text-[11px] font-semibold">Caret AI</span>
              </button>
              <CaretButton />
              <button
                onClick={() => navigate('/notifications')}
                className="relative h-7 w-7 rounded-lg flex items-center justify-center transition-colors duration-150 hover:bg-[var(--dash-hover)]"
                style={{ color: 'var(--dash-text-muted)' }}
                aria-label="Notifications"
              >
                <Bell size={14} />
                {unreadCount > 0 && (
                  <span
                    className="absolute top-1 right-1 h-1.5 w-1.5 rounded-full"
                    style={{ background: '#ef4444' }}
                  />
                )}
              </button>
            </div>
          </header>
          <SearchCommand />

          {/* Page content */}
          <AnimatePresence mode="wait">
            <motion.main
              key={location.pathname}
              initial={{ opacity: 0, y: 6 }}
              animate={{ opacity: 1, y: 0 }}
              exit={{ opacity: 0, y: -4 }}
              transition={{ duration: 0.18, ease: 'easeOut' }}
              className="flex-1 overflow-y-auto p-6"
            >
              {children}
            </motion.main>
          </AnimatePresence>
        </div>
      </div>
      <CaretPanel
        appId="dash"
        userId={userId}
        apiBaseUrl={import.meta.env.VITE_API_URL ?? 'https://api.voxire.dev/'}
        systemContext="User is in the Dash app (business command center — KPIs, orders, staff overview)"
      />
    </>
  );
};

export default DefaultLayout;
