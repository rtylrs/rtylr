import type {
  QueuedCashMovement,
  QueuedCashMovementAction,
} from '../../shared/stores/cashMovementQueueStore';
import type { CashMovementPayload } from '../../shared/services/api/posService';

interface SyncQueuedCashMovementsParams {
  entries: QueuedCashMovement[];
  syncMovement: (action: QueuedCashMovementAction, payload: CashMovementPayload) => Promise<void>;
  retry: (id: string) => void;
  markSynced: (id: string) => void;
  markFailed: (id: string, error: string) => void;
}

interface SyncQueuedCashMovementsResult {
  attempted: number;
  synced: number;
  failed: number;
}

function errorMessage(error: unknown): string {
  const err = error as { response?: { data?: { message?: string } }; message?: string };
  return err?.response?.data?.message ?? err?.message ?? 'Sync failed';
}

export async function syncQueuedCashMovements({
  entries,
  syncMovement,
  retry,
  markSynced,
  markFailed,
}: SyncQueuedCashMovementsParams): Promise<SyncQueuedCashMovementsResult> {
  const result: SyncQueuedCashMovementsResult = { attempted: 0, synced: 0, failed: 0 };

  for (const entry of entries) {
    if (entry.status === 'syncing') continue;

    result.attempted += 1;
    retry(entry.id);

    try {
      await syncMovement(entry.action, entry.payload);
      markSynced(entry.id);
      result.synced += 1;
    } catch (error) {
      markFailed(entry.id, errorMessage(error));
      result.failed += 1;
    }
  }

  return result;
}
