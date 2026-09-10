import { useCallback } from 'react';
import { useNavigate, useRouterState } from '@tanstack/react-router';

/**
 * Opens a task, in the way that suits where the reader is standing.
 *
 * From a list, it sets `?openTask=<id>` so task detail arrives as a drawer over
 * that list and the list stays mounted underneath. From the standalone
 * `/task/$id` page there is no list to preserve, so following a dependency or a
 * subtask navigates to that task's own page instead of stacking a drawer on top
 * of a page showing the same component.
 *
 * It also removes twelve `to={\`/task/${id}\` as string}` casts.
 * `docs/routing-search-params.md` forbids casting `to=`, precisely because a
 * cast means a renamed route fails at run time instead of at build time, and
 * every call site that opened a task had one (rtylr#1079).
 */
export function useOpenTask(): (taskId: string) => void {
  const navigate = useNavigate();
  const onTaskPage = useRouterState({
    select: (state) => state.location.pathname.startsWith('/task/'),
  });

  return useCallback(
    (taskId: string) => {
      if (onTaskPage) {
        void navigate({ to: '/task/$id', params: { id: taskId } });
        return;
      }
      void navigate({
        to: '.',
        search: (previous: Record<string, unknown>) => ({ ...previous, openTask: taskId }),
      });
    },
    [navigate, onTaskPage]
  );
}
