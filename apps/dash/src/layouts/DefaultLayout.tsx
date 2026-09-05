import React, { useRef } from 'react';
import { useNavigate, Link, useRouterState } from '@tanstack/react-router';
import { useForesightLinks } from '@voxire/utils';
import { motion, AnimatePresence } from 'motion/react';
import { APP_URLS, getAvatarUrl, resolveAppUrl } from '@voxire/config';
import { signOut } from '@voxire/auth';
import {
  LayoutDashboard,
  UtensilsCrossed,
  QrCode,
  Palette,
  Settings,
  Bell,
  LogOut,
  ChevronRight,
  ChevronsUpDown,
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
  Menu,
  X,
  Globe,
  Check,
  ChevronDown,
} from 'lucide-react';
import { SUPPORTED_LANGUAGES } from '@voxire/i18n';
import { selectLanguage } from '../shared/services/i18next';
import { CaretPanel, CaretButton, useCaretStore, useCaretSettingsStore } from '@voxire/caret';
import { LicenseBanner } from '@voxire/components';
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
import { isAddonEnabled, useAddons } from '../shared/hooks/queries/useAddons';
import BusinessSwitcher from '../features/business-switcher/BusinessSwitcher';
import { buildDashPageContext, buildDashSystemContext } from '../shared/lib/dashCaretContext';
import { useTranslation } from 'react-i18next';
import { licenseClient } from '../shared/lib/clients';

interface NavItem {
  key: string;
  label: string;
  path: string;
  icon: React.ComponentType<{ size?: number; className?: string }>;
  badge?: number | string;
}

const rawNavItems: NavItem[] = [
  { key: 'dashboard', label: 'Account Home', path: '/', icon: LayoutDashboard },
  { key: 'team', label: 'Team', path: '/team', icon: Users },
  { key: 'apps', label: 'Apps', path: '/apps', icon: Grid3x3 },
];

const rawMenuSubItems: NavItem[] = [
  { key: 'menus', label: 'Menus', path: '/menus', icon: UtensilsCrossed },
  { key: 'qrCodes', label: 'QR Codes', path: '/qr-codes', icon: QrCode },
  { key: 'design', label: 'Design', path: '/design', icon: Palette },
];

const routeKeys: Record<string, { key: string; fallback: string }> = {
  '/': { key: 'dashboard', fallback: 'Account Home' },
  '/chat': { key: 'chat', fallback: '__caret__' },
  '/menus': { key: 'menus', fallback: 'Menus' },
  '/orders': { key: 'orders', fallback: 'Orders' },
  '/qr-codes': { key: 'qrCodes', fallback: 'QR Codes' },
  '/design': { key: 'design', fallback: 'Design' },
  '/menu': { key: 'menu', fallback: 'Menu' },
  '/create-item': { key: 'createItem', fallback: 'Create Item' },
  '/create-banner': { key: 'createBanner', fallback: 'Create Banner' },
  '/team': { key: 'team', fallback: 'Team' },
  '/team/invite': { key: 'addMember', fallback: 'Add a member' },
  '/permissions': { key: 'permissions', fallback: 'Permissions' },
  '/roles': { key: 'roles', fallback: 'Roles' },
  '/apps': { key: 'apps', fallback: 'Apps' },
  '/audit': { key: 'auditLog', fallback: 'Audit log' },
  '/notifications': { key: 'notifications', fallback: 'Notifications' },
  '/account': { key: 'account', fallback: 'Account' },
  '/billing': { key: 'billing', fallback: 'Billing' },
  '/integrations': { key: 'integrations', fallback: 'Integrations' },
  '/api-keys': { key: 'apiKeys', fallback: 'API Keys' },
  '/webhooks': { key: 'webhooks', fallback: 'Webhooks' },
  '/security': { key: 'security', fallback: 'Security' },
  '/theme': { key: 'theme', fallback: 'Theme' },
  '/help': { key: 'helpCenter', fallback: 'Help center' },
  '/settings': { key: 'manageAccount', fallback: 'Manage account' },
  '/settings/account': { key: 'account', fallback: 'Account' },
  '/settings/general': { key: 'general', fallback: 'General' },
  '/settings/industry': { key: 'industry', fallback: 'Industry' },
  '/settings/domain': { key: 'customDomain', fallback: 'Custom Domain' },
  '/settings/email': { key: 'emailSettings', fallback: 'Email Settings' },
  '/settings/billing': { key: 'billing', fallback: 'Billing' },
  '/settings/subscription': { key: 'billing', fallback: 'Billing' },
  '/settings/addons': { key: 'addons', fallback: 'Addons' },
};

