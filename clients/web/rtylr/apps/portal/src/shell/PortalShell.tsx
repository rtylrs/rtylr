import { useEffect, useMemo, useRef, useState, lazy, Suspense, type ReactNode } from 'react';
import { Link, useRouterState } from '@tanstack/react-router';
import { useTranslation } from 'react-i18next';
import { resolveApiBase } from '@voxire/config';
import { PanelLeft, Search, Bell, Settings, Sparkles, X } from 'lucide-react';
import { HomeIcon, InsightsIcon, FlowIcon } from '@voxire/ui';
import { CaretPanel, useCaretStore, useCaretSettingsStore } from '@voxire/caret';
import { useUserStore } from '../shared/stores/userStore';
import { useCommandPalette } from '../shared/stores/commandPaletteStore';
import { useNotifications } from '../shared/hooks/queries/useNotifications';
import { useAddons, isAddonEnabled } from '../shared/hooks/queries/useAddons';
import { useKeyboardShortcuts } from '../shared/hooks/useKeyboardShortcuts';
import { useOnboardingTour } from '../shared/hooks/useOnboardingTour';
import CommandPalette from './command-palette/CommandPalette';
import { buildPortalPageContext, buildPortalSystemContext } from '../shared/lib/portalCaretContext';
import { canAccessPortalPath } from '../shared/utils/canAccessPortalPath';
import InsightsHeader from '../modules/insights/InsightsHeader';
import BusinessSwitcher from '../features/business-switcher/BusinessSwitcher';
import { insightsNavItems } from '../modules/insights/navigation';
import ProfileMenu from './ProfileMenu';
import SettingsNavigation from './SettingsNavigation';
import { PortalHeaderActionsProvider } from './PortalHeaderActions';
import {
  rawNavItems,
  rawMenuSubItems,
  settingsNavItems,
  addonSettingsItems,
  PortalNavLink,
  routeKeys,
} from './navigation';
import { useFlowStore } from '../modules/flow/shared/stores/flowStore';
const FlowQuickCreate = lazy(() => import('../modules/flow/shared/components/QuickCreate'));
import './portal.css';
import './caret.css';

function initialCollapsed() {
  try {
    return localStorage.getItem('portal-sidebar-collapsed') === 'true';
  } catch {
    return false;
  }
}

