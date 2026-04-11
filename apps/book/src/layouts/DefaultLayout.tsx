import React, { useEffect, useState } from 'react';
import { useLocation, Link } from 'react-router-dom';
import { motion, AnimatePresence } from 'motion/react';
import { APP_URLS } from '@voxire/config';
import {
  LayoutDashboard,
  Calendar,
  ClipboardList,
  Briefcase,
  Monitor,
  ListOrdered,
  Users,
  Settings as SettingsIcon,
  ChevronRight,
  Sun,
  Moon,
  LogOut,
  ArrowLeft,
  Bell,
} from 'lucide-react';
import { Toaster } from 'sonner';

interface NavItem {
  label: string;
  path: string;
  icon: React.ComponentType<{ size?: number; className?: string }>;
}

const navItems: NavItem[] = [
  { label: 'Dashboard', path: '/', icon: LayoutDashboard },
  { label: 'Calendar', path: '/calendar', icon: Calendar },
  { label: 'Appointments', path: '/appointments', icon: ClipboardList },
  { label: 'Services', path: '/services', icon: Briefcase },
  { label: 'Resources', path: '/resources', icon: Monitor },
  { label: 'Waitlist', path: '/waitlist', icon: ListOrdered },
  { label: 'Customers', path: '/customers', icon: Users },
];

const routeLabels: Record<string, string> = {
  '/': 'Dashboard',
  '/calendar': 'Calendar',
  '/appointments': 'Appointments',
  '/services': 'Services',
  '/resources': 'Resources',
  '/waitlist': 'Waitlist',
  '/customers': 'Customers',
  '/settings': 'Settings',
};

const INDIGO = '#6366f1';

function getPageTitle(pathname: string): string {
  if (routeLabels[pathname]) return routeLabels[pathname];
  const match = Object.keys(routeLabels).find((key) => pathname.startsWith(key) && key !== '/');
  return match ? routeLabels[match] : 'Book';
}

function getInitials(name?: string | null): string {
  if (!name) return '?';
  return name
    .split(' ')
    .map((n) => n[0])
    .join('')
    .toUpperCase()
    .slice(0, 2);
}

type Theme = 'dark' | 'light' | 'system';

function applyTheme(theme: Theme) {
  const root = document.documentElement;
  if (theme === 'system') {
    const prefersDark = window.matchMedia('(prefers-color-scheme: dark)').matches;
    root.setAttribute('data-theme', prefersDark ? 'dark' : 'light');
  } else {
    root.setAttribute('data-theme', theme);
  }
}

const SidebarNavItem: React.FC<NavItem> = ({ label, path, icon: Icon }) => {
  const location = useLocation();
  const isActive = path === '/' ? location.pathname === '/' : location.pathname.startsWith(path);

  return (
    <Link
      to={path}
      className="relative flex items-center gap-3 px-3 py-1.5 rounded-lg text-sm font-medium transition-colors duration-150"
      style={{
        color: isActive ? 'var(--hub-text)' : 'var(--hub-text-muted)',
        textDecoration: 'none',
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
          layoutId="sidebar-active"
          className="absolute inset-0 rounded-lg"
          style={{ background: 'var(--hub-active)' }}
          transition={{ type: 'spring', stiffness: 400, damping: 35 }}
        />
      )}
      <Icon size={14} className="relative shrink-0" />
      <span className="relative">{label}</span>
      {isActive && (
        <span
          className="relative ml-auto h-1 w-1 rounded-full shrink-0"
          style={{ background: INDIGO }}
        />
      )}
    </Link>
  );
};

interface DefaultLayoutProps {
  children?: React.ReactNode;
}

