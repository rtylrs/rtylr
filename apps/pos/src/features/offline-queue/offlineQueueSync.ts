import { unwrap } from '@voxire/api';
import { posService } from '../../shared/services/api/posService';
import { useOfflineQueueStore } from '../../shared/stores/offlineQueueStore';
import type { CreateOrderPayload } from '../../shared/services/api/posService';
import { syncQueuedOrders } from './syncOfflineQueue';

type SyncOrder = (order: CreateOrderPayload) => Promise<void>;

const syncOrderWithApi: SyncOrder = async (order) => {
  await unwrap(posService.createOrder(order));
};

export async function syncPendingOfflineOrders(
  syncOrder: SyncOrder = syncOrderWithApi
): Promise<{ attempted: number; synced: number; failed: number }> {
  const { queue, retry, markSynced, markFailed } = useOfflineQueueStore.getState();
  const pending = queue.filter((entry) => entry.status === 'pending');

  return syncQueuedOrders({
    entries: pending,
    syncOrder,
    retry,
    markSynced,
    markFailed,
  });
}
