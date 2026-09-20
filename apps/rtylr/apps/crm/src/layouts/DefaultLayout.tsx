import React, { useEffect, useRef, useState } from 'react';
import { cn, Avatar } from '@voxire/ui';
import TrialBanner from '../shared/components/TrialBanner';
import { useLocation, useRouterState, Link, Outlet } from '@tanstack/react-router';
import { motion, AnimatePresence } from 'motion/react';
import { getAvatarUrl } from '@voxire/config';
import { signOut, useUserStore } from '@voxire/auth';
import { buildCaretRouteContext, CaretPanel, CaretButton, useCaretStore } from '@voxire/caret';
import { useQuery } from '@tanstack/react-query';
import {
  LayoutDashboard,
  Users,
  Gift,
  Megaphone,
  PieChart,
  Bell,
  Sun,
  Moon,
  Monitor,
  ChevronRight,
  LogOut,
  Zap,
  FileText,
  MessageSquare,
  Settings as SettingsIcon,
  Search,
  Menu,
  X,
  TrendingUp,
} from 'lucide-react';
import { useShortcuts } from '../shared/components/ShortcutsProvider';
import { useForesightLinks } from '@voxire/utils';
import { pageTitleKey, routeLabels } from '../shared/lib/routeLabels';
import { useTranslation } from 'react-i18next';
import { createOneKitClient } from '@voxire/api';
import { LicenseBanner } from '@voxire/components';
import { AppServiceClient } from '@voxire/voxie/hub/app/v1/client';
import type { Addon } from '@voxire/voxie/hub/app/v1/types';
import { UserServiceClient } from '@voxire/voxie/hub/user/v1/client';
import { licenseClient } from '../shared/lib/clients';
import { toast } from 'sonner';
import { readLocalStorage, writeLocalStorage } from '../shared/lib/storage';

const ACCENT = '#3b82f6';

const appClient = createOneKitClient(AppServiceClient, 'crm');
const userClient = createOneKitClient(UserServiceClient, 'crm');

function useAddonEnabled(addonId: string): boolean {
  // Errors propagate (never cached as a result) and read as "enabled": hiding
  // the nav entry on a network blip made Sales disappear with no explanation
  // (#941). Clicking through lands on the guard's retryable error state.
  const { data, isError } = useQuery<Addon[]>({
    queryKey: ['crm-addons'],
    queryFn: async (): Promise<Addon[]> => {
      const res = await appClient.listAddons({});
      return res.addons ?? [];
    },
    staleTime: 5 * 60 * 1000,
    retry: 1,
  });
  if (isError) return true;
  return (data ?? []).some((a) => a.id === addonId && a.isEnabled);
}

interface NavItem {
  labelKey: string;
  path: string;
  icon: React.ComponentType<{ size?: number; className?: string }>;
}

const navItems: NavItem[] = [
  { labelKey: routeLabels['/'].nav, path: '/', icon: LayoutDashboard },
  { labelKey: routeLabels['/customers'].nav, path: '/customers', icon: Users },
  { labelKey: routeLabels['/loyalty'].nav, path: '/loyalty', icon: Gift },
  { labelKey: routeLabels['/campaigns'].nav, path: '/campaigns', icon: Megaphone },
  { labelKey: routeLabels['/segments'].nav, path: '/segments', icon: PieChart },
  { labelKey: routeLabels['/automations'].nav, path: '/automations', icon: Zap },
  { labelKey: routeLabels['/templates'].nav, path: '/templates', icon: FileText },
  { labelKey: routeLabels['/feedback'].nav, path: '/feedback', icon: MessageSquare },
  { labelKey: routeLabels['/settings'].nav, path: '/settings', icon: SettingsIcon },
];

type Theme = 'light' | 'dark' | 'system';

function getStoredTheme(): Theme {
  const saved = readLocalStorage('crm-theme');
  return saved === 'light' || saved === 'dark' || saved === 'system' ? saved : 'dark';
}

