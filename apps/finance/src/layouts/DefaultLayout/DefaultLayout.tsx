import React, { useState } from 'react';
import { useLocation, Link, Outlet } from 'react-router-dom';
import { motion, AnimatePresence } from 'framer-motion';
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
  ChevronRight,
} from 'lucide-react';

const ACCENT = '#10b981';

interface NavItem {
  label: string;
  path: string;
  icon: React.ComponentType<{ size?: number; className?: string; style?: React.CSSProperties }>;
}

const navItems: NavItem[] = [
  { label: 'Dashboard', path: '/',          icon: LayoutDashboard },
  { label: 'Invoices',  path: '/invoices',  icon: FileText },
  { label: 'Expenses',  path: '/expenses',  icon: Receipt },
  { label: 'P&L',       path: '/pl',        icon: TrendingUp },
  { label: 'Tax',       path: '/tax',       icon: Calculator },
];

const routeLabels: Record<string, string> = {
  '/':          'Dashboard',
  '/invoices':  'Invoices',
  '/expenses':  'Expenses',
  '/pl':        'Profit & Loss',
  '/tax':       'Tax',
};

function getPageTitle(pathname: string): string {
  return routeLabels[pathname] ?? 'Finance';
}

type Theme = 'light' | 'dark' | 'system';

function applyTheme(t: Theme) {
  const resolved = t === 'system'
    ? (window.matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'light')
    : t;
  document.documentElement.setAttribute('data-theme', resolved);
  localStorage.setItem('finance-theme', t);
}

const SidebarNavItem: React.FC<NavItem> = ({ label, path, icon: Icon }) => {
  const location = useLocation();
  const isActive = path === '/' ? location.pathname === '/' : location.pathname.startsWith(path);

  return (
    <Link
      to={path}
      style={{ color: isActive ? 'var(--hub-text)' : 'var(--hub-text-muted)', textDecoration: 'none' }}
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
      <Icon size={16} style={{ position: 'relative', flexShrink: 0 }} />
      <span style={{ position: 'relative' }}>{label}</span>
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
  const pageTitle = getPageTitle(location.pathname);
  const [theme, setThemeState] = useState<Theme>(
    () => (localStorage.getItem('finance-theme') as Theme) || 'dark'
  );

  const handleTheme = (t: Theme) => {
    setThemeState(t);
    applyTheme(t);
  };

  return (
    <div className="flex h-screen overflow-hidden" style={{ background: 'var(--hub-bg)' }}>
      {/* Sidebar */}
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
        {/* Logo / Header */}
        <div
          style={{
            display: 'flex',
            alignItems: 'center',
            gap: 10,
            padding: '16px',
            borderBottom: '1px solid var(--hub-border)',
          }}
        >
          <span
            style={{
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'center',
              height: 28,
              width: 28,
              borderRadius: 8,
              background: 'rgba(16,185,129,0.15)',
              color: ACCENT,
              flexShrink: 0,
            }}
          >
            <TrendingUp size={14} />
          </span>
          <div style={{ display: 'flex', flexDirection: 'column', minWidth: 0 }}>
            <span style={{ fontSize: 13, fontWeight: 600, color: 'var(--hub-text)', lineHeight: 1.2 }}>
              Finance
            </span>
            <span style={{ fontSize: 10, color: 'var(--hub-text-dim)', letterSpacing: '0.04em' }}>
              rtylr
            </span>
          </div>
        </div>

        {/* Hub back link */}
        <div style={{ padding: '10px 8px 4px' }}>
          <a
            href="http://localhost:5176"
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
              transition: 'background 0.15s',
            }}
          >
            <ArrowLeft size={12} />
            <span>Hub</span>
          </a>
        </div>

        {/* Navigation */}
        <nav style={{ flex: 1, overflowY: 'auto', padding: '8px', display: 'flex', flexDirection: 'column', gap: 2 }}>
          <p style={{ padding: '4px 12px 6px', fontSize: 10, fontWeight: 600, textTransform: 'uppercase', letterSpacing: '0.1em', color: 'var(--hub-text-dim)', margin: 0 }}>
            Finance
          </p>
          {navItems.map((item) => (
            <SidebarNavItem key={item.path} {...item} />
          ))}
        </nav>

        {/* Footer — theme selector */}
        <div style={{ padding: 8, borderTop: '1px solid var(--hub-border)' }}>
          <p style={{ fontSize: 10, fontWeight: 600, textTransform: 'uppercase', letterSpacing: '0.08em', color: 'var(--hub-text-dim)', margin: '0 0 6px 4px' }}>
            Appearance
          </p>
          <div style={{ display: 'flex', gap: 4 }}>
            {([
              { value: 'light' as Theme, icon: Sun,     label: 'Light'  },
              { value: 'dark'  as Theme, icon: Moon,    label: 'Dark'   },
              { value: 'system'as Theme, icon: Monitor, label: 'System' },
            ]).map(({ value, icon: Icon, label }) => (
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
                  transition: 'all 0.15s',
                }}
              >
                <Icon size={13} />
                {label}
              </button>
            ))}
          </div>
        </div>
      </aside>

      {/* Main content */}
      <div style={{ display: 'flex', flexDirection: 'column', flex: 1, overflow: 'hidden' }}>
        {/* Topbar */}
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
                transition: 'background 0.15s',
              }}
            >
              <Bell size={14} />
            </button>
          </div>
        </header>

        {/* Page content */}
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
    </div>
  );
};

export default DefaultLayout;
