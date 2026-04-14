import React, { Suspense } from 'react';
import { Routes, Route, Navigate } from 'react-router-dom';
import { ProtectedRoute } from '@voxire/auth';
import ErrorBoundary from './shared/components/ErrorBoundary';
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
const Onboarding = React.lazy(() => import('./features/onboarding/Onboarding'));
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

function wrap(screen: React.ReactNode): React.JSX.Element {
  return <DefaultLayout>{screen}</DefaultLayout>;
}

const App = (): React.JSX.Element => {
  return (
    <ErrorBoundary>
      <ProtectedRoute appId="hr">
        <Suspense fallback={<PageLoader />}>
          <Routes>
            {/* Setup wizard — outside HRConfigRoute */}
            <Route
              path="/setup"
              element={
                <WizardLayout>
                  <SetupWizard />
                </WizardLayout>
              }
            />

            {/* All feature routes — inside HRConfigRoute */}
            <Route element={<HRConfigRoute />}>
              <Route path="/" element={wrap(<Dashboard />)} />
              <Route path="/employees" element={wrap(<Employees />)} />
              <Route path="/employees/new" element={wrap(<NewEmployee />)} />
              <Route path="/employees/:id" element={wrap(<EmployeeDetail />)} />
              <Route path="/staff" element={wrap(<Staff />)} />
              <Route path="/staff/:id" element={wrap(<EmployeeProfile />)} />
              <Route path="/shifts" element={wrap(<Shifts />)} />
              <Route path="/schedule" element={wrap(<Schedule />)} />
              <Route path="/time" element={wrap(<TimeLogging />)} />
              <Route path="/timesheets" element={wrap(<TimesheetsList />)} />
              <Route path="/timesheets/approvals" element={wrap(<TimesheetApprovals />)} />
              <Route path="/timesheets/:id" element={wrap(<TimesheetDetail />)} />
              <Route path="/attendance" element={wrap(<AttendanceScreen />)} />
              <Route path="/time-off" element={wrap(<TimeOff />)} />
              <Route path="/time-off/new" element={wrap(<TimeOffNew />)} />
              <Route path="/calendar" element={wrap(<CalendarScreen />)} />
              <Route path="/payroll" element={wrap(<Payroll />)} />
              <Route path="/payroll/my-payslips" element={wrap(<MyPaySlips />)} />
              <Route path="/payroll/my-payslips/view" element={wrap(<PaySlip />)} />
              <Route path="/payroll/:id" element={wrap(<PayrollRunDetail />)} />
              <Route path="/settings/payroll-config" element={wrap(<PayrollConfigSettings />)} />
              <Route path="/departments" element={wrap(<Departments />)} />
              <Route path="/reviews" element={wrap(<Reviews />)} />
              <Route path="/reviews/new" element={wrap(<ReviewForm />)} />
              <Route path="/one-on-ones" element={wrap(<OneOnOnes />)} />
              <Route path="/onboarding" element={wrap(<OnboardingManagement />)} />
              <Route path="/onboarding/templates" element={wrap(<ChecklistTemplates />)} />
              <Route path="/offboarding" element={wrap(<Offboarding />)} />
              <Route path="/documents" element={wrap(<Documents />)} />
              <Route path="/equipment" element={wrap(<EquipmentScreen />)} />
              <Route path="/equipment/assets" element={wrap(<AssetRegister />)} />
              <Route path="/benefits" element={wrap(<Benefits />)} />
              <Route path="/training" element={wrap(<Training />)} />
              <Route path="/org-chart" element={wrap(<OrgChart />)} />
              <Route path="/leave" element={wrap(<LeaveScreen />)} />
              <Route path="/leave/new" element={wrap(<LeaveRequestFormScreen />)} />
              <Route path="/overtime" element={wrap(<OvertimeScreen />)} />
              <Route path="/shifts/roster" element={wrap(<RosterBuilder />)} />
              <Route path="/operations" element={wrap(<Operations />)} />
              <Route path="/settings" element={wrap(<SettingsScreen />)} />
            </Route>

            <Route path="*" element={<Navigate to="/" replace />} />
          </Routes>
        </Suspense>
      </ProtectedRoute>
    </ErrorBoundary>
  );
};

export default App;
