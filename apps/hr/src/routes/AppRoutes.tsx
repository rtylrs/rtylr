import React from 'react';
import { Routes, Route, Navigate } from 'react-router-dom';
import DefaultLayout from '../layouts/DefaultLayout/DefaultLayout';

const Dashboard        = React.lazy(() => import('../screens/Dashboard/Dashboard'));
const Staff            = React.lazy(() => import('../screens/Staff/Staff'));
const Shifts           = React.lazy(() => import('../screens/Shifts/Shifts'));
const Payroll          = React.lazy(() => import('../screens/Payroll/Payroll'));
const Departments      = React.lazy(() => import('../screens/Departments/Departments'));
const EmployeeProfile  = React.lazy(() => import('../screens/EmployeeProfile/EmployeeProfile'));
const TimeOff          = React.lazy(() => import('../screens/TimeOff/TimeOff'));
const TimeOffNew       = React.lazy(() => import('../screens/TimeOffNew/TimeOffNew'));
const CalendarScreen   = React.lazy(() => import('../screens/Calendar/Calendar'));
const Schedule         = React.lazy(() => import('../screens/Schedule/Schedule'));
const Timesheets       = React.lazy(() => import('../screens/Timesheets/Timesheets'));
const Reviews          = React.lazy(() => import('../screens/Reviews/Reviews'));
const ReviewForm       = React.lazy(() => import('../screens/ReviewForm/ReviewForm'));
const OneOnOnes        = React.lazy(() => import('../screens/OneOnOnes/OneOnOnes'));
const Onboarding       = React.lazy(() => import('../screens/Onboarding/Onboarding'));
const Offboarding      = React.lazy(() => import('../screens/Offboarding/Offboarding'));
const Documents        = React.lazy(() => import('../screens/Documents/Documents'));
const Benefits         = React.lazy(() => import('../screens/Benefits/Benefits'));
const Training         = React.lazy(() => import('../screens/Training/Training'));
const OrgChart         = React.lazy(() => import('../screens/OrgChart/OrgChart'));
const SettingsScreen   = React.lazy(() => import('../screens/Settings/Settings'));
const Operations       = React.lazy(() => import('../screens/Operations/Operations'));

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
