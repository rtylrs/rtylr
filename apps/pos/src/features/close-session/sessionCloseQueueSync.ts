import { createOneKitClient } from '@voxire/api';
import { SessionServiceClient } from '@voxire/voxie/pos/session/v1/client';
import { useSessionCloseQueueStore } from '../../shared/stores/sessionCloseQueueStore';
import { resolveOpenSessionId, type SessionClosePayload } from './closeSessionWorkflow';
import { syncQueuedSessionCloses } from './syncSessionCloseQueue';

type SyncClose = (payload: SessionClosePayload) => Promise<void>;

const sessionClient = createOneKitClient(SessionServiceClient, 'pos');

const syncCloseWithApi: SyncClose = async (payload) => {
  // Prefer the session captured at enqueue time; re-resolving could close a
  // different terminal's newly-opened session. Legacy entries without an id
  // still fall back to resolution.
  const sessionId =
    payload.sessionId != null && Number.isInteger(payload.sessionId) && payload.sessionId > 0
      ? payload.sessionId
      : await resolveOpenSessionId();
  await sessionClient.closeSession({ id: String(sessionId), closingFloat: payload.closingFloat });
};

export async function syncPendingSessionCloses(
  syncClose: SyncClose = syncCloseWithApi
): Promise<{ attempted: number; synced: number; failed: number }> {
  const { queue, retry, markSynced, markFailed } = useSessionCloseQueueStore.getState();
  const pending = queue.filter((entry) => entry.status === 'pending');

  return syncQueuedSessionCloses({
    entries: pending,
    syncClose,
    retry,
    markSynced,
    markFailed,
  });
}

export async function syncAllSessionCloses(
  syncClose: SyncClose = syncCloseWithApi
): Promise<{ attempted: number; synced: number; failed: number }> {
  const { queue, retry, markSynced, markFailed } = useSessionCloseQueueStore.getState();

  return syncQueuedSessionCloses({
    entries: queue,
    syncClose,
    retry,
    markSynced,
    markFailed,
  });
}

export async function retrySessionClose(
  id: string,
  syncClose: SyncClose = syncCloseWithApi
): Promise<boolean> {
  const entry = useSessionCloseQueueStore.getState().queue.find((item) => item.id === id);
  if (!entry) return false;

  const { retry, markSynced, markFailed } = useSessionCloseQueueStore.getState();
  const result = await syncQueuedSessionCloses({
    entries: [entry],
    syncClose,
    retry,
    markSynced,
    markFailed,
  });

  return result.synced === 1;
}
