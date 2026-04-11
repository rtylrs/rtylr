import React, { lazy, Suspense } from 'react';
import { Routes, Route, Navigate } from 'react-router-dom';
import DefaultLayout from '../layouts/DefaultLayout/DefaultLayout';

const Dashboard = lazy(() => import('../screens/Dashboard/Dashboard'));
const Products = lazy(() => import('../screens/Products/Products'));
const Orders = lazy(() => import('../screens/Orders/Orders'));
const Storefront = lazy(() => import('../screens/Storefront/Storefront'));
const Categories = lazy(() => import('../screens/Categories/Categories'));
const Shipping = lazy(() => import('../screens/Shipping/Shipping'));
const Payments = lazy(() => import('../screens/Payments/Payments'));
const Settings = lazy(() => import('../screens/Settings/Settings'));

const PageLoader: React.FC = () => (
  <div className="flex items-center justify-center py-20">
    <div
      style={{
        width: 20,
        height: 20,
        border: '2px solid var(--hub-border)',
        borderTopColor: '#14b8a6',
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
      <Suspense fallback={<PageLoader />}>
        <Routes>
          <Route path="/" element={<Dashboard />} />
          <Route path="/products" element={<Products />} />
          <Route path="/orders" element={<Orders />} />
          <Route path="/storefront" element={<Storefront />} />
          <Route path="/categories" element={<Categories />} />
          <Route path="/shipping" element={<Shipping />} />
          <Route path="/payments" element={<Payments />} />
          <Route path="/settings" element={<Settings />} />
          <Route path="*" element={<Navigate to="/" replace />} />
        </Routes>
      </Suspense>
    </DefaultLayout>
  );
};

export default AppRoutes;
