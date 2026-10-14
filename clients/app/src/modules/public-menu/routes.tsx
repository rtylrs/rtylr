import React, { Suspense } from 'react';
import { createRoute, Outlet, type AnyRoute } from '@tanstack/react-router';
import { AppFlagGuard, FLAG_KEYS } from '@voxire/config';
import { FullScreenLoader } from '@voxire/components';
import { ErrorBoundary } from '@voxire/ui';
import './public-menu.css';

const PosMenu = React.lazy(() => import('./screens/Public/PosMenu/PosMenu'));
const SlugLanding = React.lazy(() => import('./screens/Public/SlugLanding/SlugLanding'));
const LocationMenu = React.lazy(() => import('./screens/Public/SlugLanding/LocationMenu'));
const Checkout = React.lazy(() => import('./features/cart/Checkout'));
const OrderStatus = React.lazy(() => import('./features/cart/OrderStatus'));

/**
 * Public, unauthenticated customer menu, mounted on the Portal's root route so
 * it bypasses `BusinessRoute` / `PortalShell`, like the standalone POS routes.
 * A pathless layout carries the shared `ErrorBoundary`, `APP_MENU` flag guard
 * and `Outlet`; its wrapper owns the scroll container because the Portal's
 * globals set `html`, `body` and `#root` to `overflow: hidden`.
 *
 * Returned as an array so the router can spread it next to the POS routes.
 */
export function createPublicMenuRoutes<TParent extends AnyRoute>(rootRoute: TParent) {
  const layoutRoute = createRoute({
    getParentRoute: () => rootRoute,
    id: 'public-menu',
    component: () => (
      <ErrorBoundary>
        <Suspense fallback={<FullScreenLoader />}>
          <AppFlagGuard flag={FLAG_KEYS.APP_MENU}>
            <div className="public-menu-root">
              <Outlet />
            </div>
          </AppFlagGuard>
        </Suspense>
      </ErrorBoundary>
    ),
  });

  return [
    layoutRoute.addChildren([
      createRoute({
        getParentRoute: () => layoutRoute,
        path: '/$slug/menu',
        component: SlugLanding,
      }),
      createRoute({
        getParentRoute: () => layoutRoute,
        path: '/$slug/menu/$location',
        component: LocationMenu,
      }),
      createRoute({
        getParentRoute: () => layoutRoute,
        path: '/menu/checkout',
        component: Checkout,
      }),
      createRoute({
        getParentRoute: () => layoutRoute,
        path: '/menu/order-status/$id',
        component: OrderStatus,
      }),
      createRoute({
        getParentRoute: () => layoutRoute,
        path: '/b/$businessId/menu',
        component: PosMenu,
      }),
    ]),
  ];
}
