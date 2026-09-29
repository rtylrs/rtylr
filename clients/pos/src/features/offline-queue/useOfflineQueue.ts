import { useCallback } from 'react';
import { useOfflineQueueStore } from '../../shared/stores/offlineQueueStore';
import { useSessionCloseQueueStore } from '../../shared/stores/sessionCloseQueueStore';
import { useCashMovementQueueStore } from '../../shared/stores/cashMovementQueueStore';
import { syncQueuedOrders } from './syncOfflineQueue';
import { replayQueuedOrder } from './offlineQueueSync';
import {
  retrySessionClose,
  syncAllSessionCloses,
  syncPendingSessionCloses,
} from '../close-session/sessionCloseQueueSync';
import { retryCashMovement, syncAllCashMovements } from '../cash-drop/cashMovementQueueSync';

export function useOfflineQueue() {
  const store = useOfflineQueueStore();
  const sessionCloseStore = useSessionCloseQueueStore();
  const cashMovementStore = useCashMovementQueueStore();

  const flush = useCallback(async () => {
    const ordersResult = await (async () => {
      const { retry, markSynced, markFailed } = useOfflineQueueStore.getState();
      const entries = useOfflineQueueStore.getState().queue;
      return syncQueuedOrders({
        entries,
        syncOrder: replayQueuedOrder,
        retry,
        markSynced,
        markFailed,
        readEntry: (id) => useOfflineQueueStore.getState().queue.find((q) => q.id === id),
      });
    })();
    const sessionResult = await syncAllSessionCloses();
    const cashResult = await syncAllCashMovements();
    return {
      attempted: ordersResult.attempted + sessionResult.attempted + cashResult.attempted,
      synced: ordersResult.synced + sessionResult.synced + cashResult.synced,
      failed: ordersResult.failed + sessionResult.failed + cashResult.failed,
    };
  }, []);

  const retryOne = useCallback(async (id: string) => {
    const entry = useOfflineQueueStore.getState().queue.find((q) => q.id === id);
    if (!entry) return false;
    const { retry, markSynced, markFailed } = useOfflineQueueStore.getState();
    const result = await syncQueuedOrders({
      entries: [entry],
      syncOrder: replayQueuedOrder,
      retry,
      markSynced,
      markFailed,
    });
    return result.synced === 1;
  }, []);

  return {
    queue: store.queue,
    sessionCloseQueue: sessionCloseStore.queue,
    cashMovementQueue: cashMovementStore.queue,
    totalQueued: store.queue.length + sessionCloseStore.queue.length + cashMovementStore.queue.length,
    online: store.online,
    enqueue: store.enqueue,
    flush,
    retry: retryOne,
    retrySessionClose,
    retryCashMovement,
    // Discarding the orders discards the tenders and keys that would have
    // settled them, in the same write.
    clear: () => {
      store.clear();
      sessionCloseStore.clear();
      cashMovementStore.clear();
    },
  };
}
