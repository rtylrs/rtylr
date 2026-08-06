// Pure helpers for managing offline-queued orders.

import type { OfflineQueuedOrder } from '../types/pos';

/** Return true if the queued order is eligible for a retry attempt. */
export function isRetryable(entry: OfflineQueuedOrder, maxRetries = 3): boolean {
  return entry.status === 'failed' && entry.retries < maxRetries;
}

/** Return true if the queued order has permanently failed (exceeded retry limit). */
export function isPermanentlyFailed(entry: OfflineQueuedOrder, maxRetries = 3): boolean {
  return entry.status === 'failed' && entry.retries >= maxRetries;
}

/** Return true if the queued order is still pending (not yet attempted). */
export function isPending(entry: OfflineQueuedOrder): boolean {
  return entry.status === 'pending';
}

/** Return true if the queued order is currently being synced. */
export function isSyncing(entry: OfflineQueuedOrder): boolean {
  return entry.status === 'syncing';
}

/** Count entries by status in the offline queue. */
export function queueStatusCounts(
  entries: OfflineQueuedOrder[],
): Record<OfflineQueuedOrder['status'], number> {
  const counts: Record<OfflineQueuedOrder['status'], number> = {
    pending: 0,
    syncing: 0,
    failed: 0,
  };
  for (const e of entries) {
    counts[e.status] += 1;
  }
  return counts;
}

/** Return entries sorted oldest-queued-first. */
export function sortQueueOldestFirst(entries: OfflineQueuedOrder[]): OfflineQueuedOrder[] {
  return [...entries].sort(
    (a, b) => new Date(a.queuedAt).getTime() - new Date(b.queuedAt).getTime(),
  );
}

/** Return only the retryable entries from the queue. */
export function retryableEntries(
  entries: OfflineQueuedOrder[],
  maxRetries = 3,
): OfflineQueuedOrder[] {
  return entries.filter((e) => isRetryable(e, maxRetries));
}

/** Return total number of items across all queued orders. */
export function totalQueuedItemCount(entries: OfflineQueuedOrder[]): number {
  return entries.reduce((sum, e) => sum + (e.order.items ?? []).length, 0);
}
