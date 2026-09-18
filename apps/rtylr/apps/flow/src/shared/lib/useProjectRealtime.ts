import { useCallback, useMemo } from 'react';
import { useQueryClient } from '@tanstack/react-query';
import type { RealtimeEvent } from '@voxire/api';

import { getCurrentUserId } from './currentUser';
import { useRealtime } from './realtime';

/**
 * Subscribe to a project's realtime channel and invalidate the query
 * caches every consumer of that project cares about. Every project
 * view (Board, List, Timeline, Calendar, Dashboard, Overview) renders
 * the same tasks/columns/overview queries, so extracting the fanout
 * logic here keeps us from reinventing it per screen and keeps the
 * invalidation set in sync.
 *
 * Events the current user caused are skipped - react-query's mutation
 * onSuccess has already invalidated, refetching twice is wasteful and
 * produces flicker when an optimistic UI is in flight.
 *
 * Pass an empty string (or no projectId) to disable - the hook becomes
 * a no-op. This is safer than requiring callers to branch.
 */
export function useProjectRealtime(projectId: string | undefined): void {
  const qc = useQueryClient();
  const currentUserId = getCurrentUserId();

  const onEvent = useCallback(
    (e: RealtimeEvent) => {
      if (e.actor_id && currentUserId && String(e.actor_id) === currentUserId) {
        return;
      }
      // Invalidation is keyed by prefix - react-query invalidates every
      // query whose key starts with the given array, which is exactly
      // what we want here (any filter of tasks, columns for this
      // project, the overview aggregate, etc.).
      if (e.type.startsWith('task.') || e.type.startsWith('column.')) {
        qc.invalidateQueries({ queryKey: ['tasks'] });
        // `my-tasks` is a separate key from `tasks`, so the line above does
        // not cover it. Any task in this project can be assigned to the
        // current user, which means a teammate's edit can change what My
        // tasks and Today should show. Every task mutation in useTasks.ts
        // already invalidates both keys together; realtime has to match, or
        // the cache stays stale until the 60s staleTime lapses
        // (refetchOnWindowFocus is off, so nothing else refreshes it).
        qc.invalidateQueries({ queryKey: ['my-tasks'] });
        qc.invalidateQueries({ queryKey: ['columns', projectId] });
        qc.invalidateQueries({ queryKey: ['project-overview', projectId] });
        qc.invalidateQueries({ queryKey: ['milestones', projectId] });
      }
      if (e.type.startsWith('project.')) {
        qc.invalidateQueries({ queryKey: ['projects'] });
        qc.invalidateQueries({ queryKey: ['project-overview', projectId] });
      }
    },
    [qc, projectId, currentUserId]
  );

  const topics = useMemo(() => (projectId ? [`project:${projectId}`] : []), [projectId]);
  useRealtime(topics, onEvent);
}
