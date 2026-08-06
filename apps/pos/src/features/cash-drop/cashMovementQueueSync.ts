import { createOneKitClient } from '@voxire/api';
import { CashServiceClient } from '@voxire/voxie/pos/cash/v1/client';
import type { CashAdjustRequest, CashDropRequest, CashPickupRequest } from '@voxire/voxie/pos/cash/v1/types';
import {
  useCashMovementQueueStore,
  type CashMovementPayload,
  type QueuedCashMovementAction,
} from '../../shared/stores/cashMovementQueueStore';
import { resolveOpenSessionId } from '../close-session/closeSessionWorkflow';
import { syncQueuedCashMovements } from './syncCashMovementQueue';

type SyncMovement = (
  action: QueuedCashMovementAction,
  payload: CashMovementPayload
) => Promise<void>;

const cashClient = createOneKitClient(CashServiceClient, 'pos');

function id(value: unknown): string {
  return value == null || value === '' ? '0' : String(value);
}

function buildCashRequest(
  payload: CashMovementPayload
): CashDropRequest | CashPickupRequest | CashAdjustRequest {
  return {
    branchId: id(payload.branch_id),
    terminalId: id(payload.terminal_id),
    sessionId: id(payload.session_id),
    employeeId: id(payload.employee_id),
    amount: payload.amount,
    reason: payload.reason ?? '',
  };
}

export const syncCashMovementWithApi: SyncMovement = async (action, payload) => {
  const body: CashMovementPayload =
    payload.session_id > 0 ? payload : { ...payload, session_id: await resolveOpenSessionId() };
  const request = buildCashRequest(body);

  if (action === 'drop') {
    await cashClient.cashDrop(request);
    return;
  }

  if (action === 'paid_out') {
    await cashClient.cashPickup(request);
    return;
  }

  await cashClient.cashAdjust(request);
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
