import React, { Suspense } from 'react';
import {
  createRootRoute,
  createRoute,
  createRouter,
  Navigate,
  Outlet,
} from '@tanstack/react-router';
import { APP_URLS, AppFlagGuard, FLAG_KEYS } from '@voxire/config';
import { AppGuard } from '@voxire/auth';
import { CookieBanner } from '@voxire/consent';
import { ErrorBoundary } from '@voxire/ui';
import { userClient } from './shared/lib/clients';

import RootLanding from './features/public/RootLanding';
import CareersIndex from './features/public/CareersIndex';
import CareerDetail from './features/public/CareerDetail';
import ApplyPage from './features/public/ApplyPage';
import NotFound from './features/public/NotFound';

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

function redirectAdmin() {
  const page = window.location.pathname.replace(/^\/admin\/?/, '');
  const target =
    page === 'customization'
      ? '/settings/recruit'
      : ['/careers', '/submissions', '/sections'].includes(`/${page}`)
        ? `/recruit/${page}`
        : '/recruit';
  window.location.replace(`${APP_URLS.dash}${target}${window.location.search}`);
  return null;
}
const adminRedirectRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: '/admin/$',
  component: redirectAdmin,
});
const adminIndexRedirectRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: '/admin',
  component: redirectAdmin,
});
const routeTree = rootRoute.addChildren([
  indexRoute,
  careersIndexRoute,
  careerDetailRoute,
  applyRoute,
  adminIndexRedirectRoute,
  adminRedirectRoute,
  notFoundRoute,
]);
export const publicRecruitRouter = createRouter({ routeTree });
