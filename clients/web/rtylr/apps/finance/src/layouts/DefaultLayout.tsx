import React, { useEffect, useRef, useState } from 'react';
import { useTranslation } from 'react-i18next';
import { Avatar } from '@voxire/ui';
import { parseEffectiveAccess, signOut, useUserStore } from '@voxire/auth';
import { buildCaretRouteContext, CaretPanel, CaretButton, useCaretStore } from '@voxire/caret';
import TrialBanner from '../shared/components/TrialBanner';
import { userClient } from '../shared/lib/clients';
import { useForesightLinks } from '@voxire/utils';
import { useNavigate, Link, Outlet, useRouterState } from '@tanstack/react-router';
import { motion, AnimatePresence } from 'motion/react';
import { getAvatarUrl } from '@voxire/config';
import {
  LayoutDashboard,
  FileText,
  Receipt,
  TrendingUp,
  Calculator,
  Bell,
  ChevronRight,
  Sun,
  Moon,
  Monitor,
  LogOut,
  Search as SearchIcon,
  Repeat,
  ClipboardList,
  Wallet,
  Building2,
  Shuffle,
  BarChart2,
  LineChart,
  BookOpen,
  Settings as SettingsIcon,
  Tag,
  FileCheck,
  Menu,
  X,
} from 'lucide-react';
import GlobalSearch from '../shared/components/GlobalSearch';
import { canFinancePermission, type FinancePermission } from '../shared/lib/financePermissions';

const ACCENT = 'var(--finance-accent)';

interface NavItem {
  labelKey: string;
  path: string;
  icon: React.ComponentType<{ size?: number; className?: string; style?: React.CSSProperties }>;
  // Resource required to see this nav item. Bank feeds, reconciliation, the
  // chart of accounts and tax configuration are admin-tier resources in
  // packages/permissions/src/matrix.ts (not granted to billing/operator/staff
  // even for viewing), so their nav entries are hidden rather than shown to a
  // role that cannot open the underlying screen's primary action.
  requiredPermission?: FinancePermission;
}

interface NavGroup {
  titleKey: string;
  items: NavItem[];
}

const navGroups: NavGroup[] = [
  {
    titleKey: 'nav.groups.overview',
    items: [{ labelKey: 'nav.dashboard', path: '/', icon: LayoutDashboard }],
  },
  {
    titleKey: 'nav.groups.sales',
    items: [
      { labelKey: 'nav.invoices', path: '/invoices', icon: FileText },
      {
        labelKey: 'nav.recurring',
        path: '/invoices/recurring',
        icon: Repeat,
        requiredPermission: 'finance:recurring_invoices:view',
      },
      {
        labelKey: 'nav.quotes',
        path: '/quotes',
        icon: ClipboardList,
        requiredPermission: 'finance:quotes:view',
      },
    ],
  },
  {
    titleKey: 'nav.groups.spend',
    items: [
      { labelKey: 'nav.expenses', path: '/expenses', icon: Receipt },
      { labelKey: 'nav.categories', path: '/categories', icon: Tag },
      { labelKey: 'nav.bills', path: '/bills', icon: Wallet },
    ],
  },
  {
    titleKey: 'nav.groups.banking',
    items: [
      {
        labelKey: 'nav.bankFeeds',
        path: '/bank',
        icon: Building2,
        requiredPermission: 'finance:accounts',
      },
      {
        labelKey: 'nav.reconcile',
        path: '/reconcile',
        icon: Shuffle,
        requiredPermission: 'finance:reconcile',
      },
    ],
  },
  {
    titleKey: 'nav.groups.reports',
    items: [
      { labelKey: 'nav.pl', path: '/pl', icon: TrendingUp },
      { labelKey: 'nav.balanceSheet', path: '/reports/balance-sheet', icon: BarChart2 },
      { labelKey: 'nav.cashFlow', path: '/reports/cash-flow', icon: LineChart },
      { labelKey: 'nav.taxSummary', path: '/reports/tax', icon: FileCheck },
    ],
  },
  {
    titleKey: 'nav.groups.accounting',
    items: [
      { labelKey: 'nav.journal', path: '/journal', icon: BookOpen },
      {
        labelKey: 'nav.accounts',
        path: '/accounts',
        icon: FileText,
        requiredPermission: 'finance:accounts',
      },
      {
        labelKey: 'nav.taxConfig',
        path: '/tax',
        icon: Calculator,
        requiredPermission: 'finance:tax',
      },
    ],
  },
];

