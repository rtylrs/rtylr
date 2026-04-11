import React from 'react';
import { Routes, Route, Navigate } from 'react-router-dom';
import DefaultLayout from '../layouts/DefaultLayout/DefaultLayout';
import Dashboard from '../screens/Dashboard/Dashboard';
import Revenue from '../screens/Revenue/Revenue';
import Customers from '../screens/Customers/Customers';
import Products from '../screens/Products/Products';
import Operations from '../screens/Operations/Operations';
import Financial from '../screens/Financial/Financial';
import Reports from '../screens/Reports/Reports';
import Settings from '../screens/Settings/Settings';

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
