import React, { useEffect, useRef, useState } from 'react';
import { useLocation, Link } from 'react-router-dom';
import { motion, AnimatePresence } from 'motion/react';
import { useTranslation } from 'react-i18next';
import { APP_URLS, getAvatarUrl } from '@voxire/config';
import { signOut } from '@voxire/auth';
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
  PlayCircle,
  CalendarCheck,
  CheckSquare,
  Palmtree,
  ClockArrowUp,
  CalendarRange,
  Receipt,
  Package,
  HardDrive,
  BarChart3,
  Star,
  Target,
  FileBarChart,
} from 'lucide-react';
import type { HRCompanyConfigPayload } from '@voxire/contracts/hr/v1/setup';
import { useForesightLinks } from '@voxire/utils';
import { useHRConfigStore } from '../shared/stores/hrConfigStore';
import CommandPalette from '../shared/components/CommandPalette';
import GlobalShortcuts from '../shared/components/GlobalShortcuts';
import { useMediaQuery } from '../shared/components/ui';

/* ── Types ── */
interface NavItem {
  i18nKey: string;
  path: string;
  icon: React.ComponentType<{ size?: number; className?: string; style?: React.CSSProperties }>;
  /** Optional guard: item is only shown when this returns true for the current config */
  configGuard?: (timeTrackingModel: string, config: HRCompanyConfigPayload | null) => boolean;
}

/* ── Nav config ── */
const ACCENT = '#8b5cf6';

interface NavSection {
  i18nKey: string;
  items: NavItem[];
}

const navSections: NavSection[] = [
  {
    i18nKey: 'layout.sections.workspace',
    items: [
      { i18nKey: 'nav.dashboard', path: '/', icon: LayoutDashboard },
      { i18nKey: 'nav.staff', path: '/staff', icon: Users },
      { i18nKey: 'nav.departments', path: '/departments', icon: Building2 },
      { i18nKey: 'nav.orgChart', path: '/org-chart', icon: Network },
    ],
  },
  {
    i18nKey: 'layout.sections.time',
    items: [
      { i18nKey: 'nav.shifts', path: '/shifts', icon: Clock },
      { i18nKey: 'nav.schedule', path: '/schedule', icon: CalendarDays },
      { i18nKey: 'nav.rosterBuilder', path: '/shifts/roster', icon: CalendarRange },
      {
        i18nKey: 'nav.timeLogging',
        path: '/time',
        icon: PlayCircle,
        configGuard: (m) => m === 'time_logging' || m === 'both',
      },
      { i18nKey: 'nav.timesheets', path: '/timesheets', icon: Timer },
      {
        i18nKey: 'nav.attendance',
        path: '/attendance',
        icon: CalendarCheck,
        configGuard: (m) => m === 'attendance' || m === 'both',
      },
      {
        i18nKey: 'nav.approvals',
        path: '/timesheets/approvals',
        icon: CheckSquare,
      },
      { i18nKey: 'nav.timeOff', path: '/time-off', icon: Calendar },
      { i18nKey: 'nav.leave', path: '/leave', icon: Palmtree },
      { i18nKey: 'nav.overtime', path: '/overtime', icon: ClockArrowUp },
      { i18nKey: 'nav.calendar', path: '/calendar', icon: Calendar },
    ],
  },
  {
    i18nKey: 'layout.sections.people',
    items: [
      { i18nKey: 'nav.reviews', path: '/reviews', icon: ClipboardList },
      { i18nKey: 'nav.oneOnOnes', path: '/one-on-ones', icon: MessageSquare },
      { i18nKey: 'nav.performance', path: '/performance', icon: BarChart3 },
      { i18nKey: 'nav.myReview', path: '/performance/my-review', icon: Star },
      { i18nKey: 'nav.goals', path: '/performance/goals', icon: Target },
      { i18nKey: 'nav.onboarding', path: '/onboarding', icon: Rocket },
      { i18nKey: 'nav.offboarding', path: '/offboarding', icon: LogOut },
      { i18nKey: 'nav.training', path: '/training', icon: GraduationCap },
    ],
  },
  {
    i18nKey: 'layout.sections.hr',
    items: [
      {
        i18nKey: 'nav.payroll',
        path: '/payroll',
        icon: DollarSign,
        configGuard: (_m, cfg) => !!cfg?.payrollPeriod,
      },
      {
        i18nKey: 'nav.myPaySlips',
        path: '/payroll/my-payslips',
        icon: Receipt,
        configGuard: (_m, cfg) => !!cfg?.payrollPeriod,
      },
      {
        i18nKey: 'nav.payrollConfig',
        path: '/settings/payroll-config',
        icon: Settings,
        configGuard: (_m, cfg) => !!cfg?.payrollPeriod,
      },
      { i18nKey: 'nav.benefits', path: '/benefits', icon: Shield },
      {
        i18nKey: 'nav.equipment',
        path: '/equipment',
        icon: Package,
        configGuard: (_m, cfg) => !!cfg,
      },
      {
        i18nKey: 'nav.assets',
        path: '/equipment/assets',
        icon: HardDrive,
        configGuard: (_m, cfg) => !!cfg,
      },
      { i18nKey: 'nav.documents', path: '/documents', icon: FileText },
      { i18nKey: 'nav.reports', path: '/reports', icon: FileBarChart },
      { i18nKey: 'nav.operations', path: '/operations', icon: Timer },
      { i18nKey: 'nav.settings', path: '/settings', icon: Settings },
    ],
  },
];

