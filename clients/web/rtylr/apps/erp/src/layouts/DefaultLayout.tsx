import React, { useEffect, useRef, useState } from 'react';
import { cn, Avatar } from '@voxire/ui';
import TrialBanner from '../shared/components/TrialBanner';
import { useForesightLinks } from '@voxire/utils';
import { Link, useNavigate, useRouterState, Outlet } from '@tanstack/react-router';
import { motion, AnimatePresence } from 'motion/react';
import { getAvatarUrl } from '@voxire/config';
import { signOut, useUserStore } from '@voxire/auth';
import { LicenseBanner } from '@voxire/components';
import { licenseClient, userClient } from '../shared/lib/clients';
import { buildCaretRouteContext, CaretPanel, CaretButton, useCaretStore } from '@voxire/caret';
import {
  LayoutDashboard,
  Package,
  ShoppingCart,
  Users,
  Bell,
  ChevronRight,
  Sun,
  Moon,
  Monitor,
  LogOut,
  Warehouse,
  Truck,
  ClipboardList,
  AlertTriangle,
  PackageCheck,
  Sparkles,
  Tags,
  Layers,
  Barcode,
  CalendarX,
  DollarSign,
  Settings as SettingsIcon,
  Search as SearchIcon,
  Keyboard,
  Folder,
  Tag,
  Menu,
  X,
} from 'lucide-react';
import { Toaster } from 'sonner';
import CommandPalette from '../shared/components/CommandPalette';
import KeyboardShortcutsModal from '../shared/components/KeyboardShortcutsModal';
import QuickAddProductModal from '../shared/components/QuickAddProductModal';
import { useTranslation } from 'react-i18next';
import { pageTitleKey, routeLabels } from '../shared/lib/routeLabels';
import { applyErpTheme, useSettings, type ErpTheme } from '../shared/lib/settingsStore';

interface NavItem {
  labelKey: string;
  path: string;
  icon: React.ComponentType<{ size?: number; className?: string }>;
}

interface NavGroup {
  labelKey: string;
  items: NavItem[];
}

const navGroups: NavGroup[] = [
  {
    labelKey: 'nav.groups.workspace',
    items: [
      { labelKey: routeLabels['/'].nav, path: '/', icon: LayoutDashboard },
      { labelKey: routeLabels['/inventory'].nav, path: '/inventory', icon: Package },
    ],
  },
  {
    labelKey: 'nav.groups.catalog',
    items: [
      { labelKey: routeLabels['/categories'].nav, path: '/categories', icon: Folder },
      { labelKey: routeLabels['/brands'].nav, path: '/brands', icon: Tag },
      { labelKey: routeLabels['/variants'].nav, path: '/variants', icon: Layers },
      { labelKey: routeLabels['/price-lists'].nav, path: '/price-lists', icon: Tags },
      { labelKey: routeLabels['/barcodes'].nav, path: '/barcodes', icon: Barcode },
    ],
  },
  {
    labelKey: 'nav.groups.warehouseOps',
    items: [
      { labelKey: routeLabels['/warehouses'].nav, path: '/warehouses', icon: Warehouse },
      { labelKey: routeLabels['/transfers'].nav, path: '/transfers', icon: Truck },
      { labelKey: routeLabels['/stock-takes'].nav, path: '/stock-takes', icon: ClipboardList },
      { labelKey: routeLabels['/receiving'].nav, path: '/receiving', icon: PackageCheck },
      { labelKey: routeLabels['/expiry'].nav, path: '/expiry', icon: CalendarX },
    ],
  },
  {
    labelKey: 'nav.groups.insights',
    items: [
      { labelKey: routeLabels['/alerts'].nav, path: '/alerts', icon: AlertTriangle },
      { labelKey: routeLabels['/reorder'].nav, path: '/reorder', icon: Sparkles },
      { labelKey: routeLabels['/valuation'].nav, path: '/valuation', icon: DollarSign },
    ],
  },
  {
    labelKey: 'nav.groups.purchasing',
    items: [
      { labelKey: routeLabels['/purchasing'].nav, path: '/purchasing', icon: ShoppingCart },
      { labelKey: routeLabels['/suppliers'].nav, path: '/suppliers', icon: Users },
    ],
  },
];

