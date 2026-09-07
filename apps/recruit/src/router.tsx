import React, { Suspense } from 'react';
import {
  createRootRoute,
  createRoute,
  createRouter,
  Navigate,
  Outlet,
} from '@tanstack/react-router';
import { AppFlagGuard, FLAG_KEYS } from '@voxire/config';
import { AppGuard } from '@voxire/auth';
import { CookieBanner } from '@voxire/consent';
import { ErrorBoundary } from '@voxire/ui';
import { userClient } from './shared/lib/clients';

import RootLanding from './features/public/RootLanding';
import CareersIndex from './features/public/CareersIndex';
import CareerDetail from './features/public/CareerDetail';
import ApplyPage from './features/public/ApplyPage';
import NotFound from './features/public/NotFound';
import AdminLayout from './features/admin/AdminLayout';

const CareersAdmin = React.lazy(() => import('./features/admin/CareersAdmin'));
const SubmissionsAdmin = React.lazy(() => import('./features/admin/SubmissionsAdmin'));
const CustomizationAdmin = React.lazy(() => import('./features/admin/CustomizationAdmin'));
const SectionsAdmin = React.lazy(() => import('./features/admin/SectionsAdmin'));

const rootRoute = createRootRoute({
  component: () => (
    <ErrorBoundary>
      <AppFlagGuard flag={FLAG_KEYS.APP_RECRUIT}>
        <Outlet />
        <CookieBanner />
      </AppFlagGuard>
    </ErrorBoundary>
  ),
});

// Public routes
const indexRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: '/',
  component: RootLanding,
});

const careersIndexRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: '/$slug',
  component: CareersIndex,
});

const careerDetailRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: '/$slug/$careerSlug',
  component: CareerDetail,
});

const applyRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: '/$slug/$careerSlug/apply',
  component: ApplyPage,
});

const notFoundRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: '*',
  component: NotFound,
});

// Admin routes - guarded
const adminGuardRoute = createRoute({
  getParentRoute: () => rootRoute,
  id: 'admin-guard',
  component: () => (
    <AppGuard
      appId="recruit"
      fetchUser={(): Promise<Record<string, unknown>> =>
        userClient.getMe({}).then((response) => response as unknown as Record<string, unknown>)
      }
    >
      <Outlet />
    </AppGuard>
  ),
});

const adminLayoutRoute = createRoute({
  getParentRoute: () => adminGuardRoute,
  path: '/admin',
  component: AdminLayout,
});

const adminIndexRoute = createRoute({
  getParentRoute: () => adminLayoutRoute,
  path: '/',
  component: () => (
    <Suspense fallback={null}>
      <CareersAdmin />
    </Suspense>
  ),
});

const adminCareersRoute = createRoute({
  getParentRoute: () => adminLayoutRoute,
  path: '/careers',
  component: () => (
    <Suspense fallback={null}>
      <CareersAdmin />
    </Suspense>
  ),
});

const adminSubmissionsRoute = createRoute({
  getParentRoute: () => adminLayoutRoute,
  path: '/submissions',
  component: () => (
    <Suspense fallback={null}>
      <SubmissionsAdmin />
    </Suspense>
  ),
});

const adminSectionsRoute = createRoute({
  getParentRoute: () => adminLayoutRoute,
  path: '/sections',
  component: () => (
    <Suspense fallback={null}>
      <SectionsAdmin />
    </Suspense>
  ),
});

const adminCustomizationRoute = createRoute({
  getParentRoute: () => adminLayoutRoute,
  path: '/customization',
  component: () => (
    <Suspense fallback={null}>
      <CustomizationAdmin />
    </Suspense>
  ),
});

const adminCatchAllRoute = createRoute({
  getParentRoute: () => adminLayoutRoute,
  path: '*',
  component: () => <Navigate to="/admin/careers" replace />,
});

const routeTree = rootRoute.addChildren([
  indexRoute,
  adminGuardRoute.addChildren([
    adminLayoutRoute.addChildren([
      adminIndexRoute,
      adminCareersRoute,
      adminSubmissionsRoute,
      adminSectionsRoute,
      adminCustomizationRoute,
      adminCatchAllRoute,
    ]),
  ]),
  careersIndexRoute,
  careerDetailRoute,
  applyRoute,
  notFoundRoute,
]);

export const router = createRouter({ routeTree });

declare module '@tanstack/react-router' {
  interface Register {
    router: typeof router;
  }
}
