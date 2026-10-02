import { useCallback, useMemo } from 'react';
import { useQueryClient } from '@tanstack/react-query';
import type { RealtimeEvent } from '@voxire/api';

import { getCurrentUserId } from './currentUser';
import { useRealtime } from './realtime';

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
        qc.invalidateQueries({ queryKey: ['flow-notifications'] });
        return;
      }
      if (e.actor_id && currentUserId && String(e.actor_id) === currentUserId) {
        return;
      }
      if (e.type.startsWith('project.')) {
        qc.invalidateQueries({ queryKey: ['projects'] });
        qc.invalidateQueries({ queryKey: ['projects-v2'] });
        // archived / restored / purged move a project between the list and the
        // recycle bin, so the bin's own query has to refetch too.
        qc.invalidateQueries({ queryKey: ['archived-projects'] });
      }
    },
    [qc, currentUserId]
  );

  const topics = useMemo(() => {
    const t: string[] = [];
    if (businessId) {
      t.push(`business:${businessId}`);
    }
    if (userId) t.push(`user:${userId}`);
    return t;
  }, [businessId, userId]);

  useRealtime(enabled ? topics : [], onEvent);
}
