import React, { useState } from 'react';
import { useLocation, useNavigate, Link, Outlet } from 'react-router-dom';
import { motion, AnimatePresence } from 'motion/react';
import { APP_URLS } from '@voxire/config';
import {
  LayoutDashboard,
  Users,
  Gift,
  Megaphone,
  PieChart,
  ArrowLeft,
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
} from 'lucide-react';
import { useShortcuts } from '../shared/components/ShortcutsProvider';

const ACCENT = '#3b82f6';

interface NavItem {
  label: string;
  path: string;
  icon: React.ComponentType<{ size?: number; className?: string }>;
}

const navItems: NavItem[] = [
  { label: 'Dashboard',   path: '/',            icon: LayoutDashboard },
  { label: 'Customers',   path: '/customers',   icon: Users },
  { label: 'Loyalty',     path: '/loyalty',     icon: Gift },
  { label: 'Campaigns',   path: '/campaigns',   icon: Megaphone },
  { label: 'Segments',    path: '/segments',    icon: PieChart },
  { label: 'Automations', path: '/automations', icon: Zap },
  { label: 'Templates',   path: '/templates',   icon: FileText },
  { label: 'Feedback',    path: '/feedback',    icon: MessageSquare },
  { label: 'Settings',    path: '/settings',    icon: SettingsIcon },
];

const routeLabels: Record<string, string> = {
  '/':            'Dashboard',
  '/customers':   'Customers',
  '/customers/new': 'New customer',
  '/customers/import': 'Import customers',
  '/loyalty':     'Loyalty',
  '/loyalty/members': 'Loyalty members',
  '/loyalty/rewards': 'Rewards catalog',
  '/loyalty/tiers': 'Loyalty tiers',
  '/campaigns':   'Campaigns',
  '/campaigns/email': 'New email campaign',
  '/campaigns/sms': 'New SMS campaign',
  '/segments':    'Segments',
  '/segments/new': 'New segment',
  '/automations': 'Automations',
  '/automations/new': 'New automation',
  '/templates':   'Templates',
  '/feedback':    'Feedback & NPS',
  '/settings':    'Settings',
};

function getPageTitle(pathname: string): string {
  if (routeLabels[pathname]) return routeLabels[pathname];
  const match = Object.keys(routeLabels).find(
    (key) => pathname.startsWith(key) && key !== '/'
  );
  return match ? routeLabels[match] : 'CRM';
}

type Theme = 'light' | 'dark' | 'system';

function getStoredTheme(): Theme {
  return (localStorage.getItem('crm-theme') as Theme) || 'dark';
}

function applyTheme(theme: Theme) {
  const resolved = theme === 'system'
    ? (window.matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'light')
    : theme;
  document.documentElement.setAttribute('data-theme', resolved);
  localStorage.setItem('crm-theme', theme);
}

