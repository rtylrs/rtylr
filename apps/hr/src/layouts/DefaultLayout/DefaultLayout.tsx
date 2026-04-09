import React, { useState } from 'react';
import { useLocation, Link } from 'react-router-dom';
import { motion, AnimatePresence } from 'framer-motion';
import { APP_URLS } from '@voxire/config';
import {
  LayoutDashboard,
  Users,
  Clock,
  DollarSign,
  Building2,
  Bell,
  ArrowLeft,
  Sun,
  Moon,
  Monitor,
  LogOut,
  ChevronUp,
} from 'lucide-react';

/* ── Types ── */
interface NavItem {
  label: string;
  path: string;
  icon: React.ComponentType<{ size?: number; className?: string; style?: React.CSSProperties }>;
}

/* ── Nav config ── */
const ACCENT = '#8b5cf6';

const navItems: NavItem[] = [
  { label: 'Dashboard',   path: '/',            icon: LayoutDashboard },
  { label: 'Staff',       path: '/staff',       icon: Users           },
  { label: 'Shifts',      path: '/shifts',      icon: Clock           },
  { label: 'Payroll',     path: '/payroll',     icon: DollarSign      },
  { label: 'Departments', path: '/departments', icon: Building2       },
];

const routeLabels: Record<string, string> = {
  '/':            'Dashboard',
  '/staff':       'Staff',
  '/shifts':      'Shifts',
  '/payroll':     'Payroll',
  '/departments': 'Departments',
};

function getPageTitle(pathname: string): string {
  return routeLabels[pathname] ?? 'HR';
}

/* ── Theme helpers ── */
type Theme = 'light' | 'dark' | 'system';

function useTheme(): [Theme, (t: Theme) => void] {
  const [theme, setThemeState] = useState<Theme>(() => {
    return (localStorage.getItem('hr-theme') as Theme) || 'dark';
  });

  const setTheme = (t: Theme) => {
    localStorage.setItem('hr-theme', t);
    setThemeState(t);
    const systemDark = window.matchMedia('(prefers-color-scheme: dark)').matches;
    const resolved = t === 'system' ? (systemDark ? 'dark' : 'light') : t;
    document.documentElement.setAttribute('data-theme', resolved);
  };

  return [theme, setTheme];
}

/* ── Sidebar nav item ── */
const SidebarNavItem: React.FC<NavItem> = ({ label, path, icon: Icon }) => {
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
        color: isActive ? 'var(--hr-text)' : 'var(--hr-text-muted)',
        transition: 'color 0.15s',
      }}
      onMouseEnter={(e) => {
        if (!isActive) (e.currentTarget as HTMLAnchorElement).style.background = 'var(--hr-hover)';
      }}
      onMouseLeave={(e) => {
        if (!isActive) (e.currentTarget as HTMLAnchorElement).style.background = 'transparent';
      }}
    >
      {isActive && (
        <motion.span
          layoutId="sidebar-active"
          style={{
            position: 'absolute',
            inset: 0,
            borderRadius: '8px',
            background: 'var(--hr-active)',
          }}
          transition={{ type: 'spring', stiffness: 400, damping: 35 }}
        />
      )}
      <Icon size={15} style={{ position: 'relative', flexShrink: 0, color: isActive ? ACCENT : undefined }} />
      <span style={{ position: 'relative' }}>{label}</span>
      {isActive && (
        <span
          style={{
            position: 'relative',
            marginLeft: 'auto',
            height: 5,
            width: 5,
            borderRadius: '50%',
            background: ACCENT,
            flexShrink: 0,
          }}
        />
      )}
    </Link>
  );
};