const routeLabels: Record<string, string> = {
  '/': 'Dashboard',
  '/invoices': 'Invoices',
  '/invoices/new': 'New Invoice',
  '/invoices/recurring': 'Recurring',
  '/quotes': 'Quotes',
  '/expenses': 'Expenses',
  '/expenses/new': 'New Expense',
  '/categories': 'Categories',
  '/bills': 'Bills',
  '/bank': 'Bank Feeds',
  '/reconcile': 'Reconcile',
  '/pl': 'Profit & Loss',
  '/reports/balance-sheet': 'Balance Sheet',
  '/reports/cash-flow': 'Cash Flow',
  '/reports/tax': 'Tax Summary',
  '/tax': 'Tax',
  '/journal': 'Journal',
  '/accounts': 'Chart of Accounts',
  '/settings': 'Settings',
};

function getPageTitle(pathname: string): string {
  if (routeLabels[pathname]) return routeLabels[pathname];
  if (pathname.startsWith('/invoices/')) return 'Invoice';
  if (pathname.startsWith('/expenses/')) return 'Expense';
  return 'Finance';
}

type Theme = 'light' | 'dark' | 'system';

function applyTheme(t: Theme) {
  const resolved =
    t === 'system'
      ? window.matchMedia('(prefers-color-scheme: dark)').matches
        ? 'dark'
        : 'light'
      : t;
  document.documentElement.setAttribute('data-theme', resolved);
  localStorage.setItem('finance-theme', t);
}

const SidebarNavItem: React.FC<NavItem> = ({ labelKey, path, icon: Icon }) => {
  const { t } = useTranslation();
  const pathname = useRouterState({ select: (s) => s.location.pathname });
  const isActive =
    path === '/' ? pathname === '/' : pathname === path || pathname.startsWith(path + '/');

  return (
    <Link
      to={path}
      style={{
        color: isActive ? 'var(--hub-text)' : 'var(--hub-text-muted)',
        textDecoration: 'none',
      }}
      className={[
        'relative flex items-center gap-3 px-3 py-2 rounded-lg text-sm font-medium transition-colors duration-150',
        isActive ? '' : 'hover:bg-[var(--hub-hover)]',
      ].join(' ')}
    >
      {isActive && (
        <motion.span
          layoutId="sidebar-active"
          className="absolute inset-0 rounded-lg"
          style={{ background: 'var(--hub-active)' }}
          transition={{ type: 'spring', stiffness: 400, damping: 35 }}
        />
      )}
      <Icon size={15} style={{ position: 'relative', flexShrink: 0 }} />
      <span style={{ position: 'relative', fontSize: 12.5 }}>{t(labelKey)}</span>
      {isActive && (
        <span
          style={{
            position: 'relative',
            marginInlineStart: 'auto',
            height: 6,
            width: 6,
            borderRadius: '50%',
            flexShrink: 0,
            background: ACCENT,
          }}
        />
      )}
    </Link>
  );
};

