import React from 'react';
import { Routes, Route, Navigate } from 'react-router-dom';
import DefaultLayout from '../layouts/DefaultLayout/DefaultLayout';
import Dashboard from '../features/dashboard/Dashboard';
import Inventory from '../features/inventory/Inventory';
import Purchasing from '../features/purchasing/Purchasing';
import Suppliers from '../features/suppliers/Suppliers';
import Products from '../features/products/Products';
import ProductDetail from '../features/products/ProductDetail';
import Categories from '../features/categories/Categories';
import Brands from '../features/brands/Brands';
import Warehouses from '../features/warehouses/Warehouses';
import StockTransfers from '../features/stock-transfers/StockTransfers';
import StockTakes from '../features/stock-takes/StockTakes';
import LowStockAlerts from '../features/low-stock-alerts/LowStockAlerts';
import Receiving from '../features/purchasing/Receiving';
import Reorder from '../features/reorder/Reorder';
import PriceLists from '../features/price-lists/PriceLists';
import Variants from '../features/variants/Variants';
import Barcodes from '../features/barcodes/Barcodes';
import Expiry from '../features/expiry/Expiry';
import Valuation from '../features/valuation/Valuation';
import Settings from '../features/settings/Settings';

const AppRoutes: React.FC = () => {
  return (
    <DefaultLayout>
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
    </DefaultLayout>
  );
};

export default AppRoutes;
