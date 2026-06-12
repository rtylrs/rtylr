import type { SessionClosePayload } from './closeSessionWorkflow';
import type { QueuedSessionClose } from '../../shared/stores/sessionCloseQueueStore';

interface SyncQueuedSessionClosesParams {
  entries: QueuedSessionClose[];
  syncClose: (payload: SessionClosePayload) => Promise<void>;
  retry: (id: string) => void;
  markSynced: (id: string) => void;
  markFailed: (id: string, error: string) => void;
}

interface SyncQueuedSessionClosesResult {
  attempted: number;
  synced: number;
  failed: number;
}

function errorMessage(error: unknown): string {
  const err = error as { response?: { data?: { message?: string } }; message?: string };
  return err?.response?.data?.message ?? err?.message ?? 'Sync failed';
}

export async function syncQueuedSessionCloses({
  entries,
  syncClose,
  retry,
  markSynced,
  markFailed,
}: SyncQueuedSessionClosesParams): Promise<SyncQueuedSessionClosesResult> {
  const result: SyncQueuedSessionClosesResult = { attempted: 0, synced: 0, failed: 0 };

  for (const entry of entries) {
    if (entry.status === 'syncing') continue;

    result.attempted += 1;
    retry(entry.id);

    try {
      await syncClose(entry.payload);
      markSynced(entry.id);
      result.synced += 1;
    } catch (error) {
      markFailed(entry.id, errorMessage(error));
      result.failed += 1;
    }
  }

  return result;
}
