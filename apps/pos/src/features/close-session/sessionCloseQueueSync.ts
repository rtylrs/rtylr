import { unwrap } from '@voxire/api';
import { posService } from '../../shared/services/posService';
import type { SessionPayload } from '../../shared/services/posService';
import { useSessionCloseQueueStore } from '../../shared/stores/sessionCloseQueueStore';
import { syncQueuedSessionCloses } from './syncSessionCloseQueue';

type SyncClose = (payload: SessionPayload) => Promise<void>;

const syncCloseWithApi: SyncClose = async (payload) => {
  await unwrap(posService.closeSession(payload));
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
