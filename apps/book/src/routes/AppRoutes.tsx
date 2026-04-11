import React, { Suspense, lazy } from 'react';
import { Routes, Route, Navigate } from 'react-router-dom';
import DefaultLayout from '../layouts/DefaultLayout/DefaultLayout';

const Dashboard = lazy(() => import('../screens/Dashboard/Dashboard'));
const CalendarScreen = lazy(() => import('../screens/Calendar/Calendar'));
const Appointments = lazy(() => import('../screens/Appointments/Appointments'));
const Services = lazy(() => import('../screens/Services/Services'));
const Resources = lazy(() => import('../screens/Resources/Resources'));
const Waitlist = lazy(() => import('../screens/Waitlist/Waitlist'));
const Customers = lazy(() => import('../screens/Customers/Customers'));
const Settings = lazy(() => import('../screens/Settings/Settings'));

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
