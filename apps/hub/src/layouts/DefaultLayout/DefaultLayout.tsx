import React from 'react';
import { useLocation, useNavigate, Link } from 'react-router-dom';
import { motion, AnimatePresence } from 'framer-motion';
import { useQuery } from '@tanstack/react-query';
import api from '../../lib/api';
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
import { useUserStore } from '../../stores/userStore';
import { useThemeStore } from '../../stores/themeStore';
import { cn } from '../../lib/utils';
import { Avatar, DropdownMenu, DropdownMenuTrigger, DropdownMenuContent, DropdownMenuItem, DropdownMenuSeparator } from '../../components/ui';
import { SearchCommand } from '../../components/hub';
import { useCommandPalette } from '../../stores/commandPaletteStore';
import { useKeyboardShortcuts } from '../../hooks/useKeyboardShortcuts';
import { useOnboardingTour } from '../../hooks/useOnboardingTour';
import { useNotifications } from '../../hooks/queries/useNotifications';

interface NavItem {
  label: string;
  path: string;
  icon: React.ComponentType<{ size?: number; className?: string }>;
}

const navItems: NavItem[] = [
  { label: 'Dashboard',     path: '/',                  icon: LayoutDashboard },
  { label: 'Team',          path: '/team',              icon: Users },
  { label: 'Apps',          path: '/apps',              icon: Grid3x3 },
  { label: 'Activity',      path: '/activity',          icon: ActivityIcon },
];

const menuSubItems: NavItem[] = [
  { label: 'Menus',         path: '/menus',             icon: UtensilsCrossed },
  { label: 'Orders',        path: '/orders',            icon: ClipboardList },
  { label: 'QR Codes',      path: '/qr-codes',          icon: QrCode },
  { label: 'Design',        path: '/design',            icon: Palette },
];

const routeLabels: Record<string, string> = {
  '/':                       'Dashboard',
  '/chat':                   '__caret__',
  '/menus':                  'Menus',
  '/orders':                 'Orders',
  '/qr-codes':               'QR Codes',
  '/design':                 'Design',
  '/analytics':              'Analytics',
  '/menu':                   'Menu',
  '/create-item':            'Create Item',
  '/create-banner':          'Create Banner',
  '/team':                   'Team',
  '/team/invite':            'Invite member',
  '/permissions':            'Permissions',
  '/roles':                  'Roles',
  '/apps':                   'Apps',
  '/activity':               'Activity',
  '/audit':                  'Audit log',
  '/notifications':          'Notifications',
  '/account':                'Account',
  '/billing':                'Billing',
  '/integrations':           'Integrations',
  '/api-keys':               'API Keys',
  '/webhooks':               'Webhooks',
  '/security':               'Security',
  '/marketplace':            'Marketplace',
  '/changelog':              "What's new",
  '/flags':                  'Feature flags',
  '/theme':                  'Theme',
  '/referrals':              'Referrals',
  '/help':                   'Help center',
  '/export':                 'Export data',
  '/import':                 'Import data',
  '/settings':               'Settings',
  '/settings/account':       'Account',
  '/settings/general':       'General',
  '/settings/industry':      'Industry',
  '/settings/domain':        'Custom Domain',
  '/settings/email':         'Email Settings',
  '/settings/billing':       'Billing',
  '/settings/addons':        'Addons',
};

function getPageTitle(pathname: string): string {
  if (routeLabels[pathname]) return routeLabels[pathname];
  const match = Object.keys(routeLabels).find(
    (key) => pathname.startsWith(key) && key !== '/'
  );
  return match ? routeLabels[match] : 'Hub';
}

function getInitials(name?: string | null): string {
  if (!name) return '?';
  return name.split(' ').map((n) => n[0]).join('').toUpperCase().slice(0, 2);
}

