import { useCallback } from 'react';
import { useQueryClient } from '@tanstack/react-query';
import type { RealtimeEvent } from '@voxire/api';
import { useUserStore } from '../stores/userStore';
import { useRealtime } from './realtime';

/**
 * Subscribe to real-time notification events for the current user.
 * When the server publishes a `notification.created` event on `user:{id}`,
 * this hook invalidates the `['notifications']` query so the badge and
 * list refresh immediately without polling.
 *
 * Mount once at the app level (e.g. inside BusinessRoute or the Dash
 * layout) — the underlying WS connection is a singleton shared by all
 * subscribers.
 */
export function useNotificationsRealtime(): void {
  const qc = useQueryClient();
  const { user } = useUserStore();
  const userId = user?.id ? String(user.id) : null;

  const topic = userId ? `user:${userId}` : '';
  const topics = topic ? [topic] : [];

  const handleEvent = useCallback(
    (event: RealtimeEvent): void => {
      if (event.type === 'notification.created') {
        void qc.invalidateQueries({ queryKey: ['notifications'] });
      }
    },
    [qc]
  );

  useRealtime(topics, handleEvent);
}
