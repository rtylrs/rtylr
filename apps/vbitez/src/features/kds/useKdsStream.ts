import { useEffect } from 'react';
import type { RealtimeEvent } from '@voxire/api';
import { posRealtime } from '../../shared/lib/realtime';

const KDS_EVENT_TYPES = new Set([
  'pos:order:created',
  'pos:order:preparing',
  'pos:kds:item:updated',
]);

export function useKdsStream(ownerID: number | undefined, onInvalidate: () => void): void {
  useEffect(() => {
    if (!ownerID) return;
    const topic = `owner:${ownerID}`;

    const handler = (e: RealtimeEvent): void => {
      if (KDS_EVENT_TYPES.has(e.type)) {
        onInvalidate();
      }
    };

    const unsub = posRealtime.subscribe(topic, handler);
    return (): void => {
      unsub();
    };
  }, [ownerID, onInvalidate]);
}
