import type React from 'react';
import {
  LayoutDashboard,
  Users,
  Clock,
  DollarSign,
  Building2,
  Bell,
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
  GitBranch,
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
import type { HRCompanyConfigPayload } from '@voxire/voxie/hr/setup/v1/types';
import { isPayrollConfigured } from './shared/lib/payrollConfig';
import { canHrPermission, type HRPermission } from './shared/lib/hrPermissions';
import { hasEffectiveAppAccess } from '@voxire/permissions';
import type { StoredUser } from '../../shared/stores/userStore';
interface NavItem {
  i18nKey: string;
  path: string;
  icon: React.ComponentType<{ size?: number; className?: string; style?: React.CSSProperties }>;
  /** Optional guard: item is only shown when this returns true for the current config */
  configGuard?: (timeTrackingModel: string, config: HRCompanyConfigPayload | null) => boolean;
  permission?: HRPermission;
}

type EquipmentConfig = HRCompanyConfigPayload & { equipmentEnabled?: boolean | string };

const isEquipmentEnabled = (config: HRCompanyConfigPayload | null): boolean => {
  const value = (config as EquipmentConfig | null)?.equipmentEnabled;
  return value === true || value === 'true';
};

/* -- Nav config -- */
/** The themed accent, so a workspace's own colour is used (rtylr#1242). */
const ACCENT = 'var(--accent)';

interface NavSection {
  i18nKey: string;
  items: NavItem[];
}

export const hrNavSections: NavSection[] = [
  {
    i18nKey: 'hrLayout.sections.workspace',
    items: [
      { i18nKey: 'hrNav.dashboard', path: '/hr', icon: LayoutDashboard },
      { i18nKey: 'hrNav.staff', path: '/hr/staff', icon: Users, permission: 'hr:employees:view' },
      {
        i18nKey: 'hrNav.departments',
        path: '/hr/departments',
        icon: Building2,
        permission: 'hr:employees:view',
      },
      { i18nKey: 'hrNav.orgChart', path: '/hr/org-chart', icon: Network },
      {
        i18nKey: 'hrNav.reportingLines',
        path: '/hr/staff/reporting-lines',
        icon: GitBranch,
        permission: 'hr:employees:manage',
      },
    ],
  },
  {
    i18nKey: 'hrLayout.sections.time',
    items: [
      {
        i18nKey: 'hrNav.shifts',
        path: '/hr/shifts',
        icon: Clock,
        permission: 'hr:schedules:manage',
      },
      {
        i18nKey: 'hrNav.schedule',
        path: '/hr/schedule',
        icon: CalendarDays,
        permission: 'hr:schedules:manage',
      },
      {
        i18nKey: 'hrNav.rosterBuilder',
        path: '/hr/shifts/roster',
        icon: CalendarRange,
        permission: 'hr:schedules:manage',
      },
      {
        i18nKey: 'hrNav.timeLogging',
        path: '/hr/time',
        icon: PlayCircle,
        // Reads the raw config rather than the defaulted `m`, so this matches
        // the route's own guard exactly. `timeTrackingModel` is never
        // populated by the backend, and the sidebar's `?? 'both'` fallback
        // therefore advertised a route that redirects away for every user and
        // every role. The fallback stays for Attendance below, whose route has
        // no config guard at all and is always reachable.
        configGuard: (_m, cfg) =>
          cfg?.timeTrackingModel === 'time_logging' || cfg?.timeTrackingModel === 'both',
      },
      {
        i18nKey: 'hrNav.timesheets',
        path: '/hr/timesheets',
        icon: Timer,
        permission: 'hr:schedules:manage',
      },
      {
        i18nKey: 'hrNav.attendance',
        path: '/hr/attendance',
        icon: CalendarCheck,
        configGuard: (m) => m === 'attendance' || m === 'both',
      },
      {
        i18nKey: 'hrNav.approvals',
        path: '/hr/timesheets/approvals',
        icon: CheckSquare,
        permission: 'hr:time_off:approve',
      },
      {
        i18nKey: 'hrNav.timeOff',
        path: '/hr/time-off',
        icon: Calendar,
        permission: 'hr:time_off:request',
      },
      {
        i18nKey: 'hrNav.leave',
        path: '/hr/leave',
        icon: Palmtree,
        permission: 'hr:time_off:request',
      },
      {
        i18nKey: 'hrNav.overtime',
        path: '/hr/overtime',
        icon: ClockArrowUp,
        permission: 'hr:time_off:request',
      },
      {
        i18nKey: 'hrNav.calendar',
        path: '/hr/calendar',
        icon: Calendar,
        permission: 'hr:time_off:request',
      },
    ],
  },
  {
    i18nKey: 'hrLayout.sections.people',
    items: [
      {
        i18nKey: 'hrNav.reviews',
        path: '/hr/reviews',
        icon: ClipboardList,
        permission: 'hr:employees:view',
      },
      { i18nKey: 'hrNav.oneOnOnes', path: '/hr/one-on-ones', icon: MessageSquare },
      {
        i18nKey: 'hrNav.performance',
        path: '/hr/performance',
        icon: BarChart3,
        permission: 'hr:employees:view',
      },
      { i18nKey: 'hrNav.myReview', path: '/hr/performance/my-review', icon: Star },
      { i18nKey: 'hrNav.goals', path: '/hr/performance/goals', icon: Target },
      {
        i18nKey: 'hrNav.onboarding',
        path: '/hr/onboarding',
        icon: Rocket,
        permission: 'hr:employees:manage',
      },
      {
        i18nKey: 'hrNav.offboarding',
        path: '/hr/offboarding',
        icon: LogOut,
        permission: 'hr:employees:manage',
      },
      {
        i18nKey: 'hrNav.training',
        path: '/hr/training',
        icon: GraduationCap,
        permission: 'hr:employees:manage',
      },
    ],
  },
  {
    i18nKey: 'hrLayout.sections.hr',
    items: [
      {
        i18nKey: 'hrNav.payroll',
        path: '/hr/payroll',
        icon: DollarSign,
        configGuard: (_m, cfg) => isPayrollConfigured(cfg),
        permission: 'hr:payroll:view',
      },
      {
        // No `permission` here: `/hr/payroll/my-runs` already scopes to the
        // caller's own employee id server-side, so every HR app user should
        // see their own pay slips regardless of role - see the comment on
        // `myPayslipsWrap` in `router.tsx`.
        i18nKey: 'hrNav.myPaySlips',
        path: '/hr/payroll/my-payslips',
        icon: Receipt,
        configGuard: (_m, cfg) => isPayrollConfigured(cfg),
      },
      {
        i18nKey: 'hrNav.payrollConfig',
        path: '/hr/settings/payroll-config',
        icon: Settings,
        permission: 'hr:payroll:view',
      },
      { i18nKey: 'hrNav.benefits', path: '/hr/benefits', icon: Shield },
      {
        i18nKey: 'hrNav.equipment',
        path: '/hr/equipment',
        icon: Package,
        configGuard: (_m, cfg) => isEquipmentEnabled(cfg),
      },
      {
        i18nKey: 'hrNav.assets',
        path: '/hr/equipment/assets',
        icon: HardDrive,
        configGuard: (_m, cfg) => isEquipmentEnabled(cfg),
      },
      { i18nKey: 'hrNav.documents', path: '/hr/documents', icon: FileText },
      {
        i18nKey: 'hrNav.reports',
        path: '/hr/reports',
        icon: FileBarChart,
        permission: 'hr:employees:view',
      },
      {
        i18nKey: 'hrNav.operations',
        path: '/hr/operations',
        icon: Timer,
        permission: 'hr:schedules:manage',
      },
      {
        i18nKey: 'hrNav.rolesPermissions',
        path: '/hr/settings/roles-permissions',
        icon: Shield,
        permission: 'hub:business:settings:manage',
      },
      {
        i18nKey: 'hrNav.settings',
        path: '/hr/settings',
        icon: Settings,
        permission: 'hub:business:settings:manage',
      },
    ],
  },
];

export const hrNavItems = hrNavSections
  .flatMap((section) => section.items)
  .filter((item) => !item.path.startsWith('/hr/settings'))
  .map((item) => ({ ...item, key: item.i18nKey.replace('hrNav.', '') }));
export function canAccessHrPath(user: StoredUser | null, path: string): boolean {
  const access = user?.effectiveAccess;
  if (!hasEffectiveAppAccess(access, 'hr')) return false;
  if (access?.isOwner || access?.isAdmin) return true;
  if (path === '/hr') return true;
  if (path === '/hr/setup' || path === '/settings/hr' || path === '/hr/settings')
    return canHrPermission('hub:business:settings:manage', access);
  if (path.startsWith('/settings/hr/payroll-config'))
    return canHrPermission('hr:payroll:manage', access);
  if (path.startsWith('/settings/hr/leave-policy'))
    return canHrPermission('hr:time_off:approve', access);
  if (path === '/hr/employees/new') return canHrPermission('hr:employees:manage', access);
  if (path.startsWith('/hr/employees')) return canHrPermission('hr:employees:view', access);
  if (path.startsWith('/hr/onboarding/templates'))
    return canHrPermission('hr:employees:manage', access);
  if (path === '/hr/time-off/new') return true;
  const item = hrNavItems
    .filter(
      (item) => item.path !== '/hr' && (path === item.path || path.startsWith(`${item.path}/`))
    )
    .sort((a, b) => b.path.length - a.path.length)[0];
  return !!item && (!item.permission || canHrPermission(item.permission, access));
}
export function getHrNavItems(config: HRCompanyConfigPayload | null) {
  return hrNavItems.filter(
    (item) => !item.configGuard || item.configGuard(config?.timeTrackingModel ?? 'both', config)
  );
}
