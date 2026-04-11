import React, { useEffect, useState } from 'react';
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
  Calendar,
  ClipboardList,
  Rocket,
  FileText,
  Shield,
  GraduationCap,
  Network,
  Settings,
  MessageSquare,
  CalendarDays,
  Timer,
  Search,
  Command,
  Menu,
  X,
} from 'lucide-react';
import CommandPalette from '../shared/components/CommandPalette';
import GlobalShortcuts from '../shared/components/GlobalShortcuts';
import { useMediaQuery } from '../shared/components/ui';

/* ── Types ── */
interface NavItem {
  label: string;
  path: string;
  icon: React.ComponentType<{ size?: number; className?: string; style?: React.CSSProperties }>;
}

/* ── Nav config ── */
const ACCENT = '#8b5cf6';

interface NavSection {
  title: string;
  items: NavItem[];
}

const navSections: NavSection[] = [
  {
    title: 'Workspace',
    items: [
      { label: 'Dashboard',   path: '/',            icon: LayoutDashboard },
      { label: 'Staff',       path: '/staff',       icon: Users           },
      { label: 'Departments', path: '/departments', icon: Building2       },
      { label: 'Org Chart',   path: '/org-chart',   icon: Network         },
    ],
  },
  {
    title: 'Time',
    items: [
      { label: 'Shifts',     path: '/shifts',     icon: Clock      },
      { label: 'Schedule',   path: '/schedule',   icon: CalendarDays },
      { label: 'Timesheets', path: '/timesheets', icon: Timer       },
      { label: 'Time Off',   path: '/time-off',   icon: Calendar    },
      { label: 'Calendar',   path: '/calendar',   icon: Calendar    },
    ],
  },
  {
    title: 'People',
    items: [
      { label: 'Reviews',      path: '/reviews',      icon: ClipboardList  },
      { label: '1-on-1s',      path: '/one-on-ones',  icon: MessageSquare  },
      { label: 'Onboarding',   path: '/onboarding',   icon: Rocket         },
      { label: 'Offboarding',  path: '/offboarding',  icon: LogOut         },
      { label: 'Training',     path: '/training',     icon: GraduationCap  },
    ],
  },
  {
    title: 'HR',
    items: [
      { label: 'Payroll',    path: '/payroll',    icon: DollarSign },
      { label: 'Benefits',   path: '/benefits',   icon: Shield     },
      { label: 'Documents',  path: '/documents',  icon: FileText   },
      { label: 'Operations', path: '/operations', icon: Timer      },
      { label: 'Settings',   path: '/settings',   icon: Settings   },
    ],
  },
];

const navItems: NavItem[] = navSections.flatMap((s) => s.items);

const routeLabels: Record<string, string> = {
  '/':            'Dashboard',
  '/staff':       'Staff',
  '/shifts':      'Shifts',
  '/payroll':     'Payroll',
  '/departments': 'Departments',
  '/time-off':    'Time Off',
  '/time-off/new':'Request Time Off',
  '/calendar':    'Time Off Calendar',
  '/schedule':    'Shift Schedule',
  '/timesheets':  'Timesheets',
  '/reviews':     'Performance Reviews',
  '/reviews/new': 'New Review',
  '/one-on-ones': '1-on-1s',
  '/onboarding':  'Onboarding',
  '/offboarding': 'Offboarding',
  '/documents':   'Documents',
  '/benefits':    'Benefits',
  '/training':    'Training',
  '/org-chart':   'Org Chart',
  '/operations':  'Operations',
  '/settings':    'Settings',
};

