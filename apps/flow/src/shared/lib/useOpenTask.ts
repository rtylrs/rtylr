import { useCallback } from 'react';
import { useNavigate } from '@tanstack/react-router';

/**
 * Opens a task as a drawer over whatever list is showing, by setting
 * `?openTask=<id>` (rtylr#1079).
 *
 * It also removes twelve `to={\`/task/${id}\` as string}` casts.
 * `docs/routing-search-params.md` forbids casting `to=`, precisely because a
 * cast means a renamed route fails at run time rather than at build time, and
 * every call site that opened a task had one.
 *
 * ## Why this takes no view on where it is called from
 *
 * The first version read the pathname through `useRouterState`, so that the
 * standalone `/task/$id` page could navigate rather than stack a drawer over
 * itself. That broke **13 test files**: they mock `@tanstack/react-router` with
 * a partial factory supplying only `useNavigate`, and 32 files in this app do
 * the same. A hook reached from a dozen components cannot afford to need more
 * of the router than its callers' mocks provide, and updating 13 mocks would
 * only have deferred the next occurrence.
 *
 * The one caller that needs different behaviour, task detail following a
 * dependency, already knows which context it is in: it is handed a `taskId`
 * when it renders inside the drawer and reads the route when it does not. That
 * decision belongs there and not here.
 */
export function useOpenTask(): (taskId: string) => void {
  const navigate = useNavigate();

  return useCallback(
    (taskId: string) => {
      void navigate({
        to: '.',
        search: (previous: Record<string, unknown>) => ({ ...previous, openTask: taskId }),
      });
    },
    [navigate]
  );
}