const routeLabelKeys: Record<string, string> = {
  '/': 'nav.dashboard',
  '/staff': 'nav.staff',
  '/shifts': 'nav.shifts',
  '/payroll': 'nav.payroll',
  '/payroll/my-payslips': 'nav.myPaySlips',
  '/settings/payroll-config': 'layout.routes.payrollConfig',
  '/departments': 'nav.departments',
  '/time': 'nav.timeLogging',
  '/time-off': 'nav.timeOff',
  '/time-off/new': 'layout.routes.requestTimeOff',
  '/calendar': 'layout.routes.timeOffCalendar',
  '/schedule': 'layout.routes.shiftSchedule',
  '/timesheets': 'nav.timesheets',
  '/timesheets/approvals': 'nav.approvals',
  '/attendance': 'nav.attendance',
  '/reviews': 'layout.routes.performanceReviews',
  '/reviews/new': 'layout.routes.newReview',
  '/one-on-ones': 'nav.oneOnOnes',
  '/onboarding': 'nav.onboarding',
  '/onboarding/templates': 'layout.routes.checklistTemplates',
  '/offboarding': 'nav.offboarding',
  '/documents': 'nav.documents',
  '/equipment': 'nav.equipment',
  '/equipment/assets': 'nav.assets',
  '/benefits': 'nav.benefits',
  '/training': 'nav.training',
  '/org-chart': 'nav.orgChart',
  '/operations': 'nav.operations',
  '/settings': 'nav.settings',
  '/leave': 'nav.leave',
  '/leave/new': 'layout.routes.newLeaveRequest',
  '/overtime': 'nav.overtime',
  '/shifts/roster': 'nav.rosterBuilder',
  '/performance': 'nav.performance',
  '/performance/my-review': 'nav.myReview',
  '/performance/goals': 'nav.goals',
  '/reports': 'nav.reports',
};

/* ── Theme helpers ── */
type Theme = 'light' | 'dark' | 'system';

