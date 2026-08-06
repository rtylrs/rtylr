import React, { useEffect, useRef, useState } from 'react';
import { cn, Avatar } from '@voxire/ui';
import { useLocation, useRouterState, Link, Outlet } from '@tanstack/react-router';
import { useForesightLinks } from '@voxire/utils';
import { motion, AnimatePresence } from 'motion/react';
import { getAvatarUrl } from '@voxire/config';
import { signOut, useUserStore } from '@voxire/auth';
import { unwrap } from '@voxire/api';
import { LicenseBanner } from '@voxire/components';
import { buildCaretRouteContext, CaretPanel, CaretButton, useCaretStore } from '@voxire/caret';
import { safe } from '../shared/lib/api';
import {
  LayoutDashboard,
  TrendingUp,
  Users,
  Package,
  Briefcase,
  DollarSign,
  FileText,
  Settings as SettingsIcon,
  ChevronRight,
  Sun,
  Moon,
  Monitor,
  LogOut,
  Menu,
  X,
} from 'lucide-react';
import { Toaster } from 'sonner';
import { useTranslation } from 'react-i18next';

interface NavItem {
  label: string;
  path: string;
  icon: React.ComponentType<{ size?: number; className?: string }>;
}

const navItems: NavItem[] = [
  { label: 'Dashboard', path: '/', icon: LayoutDashboard },
  { label: 'Revenue', path: '/revenue', icon: TrendingUp },
  { label: 'Customers', path: '/customers', icon: Users },
  { label: 'Products', path: '/products', icon: Package },
  { label: 'Operations', path: '/operations', icon: Briefcase },
  { label: 'Financial', path: '/financial', icon: DollarSign },
  { label: 'Reports', path: '/reports', icon: FileText },
  { label: 'Settings', path: '/settings', icon: SettingsIcon },
];

const routeLabels: Record<string, string> = {
  '/': 'Dashboard',
  '/revenue': 'Revenue',
  '/customers': 'Customers',
  '/products': 'Products',
  '/operations': 'Operations',
  '/financial': 'Financial',
  '/reports': 'Reports',
  '/settings': 'Settings',
};

const ACCENT = '#6366f1';

function getPageTitle(pathname: string): string {
  if (routeLabels[pathname]) return routeLabels[pathname];
  const match = Object.keys(routeLabels).find((key) => pathname.startsWith(key) && key !== '/');
  return match ? routeLabels[match] : 'Insights';
}

type Theme = 'dark' | 'light' | 'system';

function applyTheme(theme: Theme) {
  const root = document.documentElement;
  if (theme === 'system') {
    const prefersDark = window.matchMedia('(prefers-color-scheme: dark)').matches;
    root.setAttribute('data-theme', prefersDark ? 'dark' : 'light');
  } else {
    root.setAttribute('data-theme', theme);
  }
}

const SidebarNavItem: React.FC<NavItem> = ({ label, path, icon: Icon }) => {
  const location = useLocation();
  const isActive = path === '/' ? location.pathname === '/' : location.pathname.startsWith(path);

  return (
    <Link
      to={path}
      className="relative flex items-center gap-3 px-3 py-1.5 rounded-lg text-sm font-medium transition-colors duration-150"
      style={{
        color: isActive ? 'var(--hub-text)' : 'var(--hub-text-muted)',
        textDecoration: 'none',
      }}
      onMouseEnter={(e) => {
        if (!isActive) (e.currentTarget as HTMLElement).style.background = 'var(--hub-hover)';
      }}
      onMouseLeave={(e) => {
        if (!isActive) (e.currentTarget as HTMLElement).style.background = 'transparent';
      }}
    >
      {isActive && (
        <motion.span
          layoutId="sidebar-active"
          className="absolute inset-0 rounded-lg"
          style={{ background: 'var(--hub-active)' }}
          transition={{ type: 'spring', stiffness: 400, damping: 35 }}
        />
      )}
      <Icon size={14} className="relative shrink-0" />
      <span className="relative">{label}</span>
      {isActive && (
        <span
          className="relative ml-auto h-1 w-1 rounded-full shrink-0"
          style={{ background: ACCENT }}
        />
      )}
    </Link>
  );
};

