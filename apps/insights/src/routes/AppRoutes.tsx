import React, { Suspense } from 'react';
import { Routes, Route, Navigate } from 'react-router-dom';
import { FullScreenLoader } from '@voxire/components';
import DefaultLayout from '../layouts/DefaultLayout';

// Per CLAUDE.md mandate, every route is lazy-loaded. The Suspense
// fallback is required so the lazy chunks don't fall through to a
// silent error when not yet loaded (audit insights.md P0 #4).
const Dashboard = React.lazy(() => import('../features/dashboard/Dashboard'));
const Revenue = React.lazy(() => import('../features/revenue/Revenue'));
const Customers = React.lazy(() => import('../features/customers/Customers'));
const Products = React.lazy(() => import('../features/products/Products'));
const Operations = React.lazy(() => import('../features/operations/Operations'));
const Financial = React.lazy(() => import('../features/financial/Financial'));
const Reports = React.lazy(() => import('../features/reports/Reports'));
const Settings = React.lazy(() => import('../features/settings/Settings'));

const AppRoutes: React.FC = () => {
  return (
    <DefaultLayout>
      <Suspense fallback={<FullScreenLoader />}>
        <Routes>
          <Route path="/" element={<Dashboard />} />
          <Route path="/revenue" element={<Revenue />} />
          <Route path="/customers" element={<Customers />} />
          <Route path="/products" element={<Products />} />
          <Route path="/operations" element={<Operations />} />
          <Route path="/financial" element={<Financial />} />
          <Route path="/reports" element={<Reports />} />
          <Route path="/settings" element={<Settings />} />
          <Route path="*" element={<Navigate to="/" replace />} />
        </Routes>
      </Suspense>
    </DefaultLayout>
  );
};

export default AppRoutes;
