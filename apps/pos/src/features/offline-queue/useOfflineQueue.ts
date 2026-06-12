import { unwrap } from '@voxire/api';
import { useCallback } from 'react';
import { useOfflineQueueStore } from '../../shared/stores/offlineQueueStore';
import { useSessionCloseQueueStore } from '../../shared/stores/sessionCloseQueueStore';
import { useCashMovementQueueStore } from '../../shared/stores/cashMovementQueueStore';
import { posService } from '../../shared/services/api/posService';
import { syncQueuedOrders } from './syncOfflineQueue';
import {
  retrySessionClose,
  syncAllSessionCloses,
  syncPendingSessionCloses,
} from '../close-session/sessionCloseQueueSync';
import { retryCashMovement, syncAllCashMovements } from '../cash-drop/cashMovementQueueSync';

const syncOrder = async (order: Parameters<typeof posService.createOrder>[0]): Promise<void> => {
  await unwrap(posService.createOrder(order));
};

export function useOfflineQueue() {
  const store = useOfflineQueueStore();
  const sessionCloseStore = useSessionCloseQueueStore();
  const cashMovementStore = useCashMovementQueueStore();

  const flush = useCallback(async () => {
    const { queue, retry, markSynced, markFailed } = useOfflineQueueStore.getState();
    await syncQueuedOrders({ entries: queue, syncOrder, retry, markSynced, markFailed });
    await syncAllSessionCloses();
    await syncAllCashMovements();
  }, []);

  const retryOne = useCallback(async (id: string) => {
    const entry = useOfflineQueueStore.getState().queue.find((q) => q.id === id);
    if (!entry) return false;
    const { retry, markSynced, markFailed } = useOfflineQueueStore.getState();
    const result = await syncQueuedOrders({
      entries: [entry],
      syncOrder,
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
    clear: () => {
      store.clear();
      sessionCloseStore.clear();
      cashMovementStore.clear();
    },
  };
}
