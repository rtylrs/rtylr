import { createRootRoute, createRouter, Outlet } from '@tanstack/react-router';
import { ErrorBoundary } from '@voxire/ui';
import { createRenniliRoutes } from './routes';

// No consent banner on Rennili: nothing optional is tracked without a choice,
// and a customer keeps these pages open while they wait.
const rootRoute = createRootRoute({
  component: () => (
    <ErrorBoundary>
      <Outlet />
    </ErrorBoundary>
  ),
});

export const routeTree = rootRoute.addChildren([...createRenniliRoutes(rootRoute)]);

/** rennili.rtylr.com: built by `entrypoints/rennili`, deployed as its own Pages project. */
export const renniliRouter = createRouter({ routeTree });
