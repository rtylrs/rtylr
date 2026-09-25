import React, { Suspense } from 'react';
import {
  createRootRoute,
  createRoute,
  createRouter,
  Outlet,
  useNavigate,
} from '@tanstack/react-router';
import { AppFlagGuard, FLAG_KEYS } from '@voxire/config';
import { AppGuard } from '@voxire/auth';
import { ErrorBoundary } from '@voxire/ui';
import { FullScreenLoader } from '@voxire/components';
import { userClient } from './shared/lib/clients';
import DefaultLayout from './layouts/DefaultLayout';

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
const Settings = React.lazy(() => import('./features/settings/Settings'));

function wrap(component: React.ReactNode) {
  return () => <Suspense fallback={<FullScreenLoader />}>{component}</Suspense>;
}

function CatchAllRedirect(): null {
  const navigate = useNavigate();
  React.useEffect(() => {
    void navigate({ to: '/', replace: true });
  }, [navigate]);
  return null;
}

const rootRoute = createRootRoute({
  component: () => (
    <ErrorBoundary>
      <AppGuard
        appId="erp"
        fetchUser={async (): Promise<Record<string, unknown>> =>
          (await userClient.getMe({})) as unknown as Record<string, unknown>
        }
      >
        <AppFlagGuard flag={FLAG_KEYS.APP_ERP}>
          <Outlet />
        </AppFlagGuard>
      </AppGuard>
    </ErrorBoundary>
  ),
});

const layoutRoute = createRoute({
  getParentRoute: () => rootRoute,
  id: 'layout',
  component: DefaultLayout,
});

const indexRoute = createRoute({
  getParentRoute: () => layoutRoute,
  path: '/',
  component: wrap(<Dashboard />),
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
  component: wrap(<Settings />),
});

const catchAllRoute = createRoute({
  getParentRoute: () => layoutRoute,
  path: '*',
  component: CatchAllRedirect,
});

const routeTree = rootRoute.addChildren([
  layoutRoute.addChildren([
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
  ]),
]);

export const router = createRouter({ routeTree });

declare module '@tanstack/react-router' {
  interface Register {
    router: typeof router;
  }
}
