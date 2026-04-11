import React, { Suspense, lazy } from 'react';
import { Routes, Route, Navigate } from 'react-router-dom';
import DefaultLayout from '../layouts/DefaultLayout/DefaultLayout';

const Dashboard = lazy(() => import('../features/dashboard/Dashboard'));
const CalendarScreen = lazy(() => import('../features/calendar/Calendar'));
const Appointments = lazy(() => import('../features/appointments/Appointments'));
const Services = lazy(() => import('../features/services/Services'));
const Resources = lazy(() => import('../features/resources/Resources'));
const Waitlist = lazy(() => import('../features/waitlist/Waitlist'));
const Customers = lazy(() => import('../features/customers/Customers'));
const Settings = lazy(() => import('../features/settings/Settings'));

const LazyFallback: React.FC = () => (
  <div className="flex items-center justify-center h-full">
    <div
      style={{
        width: 20,
        height: 20,
        border: '2px solid var(--hub-border)',
        borderTopColor: '#6366f1',
        borderRadius: '50%',
        animation: 'spin 0.6s linear infinite',
      }}
    />
    <style>{'@keyframes spin { to { transform: rotate(360deg) } }'}</style>
  </div>
);

const AppRoutes: React.FC = () => {
  return (
    <DefaultLayout>
      <Suspense fallback={<LazyFallback />}>
        <Routes>
          <Route path="/" element={<Dashboard />} />
          <Route path="/calendar" element={<CalendarScreen />} />
          <Route path="/appointments" element={<Appointments />} />
          <Route path="/services" element={<Services />} />
          <Route path="/resources" element={<Resources />} />
          <Route path="/waitlist" element={<Waitlist />} />
          <Route path="/customers" element={<Customers />} />
          <Route path="/settings" element={<Settings />} />
          <Route path="*" element={<Navigate to="/" replace />} />
        </Routes>
      </Suspense>
    </DefaultLayout>
  );
};

export default AppRoutes;
