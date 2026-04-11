import React from 'react';
import { Routes, Route, Navigate } from 'react-router-dom';
import DefaultLayout from '../layouts/DefaultLayout';

const Dashboard        = React.lazy(() => import('../features/dashboard/Dashboard'));
const Staff            = React.lazy(() => import('../features/staff/Staff'));
const Shifts           = React.lazy(() => import('../features/shifts/Shifts'));
const Payroll          = React.lazy(() => import('../features/payroll/Payroll'));
const Departments      = React.lazy(() => import('../features/departments/Departments'));
const EmployeeProfile  = React.lazy(() => import('../features/staff/EmployeeProfile'));
const TimeOff          = React.lazy(() => import('../features/time-off/TimeOff'));
const TimeOffNew       = React.lazy(() => import('../features/time-off/TimeOffNew'));
const CalendarScreen   = React.lazy(() => import('../features/calendar/Calendar'));
const Schedule         = React.lazy(() => import('../features/shifts/Schedule'));
const Timesheets       = React.lazy(() => import('../features/shifts/Timesheets'));
const Reviews          = React.lazy(() => import('../features/reviews/Reviews'));
const ReviewForm       = React.lazy(() => import('../features/reviews/ReviewForm'));
const OneOnOnes        = React.lazy(() => import('../features/reviews/OneOnOnes'));
const Onboarding       = React.lazy(() => import('../features/onboarding/Onboarding'));
const Offboarding      = React.lazy(() => import('../features/onboarding/Offboarding'));
const Documents        = React.lazy(() => import('../features/documents/Documents'));
const Benefits         = React.lazy(() => import('../features/benefits/Benefits'));
const Training         = React.lazy(() => import('../features/training/Training'));
const OrgChart         = React.lazy(() => import('../features/departments/OrgChart'));
const SettingsScreen   = React.lazy(() => import('../features/settings/Settings'));
const Operations       = React.lazy(() => import('../features/operations/Operations'));

function wrap(screen: React.ReactNode) {
  return <DefaultLayout>{screen}</DefaultLayout>;
}

const AppRoutes: React.FC = () => (
  <Routes>
    <Route path="/"              element={wrap(<Dashboard />)} />
    <Route path="/staff"         element={wrap(<Staff />)} />
    <Route path="/staff/:id"     element={wrap(<EmployeeProfile />)} />
    <Route path="/shifts"        element={wrap(<Shifts />)} />
    <Route path="/schedule"      element={wrap(<Schedule />)} />
    <Route path="/timesheets"    element={wrap(<Timesheets />)} />
    <Route path="/time-off"      element={wrap(<TimeOff />)} />
    <Route path="/time-off/new"  element={wrap(<TimeOffNew />)} />
    <Route path="/calendar"      element={wrap(<CalendarScreen />)} />
    <Route path="/payroll"       element={wrap(<Payroll />)} />
    <Route path="/departments"   element={wrap(<Departments />)} />
    <Route path="/reviews"       element={wrap(<Reviews />)} />
    <Route path="/reviews/new"   element={wrap(<ReviewForm />)} />
    <Route path="/one-on-ones"   element={wrap(<OneOnOnes />)} />
    <Route path="/onboarding"    element={wrap(<Onboarding />)} />
    <Route path="/offboarding"   element={wrap(<Offboarding />)} />
    <Route path="/documents"     element={wrap(<Documents />)} />
    <Route path="/benefits"      element={wrap(<Benefits />)} />
    <Route path="/training"      element={wrap(<Training />)} />
    <Route path="/org-chart"     element={wrap(<OrgChart />)} />
    <Route path="/operations"    element={wrap(<Operations />)} />
    <Route path="/settings"      element={wrap(<SettingsScreen />)} />
    <Route path="*"              element={<Navigate to="/" replace />} />
  </Routes>
);

export default AppRoutes;
