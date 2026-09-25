import type { SessionClosePayload } from './closeSessionWorkflow';
import type { QueuedSessionClose } from '../../shared/stores/sessionCloseQueueStore';
import { getOneKitErrorMessage } from '@voxire/api';

interface SyncQueuedSessionClosesParams {
  entries: QueuedSessionClose[];
  syncClose: (payload: SessionClosePayload) => Promise<void>;
  retry: (id: string) => void;
  markSynced: (id: string) => void;
  markFailed: (id: string, error: string) => void;
  /** Fresh store read so claims are checked against live state, not a
   *  snapshot that may already have been claimed by another runner. */
  readEntry?: (id: string) => QueuedSessionClose | undefined;
}

interface SyncQueuedSessionClosesResult {
  attempted: number;
  synced: number;
  failed: number;
}

export async function syncQueuedSessionCloses({
  entries,
  syncClose,
  retry,
  markSynced,
  markFailed,
  readEntry,
}: SyncQueuedSessionClosesParams): Promise<SyncQueuedSessionClosesResult> {
  const result: SyncQueuedSessionClosesResult = { attempted: 0, synced: 0, failed: 0 };

  for (const entry of entries) {
    const current = readEntry ? readEntry(entry.id) : entry;
    if (!current || current.status === 'syncing') continue;

    result.attempted += 1;
    retry(current.id);

    try {
      await syncClose(current.payload);
      markSynced(current.id);
      result.synced += 1;
    } catch (error) {
      markFailed(current.id, getOneKitErrorMessage(error, 'Sync failed'));
      result.failed += 1;
    }
  }

  return result;
}
