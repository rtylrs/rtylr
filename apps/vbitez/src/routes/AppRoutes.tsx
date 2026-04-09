import React from 'react';

const POS = React.lazy(() => import('../screens/POS'));
const Payment = React.lazy(() => import('../screens/Payment'));
const Receipt = React.lazy(() => import('../screens/Receipt'));
const MenuItems = React.lazy(() => import('../screens/MenuItems'));
const Modifiers = React.lazy(() => import('../screens/Modifiers'));
const Combos = React.lazy(() => import('../screens/Combos'));
const SplitBill = React.lazy(() => import('../screens/SplitBill'));
const MultiPayment = React.lazy(() => import('../screens/MultiPayment'));
const Refund = React.lazy(() => import('../screens/Refund'));
const Discount = React.lazy(() => import('../screens/Discount'));
const GiftCards = React.lazy(() => import('../screens/GiftCards'));
const KDS = React.lazy(() => import('../screens/KDS'));
const CustomerLookup = React.lazy(() => import('../screens/CustomerLookup'));
const Receipts = React.lazy(() => import('../screens/Receipts'));
const DaySummary = React.lazy(() => import('../screens/DaySummary'));
const Override = React.lazy(() => import('../screens/Override'));
const ProductQuickEdit = React.lazy(() => import('../screens/ProductQuickEdit'));
const OfflineQueue = React.lazy(() => import('../screens/OfflineQueue'));
const Orders = React.lazy(() => import('../screens/Orders'));
const Tables = React.lazy(() => import('../screens/Tables'));
const CloseSession = React.lazy(() => import('../screens/CloseSession'));
const Settings = React.lazy(() => import('../screens/Settings'));
const HeldOrders = React.lazy(() => import('../screens/HeldOrders'));
const CashDrop = React.lazy(() => import('../screens/CashDrop'));

const AppRoutes = [
  { path: '/', element: <POS /> },
  { path: '/menu', element: <MenuItems /> },
  { path: '/modifiers', element: <Modifiers /> },
  { path: '/combos', element: <Combos /> },
  { path: '/split-bill/:orderId', element: <SplitBill /> },
  { path: '/payment', element: <Payment /> },
  { path: '/payment/:orderId', element: <MultiPayment /> },
  { path: '/refund/:orderId', element: <Refund /> },
  { path: '/discount/:orderId', element: <Discount /> },
  { path: '/gift-cards', element: <GiftCards /> },
  { path: '/kds', element: <KDS /> },
  { path: '/customer-lookup', element: <CustomerLookup /> },
  { path: '/receipts', element: <Receipts /> },
  { path: '/receipt', element: <Receipt /> },
  { path: '/day-summary', element: <DaySummary /> },
  { path: '/override', element: <Override /> },
  { path: '/products/quick', element: <ProductQuickEdit /> },
  { path: '/offline', element: <OfflineQueue /> },
  { path: '/orders', element: <Orders /> },
  { path: '/tables', element: <Tables /> },
  { path: '/close-session', element: <CloseSession /> },
  { path: '/settings', element: <Settings /> },
  { path: '/held-orders', element: <HeldOrders /> },
  { path: '/cash-drop', element: <CashDrop /> },
];

export default AppRoutes;