function applyTheme(theme: Theme) {
  const resolved =
    theme === 'system'
      ? window.matchMedia('(prefers-color-scheme: dark)').matches
        ? 'dark'
        : 'light'
      : theme;
  document.documentElement.setAttribute('data-theme', resolved);
  writeLocalStorage('crm-theme', theme);
}

const SidebarNavItem: React.FC<NavItem> = ({ labelKey, path, icon: Icon }) => {
  const { t } = useTranslation();
  const location = useLocation();
  const isActive = path === '/' ? location.pathname === '/' : location.pathname.startsWith(path);

  return (
    <Link
      to={path}
      style={{
        position: 'relative',
        display: 'flex',
        alignItems: 'center',
        gap: '10px',
        padding: '8px 12px',
        borderRadius: '8px',
        fontSize: '13px',
        fontWeight: 500,
        textDecoration: 'none',
        color: isActive ? 'var(--hub-text)' : 'var(--hub-text-muted)',
        transition: 'color 0.15s',
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
          layoutId="crm-sidebar-active"
          style={{
            position: 'absolute',
            inset: 0,
            borderRadius: '8px',
            background: 'var(--hub-active)',
          }}
          transition={{ type: 'spring', stiffness: 400, damping: 35 }}
        />
      )}
      <Icon size={15} />
      <span style={{ position: 'relative', flex: 1 }}>{t(labelKey)}</span>
      {isActive && (
        <span
          style={{
            position: 'relative',
            height: '6px',
            width: '6px',
            borderRadius: '9999px',
            background: ACCENT,
            flexShrink: 0,
          }}
        />
      )}
    </Link>
  );
};