export default function PortalShell({ children }: { children: ReactNode }) {
  const { t } = useTranslation();
  const { user } = useUserStore();
  const pathname = useRouterState({ select: (state) => state.location.pathname });
  const search = useRouterState({ select: (state) => state.location.searchStr });
  const [collapsed, setCollapsed] = useState(initialCollapsed);
  const [mobileOpen, setMobileOpen] = useState(false);
  const [settingsQuery, setSettingsQuery] = useState('');
  const quickCreateOpen = useFlowStore((state) => state.quickCreateOpen);
  const dialogRef = useRef<HTMLDialogElement>(null);
  const mainRef = useRef<HTMLElement>(null);
  const { setOpen: setPaletteOpen } = useCommandPalette();
  const { isOpen: caretOpen, toggle: toggleCaret } = useCaretStore();
  const [caretMounted, setCaretMounted] = useState(caretOpen);
  const caretWidth = useCaretSettingsStore((state) => state.panelWidth);
  const { data: notifications = [] } = useNotifications();
  const { data: addons = [] } = useAddons();
  const unreadCount = notifications.filter((notification) => !notification.read).length;
  const isSettings =
    pathname.startsWith('/settings') || ['/security', '/api-keys', '/webhooks'].includes(pathname);
  const isInsights = pathname === '/insights' || pathname.startsWith('/insights/');
  const isFlow = pathname === '/flow' || pathname.startsWith('/flow/');
  const availableInsights = insightsNavItems.filter((item) => canAccessPortalPath(user, item.path));
  const insightsPage = insightsNavItems.find((item) => item.path === pathname);
  const pageKey =
    routeKeys[pathname] ||
    Object.keys(routeKeys)
      .filter((key) => key !== '/' && pathname.startsWith(`${key}/`))
      .sort((a, b) => b.length - a.length)
      .map((key) => routeKeys[key])[0];
  const pageTitle = isFlow
    ? t('appNames.short.flow')
    : isInsights
      ? t(`insightsNav.${insightsPage?.key || 'overview'}`)
      : pathname.startsWith('/team/groups')
        ? t('common.nav.groups')
        : pageKey
          ? t(`common.nav.${pageKey}`)
          : t('common.portal.home');
  const nav = (
    isSettings
      ? [
          ...settingsNavItems,
          ...addonSettingsItems.filter((item) => isAddonEnabled(addons, item.addonId)),
        ]
      : [...rawNavItems, ...(isAddonEnabled(addons, 'menu') ? rawMenuSubItems : [])]
  ).filter((item) => canAccessPortalPath(user, item.path));
  const apps = useMemo(
    () =>
      user?.effectiveAccess?.apps
        .filter((app) => app.licensed && app.hasAccess)
        .map((app) => app.appId),
    [user?.effectiveAccess]
  );
  const caretContext = useMemo(
    () => ({ pathname, search, businessName: user?.businessName, apps }),
    [pathname, search, user?.businessName, apps]
  );
  useKeyboardShortcuts();
  useOnboardingTour();
  useEffect(() => {
    if (caretOpen) setCaretMounted(true);
  }, [caretOpen]);
  useEffect(() => {
    try {
      localStorage.setItem('portal-sidebar-collapsed', String(collapsed));
    } catch {
      /* Preferences remain usable when storage is blocked. */
    }
  }, [collapsed]);
  useEffect(() => {
    const dialog = dialogRef.current;
    if (mobileOpen && !dialog?.open) dialog?.showModal();
    if (!mobileOpen && dialog?.open) dialog.close();
  }, [mobileOpen]);
  useEffect(() => {
    setMobileOpen(false);
    mainRef.current?.scrollTo({ top: 0 });
  }, [pathname]);
  useEffect(() => {
    const key = (event: KeyboardEvent) => {
      if (event.defaultPrevented || event.repeat) return;
      if ((event.metaKey || event.ctrlKey) && event.key.toLowerCase() === 'j') {
        event.preventDefault();
        toggleCaret();
        return;
      }
      if ((event.metaKey || event.ctrlKey) && event.key === '\\') {
        event.preventDefault();
        setCollapsed((value) => !value);
      }
    };
    window.addEventListener('keydown', key);
    return () => window.removeEventListener('keydown', key);
  }, [toggleCaret]);
  const toggleNavigation = () => {
    if (window.matchMedia('(max-width: 900px)').matches) setMobileOpen((value) => !value);
    else setCollapsed((value) => !value);
  };
  const actions = (
    <div className="portal-topbar-actions">
      <button
        type="button"
        className="portal-icon-button"
        onClick={toggleCaret}
        aria-label="Caret"
        title="Caret"
        aria-pressed={caretOpen}
        data-tour="caret"
      >
        <Sparkles size={18} />
      </button>
      <button
        type="button"
        className="portal-icon-button"
        onClick={() => setPaletteOpen(true)}
        aria-label={t('common.nav.openCommandPalette')}
        data-portal-command-trigger
        title={t('common.nav.openCommandPalette')}
      >
        <Search size={18} />
      </button>
      <Link
        to="/notifications"
        className="portal-icon-button"
        aria-label={
          unreadCount
            ? t('common.nav.notificationsUnread', { count: unreadCount })
            : t('common.nav.notifications')
        }
        title={t('common.nav.notifications')}
      >
        <Bell size={18} />
        {unreadCount > 0 && <span className="portal-notification-dot" />}
      </Link>
    </div>
  );
  const sidebar = (mobile = false) => (
    <>
      <div className="portal-sidebar-heading">
        <Link
          to="/"
          onClick={() => setMobileOpen(false)}
          className="portal-wordmark"
          aria-label="RTYLR"
        >
          <img className="portal-logo-light" src="/logo-dark.png" alt="RTYLR" />
          <img className="portal-logo-dark" src="/logo.png" alt="RTYLR" />
        </Link>
        <button
          type="button"
          className="portal-icon-button"
          onClick={mobile ? () => setMobileOpen(false) : toggleNavigation}
          aria-label={mobile ? t('common.closePanel') : t('common.portal.collapseSidebar')}
          aria-expanded={!collapsed}
          aria-controls="portal-app-navigation"
        >
          {mobile ? <X size={18} /> : <PanelLeft size={18} />}
        </button>
      </div>
      {!isSettings && <BusinessSwitcher />}
      <nav
        id={mobile ? 'portal-mobile-navigation' : 'portal-app-navigation'}
        className="portal-app-navigation"
        aria-label={t('common.portal.pageNavigation')}
      >
        {!isSettings && (
          <p className="portal-section-label">
            {isInsights
              ? t('appNames.short.insights')
              : isSettings
                ? t('common.nav.manageAccount')
                : t('common.portal.workspace')}
          </p>
        )}
        {isSettings ? (
          <SettingsNavigation
            items={nav}
            query={settingsQuery}
            onQueryChange={setSettingsQuery}
            onNavigate={() => setMobileOpen(false)}
          />
        ) : (
          nav.map((item) => (
            <PortalNavLink
              {...item}
              key={item.path}
              tourId={mobile ? undefined : item.key}
              label={t(`common.nav.${item.key}`)}
              onNavigate={() => setMobileOpen(false)}
            />
          ))
        )}
      </nav>
    </>
  );
  return (
    <div className="portal-shell">
      <a className="portal-skip-link" href="#main-content">
        {t('common.skipToMainContent')}
      </a>
      <nav className="portal-rail" aria-label={t('common.portal.appNavigation')}>
        <Link
          to="/"
          className={`portal-rail-button${!isSettings && !isInsights && !isFlow ? ' is-active' : ''}`}
          title={t('common.portal.home')}
          aria-label={t('common.portal.home')}
          aria-current={!isSettings && !isInsights && !isFlow ? 'page' : undefined}
        >
          <HomeIcon size={21} />
        </Link>
        {availableInsights[0] && (
          <Link
            to={availableInsights[0].path}
            className={`portal-rail-button${isInsights ? ' is-active' : ''}`}
            title={t('appNames.short.insights')}
            aria-label={t('appNames.short.insights')}
            aria-current={isInsights ? 'page' : undefined}
          >
            <InsightsIcon size={21} />
          </Link>
        )}
        {canAccessPortalPath(user, '/flow') && (
          <Link
            to="/flow"
            className={`portal-rail-button${isFlow ? ' is-active' : ''}`}
            title={t('appNames.short.flow')}
            aria-label={t('appNames.short.flow')}
            aria-current={isFlow ? 'page' : undefined}
          >
            <FlowIcon size={21} />
          </Link>
        )}
        <div className="portal-rail-spacer" />
        <Link
          data-tour="settings"
          to="/settings"
          className={`portal-rail-button${isSettings ? ' is-active' : ''}`}
          aria-label={t('common.nav.settings')}
          title={t('common.nav.settings')}
        >
          <Settings size={21} />
        </Link>
        <ProfileMenu />
      </nav>
      <div className="portal-workspace">
        {!isInsights && !isFlow && !collapsed && (
          <aside className="portal-sidebar">{sidebar()}</aside>
        )}
        <div className="portal-app-panel">
          {isInsights ? (
            <InsightsHeader actions={actions} />
          ) : !isFlow ? (
            <header className="portal-topbar">
              <div className="portal-topbar-start">
                {!isFlow && (
                  <button
                    type="button"
                    className={`portal-icon-button portal-show-sidebar${collapsed ? ' is-visible' : ''}`}
                    onClick={toggleNavigation}
                    aria-label={t('common.toggleNavigation')}
                    aria-expanded={!collapsed || mobileOpen}
                    aria-controls="portal-app-navigation"
                  >
                    <PanelLeft size={19} />
                  </button>
                )}
                <span>{pageTitle}</span>
              </div>
              {actions}
            </header>
          ) : null}
          <main
            id="main-content"
            ref={mainRef}
            tabIndex={-1}
            className={`portal-page${isFlow ? ' is-flow' : ''}`}
          >
            <PortalHeaderActionsProvider value={actions}>{children}</PortalHeaderActionsProvider>
          </main>
        </div>
        <div
          className="portal-caret"
          inert={!caretOpen}
          aria-hidden={!caretOpen}
          style={{
            width: caretOpen ? `min(${caretWidth}px, calc(100vw - 72px))` : 0,
            visibility: caretOpen ? 'visible' : 'hidden',
          }}
        >
          {(caretOpen || caretMounted) && (
            <CaretPanel
              apiBaseUrl={resolveApiBase()}
              variant="inline"
              appId={isFlow ? 'flow' : isInsights ? 'insights' : 'dash'}
              userId={Number(user?.id) || 0}
              systemContext={buildPortalSystemContext(caretContext)}
              pageContext={buildPortalPageContext(caretContext)}
            />
          )}
        </div>
      </div>
      <dialog
        ref={dialogRef}
        className="portal-mobile-sidebar"
        onCancel={() => setMobileOpen(false)}
        onClose={() => setMobileOpen(false)}
      >
        {mobileOpen && !isInsights && !isFlow && sidebar(true)}
      </dialog>
      <CommandPalette />
      {quickCreateOpen && (
        <Suspense fallback={null}>
          <FlowQuickCreate />
        </Suspense>
      )}
    </div>
  );
}
