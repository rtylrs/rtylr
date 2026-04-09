import React from 'react';
import { Routes, Route, Navigate } from 'react-router-dom';
import DefaultLayout from '../layouts/DefaultLayout/DefaultLayout';
import Dashboard from '../screens/Dashboard/Dashboard';
import Inventory from '../screens/Inventory/Inventory';
import Purchasing from '../screens/Purchasing/Purchasing';
import Suppliers from '../screens/Suppliers/Suppliers';
import Products from '../screens/Products/Products';
import ProductDetail from '../screens/Products/ProductDetail';
import Categories from '../screens/Categories/Categories';
import Brands from '../screens/Brands/Brands';
import Warehouses from '../screens/Warehouses/Warehouses';
import StockTransfers from '../screens/StockTransfers/StockTransfers';
import StockTakes from '../screens/StockTakes/StockTakes';
import LowStockAlerts from '../screens/LowStockAlerts/LowStockAlerts';
import Receiving from '../screens/Receiving/Receiving';
import Reorder from '../screens/Reorder/Reorder';
import PriceLists from '../screens/PriceLists/PriceLists';
import Variants from '../screens/Variants/Variants';
import Barcodes from '../screens/Barcodes/Barcodes';
import Expiry from '../screens/Expiry/Expiry';
import Valuation from '../screens/Valuation/Valuation';
import Settings from '../screens/Settings/Settings';

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
