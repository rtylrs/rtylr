import { useEffect, useRef } from 'react';
import type { RealtimeEvent } from '@voxire/api';
import { posRealtime } from '../../shared/services/api/realtime';
import { KDS_EVENT_TYPES } from '../../shared/services/api/realtimeEvents';

const INVALIDATION_COALESCE_MS = 300;

export function useKdsStream(ownerID: number | undefined, onInvalidate: () => void): void {
  // A burst of realtime events should trigger ONE refetch, not one per event.
  const timerRef = useRef<ReturnType<typeof setTimeout> | null>(null);
  const pendingRef = useRef(false);

  useEffect(() => {
    if (!ownerID) return;
    const topic = `owner:${ownerID}`;

    const flush = (): void => {
      timerRef.current = null;
      if (pendingRef.current) {
        pendingRef.current = false;
        onInvalidate();
      }
    };

    const handler = (e: RealtimeEvent): void => {
      if (!KDS_EVENT_TYPES.has(e.type as never)) return;
      pendingRef.current = true;
      if (timerRef.current === null) {
        timerRef.current = setTimeout(flush, INVALIDATION_COALESCE_MS);
      }
    };

    const unsub = posRealtime.subscribe(topic, handler);
    return (): void => {
      unsub();
      if (timerRef.current !== null) {
        clearTimeout(timerRef.current);
        timerRef.current = null;
      }
      pendingRef.current = false;
    };
  }, [ownerID, onInvalidate]);
}
