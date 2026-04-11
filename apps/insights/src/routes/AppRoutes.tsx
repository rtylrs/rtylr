import React from 'react';
import { Routes, Route, Navigate } from 'react-router-dom';
import DefaultLayout from '../layouts/DefaultLayout';
import Dashboard from '../features/dashboard/Dashboard';
import Revenue from '../features/revenue/Revenue';
import Customers from '../features/customers/Customers';
import Products from '../features/products/Products';
import Operations from '../features/operations/Operations';
import Financial from '../features/financial/Financial';
import Reports from '../features/reports/Reports';
import Settings from '../features/settings/Settings';

const AppRoutes: React.FC = () => {
  return (
    <DefaultLayout>
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
    </DefaultLayout>
  );
};

export default AppRoutes;
