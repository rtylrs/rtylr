import React, { Suspense } from 'react';
import { createRoute, Outlet, Navigate, type AnyRoute } from '@tanstack/react-router';
import { AppFlagGuard, FLAG_KEYS } from '@voxire/config';

import { ErrorBoundary } from '@voxire/ui';
import { paySlipSearchSchema, payrollRunParamsSchema } from './shared/schemas';
import HRConfigRoute from './features/setup-wizard/HRConfigRoute';

import WizardLayout from './layouts/WizardLayout';
import HRPermissionGuard from './shared/components/HRPermissionGuard';
import { PayrollConfigGuard } from './shared/components/PayrollConfigGuard';
import type { HRPermission } from './shared/lib/hrPermissions';
import { useHRConfigStore } from './shared/stores/hrConfigStore';

type FeatureConfig = {
  timeTrackingModel?: string;
  equipmentEnabled?: boolean | string;
};

const isEnabled = (value: boolean | string | undefined): boolean =>
  value === true || value === 'true';

function HRFeatureConfigGuard({
  enabled,
  children,
}: {
  enabled: (config: FeatureConfig) => boolean;
  children: React.ReactNode;
}): React.JSX.Element {
  const config = useHRConfigStore((state) => state.config) as FeatureConfig | null;
  return config && enabled(config) ? <>{children}</> : <Navigate to="/hr" replace />;
}

function PageLoader(): React.JSX.Element {
  return (
    <div
      className="flex min-h-48 items-center justify-center"
      style={{ background: 'var(--hr-bg)' }}
    >
      <div
        className="h-8 w-8 rounded-full border-2 animate-spin"
        style={{ borderColor: 'var(--hr-border)', borderTopColor: 'var(--accent)' }}
      />
    </div>
  );
}

const SetupWizard = React.lazy(() => import('./features/setup-wizard/SetupWizard'));
const Dashboard = React.lazy(() => import('./features/dashboard/Dashboard'));
const Employees = React.lazy(() => import('./features/employees/Employees'));
const EmployeeDetail = React.lazy(() => import('./features/employees/EmployeeDetail'));
const NewEmployee = React.lazy(() => import('./features/employees/NewEmployee'));
const Staff = React.lazy(() => import('./features/staff/Staff'));
const Shifts = React.lazy(() => import('./features/shifts/Shifts'));
const Payroll = React.lazy(() => import('./features/payroll/Payroll'));
const PayrollRunDetail = React.lazy(() => import('./features/payroll/PayrollRunDetail'));
const MyPaySlips = React.lazy(() => import('./features/payroll/MyPaySlips'));
const PaySlip = React.lazy(() => import('./features/payroll/PaySlip'));