function getPageTitle(pathname: string, t: (key: string, options?: any) => string): string {
  const matchEntry = routeKeys[pathname]
    ? routeKeys[pathname]
    : Object.keys(routeKeys).find((key) => pathname.startsWith(key) && key !== '/')
      ? routeKeys[Object.keys(routeKeys).find((key) => pathname.startsWith(key) && key !== '/')!]
      : null;

  if (matchEntry) {
    if (matchEntry.key === 'chat') return '__caret__';
    return t(`nav.${matchEntry.key}`, matchEntry.fallback);
  }
  return 'Dash';
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

const SidebarNavItem: React.FC<NavItem> = ({ label, path, icon: Icon, badge }) => {
  const pathname = useRouterState({ select: (s) => s.location.pathname });
  const isActive = path === '/' ? pathname === '/' : pathname.startsWith(path);

  return (
    <Link
      to={path}
      data-tour={label.toLowerCase()}
      title={label}
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
        style={{ color: isActive ? 'var(--accent)' : undefined }}
        whileHover={{ scale: 1.18, rotate: -4 }}
        transition={{ type: 'spring', stiffness: 500, damping: 18 }}
      >
        <Icon size={16} />
      </motion.span>
      <span className="dash-sidebar-label relative flex-1 truncate">{label}</span>
      {badge !== undefined && badge !== null && badge !== '' && (
        <span
          className="dash-sidebar-label relative shrink-0 rounded px-1.5 py-0.5 text-[10px] font-medium leading-none"
          style={{ background: 'var(--dash-hover)', color: 'var(--dash-text-dim)' }}
        >
          {badge}
        </span>
      )}
      {isActive && (
        <motion.span
          className="relative ml-auto h-1 w-1 rounded-full shrink-0"
          style={{ background: 'var(--accent)' }}
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
        type="button"
        onClick={() => setOpen((v) => !v)}
        className={cn(
          'flex items-center gap-3 px-3 py-2 rounded-lg text-sm w-full transition-colors duration-150',
          isActive ? '' : 'hover:bg-[var(--dash-hover)]'
        )}
        style={{ color: isActive ? 'var(--dash-text)' : 'var(--dash-text-muted)' }}
      >
        <Icon size={16} className="shrink-0" />
        <span className="dash-sidebar-label flex-1 text-left">{label}</span>
        <ChevronRight
          size={12}
          className={cn(
            'dash-sidebar-chevron shrink-0 transition-transform rtl:rotate-180',
            open && 'rotate-90'
          )}
          style={{ color: 'var(--dash-text-dim)' }}
        />
      </button>
      {open && (
        <div
          className="ms-5 ps-2 mt-0.5 space-y-0.5"
          style={{ borderInlineStart: '1px solid var(--dash-border)' }}
        >
          {items.map((item) => (
            <SidebarNavItem key={item.path} label={item.label} path={item.path} icon={item.icon} />
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
  const effectiveAccess = user?.effectiveAccess;
  const isOwner = effectiveAccess?.isOwner ?? (user?.role === 'owner' || user?.role === 'admin');

  const visible = apps.filter((a) => {
    if (!(a.key in APP_URLS)) return false;
    if (a.status !== 'active') return false;
    if (effectiveAccess) {
      return effectiveAccess.apps.some(
        (app) => app.appId === a.key && app.licensed && app.hasAccess
      );
    }
    if (isOwner) return userApps.length === 0 || userApps.includes(a.key);
    return userApps.includes(a.key);
  });

  if (visible.length === 0) return null;

  return (
    <div className="pt-3">
      <div
        className="dash-sidebar-section-label px-3 pb-1.5 text-xs font-semibold uppercase tracking-wide"
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
              href={resolveAppUrl(a.key as Parameters<typeof resolveAppUrl>[0])}
              target="_blank"
              rel="noopener noreferrer"
              title={a.label}
              className="flex items-center gap-3 px-3 py-2 rounded-lg text-sm transition-colors duration-150 hover:bg-[var(--dash-hover)]"
              style={{ color: 'var(--dash-text-muted)' }}
            >
              <span
                className="h-6 w-6 rounded-md flex items-center justify-center shrink-0"
                style={{ background: a.bg }}
              >
                <Icon size={13} style={{ color: a.color }} />
              </span>
              <span className="dash-sidebar-label flex-1 truncate">{a.label}</span>
              <ExternalLink
                size={11}
                className="dash-sidebar-external"
                style={{ color: 'var(--dash-text-dim)' }}
              />
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

// Language codes dash ships locale resources for. To support a new language,
// add its code here and add the matching locales/{code}/*.json files.
const DASH_LANGUAGE_CODES = ['en', 'fr', 'ar'] as const;

const DASH_LANGUAGES = SUPPORTED_LANGUAGES.filter((l) =>
  (DASH_LANGUAGE_CODES as readonly string[]).includes(l.code)
);

const FALLBACK_LANGUAGE = { code: 'en', name: 'English', nativeName: 'English', rtl: false };

const LanguageSelector = (): React.JSX.Element => {
  const { t, i18n } = useTranslation('common');
  const [open, setOpen] = React.useState(false);
  const containerRef = useRef<HTMLDivElement | null>(null);

  const current =
    DASH_LANGUAGES.find((l) => l.code === i18n.language) ?? DASH_LANGUAGES[0] ?? FALLBACK_LANGUAGE;

  React.useEffect(() => {
    if (!open) return;
    const handleClickOutside = (e: MouseEvent): void => {
      if (containerRef.current && !containerRef.current.contains(e.target as Node)) {
        setOpen(false);
      }
    };
    document.addEventListener('mousedown', handleClickOutside);
    return () => document.removeEventListener('mousedown', handleClickOutside);
  }, [open]);

  const handleSelect = (code: string): void => {
    selectLanguage(code);
    setOpen(false);
  };

  return (
    <div className="px-2 py-1.5">
      <p
        className="text-[10px] font-semibold uppercase tracking-wider mb-1.5"
        style={{ color: 'var(--dash-text-dim)' }}
      >
        {t('common.nav.language', 'Language')}
      </p>
      <div ref={containerRef} className="relative">
        <button
          type="button"
          onClick={() => setOpen((v) => !v)}
          aria-haspopup="menu"
          aria-expanded={open}
          className="flex h-11 w-full items-center gap-2 rounded-lg px-3 text-xs font-medium transition-colors hover:bg-[var(--dash-hover)]"
          style={{ background: 'var(--dash-hover)', color: 'var(--dash-text)' }}
        >
          <Globe size={13} className="shrink-0" style={{ color: 'var(--dash-text-muted)' }} />
          <span className="flex-1 text-start">{current.nativeName}</span>
          <ChevronDown
            size={12}
            className={cn('shrink-0 transition-transform', open && 'rotate-180')}
            style={{ color: 'var(--dash-text-dim)' }}
          />
        </button>
        {open && (
          <div
            role="menu"
            aria-label={t('common.nav.language', 'Language')}
            className="absolute start-0 end-0 top-full z-10 mt-1 max-h-48 overflow-y-auto rounded-xl py-1"
            style={{
              background: 'var(--dash-card)',
              border: '1px solid var(--dash-border)',
              boxShadow: 'var(--dash-shadow)',
            }}
          >
            {DASH_LANGUAGES.map((lang) => {
              const active = lang.code === current.code;
              return (
                <DropdownMenuItem
                  key={lang.code}
                  onClick={() => handleSelect(lang.code)}
                  className="min-h-[44px]"
                  style={active ? { color: 'var(--accent)' } : undefined}
                >
                  <span className="flex-1">{lang.nativeName}</span>
                  {active && <Check size={13} className="shrink-0" />}
                </DropdownMenuItem>
              );
            })}
          </div>
        )}
      </div>
    </div>
  );
};

const DefaultLayout: React.FC<DefaultLayoutProps> = ({ children }) => {
  const pathname = useRouterState({ select: (s) => s.location.pathname });
  const search = useRouterState({ select: (s) => s.location.searchStr });
  const navigate = useNavigate();
  const { user, logout } = useUserStore();
  const { theme, setTheme } = useThemeStore();
  const { setOpen: setCommandPaletteOpen } = useCommandPalette();
  const { data: notifications = [] } = useNotifications();
  const { data: addonsData = [] } = useAddons();
  const menuAddonEnabled = isAddonEnabled(addonsData, 'menu');
  const sidebarRef = useRef<HTMLElement | null>(null);

  useForesightLinks(sidebarRef, {
    '/': () => import('../features/dashboard/Dashboard/VendorDashboard'),
    '/team': () => import('../features/team/Team'),
    '/apps': () => import('../features/addons/Addons'),
    '/menus': () => import('../features/menu/Menus'),
    '/orders': () => import('../features/orders/Orders'),
    '/qr-codes': () => import('../features/qr-codes/QrCodes'),
    '/design': () => import('../features/design/Design'),
  });

  useKeyboardShortcuts();
  useOnboardingTour();

  const { isOpen: caretOpen } = useCaretStore();
  const caretWidth = useCaretSettingsStore((s) => s.panelWidth);
  const userId = Number(user?.id) || 0;
  const unreadCount = Array.isArray(notifications)
    ? notifications.filter((n) => !n.read).length
    : 0;
  const { t } = useTranslation('common');
  const pageTitle = getPageTitle(pathname, t);

  const navItems = rawNavItems.map((item) => ({
    ...item,
    label: t(`nav.${item.key}`, item.label),
  }));

  const menuSubItems = rawMenuSubItems.map((item) => ({
    ...item,
    label: t(`nav.${item.key}`, item.label),
  }));
  const userEmail = user?.email || null;
  const userName = user?.fullName || 'User';
  const businessName = user?.businessName || 'My Business';
  const isConsoleAdmin = user?.isConsoleAdmin === true;

  const caretSystemContext = React.useMemo(
    () =>
      buildDashSystemContext({
        pathname,
        search,
        businessName: user?.businessName,
        role: user?.role,
        apps: user?.apps,
      }),
    [pathname, search, user?.apps, user?.businessName, user?.role]
  );
  const caretPageContext = React.useMemo(
    () =>
      buildDashPageContext({
        pathname,
        search,
        businessName: user?.businessName,
        role: user?.role,
        apps: user?.apps,
      }),
    [pathname, search, user?.apps, user?.businessName, user?.role]
  );

  const [mobileSidebarOpen, setMobileSidebarOpen] = React.useState(false);

  const handleLogout = (): void => {
    useBusinessStore.getState().reset();
    useConfigStore.getState().reset();
    logout();
    signOut('dash');
  };

  return (
    <>
      <a
        href="#main-content"
        className="sr-only focus:not-sr-only focus:absolute focus:z-50 focus:p-3 focus:bg-[var(--dash-surface)] focus:text-[var(--dash-text)] focus:border focus:border-[var(--dash-border)] focus:rounded-lg focus:shadow-lg focus:top-2 focus:left-2"
      >
        Skip to main content
      </a>
      <div
        className="dash-shell flex h-screen overflow-hidden relative"
        style={{ background: 'var(--dash-bg)' }}
      >
        {/* Mobile sidebar overlay backdrop */}
        {mobileSidebarOpen && (
          <div
            className="fixed inset-0 bg-black/60 backdrop-blur-sm z-30 md:hidden"
            onClick={() => setMobileSidebarOpen(false)}
          />
        )}

        {/* Sidebar */}
        <aside
          className={cn(
            'dash-sidebar w-64 flex flex-col shrink-0 overflow-hidden transition-transform duration-200 z-40',
            'fixed md:relative inset-y-0 start-0 h-full',
            mobileSidebarOpen ? 'translate-x-0' : '-translate-x-full md:translate-x-0'
          )}
          style={{
            background: 'var(--dash-surface)',
            borderInlineEnd: '1px solid var(--dash-border)',
          }}
        >
          {/* Logo */}
          <div
            className="dash-sidebar-logo flex h-14 items-center gap-2.5 px-4"
            style={{ borderBottom: '1px solid var(--dash-border)' }}
          >
            <img
              src={
                theme === 'dark' ||
                (theme === 'system' && window.matchMedia('(prefers-color-scheme: dark)').matches)
                  ? '/logo.png'
                  : '/logo-dark.png'
              }
              alt="rtylr"
              className="h-9 w-auto"
            />
          </div>

          <BusinessSwitcher />

          {/* Navigation */}
          <nav className="flex-1 overflow-y-auto py-2 px-2 space-y-0.5">
            {navItems.map((item) => (
              <SidebarNavItem
                key={item.path}
                label={item.label}
                path={item.path}
                icon={item.icon}
              />
            ))}

            {/* Storefront section - only shown when menu addon is enabled */}
            {menuAddonEnabled && (
              <SidebarSection
                label={t('common.nav.storefront', 'Storefront')}
                icon={Store}
                items={menuSubItems}
              />
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
              <span className="dash-sidebar-label">
                {t('common.nav.manageAccount', 'Manage account')}
              </span>
            </Link>

            <DropdownMenu>
              <DropdownMenuTrigger>
                <button
                  type="button"
                  className="flex items-center gap-2.5 w-full px-3 py-2 rounded-lg transition-colors duration-150 hover:bg-[var(--dash-hover)]"
                >
                  <Avatar
                    src={getAvatarUrl(userEmail || userName)}
                    alt={userName || 'User'}
                    initials={getInitials(userName)}
                    size="sm"
                    className="shrink-0"
                    style={{ background: 'var(--accent-muted)', color: 'var(--accent)' }}
                  />
                  <div className="dash-sidebar-user flex flex-col items-start min-w-0 flex-1">
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
                  <ChevronsUpDown
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
                        type="button"
                        key={value}
                        onClick={() => setTheme(value)}
                        className="flex-1 flex flex-col items-center gap-1 py-2 rounded-lg text-[10px] font-medium transition-colors"
                        style={{
                          background: theme === value ? 'var(--accent-muted)' : 'var(--dash-hover)',
                          color: theme === value ? 'var(--accent)' : 'var(--dash-text-muted)',
                          border:
                            theme === value
                              ? '1px solid var(--accent-border)'
                              : '1px solid transparent',
                        }}
                      >
                        <Icon size={13} />
                        {label}
                      </button>
                    ))}
                  </div>
                </div>
                <LanguageSelector />
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
        <div className="flex flex-1 min-w-0 overflow-hidden">
          <div className="flex flex-col flex-1 min-w-0 overflow-hidden">
            <LicenseBanner client={licenseClient} />
            {/* Topbar */}
            <header
              className="dash-topbar h-11 flex items-center justify-between px-4 sm:px-6 shrink-0"
              style={{
                background: 'var(--dash-surface)',
                borderBottom: '1px solid var(--dash-border)',
              }}
            >
              <div className="flex items-center gap-2">
                <button
                  type="button"
                  onClick={() => setMobileSidebarOpen((v) => !v)}
                  className="md:hidden h-7 w-7 rounded-lg flex items-center justify-center transition-colors hover:bg-[var(--dash-hover)]"
                  style={{ color: 'var(--dash-text-muted)' }}
                  aria-label="Toggle navigation menu"
                >
                  <Menu size={16} />
                </button>
                {pageTitle === '__caret__' ? (
                  <div className="flex items-center gap-2">
                    <div
                      className="h-6 w-6 rounded-lg flex items-center justify-center"
                      style={{ background: 'rgba(139,92,246,0.1)' }}
                    >
                      <Sparkles size={13} style={{ color: 'var(--app-crm)' }} />
                    </div>
                    <span className="text-sm font-semibold" style={{ color: 'var(--dash-text)' }}>
                      Caret
                    </span>
                    <span
                      className="text-[9px] font-bold uppercase tracking-wider px-1.5 py-0.5 rounded"
                      style={{ background: 'rgba(139,92,246,0.1)', color: 'var(--app-crm)' }}
                    >
                      AI
                    </span>
                  </div>
                ) : (
                  <span className="text-sm font-semibold" style={{ color: 'var(--dash-text)' }}>
                    {pageTitle}
                  </span>
                )}
              </div>
              <div className="flex items-center gap-1">
                <button
                  type="button"
                  onClick={() => setCommandPaletteOpen(true)}
                  className="flex items-center gap-2 px-2.5 h-7 rounded-lg transition-colors duration-150 hover:bg-[var(--dash-hover)]"
                  style={{
                    color: 'var(--dash-text-muted)',
                    border: '1px solid var(--dash-border)',
                  }}
                  aria-label={t('common.nav.openCommandPalette', 'Open command palette')}
                >
                  <motion.span
                    className="inline-flex"
                    whileHover={{ scale: 1.18, rotate: -4 }}
                    transition={{ type: 'spring', stiffness: 500, damping: 18 }}
                  >
                    <Search size={12} />
                  </motion.span>
                  <span className="text-[11px]">{t('common.nav.search', 'Search')}</span>
                  <kbd
                    className="hidden sm:inline-flex items-center px-1 py-0.5 rounded text-[9px] font-mono"
                    style={{ background: 'var(--dash-hover)', color: 'var(--dash-text-dim)' }}
                  >
                    ⌘K
                  </kbd>
                </button>
                <CaretButton className="dash-caret-button" data-tour="caret" />
                <button
                  type="button"
                  onClick={() => navigate({ to: '/notifications' })}
                  className="relative h-7 w-7 rounded-lg flex items-center justify-center transition-colors duration-150 hover:bg-[var(--dash-hover)]"
                  style={{ color: 'var(--dash-text-muted)' }}
                  aria-label={
                    unreadCount > 0
                      ? t('common.nav.notificationsUnread', '{{count}} unread notifications', {
                          count: unreadCount,
                        })
                      : t('common.nav.notifications', 'Notifications')
                  }
                >
                  <Bell size={14} />
                  {unreadCount > 0 && (
                    <span
                      className="absolute -top-0.5 -right-0.5 flex h-3.5 min-w-3.5 items-center justify-center rounded-full px-0.5 text-[9px] font-semibold leading-none text-white"
                      style={{ background: 'var(--status-cancelled)' }}
                      aria-hidden="true"
                    >
                      {unreadCount > 9 ? '9+' : unreadCount}
                    </span>
                  )}
                </button>
              </div>
            </header>
            <SearchCommand />

            {/* Page content */}
            <AnimatePresence mode="wait">
              <motion.main
                id="main-content"
                key={pathname}
                initial={{ opacity: 0, y: 12, filter: 'blur(4px)' }}
                animate={{ opacity: 1, y: 0, filter: 'blur(0px)' }}
                exit={{ opacity: 0, y: -8, filter: 'blur(2px)' }}
                transition={{ duration: 0.28, ease: [0.22, 1, 0.36, 1] }}
                className="dash-page flex-1 min-w-0 overflow-y-auto p-6"
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
              systemContext={caretSystemContext}
              pageContext={caretPageContext}
            />
          </div>
        </div>
      </div>
    </>
  );
};

export default DefaultLayout;
