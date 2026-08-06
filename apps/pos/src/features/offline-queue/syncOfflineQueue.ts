import type { OfflineQueuedOrder } from '../../shared/types/pos';
import type { CreateOrderRequest } from '@voxire/voxie/pos/orders/v1/types';

interface SyncQueuedOrdersParams {
  entries: OfflineQueuedOrder[];
  syncOrder: (order: CreateOrderRequest) => Promise<void>;
  retry: (id: string) => void;
  markSynced: (id: string) => void;
  markFailed: (id: string, err: string) => void;
}

interface SyncQueuedOrdersResult {
  attempted: number;
  synced: number;
  failed: number;
}

function errorMessage(error: unknown): string {
  const err = error as { response?: { data?: { message?: string } }; message?: string };
  return err?.response?.data?.message ?? err?.message ?? 'Sync failed';
}

export async function syncQueuedOrders({
  entries,
  syncOrder,
  retry,
  markSynced,
  markFailed,
}: SyncQueuedOrdersParams): Promise<SyncQueuedOrdersResult> {
  const result: SyncQueuedOrdersResult = { attempted: 0, synced: 0, failed: 0 };

  for (const entry of entries) {
    if (entry.status === 'syncing') continue;

    result.attempted += 1;
    retry(entry.id);

    try {
      await syncOrder(entry.order);
      markSynced(entry.id);
      result.synced += 1;
    } catch (error) {
      markFailed(entry.id, errorMessage(error));
      result.failed += 1;
    }
  }

  return result;
}