const Departments = React.lazy(() => import('./features/departments/Departments'));
const EmployeeProfile = React.lazy(() => import('./features/staff/EmployeeProfile'));
const ReportingLines = React.lazy(() => import('./features/staff/ReportingLines'));
const TimeOff = React.lazy(() => import('./features/time-off/TimeOff'));
const TimeOffNew = React.lazy(() => import('./features/time-off/TimeOffNew'));
const CalendarScreen = React.lazy(() => import('./features/calendar/Calendar'));
const Schedule = React.lazy(() => import('./features/shifts/Schedule'));
const TimesheetsList = React.lazy(() => import('./features/timesheets/TimesheetsList'));
const TimesheetDetail = React.lazy(() => import('./features/timesheets/TimesheetDetail'));
const TimesheetApprovals = React.lazy(
  () => import('./features/timesheets/approvals/TimesheetApprovals')
);
const TimeLogging = React.lazy(() => import('./features/time-logging/TimeLogging'));
const AttendanceScreen = React.lazy(() => import('./features/attendance/Attendance'));
const Reviews = React.lazy(() => import('./features/reviews/Reviews'));
const ReviewForm = React.lazy(() => import('./features/reviews/ReviewForm'));
const OneOnOnes = React.lazy(() => import('./features/reviews/OneOnOnes'));
const OnboardingManagement = React.lazy(() => import('./features/onboarding/OnboardingManagement'));
const ChecklistTemplates = React.lazy(() => import('./features/onboarding/ChecklistTemplates'));
const Offboarding = React.lazy(() => import('./features/onboarding/Offboarding'));
const Documents = React.lazy(() => import('./features/documents/Documents'));
const EquipmentScreen = React.lazy(() => import('./features/equipment/Equipment'));
const AssetRegister = React.lazy(() => import('./features/equipment/AssetRegister'));
const Benefits = React.lazy(() => import('./features/benefits/Benefits'));
const Training = React.lazy(() => import('./features/training/Training'));
const OrgChart = React.lazy(() => import('./features/departments/OrgChart'));
const Operations = React.lazy(() => import('./features/operations/Operations'));
const LeaveScreen = React.lazy(() => import('./features/leave/Leave'));
const LeaveRequestFormScreen = React.lazy(() => import('./features/leave/LeaveRequestForm'));
const OvertimeScreen = React.lazy(() => import('./features/overtime/Overtime'));
const RosterBuilder = React.lazy(() => import('./features/shifts/RosterBuilder'));
const Performance = React.lazy(() => import('./features/performance/Performance'));
const ReviewCycleDetail = React.lazy(() => import('./features/performance/ReviewCycleDetail'));
const MyReview = React.lazy(() => import('./features/performance/MyReview'));
const GoalsList = React.lazy(() => import('./features/performance/GoalsList'));
const Reports = React.lazy(() => import('./features/reports/Reports'));

const wrap = (screen: React.ReactNode, permission?: HRPermission): React.JSX.Element => (
  <>
    {permission ? <HRPermissionGuard permission={permission}>{screen}</HRPermissionGuard> : screen}
  </>
);

const payrollWrap = (screen: React.ReactNode): React.JSX.Element =>
  wrap(<PayrollConfigGuard>{screen}</PayrollConfigGuard>, 'hr:payroll:view');

// `/hr/payroll/my-runs*` already resolves the caller's own employee id from
// the session and 404s on a run that isn't theirs (see the comment on
// `useListMyPayrollRuns` in `usePayroll.ts`), so there is nothing left for a
// role permission to enforce here. Gating this behind `hr:payroll` (an
// admin-tier permission `staff` doesn't have) blocked employees from their
// own pay slips for no security reason - `AppGuard` already confirms the
// user has HR app access before this route tree even mounts.
const myPayslipsWrap = (screen: React.ReactNode): React.JSX.Element =>
  wrap(<PayrollConfigGuard>{screen}</PayrollConfigGuard>);

const configWrap = (
  screen: React.ReactNode,
  enabled: (config: FeatureConfig) => boolean,
  permission?: HRPermission
): React.JSX.Element =>
  wrap(<HRFeatureConfigGuard enabled={enabled}>{screen}</HRFeatureConfigGuard>, permission);