const DefaultLayout: React.FC<DefaultLayoutProps> = ({ children }) => {
  const location = useLocation();
  const [theme, setThemeState] = useState<Theme>(() => {
    return (localStorage.getItem('rtylr-book-theme') as Theme) || 'dark';
  });
  const [showUserMenu, setShowUserMenu] = useState(false);

  const pageTitle = getPageTitle(location.pathname);

  useEffect(() => {
    applyTheme(theme);
  }, [theme]);

  const rawUser = (() => {
    try {
      return JSON.parse(localStorage.getItem('rtylr:voxire:user') || 'null');
    } catch {
      return null;
    }
  })();

  const userEmail = rawUser?.email || rawUser?.emailAddress || null;
  const userName = rawUser?.name || rawUser?.firstName || userEmail || 'User';

  const handleSetTheme = (t: Theme) => {
    setThemeState(t);
    applyTheme(t);
    localStorage.setItem('rtylr-book-theme', t);
  };

  const handleLogout = () => {
    document.cookie = 'token=; Max-Age=0; path=/';
    window.location.href = `${APP_URLS.auth}?redirect=book`;
  };

  return (
    <div className="flex h-screen overflow-hidden" style={{ background: 'var(--hub-bg)' }}>
      <aside
        className="w-56 flex flex-col shrink-0 overflow-hidden"
        style={{
          background: 'var(--hub-surface)',
          borderRight: '1px solid var(--hub-border)',
        }}
      >
        <div
          className="flex items-center gap-2.5 px-4 py-4"
          style={{ borderBottom: '1px solid var(--hub-border)' }}
        >
          <img src="/logo.png" alt="rtylr" className="h-6 w-auto" />
        </div>

        <div className="px-3 pt-3">
          <a
            href={APP_URLS.hub}
            className="flex items-center gap-2 px-3 py-1.5 rounded-lg text-xs transition-colors duration-150 no-underline"
            style={{ color: 'var(--hub-text-muted)' }}
            onMouseEnter={(e) => {
              (e.currentTarget as HTMLElement).style.background = 'var(--hub-hover)';
            }}
            onMouseLeave={(e) => {
              (e.currentTarget as HTMLElement).style.background = 'transparent';
            }}
          >
            <ArrowLeft size={12} />
            <span>Hub</span>
          </a>
        </div>

        <nav className="flex-1 overflow-y-auto py-2 px-2 space-y-0.5">
          {navItems.map((item) => (
            <SidebarNavItem key={item.path} {...item} />
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
          <div className="relative">
            <button
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
              <span
                className="flex items-center justify-center h-6 w-6 rounded-full text-[10px] font-bold shrink-0"
                style={{ background: 'rgba(99,102,241,0.15)', color: INDIGO }}
              >
                {getInitials(userName)}
              </span>
              <div className="flex flex-col items-start min-w-0 flex-1">
                <span
                  className="text-xs font-medium truncate w-full text-left"
                  style={{ color: 'var(--hub-text)' }}
                >
                  {userName}
                </span>
                {userEmail && (
                  <span
                    className="text-[10px] truncate w-full text-left"
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
                  className="absolute bottom-full left-0 right-0 mb-1 rounded-xl p-2 z-50"
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
                    Appearance
                  </p>
                  <div className="flex gap-1 mb-2">
                    {(
                      [
                        { value: 'light' as Theme, icon: Sun, label: 'Light' },
                        { value: 'dark' as Theme, icon: Moon, label: 'Dark' },
                        { value: 'system' as Theme, icon: Monitor, label: 'System' },
                      ] as const
                    ).map(({ value, icon: Icon, label }) => (
                      <button
                        key={value}
                        onClick={() => handleSetTheme(value)}
                        className="flex-1 flex flex-col items-center gap-1 py-2 rounded-lg text-[10px] font-medium transition-colors"
                        style={{
                          background:
                            theme === value ? 'rgba(99,102,241,0.12)' : 'var(--hub-hover)',
                          color: theme === value ? INDIGO : 'var(--hub-text-muted)',
                          border:
                            theme === value
                              ? '1px solid rgba(99,102,241,0.3)'
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
                    onClick={handleLogout}
                    className="flex items-center gap-2 w-full px-2 py-2 rounded-lg text-sm transition-colors"
                    style={{ color: '#ef4444' }}
                    onMouseEnter={(e) => {
                      (e.currentTarget as HTMLElement).style.background = 'rgba(239,68,68,0.08)';
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

      <div className="flex flex-col flex-1 overflow-hidden">
        <header
          className="h-11 flex items-center justify-between px-6 shrink-0"
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
              onClick={() => handleSetTheme(theme === 'dark' ? 'light' : 'dark')}
              className="h-7 w-7 rounded-lg flex items-center justify-center"
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
              className="h-7 w-7 rounded-lg flex items-center justify-center transition-colors duration-150"
              style={{ color: 'var(--hub-text-muted)' }}
              onMouseEnter={(e) => {
                (e.currentTarget as HTMLElement).style.background = 'var(--hub-hover)';
              }}
              onMouseLeave={(e) => {
                (e.currentTarget as HTMLElement).style.background = 'transparent';
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
            className="flex-1 overflow-y-auto p-6"
          >
            {children}
          </motion.main>
        </AnimatePresence>
      </div>

      <Toaster
        theme={theme === 'light' ? 'light' : 'dark'}
        position="bottom-right"
        richColors
        closeButton
      />
    </div>
  );
};

export default DefaultLayout;
