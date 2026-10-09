import { recruitNavItems } from '../modules/recruit/navigation';
import '../modules/recruit/recruit.css';
import { getHrNavItems, hrNavItems } from '../modules/hr/navigation';
import { useHRConfigStore } from '../modules/hr/shared/stores/hrConfigStore';
import '../modules/hr/hr.css';
import {
  useEffect,
  useMemo,
  useRef,
  useState,
  lazy,
  Suspense,
  type ReactNode,
  type CSSProperties,
} from 'react';
import { Link, useRouterState } from '@tanstack/react-router';
import { useTranslation } from 'react-i18next';
import { resolveApiBase, resolveAppUrl } from '@voxire/config';
import { ThemeScopeProvider } from '@voxire/ui';
import { useMediaQuery, deriveAccentFgAcross, deriveOnAccent, compositeOver } from '@voxire/utils';
import { useThemeStore } from '../shared/stores/themeStore';
import { PanelLeft, Search, Bell, Settings, Workflow, Plus, X } from 'lucide-react';
import {
  HomeIcon,
  InsightsIcon,
  FlowIcon,
  CrmIcon,
  ErpIcon,
  FinanceIcon,
  HrIcon,
  PosIcon,
} from '@voxire/ui';
import { CaretIcon, CaretPanel, useCaretStore, useCaretSettingsStore } from '@voxire/caret';
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
import { crmNavItems, crmSalesNavItems } from '../modules/crm/navigation';
import { insightsNavItems } from '../modules/insights/navigation';
import ProfileMenu from './ProfileMenu';
import { Tooltip } from '../shared/components/ui/tooltip';
import SettingsNavigation from './SettingsNavigation';
import SettingsModal from '../features/settings/SettingsModal';
import { useSettingsModalStore } from '../features/settings/settingsModalStore';
import { PortalHeaderActionsProvider } from './PortalHeaderActions';
import { SidebarCollapseProvider } from './SidebarCollapseContext';
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
import '../modules/crm/crm.css';
import '../modules/erp/erp.css';
import '../modules/finance/finance.css';
import RtylrLogo from '../shared/components/RtylrLogo';

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
  const actionAccent = accent ?? '#111111';
  const appStyle: CSSProperties & Record<string, string> = accent
    ? {
        '--app-action-bg': actionAccent,
        '--app-on-action': deriveOnAccent(actionAccent),
        '--app-action-hover': `color-mix(in srgb, ${actionAccent} 90%, white)`,
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
              '--flow-action-bg': accent,
              '--flow-on-accent': deriveOnAccent(accent),
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
        ...(appId === 'crm' ? { '--crm-on-accent': deriveOnAccent(accent) } : {}),
      }
    : {};
  const [settingsQuery, setSettingsQuery] = useState('');
  const settingsOpen = useSettingsModalStore((state) => state.path !== null);
  const quickCreateOpen = useFlowStore((state) => state.quickCreateOpen);
  const dialogRef = useRef<HTMLDialogElement>(null);
  const mainRef = useRef<HTMLElement>(null);
  const { setOpen: setPaletteOpen } = useCommandPalette();
  const { isOpen: caretOpen, toggle: toggleCaret } = useCaretStore();
  const [caretMounted, setCaretMounted] = useState(caretOpen);
  const caretWidth = useCaretSettingsStore((state) => state.panelWidth);
  const { data: notifications = [] } = useNotifications();
  const { data: addons = [], isError: addonsError } = useAddons();
  const unreadCount = notifications.filter((notification) => !notification.read).length;
  const isSettings =
    pathname.startsWith('/settings') || ['/security', '/api-keys', '/webhooks'].includes(pathname);
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
  const availableInsights = insightsNavItems.filter((item) => canAccessPortalPath(user, item.path));
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
          { key: 'automationWorkflows', path: '/automations', icon: Workflow },
          { key: 'newAutomation', path: '/automations/new', icon: Plus },
        ]
      : isSettings
        ? [
            ...settingsNavItems,
            ...addonSettingsItems.filter((item) => isAddonEnabled(addons, item.addonId)),
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
                  ? [
                      ...crmNavItems,
                      ...(addonsError || isAddonEnabled(addons, 'sales_pipeline')
                        ? crmSalesNavItems
                        : []),
                    ]
                  : [
                      ...rawNavItems,
                      ...(isAddonEnabled(addons, 'menu') ? rawMenuSubItems : []),
                    ].sort(
                      (a, b) => Number(a.key === 'extensions') - Number(b.key === 'extensions')
                    )
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
  useOnboardingTour(pathname === '/' && canAccessPortalPath(user, '/team'));
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
      {!isSettings && !isAutomations && !isCrm && !isErp && !isFinance && !isHr && !isRecruit && (
        <BusinessSwitcher />
      )}
      <nav
        id={mobile ? 'portal-mobile-navigation' : 'portal-app-navigation'}
        className="portal-app-navigation"
        aria-label={t('common.portal.pageNavigation')}
      >
        {!isSettings && (
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
              path={item.path}
              icon={item.icon}
              key={item.path}
              tourId={mobile ? undefined : item.key}
              label={t(
                `${isRecruit ? 'recruitAdmin.nav' : isHr ? 'hrNav' : isFinance ? 'financeNav' : isErp ? 'erpNav' : isCrm ? 'crmNav' : 'common.nav'}.${item.key}`
              )}
              onNavigate={() => setMobileOpen(false)}
            />
          ))
        )}
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
        <Tooltip
          content={t('common.portal.home')}
          side={i18n.dir() === 'rtl' ? 'left' : 'right'}
          focusable={false}
          contentClassName="portal-rail-tooltip"
        >
          <Link
            to="/"
            className={`portal-rail-button${!isSettings && !isInsights && !isFlow && !isCaret && !isAutomations && !isCrm && !isErp && !isFinance && !isHr && !isRecruit ? ' is-active' : ''}`}
            aria-label={t('common.portal.home')}
            aria-current={
              !isSettings &&
              !isInsights &&
              !isFlow &&
              !isCaret &&
              !isAutomations &&
              !isCrm &&
              !isErp &&
              !isFinance &&
              !isHr &&
              !isRecruit
                ? 'page'
                : undefined
            }
          >
            <HomeIcon size={21} />
          </Link>
        </Tooltip>
        {availableInsights[0] && (
          <Tooltip
            content={t('appNames.short.insights')}
            side={i18n.dir() === 'rtl' ? 'left' : 'right'}
            focusable={false}
            contentClassName="portal-rail-tooltip"
          >
            <Link
              to={availableInsights[0].path}
              className={`portal-rail-button${isInsights ? ' is-active' : ''}`}
              aria-label={t('appNames.short.insights')}
              aria-current={isInsights ? 'page' : undefined}
            >
              <InsightsIcon size={21} />
            </Link>
          </Tooltip>
        )}
        {canAccessPortalPath(user, '/flow') && (
          <Tooltip
            content={t('appNames.short.flow')}
            side={i18n.dir() === 'rtl' ? 'left' : 'right'}
            focusable={false}
            contentClassName="portal-rail-tooltip"
          >
            <Link
              to="/flow"
              className={`portal-rail-button${isFlow ? ' is-active' : ''}`}
              aria-label={t('appNames.short.flow')}
              aria-current={isFlow ? 'page' : undefined}
            >
              <FlowIcon size={21} />
            </Link>
          </Tooltip>
        )}
        {canAccessPortalPath(user, '/crm') && (
          <Tooltip
            content={t('appNames.short.crm')}
            side={i18n.dir() === 'rtl' ? 'left' : 'right'}
            focusable={false}
            contentClassName="portal-rail-tooltip"
          >
            <Link
              to="/crm"
              className={`portal-rail-button${isCrm ? ' is-active' : ''}`}
              aria-label={t('appNames.short.crm')}
              aria-current={isCrm ? 'page' : undefined}
            >
              <CrmIcon size={21} />
            </Link>
          </Tooltip>
        )}
        {canAccessPortalPath(user, '/erp') && (
          <Tooltip
            content={t('appNames.short.erp')}
            side={i18n.dir() === 'rtl' ? 'left' : 'right'}
            focusable={false}
            contentClassName="portal-rail-tooltip"
          >
            <Link
              to="/erp"
              className={`portal-rail-button${isErp ? ' is-active' : ''}`}
              aria-label={t('appNames.short.erp')}
              aria-current={isErp ? 'page' : undefined}
            >
              <ErpIcon size={21} />
            </Link>
          </Tooltip>
        )}
        {canAccessPortalPath(user, '/finance') && (
          <Tooltip
            content={t('appNames.short.finance')}
            side={i18n.dir() === 'rtl' ? 'left' : 'right'}
            focusable={false}
            contentClassName="portal-rail-tooltip"
          >
            <Link
              to="/finance"
              className={`portal-rail-button${isFinance ? ' is-active' : ''}`}
              aria-label={t('appNames.short.finance')}
              aria-current={isFinance ? 'page' : undefined}
            >
              <FinanceIcon size={21} />
            </Link>
          </Tooltip>
        )}
        {canAccessPortalPath(user, '/hr') && (
          <Tooltip
            content={t('appNames.short.hr')}
            side={i18n.dir() === 'rtl' ? 'left' : 'right'}
            focusable={false}
            contentClassName="portal-rail-tooltip"
          >
            <Link
              to="/hr"
              className={`portal-rail-button${isHr ? ' is-active' : ''}`}
              aria-label={t('appNames.short.hr')}
              aria-current={isHr ? 'page' : undefined}
            >
              <HrIcon size={21} />
            </Link>
          </Tooltip>
        )}
        {apps?.includes('pos') && (
          <Tooltip
            content={t('appNames.short.pos')}
            side={i18n.dir() === 'rtl' ? 'left' : 'right'}
            focusable={false}
            contentClassName="portal-rail-tooltip"
          >
            <a
              href={resolveAppUrl('pos')}
              className="portal-rail-button"
              aria-label={t('appNames.short.pos')}
            >
              <PosIcon size={21} />
            </a>
          </Tooltip>
        )}
        {canAccessPortalPath(user, '/automations') && (
          <>
            <div className="portal-rail-divider" role="separator" />
            <Tooltip
              content={t('common.nav.automations')}
              side={i18n.dir() === 'rtl' ? 'left' : 'right'}
              focusable={false}
              contentClassName="portal-rail-tooltip"
            >
              <Link
                to="/automations"
                className={`portal-rail-button${isAutomations ? ' is-active' : ''}`}
                aria-label={t('common.nav.automations')}
                aria-current={isAutomations ? 'page' : undefined}
                onClick={() => setMobileOpen(false)}
              >
                <Workflow size={20} />
              </Link>
            </Tooltip>
          </>
        )}
        <div className="portal-rail-spacer" />
        <Tooltip
          content={'Caret'}
          side={i18n.dir() === 'rtl' ? 'left' : 'right'}
          focusable={false}
          contentClassName="portal-rail-tooltip"
        >
          <Link
            to="/caret"
            className={`portal-rail-button${isCaret ? ' is-active' : ''}`}
            aria-label="Caret"
            aria-current={isCaret ? 'page' : undefined}
          >
            <CaretIcon size={21} />
          </Link>
        </Tooltip>
        <Link
          data-tour="settings"
          to="/settings"
          className={`portal-rail-button${settingsOpen ? ' is-active' : ''}`}
          aria-label={t('common.nav.settings')}
          title={t('common.nav.settings')}
        >
          <Settings size={21} />
        </Link>
        <ProfileMenu />
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
  return <ThemeScopeProvider style={appStyle}>{shell}</ThemeScopeProvider>;
}
