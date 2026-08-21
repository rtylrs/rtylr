import { createOneKitClient } from '@voxire/api';
import { OrderServiceClient } from '@voxire/voxie/pos/orders/v1/client';
import type { CreateOrderRequest } from '@voxire/voxie/pos/orders/v1/types';
import { useOfflineQueueStore } from '../../shared/stores/offlineQueueStore';
import { syncQueuedOrders } from './syncOfflineQueue';
import { runExclusiveDrain } from './drainGuard';

type SyncOrder = (order: CreateOrderRequest) => Promise<void>;

const orderClient = createOneKitClient(OrderServiceClient, 'pos');

const syncOrderWithApi: SyncOrder = async (order) => {
  await orderClient.createOrder(order);
};

export async function syncPendingOfflineOrders(
  syncOrder: SyncOrder = syncOrderWithApi
): Promise<{ attempted: number; synced: number; failed: number }> {
  return runExclusiveDrain('orders', () => {
    const store = useOfflineQueueStore.getState();
    // Recover entries stranded in 'syncing' by a crash before filtering.
    store.requeueStalled();
    const pending = useOfflineQueueStore
      .getState()
      .queue.filter((entry) => entry.status === 'pending');

    return syncQueuedOrders({
      entries: pending,
      syncOrder,
      retry: store.retry,
      markSynced: store.markSynced,
      markFailed: store.markFailed,
      readEntry: (id) => useOfflineQueueStore.getState().queue.find((q) => q.id === id),
    });
  });
}
