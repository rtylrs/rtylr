import React, { Suspense } from 'react';
import {
  createRootRoute,
  createRoute,
  createRouter,
  Outlet,
  redirect,
} from '@tanstack/react-router';
import { AppFlagGuard, FLAG_KEYS } from '@voxire/config';
import { AppGuard } from '@voxire/auth';
import { ErrorBoundary } from '@voxire/ui';
import { FullScreenLoader } from '@voxire/components';
import {
  Chart as ChartJS,
  CategoryScale,
  LinearScale,
  PointElement,
  LineElement,
  BarElement,
  Filler,
  Tooltip,
  Legend,
} from 'chart.js';
import DefaultLayout from './layouts/DefaultLayout';
import { userClient } from './shared/lib/clients';

ChartJS.register(
  CategoryScale,
  LinearScale,
  PointElement,
  LineElement,
  BarElement,
  Filler,
  Tooltip,
  Legend
);

const Dashboard = React.lazy(() => import('./features/dashboard/Dashboard'));
const Revenue = React.lazy(() => import('./features/revenue/Revenue'));
const Customers = React.lazy(() => import('./features/customers/Customers'));
const Products = React.lazy(() => import('./features/products/Products'));
const Operations = React.lazy(() => import('./features/operations/Operations'));
const Financial = React.lazy(() => import('./features/financial/Financial'));
const Reports = React.lazy(() => import('./features/reports/Reports'));
const Settings = React.lazy(() => import('./features/settings/Settings'));

const rootRoute = createRootRoute({
  component: () => (
    <ErrorBoundary>
      <AppGuard
        appId="insights"
        fetchUser={async (): Promise<Record<string, unknown>> =>
          (await userClient.getMe({})) as unknown as Record<string, unknown>
        }
      >
        <AppFlagGuard flag={FLAG_KEYS.APP_INSIGHTS}>
          <Outlet />
        </AppFlagGuard>
      </AppGuard>
    </ErrorBoundary>
  ),
});

const layoutRoute = createRoute({
  getParentRoute: () => rootRoute,
  id: 'default-layout',
  component: DefaultLayout,
});

const indexRoute = createRoute({
  getParentRoute: () => layoutRoute,
  path: '/',
  component: () => (
    <Suspense fallback={<FullScreenLoader />}>
      <Dashboard />
    </Suspense>
  ),
});

const revenueRoute = createRoute({
  getParentRoute: () => layoutRoute,
  path: '/revenue',
  component: () => (
    <Suspense fallback={<FullScreenLoader />}>
      <Revenue />
    </Suspense>
  ),
});

const customersRoute = createRoute({
  getParentRoute: () => layoutRoute,
  path: '/customers',
  component: () => (
    <Suspense fallback={<FullScreenLoader />}>
      <Customers />
    </Suspense>
  ),
});

const productsRoute = createRoute({
  getParentRoute: () => layoutRoute,
  path: '/products',
  component: () => (
    <Suspense fallback={<FullScreenLoader />}>
      <Products />
    </Suspense>
  ),
});

const operationsRoute = createRoute({
  getParentRoute: () => layoutRoute,
  path: '/operations',
  component: () => (
    <Suspense fallback={<FullScreenLoader />}>
      <Operations />
    </Suspense>
  ),
});

const financialRoute = createRoute({
  getParentRoute: () => layoutRoute,
  path: '/financial',
  component: () => (
    <Suspense fallback={<FullScreenLoader />}>
      <Financial />
    </Suspense>
  ),
});

const reportsRoute = createRoute({
  getParentRoute: () => layoutRoute,
  path: '/reports',
  component: () => (
    <Suspense fallback={<FullScreenLoader />}>
      <Reports />
    </Suspense>
  ),
});

const settingsRoute = createRoute({
  getParentRoute: () => layoutRoute,
  path: '/settings',
  component: () => (
    <Suspense fallback={<FullScreenLoader />}>
      <Settings />
    </Suspense>
  ),
});

const catchAllRoute = createRoute({
  getParentRoute: () => layoutRoute,
  path: '$',
  beforeLoad: () => {
    throw redirect({ to: '/' });
  },
});

const routeTree = rootRoute.addChildren([
  layoutRoute.addChildren([
    indexRoute,
    revenueRoute,
    customersRoute,
    productsRoute,
    operationsRoute,
    financialRoute,
    reportsRoute,
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
