import React, { Suspense } from 'react';
import { createRoute, Outlet, Navigate, type AnyRoute } from '@tanstack/react-router';
import { AppFlagGuard, FLAG_KEYS } from '@voxire/config';
import { ErrorBoundary } from '@voxire/ui';
import { Spinner } from '@voxire/ui';
import TrialBanner from './shared/components/TrialBanner';
import { getErpHomePath } from './navigation';
import { useUserStore } from '../../shared/stores/userStore';

const Dashboard = React.lazy(() => import('./features/dashboard/Dashboard'));
const Inventory = React.lazy(() => import('./features/inventory/Inventory'));
const Products = React.lazy(() => import('./features/products/Products'));
const Purchasing = React.lazy(() => import('./features/purchasing/Purchasing'));
const Suppliers = React.lazy(() => import('./features/suppliers/Suppliers'));
const ProductDetail = React.lazy(() => import('./features/products/ProductDetail'));
const Categories = React.lazy(() => import('./features/categories/Categories'));
const Brands = React.lazy(() => import('./features/brands/Brands'));
const Warehouses = React.lazy(() => import('./features/warehouses/Warehouses'));
const StockTransfers = React.lazy(() => import('./features/stock-transfers/StockTransfers'));
const StockTakes = React.lazy(() => import('./features/stock-takes/StockTakes'));
const LowStockAlerts = React.lazy(() => import('./features/low-stock-alerts/LowStockAlerts'));
const Receiving = React.lazy(() => import('./features/purchasing/Receiving'));
const Reorder = React.lazy(() => import('./features/reorder/Reorder'));
const PriceLists = React.lazy(() => import('./features/price-lists/PriceLists'));
const Variants = React.lazy(() => import('./features/variants/Variants'));
const Barcodes = React.lazy(() => import('./features/barcodes/Barcodes'));
const Expiry = React.lazy(() => import('./features/expiry/Expiry'));
const Valuation = React.lazy(() => import('./features/valuation/Valuation'));

function wrap(component: React.ReactNode) {
  return () => (
    <Suspense
      fallback={
        <div className="flex min-h-48 items-center justify-center">
          <Spinner />
        </div>
      }
    >
      {component}
    </Suspense>
  );
}

export function createErpRoutes<TParent extends AnyRoute>(parent: TParent) {
  const layoutRoute = createRoute({
    getParentRoute: () => parent,
    path: 'erp',
    component: () => (
      <AppFlagGuard flag={FLAG_KEYS.APP_ERP}>
        <ErrorBoundary>
          <TrialBanner />
          <Outlet />
        </ErrorBoundary>
      </AppFlagGuard>
    ),
  });

  const indexRoute = createRoute({
    getParentRoute: () => layoutRoute,
    path: '/',
    component: () => {
      const home = getErpHomePath(useUserStore((state) => state.user));
      return home === '/erp' ? (
        <Suspense fallback={<Spinner />}>
          <Dashboard />
        </Suspense>
      ) : (
        <Navigate to={home} replace />
      );
    },
  });

  const inventoryRoute = createRoute({
    getParentRoute: () => layoutRoute,
    path: '/inventory',
    component: wrap(<Inventory />),
  });

  const productsRoute = createRoute({
    getParentRoute: () => layoutRoute,
    path: '/products',
    component: wrap(<Products />),
  });

  const productDetailRoute = createRoute({
    getParentRoute: () => layoutRoute,
    path: '/products/$id',
    component: wrap(<ProductDetail />),
  });

  const categoriesRoute = createRoute({
    getParentRoute: () => layoutRoute,
    path: '/categories',
    component: wrap(<Categories />),
  });

  const brandsRoute = createRoute({
    getParentRoute: () => layoutRoute,
    path: '/brands',
    component: wrap(<Brands />),
  });

  const warehousesRoute = createRoute({
    getParentRoute: () => layoutRoute,
    path: '/warehouses',
    component: wrap(<Warehouses />),
  });

  const transfersRoute = createRoute({
    getParentRoute: () => layoutRoute,
    path: '/transfers',
    component: wrap(<StockTransfers />),
  });

  const stockTakesRoute = createRoute({
    getParentRoute: () => layoutRoute,
    path: '/stock-takes',
    component: wrap(<StockTakes />),
  });

  const alertsRoute = createRoute({
    getParentRoute: () => layoutRoute,
    path: '/alerts',
    component: wrap(<LowStockAlerts />),
  });

  const receivingRoute = createRoute({
    getParentRoute: () => layoutRoute,
    path: '/receiving',
    component: wrap(<Receiving />),
  });

  const reorderRoute = createRoute({
    getParentRoute: () => layoutRoute,
    path: '/reorder',
    component: wrap(<Reorder />),
  });

  const priceListsRoute = createRoute({
    getParentRoute: () => layoutRoute,
    path: '/price-lists',
    component: wrap(<PriceLists />),
  });

  const variantsRoute = createRoute({
    getParentRoute: () => layoutRoute,
    path: '/variants',
    component: wrap(<Variants />),
  });

  const barcodesRoute = createRoute({
    getParentRoute: () => layoutRoute,
    path: '/barcodes',
    component: wrap(<Barcodes />),
  });

  const expiryRoute = createRoute({
    getParentRoute: () => layoutRoute,
    path: '/expiry',
    component: wrap(<Expiry />),
  });

  const valuationRoute = createRoute({
    getParentRoute: () => layoutRoute,
    path: '/valuation',
    component: wrap(<Valuation />),
  });

  const purchasingRoute = createRoute({
    getParentRoute: () => layoutRoute,
    path: '/purchasing',
    component: wrap(<Purchasing />),
  });

  const suppliersRoute = createRoute({
    getParentRoute: () => layoutRoute,
    path: '/suppliers',
    component: wrap(<Suppliers />),
  });

  const settingsRoute = createRoute({
    getParentRoute: () => layoutRoute,
    path: '/settings',
    component: () => <Navigate to="/settings/erp" replace />,
  });

  const catchAllRoute = createRoute({
    getParentRoute: () => layoutRoute,
    path: '$',
    component: () => <Navigate to="/erp" replace />,
  });

  return layoutRoute.addChildren([
    indexRoute,
    inventoryRoute,
    productsRoute,
    productDetailRoute,
    categoriesRoute,
    brandsRoute,
    warehousesRoute,
    transfersRoute,
    stockTakesRoute,
    alertsRoute,
    receivingRoute,
    reorderRoute,
    priceListsRoute,
    variantsRoute,
    barcodesRoute,
    expiryRoute,
    valuationRoute,
    purchasingRoute,
    suppliersRoute,
    settingsRoute,
    catchAllRoute,
  ]);
}
