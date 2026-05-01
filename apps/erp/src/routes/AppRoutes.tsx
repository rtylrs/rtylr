import React, { Suspense } from 'react';
import { Routes, Route, Navigate } from 'react-router-dom';
import { FullScreenLoader } from '@voxire/components';
import DefaultLayout from '../layouts/DefaultLayout';

// Every route lazy-loaded per CLAUDE.md.
const Dashboard = React.lazy(() => import('../features/dashboard/Dashboard'));
const Inventory = React.lazy(() => import('../features/inventory/Inventory'));
const Purchasing = React.lazy(() => import('../features/purchasing/Purchasing'));
const Suppliers = React.lazy(() => import('../features/suppliers/Suppliers'));
const Products = React.lazy(() => import('../features/products/Products'));
const ProductDetail = React.lazy(() => import('../features/products/ProductDetail'));
const Categories = React.lazy(() => import('../features/categories/Categories'));
const Brands = React.lazy(() => import('../features/brands/Brands'));
const Warehouses = React.lazy(() => import('../features/warehouses/Warehouses'));
const StockTransfers = React.lazy(() => import('../features/stock-transfers/StockTransfers'));
const StockTakes = React.lazy(() => import('../features/stock-takes/StockTakes'));
const LowStockAlerts = React.lazy(() => import('../features/low-stock-alerts/LowStockAlerts'));
const Receiving = React.lazy(() => import('../features/purchasing/Receiving'));
const Reorder = React.lazy(() => import('../features/reorder/Reorder'));
const PriceLists = React.lazy(() => import('../features/price-lists/PriceLists'));
const Variants = React.lazy(() => import('../features/variants/Variants'));
const Barcodes = React.lazy(() => import('../features/barcodes/Barcodes'));
const Expiry = React.lazy(() => import('../features/expiry/Expiry'));
const Valuation = React.lazy(() => import('../features/valuation/Valuation'));
const Settings = React.lazy(() => import('../features/settings/Settings'));

const AppRoutes: React.FC = () => {
  return (
    <DefaultLayout>
      <Suspense fallback={<FullScreenLoader />}>
        <Routes>
          <Route path="/" element={<Dashboard />} />
          <Route path="/inventory" element={<Inventory />} />
          <Route path="/products" element={<Products />} />
          <Route path="/products/:id" element={<ProductDetail />} />
          <Route path="/categories" element={<Categories />} />
          <Route path="/brands" element={<Brands />} />
          <Route path="/warehouses" element={<Warehouses />} />
          <Route path="/transfers" element={<StockTransfers />} />
          <Route path="/stock-takes" element={<StockTakes />} />
          <Route path="/alerts" element={<LowStockAlerts />} />
          <Route path="/receiving" element={<Receiving />} />
          <Route path="/reorder" element={<Reorder />} />
          <Route path="/price-lists" element={<PriceLists />} />
          <Route path="/variants" element={<Variants />} />
          <Route path="/barcodes" element={<Barcodes />} />
          <Route path="/expiry" element={<Expiry />} />
          <Route path="/valuation" element={<Valuation />} />
          <Route path="/purchasing" element={<Purchasing />} />
          <Route path="/suppliers" element={<Suppliers />} />
          <Route path="/settings" element={<Settings />} />
          <Route path="*" element={<Navigate to="/" replace />} />
        </Routes>
      </Suspense>
    </DefaultLayout>
  );
};

export default AppRoutes;
