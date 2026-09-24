import React, { Suspense } from 'react';
import {
  createRootRoute,
  createRoute,
  createRouter,
  Outlet,
  useRouterState,
} from '@tanstack/react-router';
import { FullScreenLoader } from '@voxire/components';
import { ErrorBoundary } from '@voxire/ui';
import { AnimatedRoutes } from '@voxire/components';
import Public from './layouts/Public/public';

const Login = React.lazy(() => import('./features/login/Login'));
const Register = React.lazy(() => import('./features/register/Register'));
const ResetPassword = React.lazy(() => import('./features/reset-password/ResetPassword'));
const ResetPasswordConfirm = React.lazy(
  () => import('./features/reset-password/ResetPasswordConfirm')
);
const Callback = React.lazy(() => import('./features/callback/Callback'));
const EmailVerification = React.lazy(
  () => import('./features/email-verification/EmailVerification')
);
const VerifyEmail = React.lazy(() => import('./features/email-verification/VerifyEmail'));
const InviteAccept = React.lazy(() => import('./features/invite/InviteAccept'));
const OAuthConsent = React.lazy(() => import('./features/oauth/OAuthConsent'));

const NotFound = React.lazy(() => import('./features/not-found/NotFound'));

const rootRoute = createRootRoute({
  component: () => (
    <ErrorBoundary>
      <Outlet />
    </ErrorBoundary>
  ),
});

function PublicLayout(): React.JSX.Element {
  const pathname = useRouterState({ select: (state) => state.location.pathname });

  return (
    <Public hasHeader={false} hasPadding={false}>
      <AnimatedRoutes animationKey={pathname}>
        <Outlet />
      </AnimatedRoutes>
    </Public>
  );
}

const publicLayoutRoute = createRoute({
  getParentRoute: () => rootRoute,
  id: 'public-layout',
  component: PublicLayout,
});

const indexRoute = createRoute({
  getParentRoute: () => publicLayoutRoute,
  path: '/',
  component: () => (
    <Suspense fallback={<FullScreenLoader />}>
      <Login />
    </Suspense>
  ),
});

const loginRoute = createRoute({
  getParentRoute: () => publicLayoutRoute,
  path: '/login',
  component: () => (
    <Suspense fallback={<FullScreenLoader />}>
      <Login />
    </Suspense>
  ),
});

const registerRoute = createRoute({
  getParentRoute: () => publicLayoutRoute,
  path: '/register',
  component: () => (
    <Suspense fallback={<FullScreenLoader />}>
      <Register />
    </Suspense>
  ),
});

const passwordResetRoute = createRoute({
  getParentRoute: () => publicLayoutRoute,
  path: '/password-reset',
  component: () => (
    <Suspense fallback={<FullScreenLoader />}>
      <ResetPassword />
    </Suspense>
  ),
});

const passwordResetConfirmRoute = createRoute({
  getParentRoute: () => publicLayoutRoute,
  path: '/password-reset/confirm',
  component: () => (
    <Suspense fallback={<FullScreenLoader />}>
      <ResetPasswordConfirm />
    </Suspense>
  ),
});

const callbackRoute = createRoute({
  getParentRoute: () => publicLayoutRoute,
  path: '/callback',
  component: () => (
    <Suspense fallback={<FullScreenLoader />}>
      <Callback />
    </Suspense>
  ),
});

const verifyTokenRoute = createRoute({
  getParentRoute: () => publicLayoutRoute,
  path: '/verify/$token',
  component: () => (
    <Suspense fallback={<FullScreenLoader />}>
      <EmailVerification />
    </Suspense>
  ),
});

const verifyRoute = createRoute({
  getParentRoute: () => publicLayoutRoute,
  path: '/verify',
  component: () => (
    <Suspense fallback={<FullScreenLoader />}>
      <VerifyEmail />
    </Suspense>
  ),
});

const inviteRoute = createRoute({
  getParentRoute: () => publicLayoutRoute,
  path: '/invite/$token',
  component: () => (
    <Suspense fallback={<FullScreenLoader />}>
      <InviteAccept />
    </Suspense>
  ),
});

const oauthRoute = createRoute({
  getParentRoute: () => publicLayoutRoute,
  path: '/oauth',
  component: () => (
    <Suspense fallback={<FullScreenLoader />}>
      <OAuthConsent />
    </Suspense>
  ),
});

const routeTree = rootRoute.addChildren([
  publicLayoutRoute.addChildren([
    indexRoute,
    loginRoute,
    registerRoute,
    passwordResetRoute,
    passwordResetConfirmRoute,
    callbackRoute,
    verifyTokenRoute,
    verifyRoute,
    inviteRoute,
    oauthRoute,
  ]),
]);

export const router = createRouter({
  routeTree,
  defaultNotFoundComponent: () => (
    <Suspense fallback={<FullScreenLoader />}>
      <NotFound />
    </Suspense>
  ),
});

declare module '@tanstack/react-router' {
  interface Register {
    router: typeof router;
  }
}
