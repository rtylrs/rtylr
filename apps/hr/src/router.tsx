import React, { Suspense } from 'react';
import {
  createRootRoute,
  createRoute,
  createRouter,
  Outlet,
  Navigate,
} from '@tanstack/react-router';
import { AppGuard } from '@voxire/auth';
import { unwrap } from '@voxire/api';
import { safe } from './shared/lib/api';
import { ErrorBoundary } from '@voxire/ui';
import { paySlipSearchSchema } from './shared/schemas';
import HRConfigRoute from './features/setup-wizard/HRConfigRoute';
import DefaultLayout from './layouts/DefaultLayout';
import WizardLayout from './layouts/WizardLayout';

function PageLoader(): React.JSX.Element {
  return (
    <div
      className="fixed inset-0 z-50 flex items-center justify-center"
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
const PayrollConfigSettings = React.lazy(
  () => import('./features/settings/payroll-config/PayrollConfigSettings')
);
const Departments = React.lazy(() => import('./features/departments/Departments'));
const EmployeeProfile = React.lazy(() => import('./features/staff/EmployeeProfile'));
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
const SettingsScreen = React.lazy(() => import('./features/settings/Settings'));
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

const wrap = (screen: React.ReactNode): React.JSX.Element => (
  <DefaultLayout>{screen}</DefaultLayout>
);

const rootRoute = createRootRoute({
  component: () => (
    <ErrorBoundary>
      <AppGuard
        appId="hr"
        fetchUser={(): Promise<Record<string, unknown>> =>
          unwrap(safe.get<Record<string, unknown>>('/hub/user/v1/users/me'))
        }
      >
        <Suspense fallback={<PageLoader />}>
          <Outlet />
        </Suspense>
      </AppGuard>
    </ErrorBoundary>
  ),
});

const setupRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: '/setup',
  component: () => (
    <WizardLayout>
      <Suspense fallback={<PageLoader />}>
        <SetupWizard />
      </Suspense>
    </WizardLayout>
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
      </Suspense>
    ),
});

const newEmployeeRoute = createRoute({
  getParentRoute: () => hrConfigRoute,
  path: '/employees/new',
  component: () =>
    wrap(
      <Suspense fallback={<PageLoader />}>
        <NewEmployee />
      </Suspense>
    ),
});

const employeeDetailRoute = createRoute({
  getParentRoute: () => hrConfigRoute,
  path: '/employees/$id',
  component: () =>
    wrap(
      <Suspense fallback={<PageLoader />}>
        <EmployeeDetail />
      </Suspense>
    ),
});

const staffRoute = createRoute({
  getParentRoute: () => hrConfigRoute,
  path: '/staff',
  component: () =>
    wrap(
      <Suspense fallback={<PageLoader />}>
        <Staff />
      </Suspense>
    ),
});

const staffProfileRoute = createRoute({
  getParentRoute: () => hrConfigRoute,
  path: '/staff/$id',
  component: () =>
    wrap(
      <Suspense fallback={<PageLoader />}>
        <EmployeeProfile />
      </Suspense>
    ),
});

const shiftsRoute = createRoute({
  getParentRoute: () => hrConfigRoute,
  path: '/shifts',
  component: () =>
    wrap(
      <Suspense fallback={<PageLoader />}>
        <Shifts />
      </Suspense>
    ),
});

const scheduleRoute = createRoute({
  getParentRoute: () => hrConfigRoute,
  path: '/schedule',
  component: () =>
    wrap(
      <Suspense fallback={<PageLoader />}>
        <Schedule />
      </Suspense>
    ),
});

const timeRoute = createRoute({
  getParentRoute: () => hrConfigRoute,
  path: '/time',
  component: () =>
    wrap(
      <Suspense fallback={<PageLoader />}>
        <TimeLogging />
      </Suspense>
    ),
});

const timesheetsRoute = createRoute({
  getParentRoute: () => hrConfigRoute,
  path: '/timesheets',
  component: () =>
    wrap(
      <Suspense fallback={<PageLoader />}>
        <TimesheetsList />
      </Suspense>
    ),
});

const timesheetApprovalsRoute = createRoute({
  getParentRoute: () => hrConfigRoute,
  path: '/timesheets/approvals',
  component: () =>
    wrap(
      <Suspense fallback={<PageLoader />}>
        <TimesheetApprovals />
      </Suspense>
    ),
});

const timesheetDetailRoute = createRoute({
  getParentRoute: () => hrConfigRoute,
  path: '/timesheets/$id',
  component: () =>
    wrap(
      <Suspense fallback={<PageLoader />}>
        <TimesheetDetail />
      </Suspense>
    ),
});

const attendanceRoute = createRoute({
  getParentRoute: () => hrConfigRoute,
  path: '/attendance',
  component: () =>
    wrap(
      <Suspense fallback={<PageLoader />}>
        <AttendanceScreen />
      </Suspense>
    ),
});

const timeOffRoute = createRoute({
  getParentRoute: () => hrConfigRoute,
  path: '/time-off',
  component: () =>
    wrap(
      <Suspense fallback={<PageLoader />}>
        <TimeOff />
      </Suspense>
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
      </Suspense>
    ),
});

const payrollRoute = createRoute({
  getParentRoute: () => hrConfigRoute,
  path: '/payroll',
  component: () =>
    wrap(
      <Suspense fallback={<PageLoader />}>
        <Payroll />
      </Suspense>
    ),
});

