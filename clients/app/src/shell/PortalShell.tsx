import { recruitNavItems } from '../modules/recruit/navigation';
import '../modules/recruit/recruit.css';
import { getHrNavItems, hrNavItems } from '../modules/hr/navigation';
import { useHRConfigStore } from '../modules/hr/shared/stores/hrConfigStore';
import '../modules/hr/hr.css';
import {
  Fragment,
  useCallback,
  useEffect,
  useMemo,
  useRef,
  useState,
  lazy,
  Suspense,
  type ReactNode,
  type CSSProperties,
} from 'react';
import { Link, useNavigate, useRouterState } from '@tanstack/react-router';
import { useTranslation } from 'react-i18next';
import { resolveApiBase } from '@voxire/config';
import { ThemeScopeProvider } from '@voxire/ui';
import {
  useMediaQuery,
  deriveAccentFgAcross,
  deriveActionFill,
  deriveOnAccent,
  compositeOver,
  NEAR_WHITE,
} from '@voxire/utils';
import { useThemeStore } from '../shared/stores/themeStore';
import { PanelLeft, Search, Bell, Webhook, Plus, X } from 'lucide-react';
import { AutomationsIcon } from '@voxire/ui';
import {
  CaretHostProvider,
  CaretIcon,
  CaretPanel,
  onCaretToolSuccess,
  useCaretStore,
  useCaretSettingsStore,
  type CaretHost,
  type CaretPageSpec,
} from '@voxire/caret';
import { useQueryClient } from '@tanstack/react-query';
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
import { getFinanceHomePath, financeNavItems } from '../modules/finance/navigation';
import { erpNavItems } from '../modules/erp/navigation';
import { crmNavItems, crmSalesNavItems, crmSidebarItems } from '../modules/crm/navigation';
import { insightsNavItems } from '../modules/insights/navigation';
import ProfileMenu from './ProfileMenu';
import { Tooltip } from '../shared/components/ui/tooltip';
import SettingsModal from '../features/settings/SettingsModal';
import { useSettingsModalStore } from '../features/settings/settingsModalStore';
import { PortalHeaderActionsProvider } from './PortalHeaderActions';
import { SidebarCollapseProvider } from './SidebarCollapseContext';
import { rawNavItems, rawMenuSubItems, PortalNavLink, routeKeys } from './navigation';
import { arrangeItems, portalLayoutKey, usePortalLayout } from './portalLayout';
import { RAIL_BOTTOM, RAIL_ICONS, railItems, type RailItem } from './railItems';
import { spotlightControl } from './spotlight';
import PortalPageCard from '../features/caret/PortalPageCard';
import { useFlowStore } from '../modules/flow/shared/stores/flowStore';
const FlowQuickCreate = lazy(() => import('../modules/flow/shared/components/QuickCreate'));
import './portal.css';
import './caret.css';
import '../modules/crm/crm.css';
import '../modules/erp/erp.css';
import '../modules/finance/finance.css';
import RtylrLogo from '../shared/components/RtylrLogo';

interface RailEntry extends RailItem {
  active: boolean;
}

// Keep in sync with the --app-* launcher colours in globals.css.
const APP_ACCENTS: Partial<Record<string, string>> = {
  insights: '#6fa3db',
  flow: '#ef4444',
  crm: '#8b5cf6',
  erp: '#64c5b9',
  hr: '#10b981',
  finance: '#f59e0b',
};

function initialCollapsed() {
  try {
    return localStorage.getItem('portal-sidebar-collapsed') === 'true';
  } catch {
    return false;
  }
}

