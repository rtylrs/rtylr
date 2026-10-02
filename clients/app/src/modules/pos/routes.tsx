import { lazy, Suspense } from 'react';
import { createRoute, type AnyRoute } from '@tanstack/react-router';
import { Spinner } from '@voxire/ui';
const PosModule = lazy(() => import('./PosModule'));
export function createPosRoutes<TParent extends AnyRoute>(parent: TParent) {
  const route = createRoute({ getParentRoute: () => parent, path: 'pos' });
  const component = () => (
    <Suspense
      fallback={
        <div className="flex min-h-48 items-center justify-center">
          <Spinner />
        </div>
      }
    >
      <PosModule />
    </Suspense>
  );
  return route.addChildren([
    createRoute({ getParentRoute: () => route, path: '/', component }),
    createRoute({ getParentRoute: () => route, path: '$', component }),
  ]);
}
