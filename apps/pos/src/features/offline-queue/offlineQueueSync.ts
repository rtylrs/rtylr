import { createOneKitClient } from '@voxire/api';
import { OrderServiceClient } from '@voxire/voxie/pos/orders/v1/client';
import type { CreateOrderRequest } from '@voxire/voxie/pos/orders/v1/types';
import { useOfflineQueueStore } from '../../shared/stores/offlineQueueStore';
import { syncQueuedOrders } from './syncOfflineQueue';

type SyncOrder = (order: CreateOrderRequest) => Promise<void>;

const orderClient = createOneKitClient(OrderServiceClient, 'pos');

const syncOrderWithApi: SyncOrder = async (order) => {
  await orderClient.createOrder(order);
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
