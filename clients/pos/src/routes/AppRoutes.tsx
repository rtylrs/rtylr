import React from 'react';
import { Navigate } from 'react-router-dom';
import type { POSFeatureKey } from '../shared/stores/settingsStore';

const POS = React.lazy(() => import('../features/pos'));
const Payment = React.lazy(() => import('../features/payment'));
const Receipt = React.lazy(() => import('../features/receipts/Receipt'));
const MenuItems = React.lazy(() => import('../features/menu-items'));
const Modifiers = React.lazy(() => import('../features/modifiers'));
const Combos = React.lazy(() => import('../features/combos'));
const SplitBill = React.lazy(() => import('../features/payment/SplitBill'));
const MultiPayment = React.lazy(() => import('../features/payment/MultiPayment'));
const Refund = React.lazy(() => import('../features/refund'));
const Discount = React.lazy(() => import('../features/discount'));
const GiftCards = React.lazy(() => import('../features/gift-cards'));
const KDS = React.lazy(() => import('../features/kds'));
const CustomerLookup = React.lazy(() => import('../features/customer-lookup'));
const Receipts = React.lazy(() => import('../features/receipts/ReceiptList'));
const DaySummary = React.lazy(() => import('../features/day-summary'));
const Override = React.lazy(() => import('../features/override'));
const ProductQuickEdit = React.lazy(() => import('../features/product-quick-edit'));
const Orders = React.lazy(() => import('../features/orders'));
const Tables = React.lazy(() => import('../features/tables'));
const CloseSession = React.lazy(() => import('../features/close-session'));
const Settings = React.lazy(() => import('../features/settings'));
const HeldOrders = React.lazy(() => import('../features/orders/HeldOrders'));
const CashDrop = React.lazy(() => import('../features/cash-drop'));
const Inventory = React.lazy(() => import('../features/inventory'));
const Staff = React.lazy(() => import('../features/staff'));
const OfflineQueue = React.lazy(() => import('../features/offline-queue'));

export type AppRoute = {
  path: string;
  element: React.ReactElement;
  feature?: POSFeatureKey;
};

const AppRoutes: AppRoute[] = [
  { path: '/', element: <POS /> },
  { path: '/menu', element: <MenuItems /> },
  { path: '/modifiers', element: <Modifiers />, feature: 'modifiers' },
  { path: '/combos', element: <Combos />, feature: 'combos' },
  { path: '/split-bill/:orderId', element: <SplitBill /> },
  { path: '/payment', element: <Payment /> },
  { path: '/payment/:orderId', element: <MultiPayment /> },
  { path: '/refund/:orderId', element: <Refund /> },
  { path: '/discount/:orderId', element: <Discount /> },
  { path: '/gift-cards', element: <GiftCards />, feature: 'giftCards' },
  { path: '/kds', element: <KDS />, feature: 'kitchenDisplay' },
  { path: '/customer-lookup', element: <CustomerLookup />, feature: 'customers' },
  { path: '/receipts', element: <Receipts /> },
  { path: '/receipt', element: <Receipt /> },
  { path: '/day-summary', element: <DaySummary /> },
  { path: '/override', element: <Override /> },
  { path: '/products/quick', element: <ProductQuickEdit />, feature: 'inventory' },
  { path: '/offline', element: <OfflineQueue /> },
  { path: '/orders', element: <Orders /> },
  { path: '/tables', element: <Tables />, feature: 'tables' },
  { path: '/close-session', element: <CloseSession /> },
  { path: '/settings', element: <Settings /> },
  { path: '/held-orders', element: <HeldOrders /> },
  { path: '/cash-drop', element: <CashDrop />, feature: 'cashDrawer' },
  { path: '/inventory', element: <Inventory />, feature: 'inventory' },
  { path: '/staff', element: <Staff /> },
];

export default AppRoutes;
