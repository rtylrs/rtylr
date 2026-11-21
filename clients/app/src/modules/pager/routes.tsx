import React, { Suspense } from 'react';
import { createRoute, Outlet, type AnyRoute } from '@tanstack/react-router';
import { AppFlagGuard, FLAG_KEYS } from '@voxire/config';
import { FullScreenLoader } from '@voxire/components';
import { ErrorBoundary, Spinner } from '@voxire/ui';
import './pager.css';
import './rennili.css';

const Landing = React.lazy(() => import('./screens/Public/Landing'));
const HowItWorksPage = React.lazy(() => import('./screens/Public/HowItWorksPage'));
const FeaturesPage = React.lazy(() => import('./screens/Public/FeaturesPage'));
const FaqPage = React.lazy(() => import('./screens/Public/FaqPage'));
const RtylrServicesPage = React.lazy(() => import('./screens/Public/RtylrServicesPage'));
const Scan = React.lazy(() => import('./screens/Public/Scan'));
const Ticket = React.lazy(() => import('./screens/Public/Ticket'));
const Try = React.lazy(() => import('./screens/Public/Try'));
const Board = React.lazy(() => import('./screens/Staff/Board'));
const Display = React.lazy(() => import('./screens/Staff/Display'));
const Poster = React.lazy(() => import('./screens/Staff/Poster'));
const Setup = React.lazy(() => import('./screens/Staff/Setup'));
// The account checks and the Portal code behind them load only for staff.
const StaffGate = React.lazy(() => import('./shared/components/RenniliStaffGate'));

/**
 * Rennili's public pages: the landing, the venue page a QR opens, the ticket
 * and the demo. No sign in. The venue page hands over to the ticket by a
 * router navigation, so the ticket keeps the document the QR opened.
 */
function createPublicRoutes<TParent extends AnyRoute>(rootRoute: TParent) {
  const layoutRoute = createRoute({
    getParentRoute: () => rootRoute,
    id: 'rennili-public',
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

  return layoutRoute.addChildren([
    createRoute({ getParentRoute: () => layoutRoute, path: '/', component: Landing }),
    createRoute({
      getParentRoute: () => layoutRoute,
      path: '/how-it-works',
      component: HowItWorksPage,
    }),
    createRoute({
      getParentRoute: () => layoutRoute,
      path: '/features',
      component: FeaturesPage,
    }),
    createRoute({ getParentRoute: () => layoutRoute, path: '/faq', component: FaqPage }),
    createRoute({
      getParentRoute: () => layoutRoute,
      path: '/rtylr',
      component: RtylrServicesPage,
    }),
    createRoute({ getParentRoute: () => layoutRoute, path: '/try', component: Try }),
    createRoute({ getParentRoute: () => layoutRoute, path: '/t/$token', component: Ticket }),
    createRoute({ getParentRoute: () => layoutRoute, path: '/$slug', component: Scan }),
  ]);
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
 * The staff board, its counter display and poster, and the sign-up's last
 * step. They need the signed-in business, checked the way the Portal checks it.
 */
function createStaffRoutes<TParent extends AnyRoute>(rootRoute: TParent) {
  const layoutRoute = createRoute({
    getParentRoute: () => rootRoute,
    id: 'rennili-staff',
    component: () => (
      <Suspense fallback={<FullScreenLoader />}>
        <StaffGate />
      </Suspense>
    ),
  });

  const staffPage = (path: string, Screen: React.ComponentType) =>
    createRoute({
      getParentRoute: () => layoutRoute,
      path,
      component: () => (
        <StaffScreen>
          <Screen />
        </StaffScreen>
      ),
    });

  return layoutRoute.addChildren([
    staffPage('/board', Board),
    staffPage('/display', Display),
    staffPage('/poster', Poster),
    staffPage('/setup', Setup),
  ]);
}

export function createRenniliRoutes<TParent extends AnyRoute>(rootRoute: TParent) {
  return [createPublicRoutes(rootRoute), createStaffRoutes(rootRoute)] as const;
}