const DefaultLayout: React.FC = () => {
  const { t } = useTranslation();
  const pathname = useRouterState({ select: (s) => s.location.pathname });
  const search = useRouterState({ select: (s) => s.location.searchStr });
  const navigate = useNavigate();
  const pageTitle = getPageTitle(pathname);
  const caretPageContext = React.useMemo(
    () =>
      buildCaretRouteContext({
        appId: 'finance',
        appName: 'Finance',
        pathname,
        search,
        page: pageTitle,
        routeLabels,
      }),
    [pathname, search, pageTitle]
  );
  const [theme, setThemeState] = useState<Theme>(
    () => (localStorage.getItem('finance-theme') as Theme) || 'dark'
  );
  const [searchOpen, setSearchOpen] = useState(false);
  const [showUserMenu, setShowUserMenu] = useState(false);
  const [sidebarOpen, setSidebarOpen] = useState(false);
  const sidebarRef = useRef<HTMLElement | null>(null);
  const financeUser = useUserStore((s) => s.user);
  const setAuthUser = useUserStore((s) => s.setUser);
  const effectiveAccess = financeUser?.effectiveAccess;
  const userId = Number(financeUser?.id) || 0;
  const userEmail = financeUser?.email || financeUser?.emailAddress || null;
  const userName = financeUser?.name || financeUser?.firstName || userEmail || 'User';
  const { isOpen: caretOpen } = useCaretStore();
  const visibleNavGroups = React.useMemo(
    () =>
      navGroups
        .map((group) => ({
          ...group,
          items: group.items.filter(
            (item) => !item.requiredPermission || canFinancePermission(item.requiredPermission)
          ),
        }))
        .filter((group) => group.items.length > 0),
    [effectiveAccess]
  );

  useEffect(() => {
    if (financeUser?.name || financeUser?.email) return;
    userClient
      .getMe({})
      .then((res) => {
        if (res?.id) {
          setAuthUser({
            id: String(res.id),
            businessId: String(res.businessId),
            name: res.fullName ?? '',
            email: res.email ?? '',
            businessName: res.businessName ?? '',
            isConsoleAdmin: res.isConsoleAdmin,
            effectiveAccess: parseEffectiveAccess(res.effectiveAccess) ?? undefined,
          });
        }
      })
      .catch(() => undefined);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [financeUser]);

  useForesightLinks(sidebarRef, {
    '/': () => import('../features/dashboard/Dashboard'),
    '/invoices': () => import('../features/invoices/Invoices'),
    '/invoices/recurring': () => import('../features/recurring-invoices/RecurringInvoices'),
    '/quotes': () => import('../features/quotes/Quotes'),
    '/expenses': () => import('../features/expenses/Expenses'),
    '/categories': () => import('../features/categories/Categories'),
    '/bills': () => import('../features/bills/Bills'),
    '/bank': () => import('../features/bank-feeds/BankFeeds'),
    '/reconcile': () => import('../features/reconcile/Reconcile'),
    '/pl': () => import('../features/profit-loss/ProfitLoss'),
    '/reports/balance-sheet': () => import('../features/balance-sheet/BalanceSheet'),
    '/reports/cash-flow': () => import('../features/cash-flow/CashFlow'),
    '/reports/tax': () => import('../features/tax/TaxReport'),
    '/journal': () => import('../features/journal/Journal'),
    '/accounts': () => import('../features/chart-of-accounts/ChartOfAccounts'),
    '/tax': () => import('../features/tax/Tax'),
    '/settings': () => import('../features/settings/Settings'),
  });

  const handleTheme = (t: Theme) => {
    setThemeState(t);
    applyTheme(t);
  };

  // Keyboard shortcuts: cmd+k search, n new invoice, e new expense, / focus search
  useEffect(() => {
    const handler = (e: KeyboardEvent) => {
      const target = e.target as HTMLElement | null;
      const isEditing =
        target &&
        (target.tagName === 'INPUT' || target.tagName === 'TEXTAREA' || target.isContentEditable);
      if ((e.key === 'k' && (e.metaKey || e.ctrlKey)) || (e.key === '/' && !isEditing)) {
        e.preventDefault();
        setSearchOpen(true);
        return;
      }
      if (isEditing) return;
      if (e.key === 'n') {
        e.preventDefault();
        navigate({ to: '/invoices/new' });
      } else if (e.key === 'e') {
        e.preventDefault();
        navigate({ to: '/expenses/new' });
      } else if (e.key === 'Escape') {
        setSearchOpen(false);
        setSidebarOpen(false);
      }
    };
    window.addEventListener('keydown', handler);
    return () => window.removeEventListener('keydown', handler);
  }, [navigate]);

  useEffect(() => {
    setSidebarOpen(false);
  }, [pathname]);

  return (
    <>
      <div className="flex h-screen overflow-hidden" style={{ background: 'var(--hub-bg)' }}>
        <div
          className="fixed inset-x-0 top-0 z-40 flex h-11 items-center justify-between px-4 md:hidden"
          style={{
            background: 'var(--hub-surface)',
            borderBottom: '1px solid var(--hub-border)',
          }}
        >
          <button
            type="button"
            onClick={() => setSidebarOpen(true)}
            aria-label={t('common.layout.openNav')}
            aria-expanded={sidebarOpen}
            aria-controls="finance-sidebar"
            className="flex h-11 w-11 items-center justify-center rounded-lg"
            style={{ color: 'var(--hub-text-muted)' }}
          >
            <Menu size={18} />
          </button>
          <h1 style={{ margin: 0, fontSize: 14, fontWeight: 600, color: 'var(--hub-text)' }}>
            {pageTitle}
          </h1>
          <div className="h-11 w-11" aria-hidden="true" />
        </div>

        {sidebarOpen && (
          <div
            className="fixed inset-0 z-40 bg-black/50 md:hidden"
            onClick={() => setSidebarOpen(false)}
            aria-hidden="true"
          />
        )}

        <aside
          id="finance-sidebar"
          aria-label={t('common.layout.primaryNav')}
          className={[
            sidebarOpen ? 'fixed inset-y-0 start-0 z-50 flex' : 'hidden',
            'md:flex w-56 flex-col shrink-0 overflow-hidden',
          ].join(' ')}
          style={{
            background: 'var(--hub-surface)',
            borderInlineEnd: '1px solid var(--hub-border)',
          }}
        >
          <div
            style={{
              display: 'flex',
              alignItems: 'center',
              gap: 10,
              padding: 16,
              borderBottom: '1px solid var(--hub-border)',
              justifyContent: 'space-between',
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
              aria-label={t('common.layout.closeNav')}
              className="flex h-11 w-11 items-center justify-center rounded-lg md:hidden"
              style={{ color: 'var(--hub-text-muted)' }}
            >
              <X size={18} />
            </button>
          </div>

          <div style={{ padding: '6px 8px' }}>
            <button
              type="button"
              onClick={() => setSearchOpen(true)}
              style={{
                display: 'flex',
                alignItems: 'center',
                gap: 8,
                width: '100%',
                padding: '6px 12px',
                borderRadius: 8,
                background: 'var(--hub-hover)',
                border: '1px solid var(--hub-border)',
                color: 'var(--hub-text-muted)',
                fontSize: 12,
                cursor: 'pointer',
              }}
            >
              <SearchIcon size={12} />
              <span>{t('common.layout.search')}</span>
              <span
                style={{ marginInlineStart: 'auto', fontSize: 10, color: 'var(--hub-text-dim)' }}
              >
                {t('common.layout.shortcut')}
              </span>
            </button>
          </div>

          <nav
            ref={sidebarRef}
            style={{
              flex: 1,
              overflowY: 'auto',
              padding: 8,
              display: 'flex',
              flexDirection: 'column',
              gap: 8,
            }}
          >
            {visibleNavGroups.map((group) => (
              <div
                key={group.titleKey}
                style={{ display: 'flex', flexDirection: 'column', gap: 2 }}
              >
                <p
                  style={{
                    padding: '4px 12px 4px',
                    fontSize: 9,
                    fontWeight: 700,
                    textTransform: 'uppercase',
                    letterSpacing: '0.1em',
                    color: 'var(--hub-text-dim)',
                    margin: 0,
                  }}
                >
                  {t(group.titleKey)}
                </p>
                {group.items.map((item) => (
                  <SidebarNavItem key={item.path} {...item} />
                ))}
              </div>
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
            <CaretButton size="md" className="w-full justify-start mb-1" />
            <div className="relative">
              <button
                type="button"
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
                  style={{
                    background: 'color-mix(in srgb, var(--finance-accent) 15%, transparent)',
                  }}
                />
                <div className="flex flex-col items-start min-w-0 flex-1">
                  <span
                    className="text-xs font-medium truncate w-full text-start"
                    style={{ color: 'var(--hub-text)' }}
                  >
                    {userName}
                  </span>
                  {userEmail && (
                    <span
                      className="text-[10px] truncate w-full text-start"
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
                    className="absolute bottom-full start-0 end-0 mb-1 rounded-xl p-2 z-50"
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
                      {t('common.layout.appearance')}
                    </p>
                    <div className="flex gap-1 mb-2">
                      {(
                        [
                          { value: 'light' as Theme, icon: Sun, labelKey: 'nav.theme.light' },
                          { value: 'dark' as Theme, icon: Moon, labelKey: 'nav.theme.dark' },
                          { value: 'system' as Theme, icon: Monitor, labelKey: 'nav.theme.system' },
                        ] as const
                      ).map(({ value, icon: Icon, labelKey }) => (
                        <button
                          type="button"
                          key={value}
                          onClick={() => handleTheme(value)}
                          className="flex-1 flex flex-col items-center gap-1 py-2 rounded-lg text-[10px] font-medium transition-colors"
                          style={{
                            background:
                              theme === value
                                ? 'color-mix(in srgb, var(--finance-accent) 12%, transparent)'
                                : 'var(--hub-hover)',
                            color: theme === value ? ACCENT : 'var(--hub-text-muted)',
                            border:
                              theme === value
                                ? `1px solid color-mix(in srgb, ${ACCENT} 25%, transparent)`
                                : '1px solid transparent',
                          }}
                        >
                          <Icon size={13} />
                          {t(labelKey)}
                        </button>
                      ))}
                    </div>
                    <div
                      className="my-1"
                      style={{ height: '1px', background: 'var(--hub-border)' }}
                    />
                    <button
                      type="button"
                      onClick={() => signOut('finance')}
                      className="flex items-center gap-2 w-full px-2 py-2 rounded-lg text-sm transition-colors"
                      style={{ color: 'var(--finance-danger)' }}
                      onMouseEnter={(e) => {
                        (e.currentTarget as HTMLElement).style.background =
                          'color-mix(in srgb, var(--finance-danger) 8%, transparent)';
                      }}
                      onMouseLeave={(e) => {
                        (e.currentTarget as HTMLElement).style.background = 'transparent';
                      }}
                    >
                      <LogOut size={14} />
                      {t('common.layout.signOut')}
                    </button>
                  </motion.div>
                )}
              </AnimatePresence>
            </div>
          </div>
        </aside>

        {/* Main content + Caret AI inline column */}
        <div className="flex flex-1 overflow-hidden">
          <div className="flex min-w-0 flex-1 flex-col overflow-hidden pt-11 md:pt-0">
            <header
              className="hidden h-11 shrink-0 items-center justify-between px-6 md:flex"
              style={{
                background: 'var(--hub-surface)',
                borderBottom: '1px solid var(--hub-border)',
              }}
            >
              <h1 style={{ margin: 0, fontSize: 13, fontWeight: 600, color: 'var(--hub-text)' }}>
                {pageTitle}
              </h1>
              <div style={{ display: 'flex', alignItems: 'center', gap: 4 }}>
                <button
                  type="button"
                  onClick={() => setSearchOpen(true)}
                  style={{
                    height: 28,
                    width: 28,
                    borderRadius: 8,
                    display: 'flex',
                    alignItems: 'center',
                    justifyContent: 'center',
                    background: 'transparent',
                    border: 'none',
                    cursor: 'pointer',
                    color: 'var(--hub-text-muted)',
                  }}
                >
                  <SearchIcon size={14} />
                </button>
                <button
                  type="button"
                  onClick={undefined}
                  aria-label={t('common.layout.notifications')}
                  style={{
                    height: 28,
                    width: 28,
                    borderRadius: 8,
                    display: 'flex',
                    alignItems: 'center',
                    justifyContent: 'center',
                    background: 'transparent',
                    border: 'none',
                    cursor: 'pointer',
                    color: 'var(--hub-text-muted)',
                  }}
                >
                  <Bell size={14} />
                </button>
                <CaretButton />
              </div>
            </header>

            <TrialBanner />
            <AnimatePresence mode="wait">
              <motion.main
                key={pathname}
                initial={{ opacity: 0, y: 6 }}
                animate={{ opacity: 1, y: 0 }}
                exit={{ opacity: 0, y: -4 }}
                transition={{ duration: 0.18, ease: 'easeOut' }}
                style={{ flex: 1, overflowY: 'auto', padding: 24 }}
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
              appId="finance"
              userId={userId}
              apiBaseUrl={import.meta.env.VITE_API_URL ?? 'https://api.voxire.com/'}
              systemContext="User is in the Finance app (invoices, expenses, bills, P&L, tax)"
              pageContext={caretPageContext}
            />
          </div>
        </div>

        <GlobalSearch open={searchOpen} onClose={() => setSearchOpen(false)} />
      </div>
    </>
  );
};

export default DefaultLayout;