const SidebarNavItem: React.FC<NavItem> = ({ label, path, icon: Icon }) => {
  const location = useLocation();
  const isActive =
    path === '/' ? location.pathname === '/' : location.pathname.startsWith(path);

  return (
    <Link
      to={path}
      data-tour={label.toLowerCase()}
      className={cn(
        'relative flex items-center gap-3 px-3 py-2 rounded-lg text-sm font-medium transition-colors duration-150',
        isActive ? '' : 'hover:bg-[var(--hub-hover)]'
      )}
      style={{ color: isActive ? 'var(--hub-text)' : 'var(--hub-text-muted)' }}
    >
      {isActive && (
        <motion.span
          layoutId="sidebar-active"
          className="absolute inset-0 rounded-lg"
          style={{ background: 'var(--hub-active)' }}
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

const SidebarSection: React.FC<{ label: string; icon: React.ComponentType<{ size?: number; className?: string }>; items: NavItem[] }> = ({ label, icon: Icon, items }) => {
  const location = useLocation();
  const isActive = items.some((item) => location.pathname.startsWith(item.path));
  const [open, setOpen] = React.useState(isActive);

  return (
    <div className="pt-1">
      <button
        onClick={() => setOpen((v) => !v)}
        className={cn(
          'flex items-center gap-3 px-3 py-2 rounded-lg text-sm w-full transition-colors duration-150',
          isActive ? '' : 'hover:bg-[var(--hub-hover)]'
        )}
        style={{ color: isActive ? 'var(--hub-text)' : 'var(--hub-text-muted)' }}
      >
        <Icon size={16} className="shrink-0" />
        <span className="flex-1 text-left">{label}</span>
        <ChevronRight size={12} className={cn('shrink-0 transition-transform', open && 'rotate-90')} style={{ color: 'var(--hub-text-dim)' }} />
      </button>
      {open && (
        <div className="ml-5 pl-2 mt-0.5 space-y-0.5" style={{ borderLeft: '1px solid var(--hub-border)' }}>
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
  const { data: addonsData } = useQuery({
    queryKey: ['addons'],
    queryFn: async () => {
      const res = await api.get('/addons');
      return res.data?.addons ?? [];
    },
    staleTime: 60_000,
  });
  const menuAddonEnabled = (addonsData ?? []).some((a: any) => a.id === 'menu' && a.isEnabled);

  useKeyboardShortcuts();
  useOnboardingTour();

  const unreadCount = Array.isArray(notifications) ? notifications.filter((n: any) => !n.read).length : 0;
  const pageTitle = getPageTitle(location.pathname);
  const userEmail = user?.email || user?.emailAddress || null;
  const userName = user?.name || user?.firstName || userEmail || 'User';
  const businessName = user?.businessName || user?.storeName || 'My Business';

  const handleLogout = () => {
    logout();
    window.location.href = APP_URLS.auth;
  };

  return (
    <div className="flex h-screen overflow-hidden" style={{ background: 'var(--hub-bg)' }}>
      {/* Sidebar */}
      <aside
        className="w-64 flex flex-col shrink-0 overflow-hidden"
        style={{
          background: 'var(--hub-surface)',
          borderRight: '1px solid var(--hub-border)',
        }}
      >

        {/* Logo */}
        <div
          className="flex items-center gap-2.5 px-4 py-4"
          style={{ borderBottom: '1px solid var(--hub-border)' }}
        >
          <img
            src="/logo.png"
            alt="rtylr"
            className="h-6 w-auto"
            style={{ filter: (theme === 'dark' || (theme === 'system' && window.matchMedia('(prefers-color-scheme: dark)').matches)) ? 'brightness(0) invert(1)' : 'brightness(0)' }}
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
        <div className="p-2" style={{ borderTop: '1px solid var(--hub-border)' }}>
          <Link
            to="/settings/account"
            data-tour="settings"
            className={cn(
              'flex items-center gap-3 px-3 py-2 rounded-lg text-sm transition-colors duration-150 mb-0.5',
              location.pathname.startsWith('/settings') ? 'bg-[var(--hub-active)]' : 'hover:bg-[var(--hub-hover)]'
            )}
            style={{ color: location.pathname.startsWith('/settings') ? 'var(--hub-text)' : 'var(--hub-text-muted)' }}
          >
            <Settings size={16} />
            <span>Settings</span>
          </Link>

          <DropdownMenu>
            <DropdownMenuTrigger>
              <button
                className="flex items-center gap-2.5 w-full px-3 py-2 rounded-lg transition-colors duration-150 hover:bg-[var(--hub-hover)]"
              >
                <Avatar
                  initials={getInitials(userName)}
                  size="sm"
                  className="shrink-0"
                  style={{ background: 'rgba(69,131,196,0.15)', color: '#4583C4' }}
                />
                <div className="flex flex-col items-start min-w-0 flex-1">
                  <span className="text-xs font-medium truncate w-full text-left" style={{ color: 'var(--hub-text)' }}>
                    {userName}
                  </span>
                  {userEmail && (
                    <span className="text-[10px] truncate w-full text-left" style={{ color: 'var(--hub-text-muted)' }}>
                      {userEmail}
                    </span>
                  )}
                </div>
                <ChevronRight size={12} className="shrink-0" style={{ color: 'var(--hub-text-dim)' }} />
              </button>
            </DropdownMenuTrigger>
            <DropdownMenuContent align="start" side="top" className="w-52">
              {/* Theme selector */}
              <div className="px-2 py-1.5">
                <p className="text-[10px] font-semibold uppercase tracking-wider mb-1.5" style={{ color: 'var(--hub-text-dim)' }}>
                  Appearance
                </p>
                <div className="flex gap-1">
                  {([
                    { value: 'light',  icon: Sun,     label: 'Light'  },
                    { value: 'dark',   icon: Moon,    label: 'Dark'   },
                    { value: 'system', icon: Monitor, label: 'System' },
                  ] as const).map(({ value, icon: Icon, label }) => (
                    <button
                      key={value}
                      onClick={() => setTheme(value)}
                      className="flex-1 flex flex-col items-center gap-1 py-2 rounded-lg text-[10px] font-medium transition-colors"
                      style={{
                        background: theme === value ? 'rgba(69,131,196,0.15)' : 'var(--hub-hover)',
                        color: theme === value ? '#4583C4' : 'var(--hub-text-muted)',
                        border: theme === value ? '1px solid rgba(69,131,196,0.3)' : '1px solid transparent',
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
            background: 'var(--hub-surface)',
            borderBottom: '1px solid var(--hub-border)',
          }}
        >
          {pageTitle === '__caret__' ? (
            <div className="flex items-center gap-2">
              <div className="h-6 w-6 rounded-lg flex items-center justify-center" style={{ background: 'rgba(139,92,246,0.1)' }}>
                <Sparkles size={13} style={{ color: '#8b5cf6' }} />
              </div>
              <span className="text-sm font-semibold" style={{ color: 'var(--hub-text)' }}>Caret</span>
              <span className="text-[9px] font-bold uppercase tracking-wider px-1.5 py-0.5 rounded" style={{ background: 'rgba(139,92,246,0.1)', color: '#8b5cf6' }}>AI</span>
            </div>
          ) : (
            <h1 className="text-sm font-semibold" style={{ color: 'var(--hub-text)' }}>
              {pageTitle}
            </h1>
          )}
          <div className="flex items-center gap-1">
            <button
              onClick={() => setCommandPaletteOpen(true)}
              className="flex items-center gap-2 px-2.5 h-7 rounded-lg transition-colors duration-150 hover:bg-[var(--hub-hover)]"
              style={{ color: 'var(--hub-text-muted)', border: '1px solid var(--hub-border)' }}
              aria-label="Open command palette"
            >
              <Search size={12} />
              <span className="text-[11px]">Search</span>
              <kbd
                className="hidden sm:inline-flex items-center px-1 py-0.5 rounded text-[9px] font-mono"
                style={{ background: 'var(--hub-hover)', color: 'var(--hub-text-dim)' }}
              >
                ⌘K
              </kbd>
            </button>
            <button
              onClick={() => navigate('/chat')}
              className="h-7 rounded-lg flex items-center gap-1.5 px-3 transition-all duration-150"
              style={{ background: 'rgba(139,92,246,0.1)', color: '#8b5cf6', border: '1px solid rgba(139,92,246,0.2)' }}
              onMouseEnter={(e) => { (e.currentTarget as HTMLElement).style.background = 'rgba(139,92,246,0.18)'; }}
              onMouseLeave={(e) => { (e.currentTarget as HTMLElement).style.background = 'rgba(139,92,246,0.1)'; }}
              aria-label="Chat with Caret"
              data-tour="caret"
            >
              <Sparkles size={13} />
              <span className="text-[11px] font-semibold">Caret AI</span>
            </button>
            <button
              onClick={() => navigate('/notifications')}
              className="relative h-7 w-7 rounded-lg flex items-center justify-center transition-colors duration-150 hover:bg-[var(--hub-hover)]"
              style={{ color: 'var(--hub-text-muted)' }}
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
  );
};

export default DefaultLayout;