/* ── User dropdown ── */
function UserMenu({ theme, setTheme }: { theme: Theme; setTheme: (t: Theme) => void }) {
  const [open, setOpen] = useState(false);

  return (
    <div style={{ position: 'relative' }}>
      <button
        onClick={() => setOpen((o) => !o)}
        style={{
          display: 'flex',
          alignItems: 'center',
          gap: '10px',
          width: '100%',
          padding: '8px 12px',
          borderRadius: '8px',
          background: 'transparent',
          border: 'none',
          cursor: 'pointer',
          transition: 'background 0.15s',
        }}
        onMouseEnter={(e) => ((e.currentTarget as HTMLButtonElement).style.background = 'var(--hr-hover)')}
        onMouseLeave={(e) => ((e.currentTarget as HTMLButtonElement).style.background = 'transparent')}
      >
        {/* Avatar */}
        <span
          style={{
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
            height: 28,
            width: 28,
            borderRadius: '8px',
            background: 'rgba(139,92,246,0.15)',
            color: ACCENT,
            fontSize: '11px',
            fontWeight: 700,
            flexShrink: 0,
          }}
        >
          HR
        </span>
        <div style={{ display: 'flex', flexDirection: 'column', alignItems: 'flex-start', flex: 1, minWidth: 0 }}>
          <span style={{ fontSize: '12px', fontWeight: 600, color: 'var(--hr-text)', textAlign: 'left' }}>HR Admin</span>
          <span style={{ fontSize: '10px', color: 'var(--hr-text-muted)', textAlign: 'left' }}>admin@rtylr.com</span>
        </div>
        <ChevronUp size={12} style={{ color: 'var(--hr-text-dim)', flexShrink: 0, transform: open ? 'rotate(180deg)' : 'rotate(0)', transition: 'transform 0.2s' }} />
      </button>

      <AnimatePresence>
        {open && (
          <motion.div
            initial={{ opacity: 0, y: 6 }}
            animate={{ opacity: 1, y: 0 }}
            exit={{ opacity: 0, y: 6 }}
            transition={{ duration: 0.15 }}
            style={{
              position: 'absolute',
              bottom: '100%',
              left: 0,
              right: 0,
              marginBottom: '4px',
              background: 'var(--hr-card)',
              border: '1px solid var(--hr-border)',
              borderRadius: '10px',
              padding: '6px',
              zIndex: 50,
              boxShadow: 'var(--hr-shadow)',
            }}
          >
            {/* Theme picker */}
            <p style={{ fontSize: '10px', fontWeight: 600, textTransform: 'uppercase', letterSpacing: '0.08em', color: 'var(--hr-text-dim)', padding: '4px 8px 6px' }}>
              Appearance
            </p>
            <div style={{ display: 'flex', gap: '4px', marginBottom: '6px', padding: '0 4px' }}>
              {([
                { value: 'light' as Theme,  icon: Sun,     label: 'Light'  },
                { value: 'dark' as Theme,   icon: Moon,    label: 'Dark'   },
                { value: 'system' as Theme, icon: Monitor, label: 'System' },
              ]).map(({ value, icon: Icon, label }) => (
                <button
                  key={value}
                  onClick={() => setTheme(value)}
                  style={{
                    flex: 1,
                    display: 'flex',
                    flexDirection: 'column',
                    alignItems: 'center',
                    gap: '4px',
                    padding: '8px 4px',
                    borderRadius: '8px',
                    fontSize: '10px',
                    fontWeight: 500,
                    border: theme === value ? `1px solid ${ACCENT}40` : '1px solid transparent',
                    background: theme === value ? 'rgba(139,92,246,0.12)' : 'var(--hr-hover)',
                    color: theme === value ? ACCENT : 'var(--hr-text-muted)',
                    cursor: 'pointer',
                  }}
                >
                  <Icon size={13} />
                  {label}
                </button>
              ))}
            </div>
            <hr style={{ border: 'none', borderTop: '1px solid var(--hr-border)', margin: '4px 0' }} />
            <button
              onClick={() => { window.location.href = APP_URLS.auth; }}
              style={{
                display: 'flex',
                alignItems: 'center',
                gap: '8px',
                width: '100%',
                padding: '8px 8px',
                borderRadius: '7px',
                background: 'transparent',
                border: 'none',
                color: '#ef4444',
                fontSize: '13px',
                fontWeight: 500,
                cursor: 'pointer',
                textAlign: 'left',
              }}
              onMouseEnter={(e) => ((e.currentTarget as HTMLButtonElement).style.background = 'rgba(239,68,68,0.08)')}
              onMouseLeave={(e) => ((e.currentTarget as HTMLButtonElement).style.background = 'transparent')}
            >
              <LogOut size={14} />
              Sign out
            </button>
          </motion.div>
        )}
      </AnimatePresence>
    </div>
  );
}

