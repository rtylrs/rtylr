import React, { useEffect, useState } from 'react';
import { useLocation, useNavigate, Link, Outlet } from 'react-router-dom';
import { motion, AnimatePresence } from 'motion/react';
import { APP_URLS } from '@voxire/config';
import {
  LayoutDashboard,
  FileText,
  Receipt,
  TrendingUp,
  Calculator,
  Bell,
  Sun,
  Moon,
  Monitor,
  ArrowLeft,
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
} from 'lucide-react';
import GlobalSearch from '../shared/components/GlobalSearch';

const ACCENT = '#10b981';

interface NavItem {
  label: string;
  path: string;
  icon: React.ComponentType<{ size?: number; className?: string; style?: React.CSSProperties }>;
}

interface NavGroup {
  title: string;
  items: NavItem[];
}

const navGroups: NavGroup[] = [
  {
    title: 'Overview',
    items: [{ label: 'Dashboard', path: '/', icon: LayoutDashboard }],
  },
  {
    title: 'Sales',
    items: [
      { label: 'Invoices', path: '/invoices', icon: FileText },
      { label: 'Recurring', path: '/invoices/recurring', icon: Repeat },
      { label: 'Quotes', path: '/quotes', icon: ClipboardList },
    ],
  },
  {
    title: 'Spend',
    items: [
      { label: 'Expenses', path: '/expenses', icon: Receipt },
      { label: 'Categories', path: '/categories', icon: Tag },
      { label: 'Bills', path: '/bills', icon: Wallet },
    ],
  },
  {
    title: 'Banking',
    items: [
      { label: 'Bank Feeds', path: '/bank', icon: Building2 },
      { label: 'Reconcile', path: '/reconcile', icon: Shuffle },
    ],
  },
  {
    title: 'Reports',
    items: [
      { label: 'Profit & Loss', path: '/pl', icon: TrendingUp },
      { label: 'Balance Sheet', path: '/reports/balance-sheet', icon: BarChart2 },
      { label: 'Cash Flow', path: '/reports/cash-flow', icon: LineChart },
      { label: 'Tax Summary', path: '/reports/tax', icon: FileCheck },
    ],
  },
  {
    title: 'Accounting',
    items: [
      { label: 'Journal', path: '/journal', icon: BookOpen },
      { label: 'Accounts', path: '/accounts', icon: FileText },
      { label: 'Tax Config', path: '/tax', icon: Calculator },
    ],
  },
  {
    title: 'Admin',
    items: [{ label: 'Settings', path: '/settings', icon: SettingsIcon }],
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

const SidebarNavItem: React.FC<NavItem> = ({ label, path, icon: Icon }) => {
  const location = useLocation();
  const isActive =
    path === '/'
      ? location.pathname === '/'
      : location.pathname === path || location.pathname.startsWith(path + '/');

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
      <span style={{ position: 'relative', fontSize: 12.5 }}>{label}</span>
      {isActive && (
        <span
          style={{
            position: 'relative',
            marginLeft: 'auto',
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
  const location = useLocation();
  const navigate = useNavigate();
  const pageTitle = getPageTitle(location.pathname);
  const [theme, setThemeState] = useState<Theme>(
    () => (localStorage.getItem('finance-theme') as Theme) || 'dark'
  );
  const [searchOpen, setSearchOpen] = useState(false);

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
        navigate('/invoices/new');
      } else if (e.key === 'e') {
        e.preventDefault();
        navigate('/expenses/new');
      } else if (e.key === 'Escape') {
        setSearchOpen(false);
      }
    };
    window.addEventListener('keydown', handler);
    return () => window.removeEventListener('keydown', handler);
  }, [navigate]);

  return (
    <div className="flex h-screen overflow-hidden" style={{ background: 'var(--hub-bg)' }}>
      <aside
        style={{
          width: 224,
          display: 'flex',
          flexDirection: 'column',
          flexShrink: 0,
          overflow: 'hidden',
          background: 'var(--hub-surface)',
          borderRight: '1px solid var(--hub-border)',
        }}
      >
        <div
          style={{
            display: 'flex',
            alignItems: 'center',
            gap: 10,
            padding: 16,
            borderBottom: '1px solid var(--hub-border)',
          }}
        >
          <img src="/logo.png" alt="rtylr" className="h-6 w-auto" />
        </div>

        <div style={{ padding: '10px 8px 4px' }}>
          <a
            href={APP_URLS.hub}
            style={{
              display: 'flex',
              alignItems: 'center',
              gap: 6,
              padding: '6px 12px',
              borderRadius: 8,
              fontSize: 12,
              color: 'var(--hub-text-muted)',
              textDecoration: 'none',
              background: 'var(--hub-hover)',
              border: '1px solid var(--hub-border)',
            }}
          >
            <ArrowLeft size={12} />
            <span>Hub</span>
          </a>
        </div>

        <div style={{ padding: '6px 8px' }}>
          <button
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
            <span>Search</span>
            <span style={{ marginLeft: 'auto', fontSize: 10, color: 'var(--hub-text-dim)' }}>
              Cmd K
            </span>
          </button>
        </div>

        <nav
          style={{
            flex: 1,
            overflowY: 'auto',
            padding: 8,
            display: 'flex',
            flexDirection: 'column',
            gap: 8,
          }}
        >
          {navGroups.map((group) => (
            <div key={group.title} style={{ display: 'flex', flexDirection: 'column', gap: 2 }}>
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
                {group.title}
              </p>
              {group.items.map((item) => (
                <SidebarNavItem key={item.path} {...item} />
              ))}
            </div>
          ))}
        </nav>

        <div style={{ padding: 8, borderTop: '1px solid var(--hub-border)' }}>
          <p
            style={{
              fontSize: 10,
              fontWeight: 600,
              textTransform: 'uppercase',
              letterSpacing: '0.08em',
              color: 'var(--hub-text-dim)',
              margin: '0 0 6px 4px',
            }}
          >
            Appearance
          </p>
          <div style={{ display: 'flex', gap: 4 }}>
            {[
              { value: 'light' as Theme, icon: Sun, label: 'Light' },
              { value: 'dark' as Theme, icon: Moon, label: 'Dark' },
              { value: 'system' as Theme, icon: Monitor, label: 'System' },
            ].map(({ value, icon: Icon, label }) => (
              <button
                key={value}
                onClick={() => handleTheme(value)}
                style={{
                  flex: 1,
                  display: 'flex',
                  flexDirection: 'column',
                  alignItems: 'center',
                  gap: 4,
                  padding: '6px 4px',
                  borderRadius: 8,
                  fontSize: 10,
                  fontWeight: 500,
                  cursor: 'pointer',
                  border: theme === value ? `1px solid ${ACCENT}40` : '1px solid transparent',
                  background: theme === value ? 'rgba(16,185,129,0.12)' : 'var(--hub-hover)',
                  color: theme === value ? ACCENT : 'var(--hub-text-muted)',
                }}
              >
                <Icon size={13} />
                {label}
              </button>
            ))}
          </div>
        </div>
      </aside>

      <div style={{ display: 'flex', flexDirection: 'column', flex: 1, overflow: 'hidden' }}>
        <header
          style={{
            height: 44,
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'space-between',
            padding: '0 24px',
            flexShrink: 0,
            background: 'var(--hub-surface)',
            borderBottom: '1px solid var(--hub-border)',
          }}
        >
          <h1 style={{ margin: 0, fontSize: 13, fontWeight: 600, color: 'var(--hub-text)' }}>
            {pageTitle}
          </h1>
          <div style={{ display: 'flex', alignItems: 'center', gap: 4 }}>
            <button
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
          </div>
        </header>

        <AnimatePresence mode="wait">
          <motion.main
            key={location.pathname}
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

      <GlobalSearch open={searchOpen} onClose={() => setSearchOpen(false)} />
    </div>
  );
};

export default DefaultLayout;