const SidebarNavItem: React.FC<NavItem> = ({ label, path, icon: Icon }) => {
  const location = useLocation();
  const isActive =
    path === '/' ? location.pathname === '/' : location.pathname.startsWith(path);

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
      <span style={{ position: 'relative', flex: 1 }}>{label}</span>
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
  const location = useLocation();
  const navigate = useNavigate();
  const { openPalette } = useShortcuts();
  const [theme, setThemeState] = useState<Theme>(getStoredTheme);
  const [userMenuOpen, setUserMenuOpen] = useState(false);

  const pageTitle = getPageTitle(location.pathname);

  const handleTheme = (t: Theme) => {
    setThemeState(t);
    applyTheme(t);
  };

  const handleLogout = () => {
    document.cookie = 'token=; Max-Age=0; path=/';
    window.location.href = `${APP_URLS.auth}?redirect=crm`;
  };

  return (
    <div
      style={{
        display: 'flex',
        height: '100vh',
        overflow: 'hidden',
        background: 'var(--hub-bg)',
      }}
    >
      {/* Sidebar */}
      <aside
        style={{
          width: '220px',
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
            gap: '10px',
            padding: '16px',
            borderBottom: '1px solid var(--hub-border)',
          }}
        >
          <img src="/logo.png" alt="rtylr" className="h-6 w-auto" />
        </div>

        {/* Hub back link */}
        <div style={{ padding: '10px 8px 4px' }}>
          <a
            href={APP_URLS.hub}
            style={{
              display: 'flex',
              alignItems: 'center',
              gap: '6px',
              padding: '6px 10px',
              borderRadius: '8px',
              fontSize: '12px',
              color: 'var(--hub-text-muted)',
              textDecoration: 'none',
              transition: 'background 0.15s',
            }}
            onMouseEnter={(e) => { (e.currentTarget as HTMLElement).style.background = 'var(--hub-hover)'; }}
            onMouseLeave={(e) => { (e.currentTarget as HTMLElement).style.background = 'transparent'; }}
          >
            <ArrowLeft size={13} />
            <span>Hub</span>
          </a>
        </div>

        {/* Navigation */}
        <nav style={{ flex: 1, overflowY: 'auto', padding: '4px 8px', display: 'flex', flexDirection: 'column', gap: '2px' }}>
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
            Workspace
          </p>
          {navItems.map((item) => (
            <SidebarNavItem key={item.path} {...item} />
          ))}
        </nav>

        {/* Footer — user */}
        <div style={{ padding: '8px', borderTop: '1px solid var(--hub-border)', position: 'relative' }}>
          <button
            onClick={() => setUserMenuOpen((v) => !v)}
            style={{
              display: 'flex',
              alignItems: 'center',
              gap: '8px',
              width: '100%',
              padding: '8px 10px',
              borderRadius: '8px',
              background: 'transparent',
              border: 'none',
              cursor: 'pointer',
              transition: 'background 0.15s',
            }}
            onMouseEnter={(e) => { (e.currentTarget as HTMLElement).style.background = 'var(--hub-hover)'; }}
            onMouseLeave={(e) => { (e.currentTarget as HTMLElement).style.background = 'transparent'; }}
          >
            <div
              style={{
                width: '26px',
                height: '26px',
                borderRadius: '6px',
                background: 'rgba(59,130,246,0.15)',
                color: ACCENT,
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'center',
                fontSize: '11px',
                fontWeight: 700,
                flexShrink: 0,
              }}
            >
              U
            </div>
            <div style={{ flex: 1, minWidth: 0, textAlign: 'left' }}>
              <div style={{ fontSize: '12px', fontWeight: 500, color: 'var(--hub-text)', whiteSpace: 'nowrap', overflow: 'hidden', textOverflow: 'ellipsis' }}>
                My Account
              </div>
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
                  <p style={{ fontSize: '10px', fontWeight: 600, textTransform: 'uppercase', letterSpacing: '0.1em', color: 'var(--hub-text-dim)', marginBottom: '6px', padding: '0 4px' }}>
                    Appearance
                  </p>
                  <div style={{ display: 'flex', gap: '4px' }}>
                    {([
                      { value: 'light' as Theme, icon: Sun, label: 'Light' },
                      { value: 'dark' as Theme, icon: Moon, label: 'Dark' },
                      { value: 'system' as Theme, icon: Monitor, label: 'System' },
                    ]).map(({ value, icon: Icon, label }) => (
                      <button
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
                          border: theme === value ? `1px solid ${ACCENT}40` : '1px solid transparent',
                          background: theme === value ? 'rgba(59,130,246,0.12)' : 'var(--hub-hover)',
                          color: theme === value ? ACCENT : 'var(--hub-text-muted)',
                        }}
                      >
                        <Icon size={12} />
                        {label}
                      </button>
                    ))}
                  </div>
                </div>
                <div style={{ height: '1px', background: 'var(--hub-border)', margin: '4px 0' }} />
                <button
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
                  onMouseEnter={(e) => { (e.currentTarget as HTMLElement).style.background = 'rgba(239,68,68,0.08)'; }}
                  onMouseLeave={(e) => { (e.currentTarget as HTMLElement).style.background = 'transparent'; }}
                >
                  <LogOut size={13} />
                  Sign out
                </button>
              </motion.div>
            )}
          </AnimatePresence>
        </div>
      </aside>

      {/* Main area */}
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
            background: 'var(--hub-surface)',
            borderBottom: '1px solid var(--hub-border)',
          }}
        >
          <h1 style={{ fontSize: '14px', fontWeight: 600, color: 'var(--hub-text)', margin: 0 }}>
            {pageTitle}
          </h1>
          <div style={{ display: 'flex', alignItems: 'center', gap: '8px' }}>
            <button
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
              <span>Search...</span>
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
              onMouseEnter={(e) => { (e.currentTarget as HTMLElement).style.background = 'var(--hub-hover)'; }}
              onMouseLeave={(e) => { (e.currentTarget as HTMLElement).style.background = 'transparent'; }}
            >
              <Bell size={14} />
            </button>
          </div>
        </header>

        {/* Content */}
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
    </div>
  );
};

export default DefaultLayout;
