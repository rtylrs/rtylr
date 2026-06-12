import { unwrap } from '@voxire/api';
import { posService } from '../../shared/services/api/posService';
import type { CashMovementPayload } from '../../shared/services/api/posService';
import {
  useCashMovementQueueStore,
  type QueuedCashMovementAction,
} from '../../shared/stores/cashMovementQueueStore';
import { resolveOpenSessionId } from '../close-session/closeSessionWorkflow';
import { syncQueuedCashMovements } from './syncCashMovementQueue';

type SyncMovement = (
  action: QueuedCashMovementAction,
  payload: CashMovementPayload
) => Promise<void>;

export const syncCashMovementWithApi: SyncMovement = async (action, payload) => {
  const body: CashMovementPayload =
    payload.session_id > 0 ? payload : { ...payload, session_id: await resolveOpenSessionId() };

  if (action === 'drop') {
    await unwrap(posService.cashDrop(body));
    return;
  }

  if (action === 'paid_out') {
    await unwrap(posService.cashPickup(body));
    return;
  }

  await unwrap(posService.cashAdjust(body));
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
