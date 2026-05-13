import { useCallback } from 'react';
import { useQueryClient } from '@tanstack/react-query';
import type { RealtimeEvent } from '@voxire/api';
import { useUserStore } from '../stores/userStore';
import { useRealtime } from './realtime';

/**
 * Subscribe to real-time POS order events for the current owner.
 * When the server publishes pos:order:* events on owner:{id}, this hook
 * invalidates the `['dash-orders']` query so the dashboard order count
 * and list refresh immediately.
 */
export function useOrdersRealtime(): void {
  const qc = useQueryClient();
  const { user } = useUserStore();

  const businessId = (user?.businessId ?? user?.business_id) as string | number | undefined;
  const topic = businessId ? `owner:${String(businessId)}` : '';
  const topics = topic ? [topic] : [];

  const handleEvent = useCallback(
    (event: RealtimeEvent): void => {
      if (event.type === 'pos:order:created' || event.type === 'pos:order:preparing') {
        void qc.invalidateQueries({ queryKey: ['dash-orders'] });
      }
    },
    [qc]
  );

  useRealtime(topics, handleEvent);
}