export function createHrRoutes<TParent extends AnyRoute>(parent: TParent) {
  const rootRoute = createRoute({
    getParentRoute: () => parent,
    path: 'hr',
    component: () => (
      <ErrorBoundary>
        <Suspense fallback={<PageLoader />}>
          <Outlet />
        </Suspense>
      </ErrorBoundary>
    ),
  });
  const setupRoute = createRoute({
    getParentRoute: () => rootRoute,
    path: '/setup',
    component: () => (
      <HRPermissionGuard permission="hub:business:settings:manage">
        <WizardLayout>
          <Suspense fallback={<PageLoader />}>
            <SetupWizard />
          </Suspense>
        </WizardLayout>
      </HRPermissionGuard>
    ),
  });

  const hrConfigRoute = createRoute({
    getParentRoute: () => rootRoute,
    id: 'hr-config',
    component: HRConfigRoute,
  });

  const dashboardRoute = createRoute({
    getParentRoute: () => hrConfigRoute,
    path: '/',
    component: () =>
      wrap(
        <Suspense fallback={<PageLoader />}>
          <Dashboard />
        </Suspense>
      ),
  });

  const employeesRoute = createRoute({
    getParentRoute: () => hrConfigRoute,
    path: '/employees',
    component: () =>
      wrap(
        <Suspense fallback={<PageLoader />}>
          <Employees />
        </Suspense>,
        'hr:employees:view'
      ),
  });

  const newEmployeeRoute = createRoute({
    getParentRoute: () => hrConfigRoute,
    path: '/employees/new',
    component: () =>
      wrap(
        <Suspense fallback={<PageLoader />}>
          <NewEmployee />
        </Suspense>,
        'hr:employees:manage'
      ),
  });

  const employeeDetailRoute = createRoute({
    getParentRoute: () => hrConfigRoute,
    path: '/employees/$id',
    component: () =>
      wrap(
        <Suspense fallback={<PageLoader />}>
          <EmployeeDetail />
        </Suspense>,
        'hr:employees:view'
      ),
  });

  const staffRoute = createRoute({
    getParentRoute: () => hrConfigRoute,
    path: '/staff',
    component: () =>
      wrap(
        <Suspense fallback={<PageLoader />}>
          <Staff />
        </Suspense>,
        'hr:employees:view'
      ),
  });

  // Gated on hr:staff:manage rather than hr:staff: setting a reporting line widens
  // what that manager can see in Flow, so it is an administrative act, not a read.
  const reportingLinesRoute = createRoute({
    getParentRoute: () => hrConfigRoute,
    path: '/staff/reporting-lines',
    component: () =>
      wrap(
        <Suspense fallback={<PageLoader />}>
          <ReportingLines />
        </Suspense>,
        'hr:employees:manage'
      ),
  });

  const staffProfileRoute = createRoute({
    getParentRoute: () => hrConfigRoute,
    path: '/staff/$id',
    component: () =>
      wrap(
        <Suspense fallback={<PageLoader />}>
          <EmployeeProfile />
        </Suspense>,
        'hr:employees:view'
      ),
  });

  const shiftsRoute = createRoute({
    getParentRoute: () => hrConfigRoute,
    path: '/shifts',
    component: () =>
      wrap(
        <Suspense fallback={<PageLoader />}>
          <Shifts />
        </Suspense>,
        'hr:schedules:manage'
      ),
  });

  const scheduleRoute = createRoute({
    getParentRoute: () => hrConfigRoute,
    path: '/schedule',
    component: () =>
      wrap(
        <Suspense fallback={<PageLoader />}>
          <Schedule />
        </Suspense>,
        'hr:schedules:manage'
      ),
  });

  const timeRoute = createRoute({
    getParentRoute: () => hrConfigRoute,
    path: '/time',
    component: () =>
      configWrap(
        <Suspense fallback={<PageLoader />}>
          <TimeLogging />
        </Suspense>,
        (config) =>
          config.timeTrackingModel === 'time_logging' || config.timeTrackingModel === 'both',
        'hr:schedules:manage'
      ),
  });

  const timesheetsRoute = createRoute({
    getParentRoute: () => hrConfigRoute,
    path: '/timesheets',
    component: () =>
      wrap(
        <Suspense fallback={<PageLoader />}>
          <TimesheetsList />
        </Suspense>,
        'hr:schedules:manage'
      ),
  });

  const timesheetApprovalsRoute = createRoute({
    getParentRoute: () => hrConfigRoute,
    path: '/timesheets/approvals',
    component: () =>
      wrap(
        <Suspense fallback={<PageLoader />}>
          <TimesheetApprovals />
        </Suspense>,
        'hr:time_off:approve'
      ),
  });

  const timesheetDetailRoute = createRoute({
    getParentRoute: () => hrConfigRoute,
    path: '/timesheets/$id',
    component: () =>
      wrap(
        <Suspense fallback={<PageLoader />}>
          <TimesheetDetail />
        </Suspense>,
        'hr:schedules:manage'
      ),
  });

  const attendanceRoute = createRoute({
    getParentRoute: () => hrConfigRoute,
    path: '/attendance',
    component: () =>
      wrap(
        <Suspense fallback={<PageLoader />}>
          <AttendanceScreen />
        </Suspense>,
        'hr:schedules:manage'
      ),
  });

  const timeOffRoute = createRoute({
    getParentRoute: () => hrConfigRoute,
    path: '/time-off',
    component: () =>
      wrap(
        <Suspense fallback={<PageLoader />}>
          <TimeOff />
        </Suspense>,
        'hr:time_off:request'
      ),
  });

  const timeOffNewRoute = createRoute({
    getParentRoute: () => hrConfigRoute,
    path: '/time-off/new',
    component: () =>
      wrap(
        <Suspense fallback={<PageLoader />}>
          <TimeOffNew />
        </Suspense>
      ),
  });

  const calendarRoute = createRoute({
    getParentRoute: () => hrConfigRoute,
    path: '/calendar',
    component: () =>
      wrap(
        <Suspense fallback={<PageLoader />}>
          <CalendarScreen />
        </Suspense>,
        'hr:time_off:request'
      ),
  });

  const payrollRoute = createRoute({
    getParentRoute: () => hrConfigRoute,
    path: '/payroll',
    component: () =>
      payrollWrap(
        <Suspense fallback={<PageLoader />}>
          <Payroll />
        </Suspense>
      ),
  });

  const myPaySlipsRoute = createRoute({
    getParentRoute: () => hrConfigRoute,
    path: '/payroll/my-payslips',
    component: () =>
      myPayslipsWrap(
        <Suspense fallback={<PageLoader />}>
          <MyPaySlips />
        </Suspense>
      ),
  });

  const paySlipViewRoute = createRoute({
    getParentRoute: () => hrConfigRoute,
    path: '/payroll/my-payslips/view',
    validateSearch: paySlipSearchSchema,
    component: () =>
      myPayslipsWrap(
        <Suspense fallback={<PageLoader />}>
          <PaySlip />
        </Suspense>
      ),
  });

  const payrollRunDetailRoute = createRoute({
    getParentRoute: () => hrConfigRoute,
    path: '/payroll/$id',
    params: {
      parse: (raw) => payrollRunParamsSchema.parse(raw),
      stringify: ({ id }) => ({ id: String(id) }),
    },
    component: () =>
      payrollWrap(
        <Suspense fallback={<PageLoader />}>
          <PayrollRunDetail />
        </Suspense>
      ),
  });

  const payrollConfigRoute = createRoute({
    getParentRoute: () => hrConfigRoute,
    path: '/settings/payroll-config',
    component: () => <Navigate to="/settings/hr/payroll-config" replace />,
  });

  const departmentsRoute = createRoute({
    getParentRoute: () => hrConfigRoute,
    path: '/departments',
    component: () =>
      wrap(
        <Suspense fallback={<PageLoader />}>
          <Departments />
        </Suspense>,
        'hr:employees:view'
      ),
  });

  const reviewsRoute = createRoute({
    getParentRoute: () => hrConfigRoute,
    path: '/reviews',
    component: () =>
      wrap(
        <Suspense fallback={<PageLoader />}>
          <Reviews />
        </Suspense>,
        'hr:employees:view'
      ),
  });

  const reviewsNewRoute = createRoute({
    getParentRoute: () => hrConfigRoute,
    path: '/reviews/new',
    component: () =>
      wrap(
        <Suspense fallback={<PageLoader />}>
          <ReviewForm />
        </Suspense>,
        'hr:employees:view'
      ),
  });

  const oneOnOnesRoute = createRoute({
    getParentRoute: () => hrConfigRoute,
    path: '/one-on-ones',
    component: () =>
      wrap(
        <Suspense fallback={<PageLoader />}>
          <OneOnOnes />
        </Suspense>
      ),
  });

  const onboardingRoute = createRoute({
    getParentRoute: () => hrConfigRoute,
    path: '/onboarding',
    component: () =>
      wrap(
        <Suspense fallback={<PageLoader />}>
          <OnboardingManagement />
        </Suspense>,
        'hr:employees:manage'
      ),
  });

  const onboardingTemplatesRoute = createRoute({
    getParentRoute: () => hrConfigRoute,
    path: '/onboarding/templates',
    component: () =>
      wrap(
        <Suspense fallback={<PageLoader />}>
          <ChecklistTemplates />
        </Suspense>,
        'hr:employees:manage'
      ),
  });

  const offboardingRoute = createRoute({
    getParentRoute: () => hrConfigRoute,
    path: '/offboarding',
    component: () =>
      wrap(
        <Suspense fallback={<PageLoader />}>
          <Offboarding />
        </Suspense>,
        'hr:employees:manage'
      ),
  });

  const documentsRoute = createRoute({
    getParentRoute: () => hrConfigRoute,
    path: '/documents',
    component: () =>
      wrap(
        <Suspense fallback={<PageLoader />}>
          <Documents />
        </Suspense>
      ),
  });

  const equipmentRoute = createRoute({
    getParentRoute: () => hrConfigRoute,
    path: '/equipment',
    component: () =>
      configWrap(
        <Suspense fallback={<PageLoader />}>
          <EquipmentScreen />
        </Suspense>,
        (config) => isEnabled(config.equipmentEnabled)
      ),
  });

  const assetRegisterRoute = createRoute({
    getParentRoute: () => hrConfigRoute,
    path: '/equipment/assets',
    component: () =>
      configWrap(
        <Suspense fallback={<PageLoader />}>
          <AssetRegister />
        </Suspense>,
        (config) => isEnabled(config.equipmentEnabled)
      ),
  });

  const benefitsRoute = createRoute({
    getParentRoute: () => hrConfigRoute,
    path: '/benefits',
    component: () =>
      wrap(
        <Suspense fallback={<PageLoader />}>
          <Benefits />
        </Suspense>
      ),
  });

  const trainingRoute = createRoute({
    getParentRoute: () => hrConfigRoute,
    path: '/training',
    component: () =>
      wrap(
        <Suspense fallback={<PageLoader />}>
          <Training />
        </Suspense>,
        'hr:employees:manage'
      ),
  });

  const orgChartRoute = createRoute({
    getParentRoute: () => hrConfigRoute,
    path: '/org-chart',
    component: () =>
      wrap(
        <Suspense fallback={<PageLoader />}>
          <OrgChart />
        </Suspense>,
        'hr:employees:view'
      ),
  });

  const leaveRoute = createRoute({
    getParentRoute: () => hrConfigRoute,
    path: '/leave',
    component: () =>
      wrap(
        <Suspense fallback={<PageLoader />}>
          <LeaveScreen />
        </Suspense>,
        'hr:time_off:request'
      ),
  });

  const leaveNewRoute = createRoute({
    getParentRoute: () => hrConfigRoute,
    path: '/leave/new',
    component: () =>
      wrap(
        <Suspense fallback={<PageLoader />}>
          <LeaveRequestFormScreen />
        </Suspense>,
        'hr:time_off:request'
      ),
  });

  const overtimeRoute = createRoute({
    getParentRoute: () => hrConfigRoute,
    path: '/overtime',
    component: () =>
      wrap(
        <Suspense fallback={<PageLoader />}>
          <OvertimeScreen />
        </Suspense>,
        'hr:time_off:request'
      ),
  });

  const rosterRoute = createRoute({
    getParentRoute: () => hrConfigRoute,
    path: '/shifts/roster',
    component: () =>
      wrap(
        <Suspense fallback={<PageLoader />}>
          <RosterBuilder />
        </Suspense>,
        'hr:schedules:manage'
      ),
  });

  const performanceRoute = createRoute({
    getParentRoute: () => hrConfigRoute,
    path: '/performance',
    component: () =>
      wrap(
        <Suspense fallback={<PageLoader />}>
          <Performance />
        </Suspense>,
        'hr:employees:view'
      ),
  });

  const reviewCycleDetailRoute = createRoute({
    getParentRoute: () => hrConfigRoute,
    path: '/performance/cycles/$id',
    component: () =>
      wrap(
        <Suspense fallback={<PageLoader />}>
          <ReviewCycleDetail />
        </Suspense>,
        'hr:employees:view'
      ),
  });

  const myReviewRoute = createRoute({
    getParentRoute: () => hrConfigRoute,
    path: '/performance/my-review',
    component: () =>
      wrap(
        <Suspense fallback={<PageLoader />}>
          <MyReview />
        </Suspense>
      ),
  });

  const goalsRoute = createRoute({
    getParentRoute: () => hrConfigRoute,
    path: '/performance/goals',
    component: () =>
      wrap(
        <Suspense fallback={<PageLoader />}>
          <GoalsList />
        </Suspense>
      ),
  });

  const reportsRoute = createRoute({
    getParentRoute: () => hrConfigRoute,
    path: '/reports',
    component: () =>
      wrap(
        <Suspense fallback={<PageLoader />}>
          <Reports />
        </Suspense>,
        'hr:employees:view'
      ),
  });

  const operationsRoute = createRoute({
    getParentRoute: () => hrConfigRoute,
    path: '/operations',
    component: () =>
      wrap(
        <Suspense fallback={<PageLoader />}>
          <Operations />
        </Suspense>,
        'hr:schedules:manage'
      ),
  });

  const settingsRoute = createRoute({
    getParentRoute: () => hrConfigRoute,
    path: '/settings',
    component: () => <Navigate to="/settings/hr" replace />,
  });

  const rolesPermissionsSettingsRoute = createRoute({
    getParentRoute: () => hrConfigRoute,
    path: '/settings/roles-permissions',
    component: () => <Navigate to="/roles" replace />,
  });

  const leavePolicySettingsRoute = createRoute({
    getParentRoute: () => hrConfigRoute,
    path: '/settings/leave-policy',
    component: () => <Navigate to="/settings/hr/leave-policy" replace />,
  });

  const catchAllRoute = createRoute({
    getParentRoute: () => rootRoute,
    path: '$',
    component: () => <Navigate to="/hr" replace />,
  });

  return rootRoute.addChildren([
    setupRoute,
    hrConfigRoute.addChildren([
      dashboardRoute,
      employeesRoute,
      newEmployeeRoute,
      employeeDetailRoute,
      staffRoute,
      reportingLinesRoute,
      staffProfileRoute,
      shiftsRoute,
      scheduleRoute,
      timeRoute,
      timesheetsRoute,
      timesheetApprovalsRoute,
      timesheetDetailRoute,
      attendanceRoute,
      timeOffRoute,
      timeOffNewRoute,
      calendarRoute,
      payrollRoute,
      myPaySlipsRoute,
      paySlipViewRoute,
      payrollRunDetailRoute,
      payrollConfigRoute,
      departmentsRoute,
      reviewsRoute,
      reviewsNewRoute,
      oneOnOnesRoute,
      onboardingRoute,
      onboardingTemplatesRoute,
      offboardingRoute,
      documentsRoute,
      equipmentRoute,
      assetRegisterRoute,
      benefitsRoute,
      trainingRoute,
      orgChartRoute,
      leaveRoute,
      leaveNewRoute,
      overtimeRoute,
      rosterRoute,
      performanceRoute,
      reviewCycleDetailRoute,
      myReviewRoute,
      goalsRoute,
      reportsRoute,
      operationsRoute,
      settingsRoute,
      rolesPermissionsSettingsRoute,
      leavePolicySettingsRoute,
    ]),
    catchAllRoute,
  ]);
}
