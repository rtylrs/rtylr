import { lazy, Suspense } from 'react';
import { createRoute, Outlet, Navigate, type AnyRoute } from '@tanstack/react-router';
import { AppFlagGuard, FLAG_KEYS } from '@voxire/config';
import { Spinner } from '@voxire/ui';
import {
  Chart as ChartJS,
  CategoryScale,
  LinearScale,
  PointElement,
  LineElement,
  BarElement,
  Filler,
  Tooltip,
  Legend,
} from 'chart.js';

ChartJS.register(
  CategoryScale,
  LinearScale,
  PointElement,
  LineElement,
  BarElement,
  Filler,
  Tooltip,
  Legend
);

const Overview = lazy(() => import('./features/dashboard/Dashboard'));
const Revenue = lazy(() => import('./features/revenue/Revenue'));
const Customers = lazy(() => import('./features/customers/Customers'));
const Products = lazy(() => import('./features/products/Products'));
const Operations = lazy(() => import('./features/operations/Operations'));
const Financial = lazy(() => import('./features/financial/Financial'));
const Reports = lazy(() => import('./features/reports/Reports'));
const ProjectDashboard = lazy(() => import('./features/projects/ProjectDashboardPage'));

export function createInsightsRoutes<TParent extends AnyRoute>(parent: TParent) {
  const insights = createRoute({
    getParentRoute: () => parent,
    path: 'insights',
    component: () => (
      <AppFlagGuard flag={FLAG_KEYS.APP_INSIGHTS}>
        <Suspense
          fallback={
            <div className="flex min-h-48 items-center justify-center">
              <Spinner />
            </div>
          }
        >
          <Outlet />
        </Suspense>
      </AppFlagGuard>
    ),
  });
  return insights.addChildren([
    createRoute({ getParentRoute: () => insights, path: '/', component: Overview }),
    createRoute({ getParentRoute: () => insights, path: 'revenue', component: Revenue }),
    createRoute({ getParentRoute: () => insights, path: 'customers', component: Customers }),
    createRoute({ getParentRoute: () => insights, path: 'products', component: Products }),
    createRoute({ getParentRoute: () => insights, path: 'operations', component: Operations }),
    createRoute({ getParentRoute: () => insights, path: 'financial', component: Financial }),
    createRoute({ getParentRoute: () => insights, path: 'reports', component: Reports }),
    // One Work project's dashboard; the bare path opens the first project.
    createRoute({ getParentRoute: () => insights, path: 'projects', component: ProjectDashboard }),
    createRoute({
      getParentRoute: () => insights,
      path: 'projects/$projectId',
      component: ProjectDashboard,
    }),
    createRoute({
      getParentRoute: () => insights,
      path: 'settings',
      component: () => <Navigate to="/settings/insights" replace />,
    }),
  ]);
}
