import React, { useEffect, useMemo, useRef, useState } from 'react';
import { Avatar } from '@voxire/ui';
import { useLocation, useRouterState, Link, Outlet } from '@tanstack/react-router';
import { useForesightLinks } from '@voxire/utils';
import { motion, AnimatePresence } from 'motion/react';
import { getAvatarUrl } from '@voxire/config';
import { signOut, useUserStore } from '@voxire/auth';
import { LicenseBanner } from '@voxire/components';
import { CaretPanel, buildCaretRouteContext, useCaretStore } from '@voxire/caret';
import { licenseClient, userClient } from '../shared/lib/clients';
import { getPageTitle, routeLabels } from '../shared/lib/nav';
import {
  TrendingUp,
  Users,
  Package,
  Briefcase,
  DollarSign,
  FileText,
  Settings as SettingsIcon,
  LayoutGrid,
  ChevronDown,
  Sun,
  Moon,
  Monitor,
  LogOut,
} from 'lucide-react';
import { Toaster } from 'sonner';
import { useTranslation } from 'react-i18next';

interface PageItem {
  label: string;
  path: string;
  icon: React.ComponentType<{ size?: number; className?: string }>;
}

const ACCENT = '#6366f1';

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

const DefaultLayout: React.FC = () => {
  const { t } = useTranslation();
  const location = useLocation();
  const search = useRouterState({ select: (s) => s.location.searchStr });
  const [theme, setThemeState] = useState<Theme>(() => {
    return (localStorage.getItem('rtylr-insights-theme') as Theme) || 'dark';
  });
  const [showUserMenu, setShowUserMenu] = useState(false);
  const [showPagesMenu, setShowPagesMenu] = useState(false);
  const insightsUser = useUserStore((s) => s.user);
  const setAuthUser = useUserStore((s) => s.setUser);
  const userId = Number(insightsUser?.id) || 0;
  const toggleCaret = useCaretStore((s) => s.toggle);

  const pageItems = useMemo<PageItem[]>(
    () => [
      { label: t('nav.revenue'), path: '/revenue', icon: TrendingUp },
      { label: t('nav.customers'), path: '/customers', icon: Users },
      { label: t('nav.products'), path: '/products', icon: Package },
      { label: t('nav.operations'), path: '/operations', icon: Briefcase },
      { label: t('nav.financial'), path: '/financial', icon: DollarSign },
      { label: t('nav.reports'), path: '/reports', icon: FileText },
    ],
    [t]
  );

  useEffect(() => {
    if (insightsUser?.name || insightsUser?.email) return;
    userClient
      .getMe({})
      .then((res) => {
        if (res?.id) {
          setAuthUser({ id: String(res.id), name: res.fullName ?? '', email: res.email ?? '' });
        }
      })
      .catch(() => undefined);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [insightsUser]);

  const pageTitle = getPageTitle(location.pathname);
  const isHome = location.pathname === '/';
  const menuRef = useRef<HTMLDivElement | null>(null);

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

  useForesightLinks(menuRef, {
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
    if (theme === 'system') {
      const media = window.matchMedia('(prefers-color-scheme: dark)');
      const listener = () => applyTheme('system');
      media.addEventListener('change', listener);
      return () => media.removeEventListener('change', listener);
    }
  }, [theme]);

  useEffect(() => {
    setShowPagesMenu(false);
    setShowUserMenu(false);
  }, [location.pathname]);

  // Cmd/Ctrl+K opens Caret, ChatGPT/Spotlight-style, from any page.
  useEffect(() => {
    const onKey = (e: KeyboardEvent): void => {
      const mod = e.metaKey || e.ctrlKey;
      if (!mod || (e.key !== 'k' && e.key !== 'K')) return;
      const target = e.target as HTMLElement | null;
      const editable =
        !!target &&
        (target.tagName === 'INPUT' || target.tagName === 'TEXTAREA' || target.isContentEditable);
      if (editable) return;
      e.preventDefault();
      if (!useCaretStore.getState().isOpen) toggleCaret();
    };
    window.addEventListener('keydown', onKey);
    return () => window.removeEventListener('keydown', onKey);
  }, [toggleCaret]);

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
    <div className="flex flex-col h-screen overflow-hidden" style={{ background: 'var(--hub-bg)' }}>
      <header
        className="flex items-center justify-between gap-3 h-12 px-4 shrink-0"
        style={{ background: 'var(--hub-surface)', borderBottom: '1px solid var(--hub-border)' }}
      >
        <div className="flex items-center gap-2 shrink-0">
          <Link
            to="/"
            className="flex items-center gap-2.5 shrink-0"
            style={{ textDecoration: 'none' }}
          >
            <img
              src={
                theme === 'dark' ||
                (theme === 'system' &&
                  typeof window !== 'undefined' &&
                  window.matchMedia('(prefers-color-scheme: dark)').matches)
                  ? '/logo.png'
                  : '/logo-dark.png'
              }
              alt="rtylr"
              className="h-7 w-auto"
            />
            {!isHome && (
              <span
                className="hidden sm:inline text-sm font-semibold"
                style={{ color: 'var(--hub-text-dim)' }}
              >
                / {pageTitle}
              </span>
            )}
          </Link>

          <div className="relative" ref={menuRef}>
            <button
              type="button"
              onClick={() => setShowPagesMenu((v) => !v)}
              aria-haspopup="menu"
              aria-expanded={showPagesMenu}
              aria-label={t('nav.pages')}
              className="flex items-center gap-1.5 min-h-[44px] px-3 rounded-lg text-xs font-medium transition-colors duration-150"
              style={{
                color: 'var(--hub-text-muted)',
                background: showPagesMenu ? 'var(--hub-hover)' : 'transparent',
              }}
              onMouseEnter={(e) => {
                (e.currentTarget as HTMLElement).style.background = 'var(--hub-hover)';
              }}
              onMouseLeave={(e) => {
                if (!showPagesMenu)
                  (e.currentTarget as HTMLElement).style.background = 'transparent';
              }}
            >
              <LayoutGrid size={14} />
              <span className="hidden sm:inline">{t('nav.pages')}</span>
              <ChevronDown size={12} />
            </button>

            <AnimatePresence>
              {showPagesMenu && (
                <motion.div
                  initial={{ opacity: 0, y: -4 }}
                  animate={{ opacity: 1, y: 0 }}
                  exit={{ opacity: 0, y: -4 }}
                  transition={{ duration: 0.12 }}
                  role="menu"
                  className="absolute top-full start-0 mt-1 w-48 rounded-xl p-1.5 z-50"
                  style={{
                    background: 'var(--hub-card)',
                    border: '1px solid var(--hub-border)',
                    boxShadow: 'var(--hub-shadow)',
                  }}
                >
                  {pageItems.map((item) => {
                    const isActive = location.pathname.startsWith(item.path);
                    const Icon = item.icon;
                    return (
                      <Link
                        key={item.path}
                        to={item.path}
                        role="menuitem"
                        className="flex items-center gap-2.5 px-2.5 py-2.5 min-h-[44px] rounded-lg text-sm font-medium transition-colors duration-150"
                        style={{
                          color: isActive ? 'var(--hub-text)' : 'var(--hub-text-muted)',
                          background: isActive ? 'var(--hub-active)' : 'transparent',
                          textDecoration: 'none',
                        }}
                        onMouseEnter={(e) => {
                          if (!isActive)
                            (e.currentTarget as HTMLElement).style.background = 'var(--hub-hover)';
                        }}
                        onMouseLeave={(e) => {
                          if (!isActive)
                            (e.currentTarget as HTMLElement).style.background = 'transparent';
                        }}
                      >
                        <Icon size={14} className="shrink-0" />
                        {item.label}
                      </Link>
                    );
                  })}
                </motion.div>
              )}
            </AnimatePresence>
          </div>
        </div>

        <div className="flex items-center gap-1.5">
          <button
            type="button"
            onClick={() => handleSetTheme(theme === 'dark' ? 'light' : 'dark')}
            aria-label={t('nav.toggleTheme')}
            className="min-h-[44px] min-w-[44px] rounded-lg flex items-center justify-center"
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

          <div className="relative">
            <button
              type="button"
              onClick={() => setShowUserMenu((v) => !v)}
              aria-haspopup="menu"
              aria-expanded={showUserMenu}
              className="flex items-center gap-1.5 h-9 pl-1.5 pr-2 rounded-lg transition-colors duration-150"
              style={{ background: showUserMenu ? 'var(--hub-hover)' : 'transparent' }}
              onMouseEnter={(e) => {
                (e.currentTarget as HTMLElement).style.background = 'var(--hub-hover)';
              }}
              onMouseLeave={(e) => {
                if (!showUserMenu)
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
              <span
                className="hidden sm:inline text-xs font-medium truncate max-w-[140px]"
                style={{ color: 'var(--hub-text)' }}
              >
                {userName}
              </span>
              <ChevronDown
                size={12}
                className="shrink-0"
                style={{ color: 'var(--hub-text-dim)' }}
              />
            </button>

            <AnimatePresence>
              {showUserMenu && (
                <motion.div
                  initial={{ opacity: 0, y: -4 }}
                  animate={{ opacity: 1, y: 0 }}
                  exit={{ opacity: 0, y: -4 }}
                  transition={{ duration: 0.12 }}
                  className="absolute top-full right-0 mt-1 w-56 rounded-xl p-2 z-50"
                  style={{
                    background: 'var(--hub-card)',
                    border: '1px solid var(--hub-border)',
                    boxShadow: 'var(--hub-shadow)',
                  }}
                >
                  <div className="flex items-center gap-2.5 px-2 py-1.5 mb-1">
                    <Avatar
                      src={getAvatarUrl(userEmail || userName)}
                      alt={userName}
                      size="sm"
                      className="h-7 w-7"
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
                  </div>

                  <Link
                    to="/settings"
                    className="flex items-center gap-2 px-2 py-2 mb-1 rounded-lg text-sm font-medium transition-colors"
                    style={{ color: 'var(--hub-text)', textDecoration: 'none' }}
                    onMouseEnter={(e) => {
                      (e.currentTarget as HTMLElement).style.background = 'var(--hub-hover)';
                    }}
                    onMouseLeave={(e) => {
                      (e.currentTarget as HTMLElement).style.background = 'transparent';
                    }}
                  >
                    <SettingsIcon size={14} />
                    Settings
                  </Link>
                  <div
                    className="my-1"
                    style={{ height: '1px', background: 'var(--hub-border)' }}
                  />

                  <p
                    className="px-2 py-1 text-[10px] font-semibold uppercase tracking-wider"
                    style={{ color: 'var(--hub-text-dim)' }}
                  >
                    Appearance
                  </p>
                  <div className="flex gap-1 mb-2 px-1">
                    {(
                      [
                        { value: 'light' as Theme, icon: Sun, label: 'Light' },
                        { value: 'dark' as Theme, icon: Moon, label: 'Dark' },
                        { value: 'system' as Theme, icon: Monitor, label: 'System' },
                      ] as const
                    ).map(({ value, icon: Icon, label }) => (
                      <button
                        type="button"
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
                    type="button"
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
                    {t('nav.signOut')}
                  </button>
                </motion.div>
              )}
            </AnimatePresence>
          </div>
        </div>
      </header>

      <LicenseBanner client={licenseClient} />
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

      {/* Always mounted (not gated on isOpen) so it can fade in/out and keep streaming alive while closed. */}
      <CaretPanel
        variant="overlay"
        appId="insights"
        userId={userId}
        apiBaseUrl={import.meta.env.VITE_API_URL ?? 'https://api.voxire.com/'}
        systemContext="User is in the Insights app (analytics and reporting)"
        pageContext={caretPageContext}
      />

      <Toaster
        theme={theme === 'light' ? 'light' : 'dark'}
        position="bottom-right"
        richColors
        closeButton
      />
    </div>
  );
};

export default DefaultLayout;
