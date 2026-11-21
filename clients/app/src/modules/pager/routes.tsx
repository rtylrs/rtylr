import React, { Suspense } from 'react';
import { createRoute, Outlet, type AnyRoute } from '@tanstack/react-router';
import { AppFlagGuard, FLAG_KEYS } from '@voxire/config';
import { FullScreenLoader } from '@voxire/components';
import { ErrorBoundary, Spinner } from '@voxire/ui';
import './pager.css';

const Join = React.lazy(() => import('./screens/Public/Join'));
const Ticket = React.lazy(() => import('./screens/Public/Ticket'));
const Try = React.lazy(() => import('./screens/Public/Try'));
const Board = React.lazy(() => import('./screens/Staff/Board'));
const Poster = React.lazy(() => import('./screens/Staff/Poster'));

/**
 * The customer's side of the pager: public and unauthenticated, mounted on the
 * Portal's root route next to the public menu so it bypasses `BusinessRoute`.
 * The join and ticket screens share one document on purpose: the tap on the
 * join screen unlocks the sound the ticket screen plays later.
 */
export function createPagerPublicRoutes<TParent extends AnyRoute>(rootRoute: TParent) {
  const layoutRoute = createRoute({
    getParentRoute: () => rootRoute,
    id: 'pager-public',
    component: () => (
      <ErrorBoundary>
        <Suspense fallback={<FullScreenLoader />}>
          <AppFlagGuard flag={FLAG_KEYS.APP_PAGER}>
            <div className="pager-root">
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
        path: '/$slug/pager',
        component: Join,
      }),
      createRoute({
        getParentRoute: () => layoutRoute,
        path: '/pager/t/$token',
        component: Ticket,
      }),
      createRoute({
        getParentRoute: () => layoutRoute,
        path: '/pager/try',
        component: Try,
      }),
    ]),
  ];
}

function StaffScreen({ children }: { children: React.ReactNode }): React.JSX.Element {
  return (
    <AppFlagGuard flag={FLAG_KEYS.APP_PAGER}>
      <ErrorBoundary>
        <Suspense
          fallback={
            <div className="flex min-h-48 items-center justify-center">
              <Spinner />
            </div>
          }
        >
          {children}
        </Suspense>
      </ErrorBoundary>
    </AppFlagGuard>
  );
}

/**
 * The staff board and its counter poster, inside the signed-in Portal shell.
 * Reachable by URL only; the poster is linked from the board.
 */
export function createPagerStaffRoutes<TParent extends AnyRoute>(parent: TParent) {
  return [
    createRoute({
      getParentRoute: () => parent,
      path: 'pager',
      component: () => (
        <StaffScreen>
          <Board />
        </StaffScreen>
      ),
    }),
    createRoute({
      getParentRoute: () => parent,
      path: 'pager/poster',
      component: () => (
        <StaffScreen>
          <Poster />
        </StaffScreen>
      ),
    }),
  ] as const;
}