const DefaultLayout: React.FC = () => {
  const { t } = useTranslation();
  const location = useLocation();
  const search = useRouterState({ select: (s) => s.location.searchStr });
  const { openPalette } = useShortcuts();
  const [theme, setThemeState] = useState<Theme>(getStoredTheme);
  const sidebarRef = useRef<HTMLElement | null>(null);
  const crmUser = useUserStore((s) => s.user);
  const setAuthUser = useUserStore((s) => s.setUser);
  useEffect(() => {
    if (crmUser?.name || crmUser?.email) return;
    userClient
      .getMe({})
      .then((res) => {
        setAuthUser({ id: String(res.id), name: res.fullName, email: res.email });
      })
      .catch(() => undefined);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [crmUser]);
  const userId = Number(crmUser?.id) || 0;
  const userEmail = crmUser?.email || crmUser?.emailAddress || null;
  const userName = crmUser?.name || crmUser?.firstName || userEmail || 'User';
  const { isOpen: caretOpen } = useCaretStore();
  const salesPipelineEnabled = useAddonEnabled('sales_pipeline');

  useForesightLinks(sidebarRef, {
    '/': () => import('../features/dashboard/Dashboard'),
    '/customers': () => import('../features/customers/Customers'),
    '/loyalty': () => import('../features/loyalty/Loyalty'),
    '/campaigns': () => import('../features/campaigns/Campaigns'),
    '/segments': () => import('../features/segments/Segments'),
    '/automations': () => import('../features/automations/Automations'),
    '/templates': () => import('../features/templates/Templates'),
    '/feedback': () => import('../features/feedback/Feedback'),
    '/settings': () => import('../features/settings/Settings'),
  });
  const [userMenuOpen, setUserMenuOpen] = useState(false);
  const [sidebarOpen, setSidebarOpen] = useState(false);

  const pageTitle = t(pageTitleKey(location.pathname));
  // Caret describes the current route to the assistant, so it wants the
  // reader's language rather than the key or the English behind it.
  const translatedRouteLabels = React.useMemo(
    () => Object.fromEntries(Object.entries(routeLabels).map(([p, l]) => [p, t(l.title)])),
    [t]
  );
  const caretPageContext = React.useMemo(
    () =>
      buildCaretRouteContext({
        appId: 'crm',
        appName: 'CRM',
        pathname: location.pathname,
        search,
        page: pageTitle,
        routeLabels: translatedRouteLabels,
      }),
    [location.pathname, search, pageTitle, translatedRouteLabels]
  );

  useEffect(() => {
    setSidebarOpen(false);
  }, [location.pathname]);

  const handleTheme = (t: Theme) => {
    setThemeState(t);
    applyTheme(t);
  };

  const handleLogout = (): void => {
    signOut('crm');
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
            type="button"
            onClick={() => setSidebarOpen(true)}
            className="h-8 w-8 rounded-lg flex items-center justify-center"
            style={{ color: 'var(--hub-text-muted)' }}
          >
            <Menu size={18} />
          </button>
          <h1 className="text-sm font-semibold" style={{ color: 'var(--hub-text)', margin: 0 }}>
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

        {/* Sidebar */}
        <aside
          className={cn(sidebarOpen ? 'fixed inset-y-0 left-0 z-50 flex' : 'hidden', 'md:flex')}
          style={{
            width: '220px',
            flexDirection: 'column',
            flexShrink: 0,
            overflow: 'hidden',
            background: 'var(--hub-surface)',
            borderRight: '1px solid var(--hub-border)',
          }}
        >
          {/* Logo / Header */}
          <div
            style={{
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'space-between',
              gap: '10px',
              padding: '16px',
              borderBottom: '1px solid var(--hub-border)',
            }}
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
              className="h-9 w-auto"
            />
            <button
              type="button"
              onClick={() => setSidebarOpen(false)}
              className="h-7 w-7 rounded-lg flex md:hidden items-center justify-center"
              style={{ color: 'var(--hub-text-muted)' }}
            >
              <X size={16} />
            </button>
          </div>

          {/* Navigation */}
          <nav
            ref={sidebarRef}
            style={{
              flex: 1,
              overflowY: 'auto',
              padding: '4px 8px',
              display: 'flex',
              flexDirection: 'column',
              gap: '2px',
            }}
          >
            <p
              style={{
                padding: '4px 12px 6px',
                fontSize: '10px',
                fontWeight: 600,
                textTransform: 'uppercase',
                letterSpacing: '0.1em',
                color: 'var(--hub-text-dim)',
                marginTop: '4px',
              }}
            >
              {t('crm.layout.workspace')}
            </p>
            {navItems.map((item) => (
              <SidebarNavItem key={item.path} {...item} />
            ))}

            {/* Sales Pipeline - add-on gated */}
            {salesPipelineEnabled && (
              <>
                <p
                  style={{
                    padding: '4px 12px 6px',
                    fontSize: '10px',
                    fontWeight: 600,
                    textTransform: 'uppercase',
                    letterSpacing: '0.1em',
                    color: 'var(--hub-text-dim)',
                    marginTop: '8px',
                  }}
                >
                  {t('crm.layout.sales')}
                </p>
                <SidebarNavItem
                  labelKey={routeLabels['/sales'].nav}
                  path="/sales"
                  icon={TrendingUp}
                />
                <SidebarNavItem
                  labelKey={routeLabels['/sales/leads'].nav}
                  path="/sales/leads"
                  icon={Users}
                />
                <SidebarNavItem
                  labelKey={routeLabels['/sales/contracts'].nav}
                  path="/sales/contracts"
                  icon={FileText}
                />
              </>
            )}
          </nav>

          {/* Footer, user */}
          <div
            style={{
              padding: '8px',
              borderTop: '1px solid var(--hub-border)',
              position: 'relative',
            }}
          >
            <Link
              to="/settings"
              style={{
                display: 'flex',
                alignItems: 'center',
                gap: '8px',
                padding: '6px 12px',
                borderRadius: '8px',
                fontSize: '12px',
                color: 'var(--hub-text-muted)',
                textDecoration: 'none',
                marginBottom: '4px',
              }}
              onMouseEnter={(e) => {
                (e.currentTarget as HTMLElement).style.background = 'var(--hub-hover)';
              }}
              onMouseLeave={(e) => {
                (e.currentTarget as HTMLElement).style.background = 'transparent';
              }}
            >
              <SettingsIcon size={12} /> Settings
            </Link>
            <CaretButton size="md" className="w-full justify-start mb-1" />
            <button
              type="button"
              onClick={() => setUserMenuOpen((v) => !v)}
              style={{
                display: 'flex',
                alignItems: 'center',
                gap: '10px',
                width: '100%',
                padding: '8px 10px',
                borderRadius: '8px',
                background: 'transparent',
                border: 'none',
                cursor: 'pointer',
                transition: 'background 0.15s',
              }}
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
                style={{ background: 'rgba(59,130,246,0.15)' }}
              />
              <div style={{ flex: 1, minWidth: 0, textAlign: 'left' }}>
                <div
                  style={{
                    fontSize: '12px',
                    fontWeight: 500,
                    color: 'var(--hub-text)',
                    whiteSpace: 'nowrap',
                    overflow: 'hidden',
                    textOverflow: 'ellipsis',
                  }}
                >
                  {userName}
                </div>
                {userEmail && (
                  <div
                    style={{
                      fontSize: '10px',
                      color: 'var(--hub-text-muted)',
                      whiteSpace: 'nowrap',
                      overflow: 'hidden',
                      textOverflow: 'ellipsis',
                    }}
                  >
                    {userEmail}
                  </div>
                )}
              </div>
              <ChevronRight size={12} style={{ color: 'var(--hub-text-dim)', flexShrink: 0 }} />
            </button>

            <AnimatePresence>
              {userMenuOpen && (
                <motion.div
                  initial={{ opacity: 0, y: 6, scale: 0.97 }}
                  animate={{ opacity: 1, y: 0, scale: 1 }}
                  exit={{ opacity: 0, y: 6, scale: 0.97 }}
                  transition={{ duration: 0.14 }}
                  style={{
                    position: 'absolute',
                    bottom: '100%',
                    left: '8px',
                    right: '8px',
                    background: 'var(--hub-card)',
                    border: '1px solid var(--hub-border)',
                    borderRadius: '10px',
                    padding: '8px',
                    boxShadow: 'var(--hub-shadow)',
                    zIndex: 50,
                  }}
                >
                  <div style={{ marginBottom: '8px' }}>
                    <p
                      style={{
                        fontSize: '10px',
                        fontWeight: 600,
                        textTransform: 'uppercase',
                        letterSpacing: '0.1em',
                        color: 'var(--hub-text-dim)',
                        marginBottom: '6px',
                        padding: '0 4px',
                      }}
                    >
                      {t('crm.layout.appearance')}
                    </p>
                    <div style={{ display: 'flex', gap: '4px' }}>
                      {[
                        { value: 'light' as Theme, icon: Sun, label: t('crm.layout.light') },
                        { value: 'dark' as Theme, icon: Moon, label: t('crm.layout.dark') },
                        { value: 'system' as Theme, icon: Monitor, label: t('crm.layout.system') },
                      ].map(({ value, icon: Icon, label }) => (
                        <button
                          type="button"
                          key={value}
                          onClick={() => handleTheme(value)}
                          style={{
                            flex: 1,
                            display: 'flex',
                            flexDirection: 'column',
                            alignItems: 'center',
                            gap: '3px',
                            padding: '7px 4px',
                            borderRadius: '8px',
                            fontSize: '10px',
                            fontWeight: 500,
                            cursor: 'pointer',
                            border:
                              theme === value ? `1px solid ${ACCENT}40` : '1px solid transparent',
                            background:
                              theme === value ? 'rgba(59,130,246,0.12)' : 'var(--hub-hover)',
                            color: theme === value ? ACCENT : 'var(--hub-text-muted)',
                          }}
                        >
                          <Icon size={12} />
                          {label}
                        </button>
                      ))}
                    </div>
                  </div>
                  <div
                    style={{ height: '1px', background: 'var(--hub-border)', margin: '4px 0' }}
                  />
                  <button
                    type="button"
                    onClick={handleLogout}
                    style={{
                      display: 'flex',
                      alignItems: 'center',
                      gap: '8px',
                      width: '100%',
                      padding: '7px 8px',
                      borderRadius: '6px',
                      fontSize: '13px',
                      color: '#ef4444',
                      background: 'transparent',
                      border: 'none',
                      cursor: 'pointer',
                      transition: 'background 0.15s',
                    }}
                    onMouseEnter={(e) => {
                      (e.currentTarget as HTMLElement).style.background = 'rgba(239,68,68,0.08)';
                    }}
                    onMouseLeave={(e) => {
                      (e.currentTarget as HTMLElement).style.background = 'transparent';
                    }}
                  >
                    <LogOut size={13} />
                    {t('crm.layout.signOut')}
                  </button>
                </motion.div>
              )}
            </AnimatePresence>
          </div>
        </aside>

        {/* Main content + Caret AI inline column */}
        <div className="flex flex-1 overflow-hidden">
          {/* Main area */}
          <div className="flex flex-col flex-1 overflow-hidden pt-11 md:pt-0">
            {/* Topbar */}
            <header
              className="hidden md:flex"
              style={{
                height: '44px',
                alignItems: 'center',
                justifyContent: 'space-between',
                padding: '0 24px',
                flexShrink: 0,
                background: 'var(--hub-surface)',
                borderBottom: '1px solid var(--hub-border)',
              }}
            >
              <h1
                style={{ fontSize: '14px', fontWeight: 600, color: 'var(--hub-text)', margin: 0 }}
              >
                {pageTitle}
              </h1>
              <div style={{ display: 'flex', alignItems: 'center', gap: '8px' }}>
                <button
                  aria-label={t('crm.layout.openCommandPalette')}
                  type="button"
                  onClick={openPalette}
                  style={{
                    height: '28px',
                    padding: '0 10px',
                    borderRadius: '8px',
                    display: 'flex',
                    alignItems: 'center',
                    gap: '8px',
                    background: 'var(--hub-input-bg)',
                    border: '1px solid var(--hub-border)',
                    cursor: 'pointer',
                    color: 'var(--hub-text-muted)',
                    fontSize: '12px',
                    transition: 'all 0.15s',
                  }}
                >
                  <Search size={12} />
                  <span>{t('crm.components.searchPlaceholder')}</span>
                  <kbd
                    style={{
                      fontSize: '9px',
                      padding: '1px 5px',
                      borderRadius: '3px',
                      border: '1px solid var(--hub-border)',
                      background: 'var(--hub-card)',
                      color: 'var(--hub-text-dim)',
                    }}
                  >
                    ⌘K
                  </kbd>
                </button>
                <button
                  type="button"
                  aria-label={t('crm.layout.notifications')}
                  onClick={() => toast.info(t('crm.layout.noNotifications'))}
                  style={{
                    height: '28px',
                    width: '28px',
                    borderRadius: '8px',
                    display: 'flex',
                    alignItems: 'center',
                    justifyContent: 'center',
                    background: 'transparent',
                    border: 'none',
                    cursor: 'pointer',
                    color: 'var(--hub-text-muted)',
                    transition: 'background 0.15s',
                  }}
                  onMouseEnter={(e) => {
                    (e.currentTarget as HTMLElement).style.background = 'var(--hub-hover)';
                  }}
                  onMouseLeave={(e) => {
                    (e.currentTarget as HTMLElement).style.background = 'transparent';
                  }}
                >
                  <Bell size={14} />
                </button>
                <CaretButton />
              </div>
            </header>

            {/* Content */}
            <LicenseBanner client={licenseClient} />
            <TrialBanner />
            <AnimatePresence mode="wait">
              <motion.main
                key={location.pathname}
                initial={{ opacity: 0, y: 6 }}
                animate={{ opacity: 1, y: 0 }}
                exit={{ opacity: 0, y: -4 }}
                transition={{ duration: 0.18, ease: 'easeOut' }}
                style={{
                  flex: 1,
                  overflowY: 'auto',
                  padding: '24px',
                }}
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
              appId="crm"
              userId={userId}
              apiBaseUrl={import.meta.env.VITE_API_URL ?? 'https://api.voxire.com/'}
              systemContext="User is in the CRM app (customers, loyalty, campaigns, segments)"
              pageContext={caretPageContext}
            />
          </div>
        </div>
      </div>
    </>
  );
};

export default DefaultLayout;