/* ── DefaultLayout ── */
interface DefaultLayoutProps {
  children?: React.ReactNode;
}

const DefaultLayout: React.FC<DefaultLayoutProps> = ({ children }) => {
  const location = useLocation();
  const [theme, setTheme] = useTheme();
  const pageTitle = getPageTitle(location.pathname);

  return (
    <div style={{ display: 'flex', height: '100vh', overflow: 'hidden', background: 'var(--hr-bg)' }}>
      {/* Sidebar */}
      <aside
        style={{
          width: '220px',
          display: 'flex',
          flexDirection: 'column',
          flexShrink: 0,
          overflow: 'hidden',
          background: 'var(--hr-surface)',
          borderRight: '1px solid var(--hr-border)',
        }}
      >
        {/* Logo + Hub back link */}
        <div
          style={{
            display: 'flex',
            flexDirection: 'column',
            gap: '4px',
            padding: '16px 12px 12px',
            borderBottom: '1px solid var(--hr-border)',
          }}
        >
          <div style={{ display: 'flex', alignItems: 'center', gap: '8px' }}>
            <span
              style={{
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'center',
                height: 28,
                width: 28,
                borderRadius: '8px',
                background: 'rgba(139,92,246,0.15)',
                color: ACCENT,
                fontSize: '12px',
                fontWeight: 800,
                flexShrink: 0,
              }}
            >
              HR
            </span>
            <span style={{ fontSize: '14px', fontWeight: 700, color: 'var(--hr-text)' }}>
              HR Portal
            </span>
          </div>
          <a
            href="http://localhost:5175"
            style={{
              display: 'inline-flex',
              alignItems: 'center',
              gap: '4px',
              fontSize: '11px',
              color: 'var(--hr-text-dim)',
              textDecoration: 'none',
              marginTop: '2px',
              transition: 'color 0.15s',
            }}
            onMouseEnter={(e) => ((e.currentTarget as HTMLAnchorElement).style.color = 'var(--hr-text-muted)')}
            onMouseLeave={(e) => ((e.currentTarget as HTMLAnchorElement).style.color = 'var(--hr-text-dim)')}
          >
            <ArrowLeft size={11} />
            Hub
          </a>
        </div>

        {/* Navigation */}
        <nav
          style={{
            flex: 1,
            overflowY: 'auto',
            padding: '8px 8px',
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
              color: 'var(--hr-text-dim)',
            }}
          >
            Workspace
          </p>
          {navItems.map((item) => (
            <SidebarNavItem key={item.path} {...item} />
          ))}
        </nav>

        {/* Footer */}
        <div style={{ padding: '8px', borderTop: '1px solid var(--hr-border)' }}>
          <UserMenu theme={theme} setTheme={setTheme} />
        </div>
      </aside>

      {/* Main */}
      <div style={{ display: 'flex', flexDirection: 'column', flex: 1, overflow: 'hidden' }}>
        {/* Topbar */}
        <header
          style={{
            height: '44px',
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'space-between',
            padding: '0 24px',
            flexShrink: 0,
            background: 'var(--hr-surface)',
            borderBottom: '1px solid var(--hr-border)',
          }}
        >
          <h1 style={{ fontSize: '13px', fontWeight: 600, color: 'var(--hr-text)', margin: 0 }}>
            {pageTitle}
          </h1>
          <div style={{ display: 'flex', alignItems: 'center', gap: '4px' }}>
            <button
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
                color: 'var(--hr-text-muted)',
                transition: 'background 0.15s',
              }}
              onMouseEnter={(e) => ((e.currentTarget as HTMLButtonElement).style.background = 'var(--hr-hover)')}
              onMouseLeave={(e) => ((e.currentTarget as HTMLButtonElement).style.background = 'transparent')}
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
            style={{ flex: 1, overflowY: 'auto', padding: '24px' }}
          >
            {children}
          </motion.main>
        </AnimatePresence>
      </div>
    </div>
  );
};

export default DefaultLayout;
