import type {
  CashMovementPayload,
  QueuedCashMovement,
  QueuedCashMovementAction,
} from '../../shared/stores/cashMovementQueueStore';
import { getOneKitErrorMessage } from '@voxire/api';

interface SyncQueuedCashMovementsParams {
  entries: QueuedCashMovement[];
  syncMovement: (action: QueuedCashMovementAction, payload: CashMovementPayload) => Promise<void>;
  retry: (id: string) => void;
  markSynced: (id: string) => void;
  markFailed: (id: string, error: string) => void;
  /** Fresh store read so claims are checked against live state, not a
   *  snapshot that may already have been claimed by another runner. */
  readEntry?: (id: string) => QueuedCashMovement | undefined;
}

interface SyncQueuedCashMovementsResult {
  attempted: number;
  synced: number;
  failed: number;
}

export async function syncQueuedCashMovements({
  entries,
  syncMovement,
  retry,
  markSynced,
  markFailed,
  readEntry,
}: SyncQueuedCashMovementsParams): Promise<SyncQueuedCashMovementsResult> {
  const result: SyncQueuedCashMovementsResult = { attempted: 0, synced: 0, failed: 0 };

  for (const entry of entries) {
    const current = readEntry ? readEntry(entry.id) : entry;
    if (!current || current.status === 'syncing') continue;

    result.attempted += 1;
    retry(current.id);

    try {
      await syncMovement(current.action, current.payload);
      markSynced(current.id);
      result.synced += 1;
    } catch (error) {
      markFailed(current.id, getOneKitErrorMessage(error, 'Sync failed'));
      result.failed += 1;
    }
  }

  return result;
}