function useTheme(): [Theme, (t: Theme) => void] {
  const [theme, setThemeState] = useState<Theme>(() => {
    return (localStorage.getItem('hr-theme') as Theme) || 'light';
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
const SidebarNavItem: React.FC<NavItem> = ({ i18nKey, path, icon: Icon }) => {
  const { t } = useTranslation();
  const location = useLocation();
  const isActive = path === '/' ? location.pathname === '/' : location.pathname.startsWith(path);
  const label = t(i18nKey);

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
      <Icon
        size={15}
        style={{ position: 'relative', flexShrink: 0, color: isActive ? ACCENT : undefined }}
      />
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
  const { t } = useTranslation();
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
        onMouseEnter={(e) =>
          ((e.currentTarget as HTMLButtonElement).style.background = 'var(--hr-hover)')
        }
        onMouseLeave={(e) =>
          ((e.currentTarget as HTMLButtonElement).style.background = 'transparent')
        }
      >
        <img
          src={getAvatarUrl(userEmail || userName)}
          alt={userName}
          style={{
            height: 28,
            width: 28,
            borderRadius: '8px',
            background: 'rgba(139,92,246,0.15)',
            flexShrink: 0,
            objectFit: 'cover',
          }}
        />
        <div
          style={{
            display: 'flex',
            flexDirection: 'column',
            alignItems: 'flex-start',
            flex: 1,
            minWidth: 0,
          }}
        >
          <span
            style={{
              fontSize: '12px',
              fontWeight: 600,
              color: 'var(--hr-text)',
              textAlign: 'left',
            }}
          >
            {userName}
          </span>
          {userEmail && (
            <span style={{ fontSize: '10px', color: 'var(--hr-text-muted)', textAlign: 'left' }}>
              {userEmail}
            </span>
          )}
        </div>
        <ChevronUp
          size={12}
          style={{
            color: 'var(--hr-text-dim)',
            flexShrink: 0,
            transform: open ? 'rotate(180deg)' : 'rotate(0)',
            transition: 'transform 0.2s',
          }}
        />
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
            <p
              style={{
                fontSize: '10px',
                fontWeight: 600,
                textTransform: 'uppercase',
                letterSpacing: '0.08em',
                color: 'var(--hr-text-dim)',
                padding: '4px 8px 6px',
              }}
            >
              {t('layout.appearance')}
            </p>
            <div style={{ display: 'flex', gap: '4px', marginBottom: '6px', padding: '0 4px' }}>
              {[
                { value: 'light' as Theme, icon: Sun, i18nKey: 'layout.theme.light' },
                { value: 'dark' as Theme, icon: Moon, i18nKey: 'layout.theme.dark' },
                { value: 'system' as Theme, icon: Monitor, i18nKey: 'layout.theme.system' },
              ].map(({ value, icon: Icon, i18nKey }) => (
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
                  {t(i18nKey)}
                </button>
              ))}
            </div>
            <hr
              style={{ border: 'none', borderTop: '1px solid var(--hr-border)', margin: '4px 0' }}
            />
            <button
              onClick={() => {
                signOut('hr');
              }}
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
              onMouseEnter={(e) =>
                ((e.currentTarget as HTMLButtonElement).style.background = 'rgba(239,68,68,0.08)')
              }
              onMouseLeave={(e) =>
                ((e.currentTarget as HTMLButtonElement).style.background = 'transparent')
              }
            >
              <LogOut size={14} />
              {t('layout.signOut')}
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
  const { t } = useTranslation();
  const location = useLocation();
  const [theme, setTheme] = useTheme();
  const isMobile = useMediaQuery('(max-width: 960px)');
  const [mobileNavOpen, setMobileNavOpen] = useState(false);
  const [collapsedSections, setCollapsedSections] = useState<Record<string, boolean>>({});
  const config = useHRConfigStore((s) => s.config);
  const timeTrackingModel = config?.timeTrackingModel ?? 'both';

  const sidebarRef = useRef<HTMLElement | null>(null);
  useForesightLinks(sidebarRef, {
    '/': () => import('../features/dashboard/Dashboard'),
    '/staff': () => import('../features/staff/Staff'),
    '/departments': () => import('../features/departments/Departments'),
    '/org-chart': () => import('../features/departments/OrgChart'),
    '/shifts': () => import('../features/shifts/Shifts'),
    '/schedule': () => import('../features/shifts/Schedule'),
    '/shifts/roster': () => import('../features/shifts/RosterBuilder'),
    '/time': () => import('../features/time-logging/TimeLogging'),
    '/timesheets': () => import('../features/timesheets/TimesheetsList'),
    '/attendance': () => import('../features/attendance/Attendance'),
    '/timesheets/approvals': () => import('../features/timesheets/approvals/TimesheetApprovals'),
    '/time-off': () => import('../features/time-off/TimeOff'),
    '/leave': () => import('../features/leave/Leave'),
    '/overtime': () => import('../features/overtime/Overtime'),
    '/calendar': () => import('../features/calendar/Calendar'),
    '/reviews': () => import('../features/reviews/Reviews'),
    '/one-on-ones': () => import('../features/reviews/OneOnOnes'),
    '/performance': () => import('../features/performance/Performance'),
    '/performance/my-review': () => import('../features/performance/MyReview'),
    '/performance/goals': () => import('../features/performance/GoalsList'),
    '/onboarding': () => import('../features/onboarding/Onboarding'),
    '/offboarding': () => import('../features/onboarding/Offboarding'),
    '/training': () => import('../features/training/Training'),
    '/payroll': () => import('../features/payroll/Payroll'),
    '/payroll/my-payslips': () => import('../features/payroll/MyPaySlips'),
    '/settings/payroll-config': () =>
      import('../features/settings/payroll-config/PayrollConfigSettings'),
    '/benefits': () => import('../features/benefits/Benefits'),
    '/equipment': () => import('../features/equipment/Equipment'),
    '/equipment/assets': () => import('../features/equipment/AssetRegister'),
    '/documents': () => import('../features/documents/Documents'),
    '/reports': () => import('../features/reports/Reports'),
    '/operations': () => import('../features/operations/Operations'),
    '/settings': () => import('../features/settings/Settings'),
  });

  const toggleSection = (key: string): void => {
    setCollapsedSections((prev) => ({ ...prev, [key]: !prev[key] }));
  };

  const getPageTitle = (pathname: string): string => {
    if (routeLabelKeys[pathname]) return t(routeLabelKeys[pathname]);
    if (pathname.startsWith('/staff/')) return t('layout.routes.employeeProfile');
    if (pathname.startsWith('/reviews/')) return t('layout.routes.performanceReviews');
    if (pathname.startsWith('/timesheets/')) return t('layout.routes.timesheetDetail');
    if (pathname.startsWith('/time-off/')) return t('nav.timeOff');
    if (pathname.startsWith('/payroll/my-payslips')) return t('nav.myPaySlips');
    if (pathname.startsWith('/payroll/')) return t('layout.routes.payrollRunDetail');
    if (pathname.startsWith('/performance/cycles/')) return t('layout.routes.reviewCycleDetail');
    if (pathname.startsWith('/equipment/')) return t('nav.assets');
    if (pathname.startsWith('/onboarding/')) return t('layout.routes.checklistTemplates');
    return t('layout.hr');
  };
  const pageTitle = getPageTitle(location.pathname);

  useEffect(() => {
    setMobileNavOpen(false);
  }, [location.pathname, isMobile]);

  return (
    <div
      style={{
        display: 'flex',
        height: '100dvh',
        minHeight: '100vh',
        overflow: 'hidden',
        background: 'var(--hr-bg)',
      }}
    >
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
          transform: isMobile
            ? mobileNavOpen
              ? 'translateX(0)'
              : 'translateX(-100%)'
            : 'translateX(0)',
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
            href={APP_URLS.dash}
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
            onMouseEnter={(e) =>
              ((e.currentTarget as HTMLAnchorElement).style.color = 'var(--hr-text-muted)')
            }
            onMouseLeave={(e) =>
              ((e.currentTarget as HTMLAnchorElement).style.color = 'var(--hr-text-dim)')
            }
          >
            <ArrowLeft size={11} />
            {t('layout.hubLink')}
          </a>
        </div>

        {/* Navigation */}
        <nav
          ref={sidebarRef}
          style={{
            flex: 1,
            overflowY: 'auto',
            padding: '8px 8px',
            display: 'flex',
            flexDirection: 'column',
            gap: '2px',
          }}
        >
          {navSections.map((section) => {
            const visibleItems = section.items.filter(
              (item) => !item.configGuard || item.configGuard(timeTrackingModel, config)
            );
            if (visibleItems.length === 0) return null;
            const isCollapsed = collapsedSections[section.i18nKey] ?? false;
            return (
              <div
                key={section.i18nKey}
                style={{ display: 'flex', flexDirection: 'column', gap: 2, marginBottom: 4 }}
              >
                <button
                  onClick={() => toggleSection(section.i18nKey)}
                  style={{
                    display: 'flex',
                    alignItems: 'center',
                    justifyContent: 'space-between',
                    width: '100%',
                    padding: '6px 12px 4px',
                    background: 'transparent',
                    border: 'none',
                    cursor: 'pointer',
                  }}
                >
                  <span
                    style={{
                      fontSize: '10px',
                      fontWeight: 600,
                      textTransform: 'uppercase',
                      letterSpacing: '0.1em',
                      color: 'var(--hr-text-dim)',
                    }}
                  >
                    {t(section.i18nKey)}
                  </span>
                  <ChevronUp
                    size={11}
                    style={{
                      color: 'var(--hr-text-dim)',
                      transform: isCollapsed ? 'rotate(180deg)' : 'rotate(0)',
                      transition: 'transform 0.2s',
                      flexShrink: 0,
                    }}
                  />
                </button>
                <AnimatePresence initial={false}>
                  {!isCollapsed && (
                    <motion.div
                      key="section-items"
                      initial={{ height: 0, opacity: 0 }}
                      animate={{ height: 'auto', opacity: 1 }}
                      exit={{ height: 0, opacity: 0 }}
                      transition={{ duration: 0.18, ease: 'easeOut' }}
                      style={{
                        overflow: 'hidden',
                        display: 'flex',
                        flexDirection: 'column',
                        gap: 2,
                      }}
                    >
                      {visibleItems.map((item) => (
                        <SidebarNavItem key={item.path} {...item} />
                      ))}
                    </motion.div>
                  )}
                </AnimatePresence>
              </div>
            );
          })}
        </nav>

        {/* Footer */}
        <div style={{ padding: '8px', borderTop: '1px solid var(--hr-border)' }}>
          <UserMenu theme={theme} setTheme={setTheme} />
        </div>
      </aside>

      {/* Main */}
      <div
        style={{
          display: 'flex',
          flexDirection: 'column',
          flex: 1,
          overflow: 'hidden',
          minWidth: 0,
        }}
      >
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
            <h1
              style={{
                fontSize: '13px',
                fontWeight: 600,
                color: 'var(--hr-text)',
                margin: 0,
                overflow: 'hidden',
                textOverflow: 'ellipsis',
                whiteSpace: 'nowrap',
              }}
            >
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
              onMouseEnter={(e) =>
                ((e.currentTarget as HTMLButtonElement).style.background = 'var(--hr-active)')
              }
              onMouseLeave={(e) =>
                ((e.currentTarget as HTMLButtonElement).style.background = 'var(--hr-hover)')
              }
              aria-label="Open command palette"
            >
              <Search size={12} />
              {!isMobile && 'Search'}
              {!isMobile && (
                <kbd
                  style={{
                    display: 'inline-flex',
                    alignItems: 'center',
                    gap: 2,
                    padding: '1px 4px',
                    borderRadius: 4,
                    border: '1px solid var(--hr-border)',
                    fontSize: 9,
                  }}
                >
                  <Command size={9} />K
                </kbd>
              )}
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
              onMouseEnter={(e) =>
                ((e.currentTarget as HTMLButtonElement).style.background = 'var(--hr-hover)')
              }
              onMouseLeave={(e) =>
                ((e.currentTarget as HTMLButtonElement).style.background = 'transparent')
              }
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
