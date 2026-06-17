import React, { useRef } from 'react';
import { useNavigate, Link, useRouterState } from '@tanstack/react-router';
import { useForesightLinks } from '@voxire/utils';
import { motion, AnimatePresence } from 'motion/react';
import { useQuery } from '@tanstack/react-query';
import { unwrap } from '@voxire/api';
import { safe } from '../shared/lib/api';
import { APP_URLS, getAvatarUrl } from '@voxire/config';
import { signOut } from '@voxire/auth';
import {
  LayoutDashboard,
  UtensilsCrossed,
  ClipboardList,
  QrCode,
  Palette,
  Settings,
  Bell,
  LogOut,
  ChevronRight,
  Sun,
  Moon,
  Monitor,
  ExternalLink,
  Users,
  Grid3x3,
  Store,
  MessageSquare,
  Search,
  Sparkles,
} from 'lucide-react';
import { CaretPanel, CaretButton, useCaretStore, useCaretSettingsStore } from '@voxire/caret';
import { useUserStore } from '../shared/stores/userStore';
import { useBusinessStore } from '../shared/stores/businessStore';
import { useConfigStore } from '../shared/stores/configStore';
import { useThemeStore } from '../shared/stores/themeStore';
import { useApps } from '../shared/hooks/queries/useApps';
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
import BusinessSwitcher from '../features/business-switcher/BusinessSwitcher';

interface NavItem {
  label: string;
  path: string;
  icon: React.ComponentType<{ size?: number; className?: string }>;
}