export default function PortalShell({ children }: { children: ReactNode }) {
  const { t, i18n } = useTranslation();
  const { user } = useUserStore();
  const pathname = useRouterState({ select: (state) => state.location.pathname });
  const search = useRouterState({ select: (state) => state.location.searchStr });
  const [collapsed, setCollapsed] = useState(initialCollapsed);
  const [mobileOpen, setMobileOpen] = useState(false);
  const isMobile = useMediaQuery('(max-width: 900px)');
  const systemDark = useMediaQuery('(prefers-color-scheme: dark)');
  const theme = useThemeStore((state) => state.theme);
  const appId = (
    pathname.startsWith('/settings/') ? pathname.split('/')[2] : pathname.split('/')[1]
  )?.split('-')[0];
  const accent = APP_ACCENTS[appId];
  const grounds =
    theme === 'dark' || (theme === 'system' && systemDark)
      ? ['#0a0a0a', '#111111', '#161616']
      : ['#f8f8f8', '#ffffff'];
  // Solid buttons take a fill white text reads on wherever the accent is a
  // mid-tone (CRM violet, Flow red); see deriveActionFill.
  const actionFill = deriveActionFill(accent ?? '#111111');
  const onAction = deriveOnAccent(actionFill);
  const appStyle: CSSProperties & Record<string, string> = accent
    ? {
        '--app-action-bg': actionFill,
        '--app-on-action': onAction,
        '--app-action-hover': `color-mix(in srgb, ${actionFill} 90%, ${onAction === NEAR_WHITE ? 'black' : 'white'})`,
        '--app-selection-bg': `color-mix(in srgb, ${accent} 12%, transparent)`,
        '--app-on-selection': deriveAccentFgAcross(
          accent,
          grounds.map((ground) => compositeOver(accent, 0.12, ground))
        ),
        '--app-focus': deriveAccentFgAcross(accent, grounds),
        '--accent': accent,
        '--app-logo-fold': accent,
        '--accent-foreground': deriveOnAccent(accent),
        '--color-accent-foreground': deriveOnAccent(accent),
        '--color-accent': accent,
        '--color-ring': deriveAccentFgAcross(accent, grounds),
        '--accent-muted': `color-mix(in srgb, ${accent} 12%, transparent)`,
        '--accent-border': `color-mix(in srgb, ${accent} 30%, transparent)`,
        '--accent-focus': deriveAccentFgAcross(accent, grounds),
        '--accent-fg': deriveAccentFgAcross(accent, grounds),
        '--accent-on-muted': deriveAccentFgAcross(
          accent,
          grounds.map((ground) => compositeOver(accent, 0.12, ground))
        ),
        '--dash-on-accent': deriveOnAccent(accent),
        '--hub-on-accent': deriveOnAccent(accent),
        ...(appId === 'flow'
          ? {
              '--flow-action-bg': actionFill,
              '--flow-on-accent': onAction,
              '--flow-active': `color-mix(in srgb, ${accent} 12%, transparent)`,
              '--flow-border-focus': deriveAccentFgAcross(accent, grounds),
              '--flow-focus-ring': deriveAccentFgAcross(accent, grounds),
            }
          : {}),
        ...(appId === 'hr'
          ? {
              '--hr-accent': accent,
              '--hr-payroll-accent': accent,
              '--hr-on-accent': deriveOnAccent(accent),
            }
          : {}),
        ...(appId === 'crm' ? { '--crm-action': actionFill, '--crm-on-action': onAction } : {}),
      }
    : {};
  const settingsOpen = useSettingsModalStore((state) => state.path !== null);
  const quickCreateOpen = useFlowStore((state) => state.quickCreateOpen);
  const dialogRef = useRef<HTMLDialogElement>(null);
  const mainRef = useRef<HTMLElement>(null);
  const { setOpen: setPaletteOpen } = useCommandPalette();
  const { isOpen: caretOpen, toggle: toggleCaret } = useCaretStore();
  const [caretMounted, setCaretMounted] = useState(caretOpen);
  const caretWidth = useCaretSettingsStore((state) => state.panelWidth);
  const { data: notifications = [] } = useNotifications();
  const { data: addons = [] } = useAddons();
  const portalLayout = usePortalLayout(user?.id);
  const unreadCount = notifications.filter((notification) => !notification.read).length;
  const isInsights = pathname === '/insights' || pathname.startsWith('/insights/');
  const isAutomations = pathname === '/automations' || pathname.startsWith('/automations/');
  const isCaret = pathname === '/caret';
  const isRecruit = pathname === '/recruit' || pathname.startsWith('/recruit/');
  const isHr = pathname === '/hr' || pathname.startsWith('/hr/');
  const hrConfig = useHRConfigStore((state) => state.config);
  const isFinance = pathname === '/finance' || pathname.startsWith('/finance/');
  const isErp = pathname === '/erp' || pathname.startsWith('/erp/');
  const isCrm = pathname === '/crm' || pathname.startsWith('/crm/');
  const isFlow = pathname === '/flow' || pathname.startsWith('/flow/');
  const insightsPage = insightsNavItems.find((item) => item.path === pathname);
  const pageKey =
    routeKeys[pathname] ||
    Object.keys(routeKeys)
      .filter((key) => key !== '/' && pathname.startsWith(`${key}/`))
      .sort((a, b) => b.length - a.length)
      .map((key) => routeKeys[key])[0];
  const crmPage = [...crmNavItems, ...crmSalesNavItems]
    .filter(
      (item) =>
        pathname === item.path || (item.path !== '/crm' && pathname.startsWith(`${item.path}/`))
    )
    .sort((a, b) => b.path.length - a.path.length)[0];
  const erpPage = erpNavItems
    .filter(
      (item) =>
        pathname === item.path || (item.path !== '/erp' && pathname.startsWith(`${item.path}/`))
    )
    .sort((a, b) => b.path.length - a.path.length)[0];
  const financePage = financeNavItems
    .filter(
      (item) =>
        pathname === item.path || (item.path !== '/finance' && pathname.startsWith(`${item.path}/`))
    )
    .sort((a, b) => b.path.length - a.path.length)[0];
  const hrPage = hrNavItems
    .filter(
      (item) =>
        pathname === item.path || (item.path !== '/hr' && pathname.startsWith(`${item.path}/`))
    )
    .sort((a, b) => b.path.length - a.path.length)[0];
  const recruitPage = recruitNavItems.find((item) => pathname === item.path);
  const pageTitle = isRecruit
    ? t(`recruitAdmin.nav.${recruitPage?.key ?? 'careers'}`)
    : isHr
      ? t(`hrNav.${hrPage?.key ?? 'appName'}`)
      : isFinance
        ? t(`financeNav.${financePage?.key ?? 'appName'}`)
        : isErp
          ? t(`erpNav.${erpPage?.key ?? 'appName'}`)
          : isCrm
            ? t(`crmNav.${crmPage?.key ?? 'appName'}`)
            : isFlow
              ? t('appNames.short.flow')
              : isInsights
                ? t(`insightsNav.${insightsPage?.key || 'overview'}`)
                : pathname.startsWith('/team/groups')
                  ? t('common.nav.groups')
                  : pageKey
                    ? t(`common.nav.${pageKey}`)
                    : t('common.portal.home');
  const nav = (
    isAutomations
      ? [
          { key: 'automationWorkflows', path: '/automations', icon: AutomationsIcon },
          { key: 'newAutomation', path: '/automations/new', icon: Plus },
          { key: 'webhooks', path: '/automations/webhooks', icon: Webhook },
        ]
      : isRecruit
        ? [...recruitNavItems]
        : isHr
          ? getHrNavItems(hrConfig)
          : isFinance
            ? financeNavItems.filter(
                (item) => item.path !== '/finance' || getFinanceHomePath(user) === '/finance'
              )
            : isErp
              ? [...erpNavItems]
              : isCrm
                ? [...crmSidebarItems]
                : arrangeItems(
                    [...rawNavItems, ...(isAddonEnabled(addons, 'menu') ? rawMenuSubItems : [])]
                      .sort(
                        (a, b) => Number(a.key === 'extensions') - Number(b.key === 'extensions')
                      )
                      .map((item) => ({ ...item, id: item.key ?? item.path })),
                    portalLayout?.workspace
                  ).top
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
  // The tour's overlay would cover the settings modal, so it waits for it to close.
  useOnboardingTour(pathname === '/' && !settingsOpen && canAccessPortalPath(user, '/team'));
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
        if (!isCaret) toggleCaret();
        return;
      }
      if ((event.metaKey || event.ctrlKey) && event.key === '\\') {
        event.preventDefault();
        setCollapsed((value) => !value);
      }
    };
    window.addEventListener('keydown', key);
    return () => window.removeEventListener('keydown', key);
  }, [toggleCaret, isCaret]);
  const toggleNavigation = () => {
    if (isMobile) setMobileOpen((value) => !value);
    else setCollapsed((value) => !value);
  };
  const actions = (
    <div className="portal-topbar-actions">
      {!isCaret && (
        <button
          type="button"
          className="portal-icon-button"
          onClick={toggleCaret}
          aria-label="Caret"
          title="Caret"
          aria-pressed={caretOpen}
          data-tour="caret"
        >
          <CaretIcon size={18} />
        </button>
      )}
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
  const isHomeApp =
    !isInsights &&
    !isFlow &&
    !isCaret &&
    !isAutomations &&
    !isCrm &&
    !isErp &&
    !isFinance &&
    !isHr &&
    !isRecruit;
  const activeRail = isInsights
    ? 'insights'
    : isFlow
      ? 'flow'
      : isCrm
        ? 'crm'
        : isErp
          ? 'erp'
          : isFinance
            ? 'finance'
            : isHr
              ? 'hr'
              : isAutomations
                ? 'automations'
                : isCaret
                  ? 'caret'
                  : isHomeApp
                    ? 'home'
                    : '';
  const railEntries: RailEntry[] = railItems(user, t).map((item) => ({
    ...item,
    active: item.id === 'settings' ? settingsOpen : item.id === activeRail,
  }));
  const rail = arrangeItems(railEntries, portalLayout?.rail, RAIL_BOTTOM);
  const railButton = (entry: RailEntry, group: 'top' | 'bottom') =>
    entry.id === 'profile' ? (
      <ProfileMenu key={entry.id} opensDown={group === 'top'} />
    ) : (
      <Tooltip
        key={entry.id}
        content={entry.label}
        side={i18n.dir() === 'rtl' ? 'left' : 'right'}
        focusable={false}
        contentClassName={`portal-rail-tooltip portal-rail-tooltip-${entry.id}`}
      >
        {entry.href ? (
          <a href={entry.href} className="portal-rail-button" aria-label={entry.label}>
            <entry.icon size={21} />
          </a>
        ) : (
          <Link
            to={entry.to}
            data-tour={entry.id === 'settings' ? 'settings' : undefined}
            className={`portal-rail-button${entry.active ? ' is-active' : ''}`}
            aria-label={entry.label}
            aria-current={entry.active && entry.id !== 'settings' ? 'page' : undefined}
            onClick={() => setMobileOpen(false)}
          >
            <entry.icon size={21} />
          </Link>
        )}
      </Tooltip>
    );
  const navigate = useNavigate();
  const queryClient = useQueryClient();
  // A page Caret points at opens with Caret still beside it; on a phone the
  // panel would cover the page, so it steps aside.
  const openPage = useCallback(
    ({ path, action }: CaretPageSpec) => {
      const caret = useCaretStore.getState();
      if (isMobile) caret.close();
      else if (!caret.isOpen) caret.open();
      void navigate({ href: path });
      if (action) spotlightControl(action);
    },
    [isMobile, navigate]
  );
  // The card's sketch only needs the order, which changes far less often
  // than the entries are rebuilt.
  const railOrder = [rail.top, rail.bottom]
    .map((group) => group.map((e) => e.id).join(','))
    .join('|');
  const railPreview = useMemo(() => {
    const [top, bottom] = railOrder.split('|').map((ids) =>
      ids
        .split(',')
        .filter(Boolean)
        .map((id) => ({ id, icon: RAIL_ICONS[id] }))
    );
    return { top, bottom };
  }, [railOrder]);
  const caretHost = useMemo<CaretHost>(
    () => ({
      openPage,
      renderPage: (spec) => <PortalPageCard spec={spec} rail={railPreview} onOpen={openPage} />,
    }),
    [openPage, railPreview]
  );
  // What Caret changes shows at once: its Portal tools refresh the layout,
  // and an approved change refreshes whatever is on screen.
  useEffect(
    () =>
      onCaretToolSuccess((event) => {
        if (event.name.startsWith('portal.')) {
          void queryClient.invalidateQueries({ queryKey: portalLayoutKey });
        } else if (event.approved) {
          void queryClient.invalidateQueries();
        }
      }),
    [queryClient]
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
          <RtylrLogo />
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
      {!isAutomations && !isCrm && !isErp && !isFinance && !isHr && !isRecruit && (
        <BusinessSwitcher />
      )}
      <nav
        id={mobile ? 'portal-mobile-navigation' : 'portal-app-navigation'}
        className="portal-app-navigation"
        aria-label={t('common.portal.pageNavigation')}
      >
        <p className="portal-section-label">
          {isAutomations
            ? t('common.nav.automations')
            : isRecruit
              ? t('appNames.short.recruit')
              : isHr
                ? t('appNames.short.hr')
                : isFinance
                  ? t('appNames.short.finance')
                  : isErp
                    ? t('appNames.short.erp')
                    : isCrm
                      ? t('appNames.short.crm')
                      : isInsights
                        ? t('appNames.short.insights')
                        : t('common.portal.workspace')}
        </p>
        {nav.map((item, index) => {
          const group = 'group' in item ? item.group : undefined;
          const previous = index > 0 ? nav[index - 1] : undefined;
          const startsGroup =
            group && (!previous || !('group' in previous) || previous.group !== group);
          return (
            <Fragment key={item.path}>
              {startsGroup && (
                <p className="portal-section-label portal-nav-group">
                  {t(`crmNav.sections.${group}`)}
                </p>
              )}
              <PortalNavLink
                path={item.path}
                icon={item.icon}
                tourId={mobile ? undefined : item.key}
                label={t(
                  `${isRecruit ? 'recruitAdmin.nav' : isHr ? 'hrNav' : isFinance ? 'financeNav' : isErp ? 'erpNav' : isCrm ? 'crmNav' : 'common.nav'}.${item.key}`
                )}
                onNavigate={() => setMobileOpen(false)}
              />
            </Fragment>
          );
        })}
      </nav>
    </>
  );
  const shell = (
    <div className="portal-shell" data-app={accent ? appId : undefined} style={appStyle}>
      <SettingsModal />
      <a className="portal-skip-link" href="#main-content">
        {t('common.skipToMainContent')}
      </a>
      <nav className="portal-rail" aria-label={t('common.portal.appNavigation')}>
        {rail.top.map((entry, index) => (
          <Fragment key={entry.id}>
            {entry.id === 'automations' && index > 0 && (
              <div className="portal-rail-divider" role="separator" />
            )}
            {railButton(entry, 'top')}
          </Fragment>
        ))}
        <div className="portal-rail-spacer" />
        {rail.bottom.map((entry) => railButton(entry, 'bottom'))}
      </nav>
      <div className="portal-workspace">
        {!isInsights && !isFlow && !isCaret && !collapsed && (
          <aside className="portal-sidebar">{sidebar()}</aside>
        )}
        <div className="portal-app-panel">
          {isInsights ? (
            <InsightsHeader actions={actions} />
          ) : !isFlow && !isCaret ? (
            <header className="portal-topbar">
              <div className="portal-topbar-start">
                {!isFlow && (
                  <button
                    type="button"
                    className={`portal-icon-button portal-show-sidebar${collapsed ? ' is-visible' : ''}`}
                    onClick={toggleNavigation}
                    aria-label={t('common.toggleNavigation')}
                    aria-expanded={isMobile ? mobileOpen : !collapsed}
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
            className={`portal-page${isRecruit ? ' is-recruit' : ''}${isHr ? ' is-hr' : ''}${isFinance ? ' is-finance' : ''}${isErp ? ' is-erp' : ''}${isCrm ? ' is-crm' : ''}${isFlow ? ' is-flow' : ''}${isCaret ? ' is-caret' : ''}`}
          >
            <PortalHeaderActionsProvider value={actions}>
              <SidebarCollapseProvider value={{ collapsed, setCollapsed }}>
                {children}
              </SidebarCollapseProvider>
            </PortalHeaderActionsProvider>
          </main>
        </div>
        <div
          className="portal-caret"
          inert={!caretOpen || isCaret}
          aria-hidden={!caretOpen || isCaret}
          style={{
            width: caretOpen && !isCaret ? `min(${caretWidth}px, calc(100vw - 72px))` : 0,
            visibility: caretOpen && !isCaret ? 'visible' : 'hidden',
          }}
        >
          {!isCaret && (caretOpen || caretMounted) && (
            <CaretPanel
              apiBaseUrl={resolveApiBase()}
              variant="inline"
              shortcutsEnabled={false}
              appId={
                isRecruit
                  ? 'recruit'
                  : isHr
                    ? 'hr'
                    : isFinance
                      ? 'finance'
                      : isErp
                        ? 'erp'
                        : isCrm
                          ? 'crm'
                          : isFlow
                            ? 'flow'
                            : isInsights
                              ? 'insights'
                              : 'dash'
              }
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
        {mobileOpen && !isInsights && !isFlow && !isCaret && sidebar(true)}
      </dialog>
      <CommandPalette />
      {quickCreateOpen && (
        <Suspense fallback={null}>
          <FlowQuickCreate />
        </Suspense>
      )}
    </div>
  );
  return (
    <ThemeScopeProvider style={appStyle}>
      <CaretHostProvider value={caretHost}>{shell}</CaretHostProvider>
    </ThemeScopeProvider>
  );
}