const myPaySlipsRoute = createRoute({
  getParentRoute: () => hrConfigRoute,
  path: '/payroll/my-payslips',
  component: () =>
    wrap(
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
    wrap(
      <Suspense fallback={<PageLoader />}>
        <PaySlip />
      </Suspense>
    ),
});

const payrollRunDetailRoute = createRoute({
  getParentRoute: () => hrConfigRoute,
  path: '/payroll/$id',
  component: () =>
    wrap(
      <Suspense fallback={<PageLoader />}>
        <PayrollRunDetail />
      </Suspense>
    ),
});

const payrollConfigRoute = createRoute({
  getParentRoute: () => hrConfigRoute,
  path: '/settings/payroll-config',
  component: () =>
    wrap(
      <Suspense fallback={<PageLoader />}>
        <PayrollConfigSettings />
      </Suspense>
    ),
});

const departmentsRoute = createRoute({
  getParentRoute: () => hrConfigRoute,
  path: '/departments',
  component: () =>
    wrap(
      <Suspense fallback={<PageLoader />}>
        <Departments />
      </Suspense>
    ),
});

const reviewsRoute = createRoute({
  getParentRoute: () => hrConfigRoute,
  path: '/reviews',
  component: () =>
    wrap(
      <Suspense fallback={<PageLoader />}>
        <Reviews />
      </Suspense>
    ),
});

const reviewsNewRoute = createRoute({
  getParentRoute: () => hrConfigRoute,
  path: '/reviews/new',
  component: () =>
    wrap(
      <Suspense fallback={<PageLoader />}>
        <ReviewForm />
      </Suspense>
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
      </Suspense>
    ),
});

const onboardingTemplatesRoute = createRoute({
  getParentRoute: () => hrConfigRoute,
  path: '/onboarding/templates',
  component: () =>
    wrap(
      <Suspense fallback={<PageLoader />}>
        <ChecklistTemplates />
      </Suspense>
    ),
});

const offboardingRoute = createRoute({
  getParentRoute: () => hrConfigRoute,
  path: '/offboarding',
  component: () =>
    wrap(
      <Suspense fallback={<PageLoader />}>
        <Offboarding />
      </Suspense>
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
    wrap(
      <Suspense fallback={<PageLoader />}>
        <EquipmentScreen />
      </Suspense>
    ),
});

const assetRegisterRoute = createRoute({
  getParentRoute: () => hrConfigRoute,
  path: '/equipment/assets',
  component: () =>
    wrap(
      <Suspense fallback={<PageLoader />}>
        <AssetRegister />
      </Suspense>
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
      </Suspense>
    ),
});

const orgChartRoute = createRoute({
  getParentRoute: () => hrConfigRoute,
  path: '/org-chart',
  component: () =>
    wrap(
      <Suspense fallback={<PageLoader />}>
        <OrgChart />
      </Suspense>
    ),
});

const leaveRoute = createRoute({
  getParentRoute: () => hrConfigRoute,
  path: '/leave',
  component: () =>
    wrap(
      <Suspense fallback={<PageLoader />}>
        <LeaveScreen />
      </Suspense>
    ),
});

const leaveNewRoute = createRoute({
  getParentRoute: () => hrConfigRoute,
  path: '/leave/new',
  component: () =>
    wrap(
      <Suspense fallback={<PageLoader />}>
        <LeaveRequestFormScreen />
      </Suspense>
    ),
});

const overtimeRoute = createRoute({
  getParentRoute: () => hrConfigRoute,
  path: '/overtime',
  component: () =>
    wrap(
      <Suspense fallback={<PageLoader />}>
        <OvertimeScreen />
      </Suspense>
    ),
});

const rosterRoute = createRoute({
  getParentRoute: () => hrConfigRoute,
  path: '/shifts/roster',
  component: () =>
    wrap(
      <Suspense fallback={<PageLoader />}>
        <RosterBuilder />
      </Suspense>
    ),
});

const performanceRoute = createRoute({
  getParentRoute: () => hrConfigRoute,
  path: '/performance',
  component: () =>
    wrap(
      <Suspense fallback={<PageLoader />}>
        <Performance />
      </Suspense>
    ),
});

const reviewCycleDetailRoute = createRoute({
  getParentRoute: () => hrConfigRoute,
  path: '/performance/cycles/$id',
  component: () =>
    wrap(
      <Suspense fallback={<PageLoader />}>
        <ReviewCycleDetail />
      </Suspense>
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
      </Suspense>
    ),
});

const operationsRoute = createRoute({
  getParentRoute: () => hrConfigRoute,
  path: '/operations',
  component: () =>
    wrap(
      <Suspense fallback={<PageLoader />}>
        <Operations />
      </Suspense>
    ),
});

const settingsRoute = createRoute({
  getParentRoute: () => hrConfigRoute,
  path: '/settings',
  component: () =>
    wrap(
      <Suspense fallback={<PageLoader />}>
        <SettingsScreen />
      </Suspense>
    ),
});

const catchAllRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: '*',
  component: () => <Navigate to="/" replace />,
});

const routeTree = rootRoute.addChildren([
  setupRoute,
  hrConfigRoute.addChildren([
    dashboardRoute,
    employeesRoute,
    newEmployeeRoute,
    employeeDetailRoute,
    staffRoute,
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
  ]),
  catchAllRoute,
]);

export const router = createRouter({ routeTree });

declare module '@tanstack/react-router' {
  interface Register {
    router: typeof router;
  }
}
