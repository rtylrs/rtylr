import React from 'react';
import { useLocation, useNavigate, Link } from 'react-router-dom';
import { motion, AnimatePresence } from 'framer-motion';
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
import { useNotifications } from '../../hooks/queries/useNotifications';

interface NavItem {
  label: string;
  path: string;
  icon: React.ComponentType<{ size?: number; className?: string }>;
}

const navItems: NavItem[] = [
  { label: 'Dashboard', path: '/',          icon: LayoutDashboard },
  { label: 'Team',      path: '/team',      icon: Users },
  { label: 'Apps',      path: '/apps',      icon: Grid3x3 },
  { label: 'Activity',  path: '/activity',  icon: ActivityIcon },
  { label: 'Menus',     path: '/menus',     icon: UtensilsCrossed },
  { label: 'Orders',    path: '/orders',    icon: ClipboardList },
  { label: 'QR Codes',  path: '/qr-codes',  icon: QrCode },
  { label: 'Design',    path: '/design',    icon: Palette },
  { label: 'Analytics', path: '/analytics', icon: BarChart3 },
];

const routeLabels: Record<string, string> = {
  '/':                       'Dashboard',
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

  useKeyboardShortcuts();

  const unreadCount = notifications.filter((n) => !n.read).length;
  const pageTitle = getPageTitle(location.pathname);
  const userEmail = user?.email || user?.emailAddress || null;
  const userName = user?.name || user?.firstName || userEmail || 'User';
  const businessName = user?.businessName || user?.storeName || 'My Business';

  const handleLogout = () => {
    logout();
    navigate('/auth/login', { replace: true });
  };

  return (
    <div className="flex h-screen overflow-hidden" style={{ background: 'var(--hub-bg)' }}>
      {/* Sidebar */}
      <aside
        className="w-56 flex flex-col shrink-0 overflow-hidden"
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
            style={{ filter: theme === 'dark' ? 'brightness(0) invert(1)' : 'brightness(0)' }}
          />
        </div>

        {/* Business context pill */}
        <div
          className="mx-3 mt-3 mb-1 px-3 py-2 rounded-lg flex items-center gap-2"
          style={{ background: 'var(--hub-hover)', border: '1px solid var(--hub-border)' }}
        >
          <span
            className="flex items-center justify-center h-5 w-5 rounded text-[10px] font-bold shrink-0"
            style={{ background: 'rgba(69,131,196,0.15)', color: '#4583C4' }}
          >
            {businessName.charAt(0).toUpperCase()}
          </span>
          <span className="text-xs font-medium truncate" style={{ color: 'var(--hub-text-muted)' }}>
            {businessName}
          </span>
        </div>

        {/* Navigation */}
        <nav className="flex-1 overflow-y-auto py-2 px-2 space-y-0.5">
          <p className="px-3 mb-1.5 mt-1 text-[10px] font-semibold uppercase tracking-widest" style={{ color: 'var(--hub-text-dim)' }}>
            Workspace
          </p>
          {navItems.map((item) => (
            <SidebarNavItem key={item.path} {...item} />
          ))}

          {/* Apps quick-jump */}
          <div className="pt-3 pb-1">
            <p className="px-3 mb-1.5 text-[10px] font-semibold uppercase tracking-widest" style={{ color: 'var(--hub-text-dim)' }}>
              Apps
            </p>
            {[
              { label: 'POS',     href: 'https://pos.rtylr.com',     icon: MonitorSpeaker, color: '#4583C4' },
              { label: 'ERP',     href: 'https://erp.rtylr.com',     icon: Package,        color: '#10b981' },
              { label: 'HR',      href: 'https://hr.rtylr.com',      icon: UserCheck,      color: '#f59e0b' },
              { label: 'Finance', href: 'https://finance.rtylr.com', icon: CreditCard,     color: '#ef4444' },
              { label: 'Tracker', href: 'https://tracker.rtylr.com', icon: Layers,         color: '#64C5B9' },
            ].map(({ label, href, icon: Icon, color }) => (
              <a
                key={label}
                href={href}
                target="_blank"
                rel="noopener noreferrer"
                className="group flex items-center gap-3 px-3 py-2 rounded-lg text-sm transition-colors duration-150 no-underline hover:bg-[var(--hub-hover)]"
                style={{ color: 'var(--hub-text-muted)' }}
              >
                <Icon size={15} style={{ color }} />
                <span className="flex-1 truncate">{label}</span>
                <ExternalLink size={11} className="opacity-0 group-hover:opacity-100 transition-opacity" style={{ color: 'var(--hub-text-dim)' }} />
              </a>
            ))}
          </div>
        </nav>

        {/* Footer */}
        <div className="p-2" style={{ borderTop: '1px solid var(--hub-border)' }}>
          <Link
            to="/settings"
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
              <DropdownMenuItem onClick={() => navigate('/settings/account')}>
                <Settings size={14} />
                Account Settings
              </DropdownMenuItem>
              <DropdownMenuSeparator />
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
          <h1 className="text-sm font-semibold" style={{ color: 'var(--hub-text)' }}>
            {pageTitle}
          </h1>
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
              onClick={() => navigate('/changelog')}
              className="h-7 w-7 rounded-lg flex items-center justify-center transition-colors duration-150 hover:bg-[var(--hub-hover)]"
              style={{ color: 'var(--hub-text-muted)' }}
              aria-label="What's new"
            >
              <Sparkles size={14} />
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
