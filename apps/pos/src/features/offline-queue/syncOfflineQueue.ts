import type { OfflineQueuedOrder } from '../../shared/types/pos';
import type { CreateOrderRequest } from '@voxire/voxie/pos/orders/v1/types';
import { getOneKitErrorMessage } from '@voxire/api';

interface SyncQueuedOrdersParams {
  entries: OfflineQueuedOrder[];
  syncOrder: (order: CreateOrderRequest) => Promise<void>;
  retry: (id: string) => void;
  markSynced: (id: string) => void;
  markFailed: (id: string, err: string) => void;
  /** Fresh store read so claims are checked against live state, not a
   *  snapshot that may already have been claimed by another runner. */
  readEntry?: (id: string) => OfflineQueuedOrder | undefined;
}

interface SyncQueuedOrdersResult {
  attempted: number;
  synced: number;
  failed: number;
}

export async function syncQueuedOrders({
  entries,
  syncOrder,
  retry,
  markSynced,
  markFailed,
  readEntry,
}: SyncQueuedOrdersParams): Promise<SyncQueuedOrdersResult> {
  const result: SyncQueuedOrdersResult = { attempted: 0, synced: 0, failed: 0 };

  for (const entry of entries) {
    const current = readEntry ? readEntry(entry.id) : entry;
    if (!current || current.status === 'syncing') continue;

    result.attempted += 1;
    retry(current.id);

    try {
      await syncOrder(current.order);
      markSynced(current.id);
      result.synced += 1;
    } catch (error) {
      markFailed(current.id, getOneKitErrorMessage(error, 'Sync failed'));
      result.failed += 1;
    }
  }

  return result;
}
