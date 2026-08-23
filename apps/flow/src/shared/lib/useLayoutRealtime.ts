import { useCallback, useMemo } from 'react';
import { useQueryClient } from '@tanstack/react-query';
import type { RealtimeEvent } from '@voxire/api';

import { getCurrentUserId } from './currentUser';
import { useRealtime } from './realtime';

/**
 * Subscribe the app shell to the current workspace's topic AND the user's
 * personal topic. Any project/member change triggers a fresh fetch of the
 * sidebar data so teammates see new projects instantly, and notifications
 * pushed on `user:<id>` refresh the inbox list and the header badge. We rely
 * on react-query invalidation (cheap) rather than patching state by hand.
 *
 * Self-authored events are ignored (the mutation's onSuccess already
 * invalidated), which avoids the redundant refetch and the flicker that
 * would otherwise happen when optimistic UI is in flight. Same rule as
 * useTaskRealtime and useProjectRealtime.
 *
 * Pass `enabled: false` to disable - the hook becomes a no-op. Settings does
 * this because none of the sidebar data is on screen there.
 */
export function useLayoutRealtime(
  businessId: string | undefined,
  userId: string | undefined,
  enabled: boolean
): void {
  const qc = useQueryClient();
  const currentUserId = getCurrentUserId();

  const onEvent = useCallback(
    (e: RealtimeEvent): void => {
      if (e.type.startsWith('notification.')) {
        // Refresh the inbox list + header badge across all user tabs.
        qc.invalidateQueries({ queryKey: ['notifications'] });
        return;
      }
      if (e.actor_id && currentUserId && String(e.actor_id) === currentUserId) {
        return;
      }
      if (e.type.startsWith('project.')) {
        qc.invalidateQueries({ queryKey: ['projects'] });
        // archived / restored / purged move a project between the list and the
        // recycle bin, so the bin's own query has to refetch too.
        qc.invalidateQueries({ queryKey: ['archived-projects'] });
      }
      if (e.type.startsWith('workspace.')) {
        qc.invalidateQueries({ queryKey: ['workspaces'] });
        qc.invalidateQueries({ queryKey: ['workspace-members'] });
      }
    },
    [qc, currentUserId]
  );

  const topics = useMemo(() => {
    const t: string[] = [];
    // business:<id>, not workspace:<id>. v2 fans every project lifecycle event
    // (created, updated, archived, restored, purged) to the business topic via
    // publishBusinessEvent, and nothing was listening there - so none of those
    // events reached the sidebar. Subscribing to the workspace topic instead
    // also kept a dying concept alive on the client.
    if (businessId) {
      t.push(`business:${businessId}`);
    }
    if (userId) t.push(`user:${userId}`);
    return t;
  }, [businessId, userId]);

  useRealtime(enabled ? topics : [], onEvent);
}