const DefaultLayout: React.FC = () => {
  const { t } = useTranslation();
  const location = useLocation();
  const search = useRouterState({ select: (s) => s.location.searchStr });
  const [theme, setThemeState] = useState<Theme>(() => {
    return (localStorage.getItem('rtylr-insights-theme') as Theme) || 'light';
  });
  const [showUserMenu, setShowUserMenu] = useState(false);
  const [sidebarOpen, setSidebarOpen] = useState(false);
  const insightsUser = useUserStore((s) => s.user);
  const setAuthUser = useUserStore((s) => s.setUser);
  const userId = Number(insightsUser?.id) || 0;

  useEffect(() => {
    if (insightsUser?.name || insightsUser?.email) return;
    unwrap(safe.get<{ user: { id: number; name: string; email: string } | null }>('/account'))
      .then((res) => {
        if (res?.user) {
          setAuthUser({ id: String(res.user.id), name: res.user.name, email: res.user.email });
        }
      })
      .catch(() => undefined);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [insightsUser]);

  const { isOpen: caretOpen } = useCaretStore();
  const pageTitle = getPageTitle(location.pathname);
  const caretPageContext = React.useMemo(
    () =>
      buildCaretRouteContext({
        appId: 'insights',
        appName: 'Insights',
        pathname: location.pathname,
        search,
        page: pageTitle,
        routeLabels,
      }),
    [location.pathname, search, pageTitle]
  );
  const sidebarRef = useRef<HTMLElement | null>(null);

  useForesightLinks(sidebarRef, {
    '/': () => import('../features/dashboard/Dashboard'),
    '/revenue': () => import('../features/revenue/Revenue'),
    '/customers': () => import('../features/customers/Customers'),
    '/products': () => import('../features/products/Products'),
    '/operations': () => import('../features/operations/Operations'),
    '/financial': () => import('../features/financial/Financial'),
    '/reports': () => import('../features/reports/Reports'),
    '/settings': () => import('../features/settings/Settings'),
  });

  useEffect(() => {
    applyTheme(theme);
  }, [theme]);

  useEffect(() => {
    setSidebarOpen(false);
  }, [location.pathname]);

  const userEmail = insightsUser?.email || insightsUser?.emailAddress || null;
  const userName = insightsUser?.name || insightsUser?.firstName || userEmail || 'User';

  const handleSetTheme = (t: Theme) => {
    setThemeState(t);
    applyTheme(t);
    localStorage.setItem('rtylr-insights-theme', t);
  };

  const handleLogout = (): void => {
    signOut('insights');
  };

  return (
    <>
      <div className="flex h-screen overflow-hidden" style={{ background: 'var(--hub-bg)' }}>
        {/* Mobile top bar */}
        <div
          className="fixed top-0 left-0 right-0 z-40 flex md:hidden items-center justify-between h-11 px-4 shrink-0"
          style={{
            background: 'var(--hub-surface)',
            borderBottom: '1px solid var(--hub-border)',
          }}
        >
          <button
            onClick={() => setSidebarOpen(true)}
            className="h-8 w-8 rounded-lg flex items-center justify-center"
            style={{ color: 'var(--hub-text-muted)' }}
          >
            <Menu size={18} />
          </button>
          <h1 className="text-sm font-semibold" style={{ color: 'var(--hub-text)' }}>
            {pageTitle}
          </h1>
          <div className="w-8" />
        </div>

        {/* Mobile overlay */}
        {sidebarOpen && (
          <div
            className="fixed inset-0 z-40 bg-black/50 md:hidden"
            onClick={() => setSidebarOpen(false)}
          />
        )}

        <aside
          className={cn(
            sidebarOpen ? 'fixed inset-y-0 left-0 z-50 flex' : 'hidden',
            'md:flex w-56 flex-col shrink-0 overflow-hidden'
          )}
          style={{
            background: 'var(--hub-surface)',
            borderRight: '1px solid var(--hub-border)',
          }}
        >
          <div
            className="flex items-center justify-between gap-2.5 px-4 py-4"
            style={{ borderBottom: '1px solid var(--hub-border)' }}
          >
            <img src="/logo.png?v=2" alt="rtylr" className="h-6 w-auto" />
            <button
              onClick={() => setSidebarOpen(false)}
              className="h-7 w-7 rounded-lg flex md:hidden items-center justify-center"
              style={{ color: 'var(--hub-text-muted)' }}
            >
              <X size={16} />
            </button>
          </div>

          <nav ref={sidebarRef} className="flex-1 overflow-y-auto py-2 px-2 space-y-0.5">
            {navItems.map((item) => (
              <SidebarNavItem key={item.path} {...item} />
            ))}
          </nav>

          <div className="p-2" style={{ borderTop: '1px solid var(--hub-border)' }}>
            <Link
              to="/settings"
              className="flex items-center gap-2 px-3 py-1.5 rounded-lg text-xs mb-1"
              style={{ color: 'var(--hub-text-muted)', textDecoration: 'none' }}
              onMouseEnter={(e) => {
                (e.currentTarget as HTMLElement).style.background = 'var(--hub-hover)';
              }}
              onMouseLeave={(e) => {
                (e.currentTarget as HTMLElement).style.background = 'transparent';
              }}
            >
              <SettingsIcon size={12} /> Settings
            </Link>
            <CaretButton className="w-full justify-start mb-1" />
            <div className="relative">
              <button
                onClick={() => setShowUserMenu((v) => !v)}
                className="flex items-center gap-2.5 w-full px-3 py-2 rounded-lg transition-colors duration-150"
                style={{ background: 'transparent' }}
                onMouseEnter={(e) => {
                  (e.currentTarget as HTMLElement).style.background = 'var(--hub-hover)';
                }}
                onMouseLeave={(e) => {
                  (e.currentTarget as HTMLElement).style.background = 'transparent';
                }}
              >
                <Avatar
                  src={getAvatarUrl(userEmail || userName)}
                  alt={userName}
                  size="sm"
                  className="h-6 w-6"
                  style={{ background: 'rgba(99,102,241,0.15)' }}
                />
                <div className="flex flex-col items-start min-w-0 flex-1">
                  <span
                    className="text-xs font-medium truncate w-full text-left"
                    style={{ color: 'var(--hub-text)' }}
                  >
                    {userName}
                  </span>
                  {userEmail && (
                    <span
                      className="text-[10px] truncate w-full text-left"
                      style={{ color: 'var(--hub-text-muted)' }}
                    >
                      {userEmail}
                    </span>
                  )}
                </div>
                <ChevronRight
                  size={12}
                  className="shrink-0"
                  style={{ color: 'var(--hub-text-dim)' }}
                />
              </button>

              <AnimatePresence>
                {showUserMenu && (
                  <motion.div
                    initial={{ opacity: 0, y: 4 }}
                    animate={{ opacity: 1, y: 0 }}
                    exit={{ opacity: 0, y: 4 }}
                    transition={{ duration: 0.12 }}
                    className="absolute bottom-full left-0 right-0 mb-1 rounded-xl p-2 z-50"
                    style={{
                      background: 'var(--hub-card)',
                      border: '1px solid var(--hub-border)',
                      boxShadow: 'var(--hub-shadow)',
                    }}
                  >
                    <p
                      className="px-2 py-1 text-[10px] font-semibold uppercase tracking-wider"
                      style={{ color: 'var(--hub-text-dim)' }}
                    >
                      Appearance
                    </p>
                    <div className="flex gap-1 mb-2">
                      {(
                        [
                          { value: 'light' as Theme, icon: Sun, label: 'Light' },
                          { value: 'dark' as Theme, icon: Moon, label: 'Dark' },
                          { value: 'system' as Theme, icon: Monitor, label: 'System' },
                        ] as const
                      ).map(({ value, icon: Icon, label }) => (
                        <button
                          key={value}
                          onClick={() => handleSetTheme(value)}
                          className="flex-1 flex flex-col items-center gap-1 py-2 rounded-lg text-[10px] font-medium transition-colors"
                          style={{
                            background:
                              theme === value ? 'rgba(99,102,241,0.12)' : 'var(--hub-hover)',
                            color: theme === value ? ACCENT : 'var(--hub-text-muted)',
                            border:
                              theme === value
                                ? '1px solid rgba(99,102,241,0.3)'
                                : '1px solid transparent',
                          }}
                        >
                          <Icon size={13} />
                          {label}
                        </button>
                      ))}
                    </div>
                    <div
                      className="my-1"
                      style={{ height: '1px', background: 'var(--hub-border)' }}
                    />
                    <button
                      onClick={handleLogout}
                      className="flex items-center gap-2 w-full px-2 py-2 rounded-lg text-sm transition-colors"
                      style={{ color: '#ef4444' }}
                      onMouseEnter={(e) => {
                        (e.currentTarget as HTMLElement).style.background = 'rgba(239,68,68,0.08)';
                      }}
                      onMouseLeave={(e) => {
                        (e.currentTarget as HTMLElement).style.background = 'transparent';
                      }}
                    >
                      <LogOut size={14} />
                      {t('auth.signOut', 'Sign out')}
                    </button>
                  </motion.div>
                )}
              </AnimatePresence>
            </div>
          </div>
        </aside>

        {/* Main content + Caret AI inline column */}
        <div className="flex flex-1 overflow-hidden">
          <div className="flex flex-col flex-1 overflow-hidden pt-11 md:pt-0">
            <header
              className="hidden md:flex h-11 items-center justify-between px-6 shrink-0"
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
                  onClick={() => handleSetTheme(theme === 'dark' ? 'light' : 'dark')}
                  className="h-7 w-7 rounded-lg flex items-center justify-center"
                  style={{ color: 'var(--hub-text-muted)' }}
                  onMouseEnter={(e) => {
                    (e.currentTarget as HTMLElement).style.background = 'var(--hub-hover)';
                  }}
                  onMouseLeave={(e) => {
                    (e.currentTarget as HTMLElement).style.background = 'transparent';
                  }}
                >
                  {theme === 'dark' ? <Sun size={14} /> : <Moon size={14} />}
                </button>
                <CaretButton />
              </div>
            </header>

            <LicenseBanner safe={safe} />
            <AnimatePresence mode="wait">
              <motion.main
                key={location.pathname}
                initial={{ opacity: 0, y: 6 }}
                animate={{ opacity: 1, y: 0 }}
                exit={{ opacity: 0, y: -4 }}
                transition={{ duration: 0.18, ease: 'easeOut' }}
                className="flex-1 overflow-y-auto p-6"
              >
                <Outlet />
              </motion.main>
            </AnimatePresence>
          </div>

          {/* Caret AI - inline third column, pushes content left */}
          <div
            style={{
              width: caretOpen ? '360px' : '0px',
              flexShrink: 0,
              overflow: 'hidden',
              transition: 'width 0.25s cubic-bezier(0.4, 0, 0.2, 1)',
            }}
          >
            <CaretPanel
              variant="inline"
              appId="insights"
              userId={userId}
              apiBaseUrl={import.meta.env.VITE_API_URL ?? 'https://api.voxire.com/'}
              systemContext="User is in the Insights app (analytics and reporting)"
              pageContext={caretPageContext}
            />
          </div>
        </div>

        <Toaster
          theme={theme === 'light' ? 'light' : 'dark'}
          position="bottom-right"
          richColors
          closeButton
        />
      </div>
    </>
  );
};

export default DefaultLayout;