function getPageTitle(pathname: string): string {
  if (routeLabels[pathname]) return routeLabels[pathname];
  if (pathname.startsWith('/staff/')) return 'Employee Profile';
  if (pathname.startsWith('/reviews/')) return 'Performance Reviews';
  if (pathname.startsWith('/time-off/')) return 'Time Off';
  return 'HR';
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

  const rawUser = (() => {
    try {
      return JSON.parse(localStorage.getItem('rtylr:voxire:user') || 'null');
    } catch {
      return null;
    }
  })();
  const userEmail = rawUser?.email || rawUser?.emailAddress || null;
  const userName = rawUser?.name || rawUser?.firstName || userEmail || 'User';

  function getInitials(name: string) {
    return name.split(' ').map((n: string) => n[0]).join('').toUpperCase().slice(0, 2);
  }

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
          {getInitials(userName)}
        </span>
        <div style={{ display: 'flex', flexDirection: 'column', alignItems: 'flex-start', flex: 1, minWidth: 0 }}>
          <span style={{ fontSize: '12px', fontWeight: 600, color: 'var(--hr-text)', textAlign: 'left' }}>{userName}</span>
          {userEmail && <span style={{ fontSize: '10px', color: 'var(--hr-text-muted)', textAlign: 'left' }}>{userEmail}</span>}
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
  const isMobile = useMediaQuery('(max-width: 960px)');
  const [mobileNavOpen, setMobileNavOpen] = useState(false);
  const pageTitle = getPageTitle(location.pathname);

  useEffect(() => {
    setMobileNavOpen(false);
  }, [location.pathname, isMobile]);

  return (
    <div style={{ display: 'flex', height: '100dvh', minHeight: '100vh', overflow: 'hidden', background: 'var(--hr-bg)' }}>
      <CommandPalette />
      <GlobalShortcuts />
      {isMobile && mobileNavOpen && (
        <div
          onClick={() => setMobileNavOpen(false)}
          style={{
            position: 'fixed',
            inset: 0,
            background: 'rgba(0,0,0,0.45)',
            zIndex: 70,
          }}
        />
      )}
      {/* Sidebar */}
      <aside
        style={{
          width: isMobile ? '280px' : '220px',
          display: 'flex',
          flexDirection: 'column',
          flexShrink: 0,
          overflow: 'hidden',
          background: 'var(--hr-surface)',
          borderRight: '1px solid var(--hr-border)',
          position: isMobile ? 'fixed' : 'relative',
          inset: isMobile ? '0 auto 0 0' : 'auto',
          zIndex: 80,
          transform: isMobile ? (mobileNavOpen ? 'translateX(0)' : 'translateX(-100%)') : 'translateX(0)',
          transition: 'transform 0.22s ease',
          boxShadow: isMobile ? 'var(--hr-shadow)' : 'none',
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
            <img src="/logo.png" alt="rtylr" style={{ height: 24, width: 'auto' }} />
          </div>
          <a
            href={APP_URLS.hub}
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
          {navSections.map((section) => (
            <div key={section.title} style={{ display: 'flex', flexDirection: 'column', gap: 2, marginBottom: 6 }}>
              <p
                style={{
                  padding: '6px 12px 4px',
                  fontSize: '10px',
                  fontWeight: 600,
                  textTransform: 'uppercase',
                  letterSpacing: '0.1em',
                  color: 'var(--hr-text-dim)',
                  margin: 0,
                }}
              >
                {section.title}
              </p>
              {section.items.map((item) => (
                <SidebarNavItem key={item.path} {...item} />
              ))}
            </div>
          ))}
        </nav>

        {/* Footer */}
        <div style={{ padding: '8px', borderTop: '1px solid var(--hr-border)' }}>
          <UserMenu theme={theme} setTheme={setTheme} />
        </div>
      </aside>

      {/* Main */}
      <div style={{ display: 'flex', flexDirection: 'column', flex: 1, overflow: 'hidden', minWidth: 0 }}>
        {/* Topbar */}
        <header
          style={{
            height: isMobile ? '56px' : '44px',
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'space-between',
            padding: isMobile ? '0 16px' : '0 24px',
            flexShrink: 0,
            background: 'var(--hr-surface)',
            borderBottom: '1px solid var(--hr-border)',
          }}
        >
          <div style={{ display: 'flex', alignItems: 'center', gap: 10, minWidth: 0 }}>
            {isMobile && (
              <button
                onClick={() => setMobileNavOpen((open) => !open)}
                style={{
                  height: 36,
                  width: 36,
                  borderRadius: 10,
                  display: 'flex',
                  alignItems: 'center',
                  justifyContent: 'center',
                  background: 'var(--hr-hover)',
                  border: '1px solid var(--hr-border)',
                  cursor: 'pointer',
                  color: 'var(--hr-text)',
                  flexShrink: 0,
                }}
                aria-label={mobileNavOpen ? 'Close navigation' : 'Open navigation'}
              >
                {mobileNavOpen ? <X size={16} /> : <Menu size={16} />}
              </button>
            )}
            <h1 style={{ fontSize: '13px', fontWeight: 600, color: 'var(--hr-text)', margin: 0, overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>
              {pageTitle}
            </h1>
          </div>
          <div style={{ display: 'flex', alignItems: 'center', gap: '6px' }}>
            <button
              onClick={() => {
                window.dispatchEvent(new CustomEvent('hr:open-palette'));
              }}
              style={{
                height: isMobile ? '36px' : '28px',
                display: 'flex',
                alignItems: 'center',
                gap: 6,
                padding: isMobile ? '0 12px' : '0 10px',
                borderRadius: '8px',
                background: 'var(--hr-hover)',
                border: '1px solid var(--hr-border)',
                cursor: 'pointer',
                color: 'var(--hr-text-muted)',
                fontSize: 11,
              }}
              onMouseEnter={(e) => ((e.currentTarget as HTMLButtonElement).style.background = 'var(--hr-active)')}
              onMouseLeave={(e) => ((e.currentTarget as HTMLButtonElement).style.background = 'var(--hr-hover)')}
              aria-label="Open command palette"
            >
              <Search size={12} />
              {!isMobile && 'Search'}
              {!isMobile && <kbd style={{ display: 'inline-flex', alignItems: 'center', gap: 2, padding: '1px 4px', borderRadius: 4, border: '1px solid var(--hr-border)', fontSize: 9 }}>
                <Command size={9} />K
              </kbd>}
            </button>
            <button
              style={{
                height: isMobile ? '36px' : '28px',
                width: isMobile ? '36px' : '28px',
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
            style={{ flex: 1, overflowY: 'auto', padding: isMobile ? '16px' : '24px' }}
          >
            {children}
          </motion.main>
        </AnimatePresence>
      </div>
    </div>
  );
};

export default DefaultLayout;
