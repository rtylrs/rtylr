import { unwrap } from '@voxire/api';
import { posService } from '../../shared/services/posService';
import type { CashMovementPayload } from '../../shared/services/posService';
import {
  useCashMovementQueueStore,
  type QueuedCashMovementAction,
} from '../../shared/stores/cashMovementQueueStore';
import { syncQueuedCashMovements } from './syncCashMovementQueue';

type SyncMovement = (
  action: QueuedCashMovementAction,
  payload: CashMovementPayload
) => Promise<void>;

export const syncCashMovementWithApi: SyncMovement = async (action, payload) => {
  if (action === 'drop') {
    await unwrap(posService.cashDrop(payload));
    return;
  }

  if (action === 'paid_out') {
    await unwrap(posService.cashPickup(payload));
    return;
  }

  await unwrap(posService.cashAdjust(payload));
};

export async function syncPendingCashMovements(
  syncMovement: SyncMovement = syncCashMovementWithApi
): Promise<{ attempted: number; synced: number; failed: number }> {
  const { queue, retry, markSynced, markFailed } = useCashMovementQueueStore.getState();
  const pending = queue.filter((entry) => entry.status === 'pending');

  return syncQueuedCashMovements({ entries: pending, syncMovement, retry, markSynced, markFailed });
}

export async function syncAllCashMovements(
  syncMovement: SyncMovement = syncCashMovementWithApi
): Promise<{ attempted: number; synced: number; failed: number }> {
  const { queue, retry, markSynced, markFailed } = useCashMovementQueueStore.getState();

  return syncQueuedCashMovements({ entries: queue, syncMovement, retry, markSynced, markFailed });
}

export async function retryCashMovement(
  id: string,
  syncMovement: SyncMovement = syncCashMovementWithApi
): Promise<boolean> {
  const entry = useCashMovementQueueStore.getState().queue.find((item) => item.id === id);
  if (!entry) return false;

  const { retry, markSynced, markFailed } = useCashMovementQueueStore.getState();
  const result = await syncQueuedCashMovements({
    entries: [entry],
    syncMovement,
    retry,
    markSynced,
    markFailed,
  });

  return result.synced === 1;
}
