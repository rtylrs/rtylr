import React, { Suspense } from 'react';
import { createRootRoute, createRoute, createRouter, Outlet } from '@tanstack/react-router';
import { AppFlagGuard, FLAG_KEYS } from '@voxire/config';
import { FullScreenLoader } from '@voxire/components';
import { ErrorBoundary } from '@voxire/ui';

const PosMenu = React.lazy(() => import('./screens/Public/PosMenu/PosMenu'));
const SlugLanding = React.lazy(() => import('./screens/Public/SlugLanding/SlugLanding'));
const LocationMenu = React.lazy(() => import('./screens/Public/SlugLanding/LocationMenu'));
const NoMatch = React.lazy(() => import('./screens/Public/Extras/NoMatch'));
const DevMenu = React.lazy(() => import('./screens/Dev/DevMenu'));
const Checkout = React.lazy(() => import('./features/cart/Checkout'));
const OrderStatus = React.lazy(() => import('./features/cart/OrderStatus'));

const isLocalhost =
  typeof window !== 'undefined' &&
  (window.location.hostname === 'localhost' || window.location.hostname === '127.0.0.1');

const rootRoute = createRootRoute({
  component: () => (
    <ErrorBoundary>
      <Suspense fallback={<FullScreenLoader />}>
        <AppFlagGuard flag={FLAG_KEYS.APP_MENU}>
          <Outlet />
        </AppFlagGuard>
      </Suspense>
    </ErrorBoundary>
  ),
});

const posMenuRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: '/b/$businessId',
  component: () => (
    <Suspense fallback={<FullScreenLoader />}>
      <PosMenu />
    </Suspense>
  ),
});

const checkoutRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: '/checkout',
  component: () => (
    <Suspense fallback={<FullScreenLoader />}>
      <Checkout />
    </Suspense>
  ),
});

const orderStatusRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: '/order-status/$id',
  component: () => (
    <Suspense fallback={<FullScreenLoader />}>
      <OrderStatus />
    </Suspense>
  ),
});

const locationMenuRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: '/$slug/$location',
  component: () => (
    <Suspense fallback={<FullScreenLoader />}>
      <LocationMenu />
    </Suspense>
  ),
});

const slugLandingRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: '/$slug',
  component: () => (
    <Suspense fallback={<FullScreenLoader />}>
      <SlugLanding />
    </Suspense>
  ),
});

const indexRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: '/',
  component: () => (
    <Suspense fallback={<FullScreenLoader />}>{isLocalhost ? <DevMenu /> : <NoMatch />}</Suspense>
  ),
});

const catchAllRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: '*',
  component: () => (
    <Suspense fallback={<FullScreenLoader />}>
      <NoMatch />
    </Suspense>
  ),
});

const routeTree = rootRoute.addChildren([
  posMenuRoute,
  checkoutRoute,
  orderStatusRoute,
  locationMenuRoute,
  slugLandingRoute,
  indexRoute,
  catchAllRoute,
]);

export const router = createRouter({ routeTree });

declare module '@tanstack/react-router' {
  interface Register {
    router: typeof router;
  }
}
