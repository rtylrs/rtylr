import { useCallback } from 'react';
import { useOfflineQueueStore } from '../../shared/stores/offlineQueueStore';
import { posService } from '../../shared/services/posService';

export function useOfflineQueue() {
  const store = useOfflineQueueStore();

  const flush = useCallback(async () => {
    const { queue } = useOfflineQueueStore.getState();
    for (const entry of queue) {
      try {
        useOfflineQueueStore.getState().retry(entry.id);
        await posService.createOrder(entry.order);
        useOfflineQueueStore.getState().markSynced(entry.id);
      } catch (e: any) {
        useOfflineQueueStore
          .getState()
          .markFailed(entry.id, e?.response?.data?.message ?? e?.message ?? 'Sync failed');
      }
    }
  }, []);

  const retryOne = useCallback(async (id: string) => {
    const entry = useOfflineQueueStore.getState().queue.find((q) => q.id === id);
    if (!entry) return false;
    useOfflineQueueStore.getState().retry(id);
    try {
      await posService.createOrder(entry.order);
      useOfflineQueueStore.getState().markSynced(id);
      return true;
    } catch (e: any) {
      useOfflineQueueStore
        .getState()
        .markFailed(id, e?.response?.data?.message ?? e?.message ?? 'Sync failed');
      return false;
    }
  }, []);

  return {
    queue: store.queue,
    online: store.online,
    enqueue: store.enqueue,
    flush,
    retry: retryOne,
    clear: store.clear,
  };
}
