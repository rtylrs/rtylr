import { useCallback, useMemo } from 'react';
import { useQueryClient } from '@tanstack/react-query';
import type { RealtimeEvent } from '@voxire/api';

import { getCurrentUserId } from './currentUser';
import { useRealtime } from './realtime';

/**
 * Subscribe to a single task's realtime channel. Used by TaskDetail so
 * an open detail screen refreshes comments, activity, and the task body
 * whenever someone else on the team edits it - without polling and
 * without subscribing to the whole project firehose.
 *
 * Self-authored events are ignored (the mutation's onSuccess already
 * invalidated), which avoids the flicker that would otherwise happen
 * when optimistic UI is in flight.
 *
 * Pass an empty string (or no taskId) to disable - the hook becomes a
 * no-op. Safer than forcing callers to branch.
 */
export function useTaskRealtime(taskId: string | undefined): void {
  const qc = useQueryClient();
  const currentUserId = getCurrentUserId();

  const onEvent = useCallback(
    (e: RealtimeEvent) => {
      if (e.actor_id && currentUserId && String(e.actor_id) === currentUserId) {
        return;
      }
      if (e.type.startsWith('task.')) {
        qc.invalidateQueries({ queryKey: ['task', taskId] });
        qc.invalidateQueries({ queryKey: ['tasks'] });
        // Distinct key from `tasks`, so it needs its own call. Matters most
        // when the edit reassigns this task away from the current user: the
        // detail screen refreshes, but My tasks and Today would keep listing
        // it as theirs until the cache expires.
        qc.invalidateQueries({ queryKey: ['my-tasks'] });
        qc.invalidateQueries({ queryKey: ['task-activity', taskId] });
      }
      if (e.type.startsWith('comment.')) {
        qc.invalidateQueries({ queryKey: ['task', taskId] });
        qc.invalidateQueries({ queryKey: ['task-activity', taskId] });
      }
    },
    [qc, taskId, currentUserId]
  );

  const topics = useMemo(() => (taskId ? [`task:${taskId}`] : []), [taskId]);
  useRealtime(topics, onEvent);
}