const SidebarNavItem: React.FC<NavItem> = ({ labelKey, path, icon: Icon }) => {
  const { t } = useTranslation();
  const pathname = useRouterState({ select: (s) => s.location.pathname });
  const isActive =
    path === '/' ? pathname === '/' : pathname === path || pathname.startsWith(`${path}/`);

  return (
    <Link
      to={path}
      className={cn(
        'relative flex items-center gap-3 px-3 py-1.5 rounded-lg text-sm font-medium transition-colors duration-150 focus:outline-none focus-visible:ring-1 focus-visible:ring-[var(--hub-border)]',
        !isActive && 'hover:bg-[var(--hub-hover)]'
      )}
      style={{
        color: isActive ? 'var(--hub-text)' : 'var(--hub-text-muted)',
        textDecoration: 'none',
      }}
    >
      {isActive && (
        <span className="absolute inset-0 rounded-lg" style={{ background: 'var(--hub-active)' }} />
      )}
      <Icon size={14} className="relative shrink-0" />
      <span className="relative">{t(labelKey)}</span>
      {isActive && (
        <span
          className="relative h-1 w-1 rounded-full shrink-0"
          style={{ background: 'var(--erp-accent)', marginInlineStart: 'auto' }}
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
  const theme = useSettings((state) => state.theme);
  const setTheme = useSettings((state) => state.setTheme);
  const [showUserMenu, setShowUserMenu] = useState(false);
  const [paletteOpen, setPaletteOpen] = useState(false);
  const [shortcutsOpen, setShortcutsOpen] = useState(false);
  const [quickAddOpen, setQuickAddOpen] = useState(false);
  const [sidebarOpen, setSidebarOpen] = useState(false);
  const sidebarRef = useRef<HTMLElement | null>(null);

  useForesightLinks(sidebarRef, {
    '/': () => import('../features/dashboard/Dashboard'),
    '/inventory': () => import('../features/inventory/Inventory'),
    '/categories': () => import('../features/categories/Categories'),
    '/brands': () => import('../features/brands/Brands'),
    '/variants': () => import('../features/variants/Variants'),
    '/price-lists': () => import('../features/price-lists/PriceLists'),
    '/barcodes': () => import('../features/barcodes/Barcodes'),
    '/warehouses': () => import('../features/warehouses/Warehouses'),
    '/transfers': () => import('../features/stock-transfers/StockTransfers'),
    '/stock-takes': () => import('../features/stock-takes/StockTakes'),
    '/receiving': () => import('../features/purchasing/Receiving'),
    '/expiry': () => import('../features/expiry/Expiry'),
    '/alerts': () => import('../features/low-stock-alerts/LowStockAlerts'),
    '/reorder': () => import('../features/reorder/Reorder'),
    '/valuation': () => import('../features/valuation/Valuation'),
    '/purchasing': () => import('../features/purchasing/Purchasing'),
    '/suppliers': () => import('../features/suppliers/Suppliers'),
    '/settings': () => import('../features/settings/Settings'),
  });

  const erpUser = useUserStore((s) => s.user);
  const setAuthUser = useUserStore((s) => s.setUser);
  useEffect(() => {
    if (erpUser?.name || erpUser?.email) return;
    userClient
      .getMe({})
      .then((res) => {
        if (res?.id) {
          setAuthUser({ id: String(res.id), name: res.fullName ?? '', email: res.email ?? '' });
        }
      })
      .catch(() => undefined);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [erpUser]);
  const userId = Number(erpUser?.id) || 0;
  const { isOpen: caretOpen } = useCaretStore();
  const pageTitle = t(pageTitleKey(pathname));
  // Caret describes the current route to the assistant, so it wants the
  // reader's language rather than the key or the English behind it.
  const translatedRouteLabels = React.useMemo(
    () => Object.fromEntries(Object.entries(routeLabels).map(([p, l]) => [p, t(l.title)])),
    [t]
  );
  const caretPageContext = React.useMemo(
    () =>
      buildCaretRouteContext({
        appId: 'erp',
        appName: 'ERP',
        pathname,
        search,
        page: pageTitle,
        routeLabels: translatedRouteLabels,
      }),
    [pathname, search, pageTitle, translatedRouteLabels]
  );

  useEffect(() => {
    applyErpTheme(theme);
  }, [theme]);

  useEffect(() => {
    setSidebarOpen(false);
  }, [pathname]);

  const userEmail = erpUser?.email || erpUser?.emailAddress || null;
  const userName = erpUser?.name || erpUser?.firstName || userEmail || 'User';

  const handleSetTheme = (theme: ErpTheme) => {
    setTheme(theme);
  };

  const handleLogout = (): void => {
    signOut('erp');
  };

  // Global keyboard shortcuts
  useEffect(() => {
    const handler = (e: KeyboardEvent) => {
      const target = e.target as HTMLElement;
      const isTyping =
        target.tagName === 'INPUT' || target.tagName === 'TEXTAREA' || target.isContentEditable;

      // cmd+k or ctrl+k
      if ((e.metaKey || e.ctrlKey) && e.key === 'k') {
        e.preventDefault();
        setPaletteOpen((v) => !v);
        return;
      }

      if (isTyping) return;

      // Single-key shortcuts
      if (e.key === '/' && !e.metaKey && !e.ctrlKey) {
        e.preventDefault();
        const searchInput = document.querySelector('input[type="text"]') as HTMLInputElement;
        searchInput?.focus();
      } else if (e.key === '?' && e.shiftKey) {
        e.preventDefault();
        setShortcutsOpen(true);
      } else if (e.key === 'n') {
        e.preventDefault();
        setQuickAddOpen(true);
      } else if (e.key === 'Escape') {
        setPaletteOpen(false);
        setShortcutsOpen(false);
        setQuickAddOpen(false);
      } else if (e.shiftKey) {
        if (e.key === 'I' || e.key === 'i') {
          e.preventDefault();
          navigate({ to: '/inventory' });
        } else if (e.key === 'P' || e.key === 'p') {
          e.preventDefault();
          navigate({ to: '/purchasing' });
        } else if (e.key === 'S' || e.key === 's') {
          e.preventDefault();
          navigate({ to: '/suppliers' });
        } else if (e.key === 'D' || e.key === 'd') {
          e.preventDefault();
          navigate({ to: '/' });
        }
      }
    };
    window.addEventListener('keydown', handler);
    return () => window.removeEventListener('keydown', handler);
  }, [navigate]);

  return (
    <>
      <div className="flex h-screen overflow-hidden" style={{ background: 'var(--hub-bg)' }}>
        {/* Mobile top bar */}
        <div
          className="fixed top-0 z-40 flex md:hidden items-center justify-between h-11 px-4 shrink-0"
          style={{
            background: 'var(--hub-surface)',
            borderBottom: '1px solid var(--hub-border)',
            insetInline: 0,
          }}
        >
          <button
            type="button"
            onClick={() => setSidebarOpen(true)}
            aria-label="Open navigation menu"
            aria-expanded={sidebarOpen}
            aria-controls="erp-sidebar"
            className="h-11 w-11 rounded-lg flex items-center justify-center"
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
            aria-hidden="true"
          />
        )}

        <aside
          id="erp-sidebar"
          aria-label="Primary navigation"
          className={cn(
            sidebarOpen ? 'fixed inset-y-0 z-50 flex' : 'hidden',
            'md:flex w-56 flex-col shrink-0 overflow-hidden'
          )}
          style={{
            background: 'var(--hub-surface)',
            borderInlineEnd: '1px solid var(--hub-border)',
            insetInlineStart: sidebarOpen ? 0 : undefined,
          }}
        >
          <div
            className="flex items-center justify-between gap-2.5 px-4 py-4"
            style={{ borderBottom: '1px solid var(--hub-border)' }}
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
              aria-label="Close navigation menu"
              className="h-11 w-11 rounded-lg flex md:hidden items-center justify-center"
              style={{ color: 'var(--hub-text-muted)' }}
            >
              <X size={16} />
            </button>
          </div>

          <nav ref={sidebarRef} className="flex-1 overflow-y-auto py-2 px-2 space-y-3">
            {navGroups.map((group) => (
              <div key={group.labelKey}>
                <p
                  className="px-3 mb-1 text-[10px] font-semibold uppercase tracking-widest"
                  style={{ color: 'var(--hub-text-dim)' }}
                >
                  {t(group.labelKey)}
                </p>
                <div className="space-y-0.5">
                  {group.items.map((item) => (
                    <SidebarNavItem key={item.path} {...item} />
                  ))}
                </div>
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
                aria-label={showUserMenu ? 'Close user menu' : 'Open user menu'}
                aria-expanded={showUserMenu}
                aria-haspopup="menu"
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
                  style={{ background: 'color-mix(in srgb, var(--erp-accent) 15%, transparent)' }}
                />
                <div className="flex flex-col items-start min-w-0 flex-1">
                  <span
                    className="text-xs font-medium truncate w-full"
                    style={{ color: 'var(--hub-text)', textAlign: 'start' }}
                  >
                    {userName}
                  </span>
                  {userEmail && (
                    <span
                      className="text-[10px] truncate w-full"
                      style={{ color: 'var(--hub-text-muted)', textAlign: 'start' }}
                    >
                      {userEmail}
                    </span>
                  )}
                </div>
                <ChevronRight
                  size={12}
                  className="erp-directional-icon shrink-0"
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
                    className="absolute bottom-full mb-1 rounded-xl p-2 z-50"
                    style={{
                      background: 'var(--hub-card)',
                      border: '1px solid var(--hub-border)',
                      boxShadow: 'var(--hub-shadow)',
                      insetInline: 0,
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
                          { value: 'light' as ErpTheme, icon: Sun, label: 'Light' },
                          { value: 'dark' as ErpTheme, icon: Moon, label: 'Dark' },
                          { value: 'system' as ErpTheme, icon: Monitor, label: 'System' },
                        ] as const
                      ).map(({ value, icon: Icon, label }) => (
                        <button
                          type="button"
                          key={value}
                          onClick={() => handleSetTheme(value)}
                          className="flex-1 flex flex-col items-center gap-1 py-2 rounded-lg text-[10px] font-medium transition-colors"
                          style={{
                            background:
                              theme === value
                                ? 'color-mix(in srgb, var(--erp-accent) 12%, transparent)'
                                : 'var(--hub-hover)',
                            color: theme === value ? 'var(--erp-accent)' : 'var(--hub-text-muted)',
                            border:
                              theme === value
                                ? '1px solid color-mix(in srgb, var(--erp-accent) 30%, transparent)'
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
                      style={{ color: 'var(--erp-danger-strong)' }}
                      onMouseEnter={(e) => {
                        (e.currentTarget as HTMLElement).style.background =
                          'color-mix(in srgb, var(--erp-danger-strong) 8%, transparent)';
                      }}
                      onMouseLeave={(e) => {
                        (e.currentTarget as HTMLElement).style.background = 'transparent';
                      }}
                    >
                      <LogOut size={14} />
                      Sign out
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
                  type="button"
                  onClick={() => setPaletteOpen(true)}
                  aria-label="Open search"
                  className="flex items-center gap-2 h-11 px-3 rounded-lg text-xs"
                  style={{
                    background: 'var(--hub-hover)',
                    color: 'var(--hub-text-muted)',
                    border: '1px solid var(--hub-border)',
                  }}
                >
                  <SearchIcon size={11} />
                  <span>Search...</span>
                  <kbd
                    className="text-[9px] font-mono px-1 py-0.5 rounded"
                    style={{ background: 'var(--hub-active)', color: 'var(--hub-text-dim)' }}
                  >
                    cmd+k
                  </kbd>
                </button>
                <button
                  type="button"
                  onClick={() => handleSetTheme(theme === 'dark' ? 'light' : 'dark')}
                  aria-label={`Switch to ${theme === 'dark' ? 'light' : 'dark'} theme`}
                  className="h-11 w-11 rounded-lg flex items-center justify-center"
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
                <button
                  type="button"
                  onClick={() => setShortcutsOpen(true)}
                  aria-label="Open keyboard shortcuts"
                  className="h-11 w-11 rounded-lg flex items-center justify-center"
                  style={{ color: 'var(--hub-text-muted)' }}
                  onMouseEnter={(e) => {
                    (e.currentTarget as HTMLElement).style.background = 'var(--hub-hover)';
                  }}
                  onMouseLeave={(e) => {
                    (e.currentTarget as HTMLElement).style.background = 'transparent';
                  }}
                >
                  <Keyboard size={14} />
                </button>
                <Link
                  to="/alerts"
                  aria-label="Open notifications"
                  className="h-11 w-11 rounded-lg flex items-center justify-center transition-colors duration-150"
                  style={{ color: 'var(--hub-text-muted)' }}
                  onMouseEnter={(e) => {
                    (e.currentTarget as HTMLElement).style.background = 'var(--hub-hover)';
                  }}
                  onMouseLeave={(e) => {
                    (e.currentTarget as HTMLElement).style.background = 'transparent';
                  }}
                >
                  <Bell size={14} />
                </Link>
                <CaretButton />
              </div>
            </header>

            <LicenseBanner client={licenseClient} />
            <TrialBanner />
            <AnimatePresence mode="wait">
              <motion.main
                key={pathname}
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
              appId="erp"
              userId={userId}
              apiBaseUrl={import.meta.env.VITE_API_URL ?? 'https://api.voxire.com/'}
              systemContext="User is in the ERP app (inventory, products, suppliers, purchasing)"
              pageContext={caretPageContext}
            />
          </div>
        </div>

        <CommandPalette open={paletteOpen} onClose={() => setPaletteOpen(false)} />
        <KeyboardShortcutsModal open={shortcutsOpen} onClose={() => setShortcutsOpen(false)} />
        <QuickAddProductModal open={quickAddOpen} onClose={() => setQuickAddOpen(false)} />

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
