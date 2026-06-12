import { useEffect } from 'react';
import type { RealtimeEvent } from '@voxire/api';
import { posRealtime } from '../../shared/services/api/realtime';
import { KDS_EVENT_TYPES } from '../../shared/services/api/realtimeEvents';

export function useKdsStream(ownerID: number | undefined, onInvalidate: () => void): void {
  useEffect(() => {
    if (!ownerID) return;
    const topic = `owner:${ownerID}`;

    const handler = (e: RealtimeEvent): void => {
      if (KDS_EVENT_TYPES.has(e.type as never)) {
        onInvalidate();
      }
    };

    const unsub = posRealtime.subscribe(topic, handler);
    return (): void => {
      unsub();
    };
  }, [ownerID, onInvalidate]);
}
