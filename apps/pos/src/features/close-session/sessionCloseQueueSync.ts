import { createOneKitClient } from '@voxire/api';
import { SessionServiceClient } from '@voxire/voxie/pos/session/v1/client';
import { useSessionCloseQueueStore } from '../../shared/stores/sessionCloseQueueStore';
import { resolveOpenSessionId, type SessionClosePayload } from './closeSessionWorkflow';
import { syncQueuedSessionCloses } from './syncSessionCloseQueue';
import { runExclusiveDrain } from '../offline-queue/drainGuard';

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
  return runExclusiveDrain('session-closes', () => {
    const store = useSessionCloseQueueStore.getState();
    // Recover entries stranded in 'syncing' by a crash before filtering.
    store.requeueStalled();
    const pending = useSessionCloseQueueStore
      .getState()
      .queue.filter((entry) => entry.status === 'pending');

    return syncQueuedSessionCloses({
      entries: pending,
      syncClose,
      retry: store.retry,
      markSynced: store.markSynced,
      markFailed: store.markFailed,
      readEntry: (id) => useSessionCloseQueueStore.getState().queue.find((q) => q.id === id),
    });
  });
}

export async function syncAllSessionCloses(
  syncClose: SyncClose = syncCloseWithApi
): Promise<{ attempted: number; synced: number; failed: number }> {
  return runExclusiveDrain('session-closes', () => {
    const store = useSessionCloseQueueStore.getState();
    store.requeueStalled();

    return syncQueuedSessionCloses({
      entries: useSessionCloseQueueStore.getState().queue,
      syncClose,
      retry: store.retry,
      markSynced: store.markSynced,
      markFailed: store.markFailed,
      readEntry: (id) => useSessionCloseQueueStore.getState().queue.find((q) => q.id === id),
    });
  });
}

export async function retrySessionClose(
  id: string,
  syncClose: SyncClose = syncCloseWithApi
): Promise<boolean> {
  const store = useSessionCloseQueueStore.getState();
  const entry = store.queue.find((item) => item.id === id);
  if (!entry) return false;

  const result = await syncQueuedSessionCloses({
    entries: [entry],
    syncClose,
    retry: store.retry,
    markSynced: store.markSynced,
    markFailed: store.markFailed,
    readEntry: (entryId) => useSessionCloseQueueStore.getState().queue.find((q) => q.id === entryId),
  });

  return result.synced === 1;
}
