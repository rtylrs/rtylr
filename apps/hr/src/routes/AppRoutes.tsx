import React from 'react';
import { Routes, Route, Navigate } from 'react-router-dom';
import DefaultLayout from '../layouts/DefaultLayout/DefaultLayout';

const Dashboard   = React.lazy(() => import('../screens/Dashboard/Dashboard'));
const Staff       = React.lazy(() => import('../screens/Staff/Staff'));
const Shifts      = React.lazy(() => import('../screens/Shifts/Shifts'));
const Payroll     = React.lazy(() => import('../screens/Payroll/Payroll'));
const Departments = React.lazy(() => import('../screens/Departments/Departments'));

function wrap(screen: React.ReactNode) {
  return <DefaultLayout>{screen}</DefaultLayout>;
}

const AppRoutes: React.FC = () => (
  <Routes>
    <Route path="/"            element={wrap(<Dashboard />)} />
    <Route path="/staff"       element={wrap(<Staff />)} />
    <Route path="/shifts"      element={wrap(<Shifts />)} />
    <Route path="/payroll"     element={wrap(<Payroll />)} />
    <Route path="/departments" element={wrap(<Departments />)} />
    <Route path="*"            element={<Navigate to="/" replace />} />
  </Routes>
);

export default AppRoutes;
