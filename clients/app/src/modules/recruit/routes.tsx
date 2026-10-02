import React, { Suspense } from 'react';
import { createRoute, Navigate, Outlet, type AnyRoute } from '@tanstack/react-router';
import { AppFlagGuard, FLAG_KEYS } from '@voxire/config';
import { Spinner, ErrorBoundary } from '@voxire/ui';
import { useUserStore } from '../../shared/stores/userStore';
import { getRecruitHomePath } from './navigation';
const Careers = React.lazy(() => import('./features/admin/CareersAdmin'));
const Submissions = React.lazy(() => import('./features/admin/SubmissionsAdmin'));
const Sections = React.lazy(() => import('./features/admin/SectionsAdmin'));

export function createRecruitRoutes<TParent extends AnyRoute>(parent: TParent) {
  const root = createRoute({
    getParentRoute: () => parent,
    path: 'recruit',
    component: () => (
      <AppFlagGuard flag={FLAG_KEYS.APP_RECRUIT}>
        <ErrorBoundary>
          <Suspense
            fallback={
              <div className="flex min-h-48 items-center justify-center">
                <Spinner />
              </div>
            }
          >
            <Outlet />
          </Suspense>
        </ErrorBoundary>
      </AppFlagGuard>
    ),
  });
  return root.addChildren([
    createRoute({
      getParentRoute: () => root,
      path: '/',
      component: () => {
        const user = useUserStore((state) => state.user);
        return <Navigate to={getRecruitHomePath(user)} replace />;
      },
    }),
    createRoute({ getParentRoute: () => root, path: 'careers', component: Careers }),
    createRoute({ getParentRoute: () => root, path: 'submissions', component: Submissions }),
    createRoute({ getParentRoute: () => root, path: 'sections', component: Sections }),
    createRoute({
      getParentRoute: () => root,
      path: 'customization',
      component: () => <Navigate to="/settings/recruit" replace />,
    }),
  ]);
}
