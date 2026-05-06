import { unwrap } from '@voxire/api';
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
        await unwrap(posService.createOrder(entry.order));
        useOfflineQueueStore.getState().markSynced(entry.id);
      } catch (e) {
        const err = e as { response?: { data?: { message?: string } }; message?: string };
        useOfflineQueueStore
          .getState()
          .markFailed(entry.id, err?.response?.data?.message ?? err?.message ?? 'Sync failed');
      }
    }
  }, []);

  const retryOne = useCallback(async (id: string) => {
    const entry = useOfflineQueueStore.getState().queue.find((q) => q.id === id);
    if (!entry) return false;
    useOfflineQueueStore.getState().retry(id);
    try {
      await unwrap(posService.createOrder(entry.order));
      useOfflineQueueStore.getState().markSynced(id);
      return true;
    } catch (e) {
      const err = e as { response?: { data?: { message?: string } }; message?: string };
      useOfflineQueueStore
        .getState()
        .markFailed(id, err?.response?.data?.message ?? err?.message ?? 'Sync failed');
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