const navItems: NavItem[] = [
  { label: 'Dashboard', path: '/', icon: LayoutDashboard },
  { label: 'Team', path: '/team', icon: Users },
  { label: 'Apps', path: '/apps', icon: Grid3x3 },
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
  '/settings/subscription': 'Billing',
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
  const pathname = useRouterState({ select: (s) => s.location.pathname });
  const isActive = path === '/' ? pathname === '/' : pathname.startsWith(path);

  return (
    <Link
      to={path}
      data-tour={label.toLowerCase()}
      className={cn(
        'group relative flex items-center gap-3 px-3 py-2 rounded-lg text-sm font-medium transition-colors duration-150',
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
      <motion.span
        className="relative inline-flex shrink-0"
        whileHover={{ scale: 1.18, rotate: -4 }}
        transition={{ type: 'spring', stiffness: 500, damping: 18 }}
      >
        <Icon size={16} />
      </motion.span>
      <span className="relative">{label}</span>
      {isActive && (
        <motion.span
          className="relative ml-auto h-1 w-1 rounded-full shrink-0"
          style={{ background: '#4583C4' }}
          initial={{ scale: 0 }}
          animate={{ scale: 1 }}
          transition={{ type: 'spring', stiffness: 600, damping: 20 }}
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
  const pathname = useRouterState({ select: (s) => s.location.pathname });
  const isActive = items.some((item) => pathname.startsWith(item.path));
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

const AppsSection: React.FC = () => {
  const { user } = useUserStore();
  const { data: apps = [] } = useApps();
  const userApps = (user?.apps as string[] | undefined) ?? [];
  const isOwner = user?.role === 'owner' || user?.role === 'admin';

  const visible = apps.filter((a) => {
    if (!(a.key in APP_URLS)) return false;
    if (a.status !== 'active') return false;
    if (isOwner) return userApps.length === 0 || userApps.includes(a.key);
    return userApps.includes(a.key);
  });

  if (visible.length === 0) return null;

  return (
    <div className="pt-3">
      <div
        className="px-3 pb-1.5 text-xs font-semibold uppercase tracking-wide"
        style={{ color: 'var(--dash-text-dim)' }}
      >
        Apps
      </div>
      <div className="space-y-0.5">
        {visible.map((a) => {
          const Icon = a.icon;
          return (
            <a
              key={a.key}
              href={APP_URLS[a.key as keyof typeof APP_URLS]}
              target="_blank"
              rel="noopener noreferrer"
              className="flex items-center gap-3 px-3 py-2 rounded-lg text-sm transition-colors duration-150 hover:bg-[var(--dash-hover)]"
              style={{ color: 'var(--dash-text-muted)' }}
            >
              <span
                className="h-6 w-6 rounded-md flex items-center justify-center shrink-0"
                style={{ background: a.bg }}
              >
                <Icon size={13} style={{ color: a.color }} />
              </span>
              <span className="flex-1 truncate">{a.label}</span>
              <ExternalLink size={11} style={{ color: 'var(--dash-text-dim)' }} />
            </a>
          );
        })}
      </div>
    </div>
  );
};

interface DefaultLayoutProps {
  children?: React.ReactNode;
}

const DefaultLayout: React.FC<DefaultLayoutProps> = ({ children }) => {
  const pathname = useRouterState({ select: (s) => s.location.pathname });
  const navigate = useNavigate();
  const { user, logout } = useUserStore();
  const { theme, setTheme } = useThemeStore();
  const { setOpen: setCommandPaletteOpen } = useCommandPalette();
  const { data: notifications = [] } = useNotifications();
  const { data: addonsData } = useQuery<Array<{ id: string; isEnabled: boolean }>>({
    queryKey: ['addons'],
    queryFn: async () => {
      try {
        const res = await unwrap(
          safe.get<{ addons?: Array<{ id: string; isEnabled: boolean }> }>('/addons')
        );
        return res.addons ?? [];
      } catch {
        return [];
      }
    },
    staleTime: 60_000,
  });
  const menuAddonEnabled = (addonsData ?? []).some((a) => a.id === 'menu' && a.isEnabled);
  const sidebarRef = useRef<HTMLElement | null>(null);

  useForesightLinks(sidebarRef, {
    '/': () => import('../features/dashboard/Dashboard/VendorDashboard'),
    '/team': () => import('../features/team/Team'),
    '/apps': () => import('../features/addons/Addons'),
    '/menus': () => import('../features/menu/Menus'),
    '/orders': () => import('../features/orders/Orders'),
    '/qr-codes': () => import('../features/qr-codes/QrCodes'),
    '/design': () => import('../features/design/Design'),
    '/analytics': () => import('../features/analytics/Analytics'),
  });

  useKeyboardShortcuts();
  useOnboardingTour();

  const { isOpen: caretOpen } = useCaretStore();
  const caretWidth = useCaretSettingsStore((s) => s.panelWidth);
  const userId = Number(user?.id) || 0;
  const unreadCount = Array.isArray(notifications)
    ? notifications.filter((n) => !n.read).length
    : 0;
  const pageTitle = getPageTitle(pathname);
  const userEmail = user?.email || null;
  const userName = user?.fullName || 'User';
  const businessName = user?.businessName || 'My Business';
  const isConsoleAdmin = user?.isConsoleAdmin === true || user?.is_console_admin === true;

  const handleLogout = (): void => {
    useBusinessStore.getState().reset();
    useConfigStore.getState().reset();
    logout();
    signOut('dash');
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
            className="flex h-14 items-center gap-2.5 px-4"
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

          <BusinessSwitcher />

          {/* Navigation */}
          <nav className="flex-1 overflow-y-auto py-2 px-2 space-y-0.5">
            {navItems.map((item) => (
              <SidebarNavItem key={item.path} {...item} />
            ))}

            {/* Storefront section - only shown when menu addon is enabled */}
            {menuAddonEnabled && (
              <SidebarSection label="Storefront" icon={Store} items={menuSubItems} />
            )}

            <AppsSection />
          </nav>

          {/* Footer */}
          <div className="p-2" style={{ borderTop: '1px solid var(--dash-border)' }}>
            <Link
              to="/settings/account"
              data-tour="settings"
              className={cn(
                'flex items-center gap-3 px-3 py-2 rounded-lg text-sm transition-colors duration-150 mb-0.5',
                pathname.startsWith('/settings')
                  ? 'bg-[var(--dash-active)]'
                  : 'hover:bg-[var(--dash-hover)]'
              )}
              style={{
                color: pathname.startsWith('/settings')
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
                    src={getAvatarUrl(userEmail || userName)}
                    alt={userName || 'User'}
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
                {isConsoleAdmin && (
                  <>
                    <DropdownMenuItem
                      onClick={() => {
                        window.location.href = APP_URLS.console;
                      }}
                    >
                      <ExternalLink size={14} />
                      Open console
                    </DropdownMenuItem>
                    <DropdownMenuSeparator />
                  </>
                )}
                <DropdownMenuItem destructive onClick={handleLogout}>
                  <LogOut size={14} />
                  Sign out
                </DropdownMenuItem>
              </DropdownMenuContent>
            </DropdownMenu>
          </div>
        </aside>

        {/* Main content + Caret AI inline column */}
        <div className="flex flex-1 overflow-hidden">
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
                  style={{
                    color: 'var(--dash-text-muted)',
                    border: '1px solid var(--dash-border)',
                  }}
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
                <CaretButton data-tour="caret" />
                <button
                  onClick={() => navigate({ to: '/notifications' })}
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
                key={pathname}
                initial={{ opacity: 0, y: 12, filter: 'blur(4px)' }}
                animate={{ opacity: 1, y: 0, filter: 'blur(0px)' }}
                exit={{ opacity: 0, y: -8, filter: 'blur(2px)' }}
                transition={{ duration: 0.28, ease: [0.22, 1, 0.36, 1] }}
                className="flex-1 overflow-y-auto p-6"
              >
                {children}
              </motion.main>
            </AnimatePresence>
          </div>

          {/* Caret AI - inline third column, pushes content left */}
          <div
            style={{
              width: caretOpen ? `${caretWidth}px` : '0px',
              flexShrink: 0,
              overflow: 'visible',
              transition: caretOpen ? 'none' : 'width 0.25s cubic-bezier(0.4, 0, 0.2, 1)',
            }}
          >
            <CaretPanel
              variant="inline"
              appId="dash"
              userId={userId}
              apiBaseUrl={import.meta.env.VITE_API_URL ?? 'https://api.voxire.com/'}
              systemContext="User is in the Dash app (business command center - KPIs, orders, staff overview)"
            />
          </div>
        </div>
      </div>
    </>
  );
};

export default DefaultLayout;
