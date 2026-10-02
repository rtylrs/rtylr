import React, { Suspense } from 'react';
import { createRoute, Outlet, type AnyRoute } from '@tanstack/react-router';
import { Spinner } from '@voxire/ui';
import HRPermissionGuard from './shared/components/HRPermissionGuard';
const General = React.lazy(() => import('./features/settings/Settings'));
const Payroll = React.lazy(
  () => import('./features/settings/payroll-config/PayrollConfigSettings')
);
const Leave = React.lazy(() => import('./features/settings/LeavePolicySettings'));
export function createHrSettingsRoutes<TParent extends AnyRoute>(parent: TParent) {
  const route = createRoute({
    getParentRoute: () => parent,
    path: 'hr',
    component: () => (
      <Suspense
        fallback={
          <div className="flex min-h-48 items-center justify-center">
            <Spinner />
          </div>
        }
      >
        <Outlet />
      </Suspense>
    ),
  });
  return route.addChildren([
    createRoute({
      getParentRoute: () => route,
      path: '/',
      component: () => (
        <HRPermissionGuard permission="hub:business:settings:manage">
          <General />
        </HRPermissionGuard>
      ),
    }),
    createRoute({
      getParentRoute: () => route,
      path: 'payroll-config',
      component: () => (
        <HRPermissionGuard permission="hr:payroll:manage">
          <Payroll />
        </HRPermissionGuard>
      ),
    }),
    createRoute({
      getParentRoute: () => route,
      path: 'leave-policy',
      component: () => (
        <HRPermissionGuard permission="hr:time_off:approve">
          <Leave />
        </HRPermissionGuard>
      ),
    }),
  ]);
}
