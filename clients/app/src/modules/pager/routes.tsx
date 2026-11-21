import React, { Suspense } from 'react';
import { createRoute, Outlet, type AnyRoute } from '@tanstack/react-router';
import { AppFlagGuard, FLAG_KEYS } from '@voxire/config';
import { FullScreenLoader } from '@voxire/components';
import { ErrorBoundary, Spinner } from '@voxire/ui';
import './pager.css';

const Scan = React.lazy(() => import('./screens/Public/Scan'));
const Ticket = React.lazy(() => import('./screens/Public/Ticket'));
const Try = React.lazy(() => import('./screens/Public/Try'));
const Board = React.lazy(() => import('./screens/Staff/Board'));
const Display = React.lazy(() => import('./screens/Staff/Display'));
const Poster = React.lazy(() => import('./screens/Staff/Poster'));
const Setup = React.lazy(() => import('./screens/Staff/Setup'));

/**
 * The customer's side of the pager: public and unauthenticated, mounted on the
 * Portal's root route next to the public menu so it bypasses `BusinessRoute`.
 * The scan screen hands over to the ticket screen by a router navigation, so
 * the ticket keeps the document the QR opened.
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
        component: Scan,
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
 * The staff board, its counter poster and counter display, inside the
 * signed-in Portal shell. Reachable by URL only; the other two are linked from
 * the board.
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
      path: 'pager/display',
      component: () => (
        <StaffScreen>
          <Display />
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

/** The last step of the pager sign-up, where a new venue switches its pager on. */
export function createPagerSetupRoute<TParent extends AnyRoute>(parent: TParent) {
  return createRoute({
    getParentRoute: () => parent,
    path: 'pager/setup',
    component: () => (
      <StaffScreen>
        <Setup />
      </StaffScreen>
    ),
  });
}
