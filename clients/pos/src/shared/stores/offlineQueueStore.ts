import { create } from 'zustand';
import { persist } from 'zustand/middleware';
import type { OfflineQueuedOrder, OfflineTender } from '../types/pos';
import type { CreateOrderRequest } from '@voxire/voxie/pos/orders/v1/types';

/**
 * What checkout already knows about the sale when it hands it to the queue.
 * Both halves must land in the SAME write as the entry: a queue entry with no
 * tender is a sale that took cash and can never be charged, and one with no key
 * is a charge the server refuses.
 */
export interface OfflineQueueEntryMeta {
  tender: OfflineTender;
  idempotencyKey: string;
}

interface OfflineQueueState {
  queue: OfflineQueuedOrder[];
  online: boolean;
}

interface OfflineQueueActions {
  /**
   * `meta` is optional so callers that predate the tender, and entries restored
   * from an older build, still enqueue: the store fills in a placeholder
   * tender the drain then repairs from the payload, and a per-entry key. One
   * entry is one user action, so a per-entry key is the right granularity for
   * everything the drain replays.
   */
  enqueue: (order: CreateOrderRequest, meta?: Partial<OfflineQueueEntryMeta>) => OfflineQueuedOrder;
  setOnline: (online: boolean) => void;
  retry: (id: string) => void;
  markSynced: (id: string) => void;
  markFailed: (id: string, err: string) => void;
  /** Reset entries stranded in 'syncing' by a crash back to 'pending'. */
  requeueStalled: () => void;
  /** Persist the order id from createOrder before the charge is attempted. */
  recordServerOrderId: (id: string, serverOrderId: string) => void;
  /**
   * Complete a pre-upgrade entry in place. The replay calls this before its
   * first send so the key it mints is persisted, and the retry after a failure
   * reuses it instead of minting a fresh one for the same sale.
   */
  ensureReplayable: (id: string, meta: OfflineQueueEntryMeta) => void;
  clear: () => void;
}

export const useOfflineQueueStore = create<OfflineQueueState & OfflineQueueActions>()(
  persist(
    (set) => ({
      queue: [],
      online: typeof navigator !== 'undefined' ? navigator.onLine : true,
      enqueue: (order, meta) => {
        const queued: OfflineQueuedOrder = {
          id: 'oq_' + Math.random().toString(36).slice(2, 10),
          order,
          queuedAt: new Date().toISOString(),
          retries: 0,
          status: 'pending',
          tender: meta?.tender ?? {
            paymentMethod: order.paymentMethod ?? 'cash',
            amountPaid: 0,
            localOrderId: '',
          },
          idempotencyKey: meta?.idempotencyKey ?? crypto.randomUUID(),
        };
        set((s) => ({ queue: [queued, ...s.queue] }));
        return queued;
      },
      setOnline: (online) => set({ online }),
      retry: (id) =>
        set((s) => ({
          queue: s.queue.map((q) =>
            q.id === id ? { ...q, retries: q.retries + 1, status: 'syncing' as const } : q
          ),
        })),
      markSynced: (id) => set((s) => ({ queue: s.queue.filter((q) => q.id !== id) })),
      markFailed: (id, err) =>
        set((s) => ({
          queue: s.queue.map((q) =>
            q.id === id ? { ...q, status: 'failed' as const, lastError: err } : q
          ),
        })),
      requeueStalled: () =>
        set((s) => ({
          queue: s.queue.map((q) => (q.status === 'syncing' ? { ...q, status: 'pending' as const } : q)),
        })),
      recordServerOrderId: (id, serverOrderId) =>
        set((s) => ({
          queue: s.queue.map((q) => (q.id === id ? { ...q, serverOrderId } : q)),
        })),
      ensureReplayable: (id, meta) =>
        set((s) => ({
          queue: s.queue.map((q) =>
            q.id === id ? { ...q, tender: meta.tender, idempotencyKey: meta.idempotencyKey } : q
          ),
        })),
      clear: () => set({ queue: [] }),
    }),
    { name: 'pos-offline-queue' }
  )
);

// Initialize online listeners once at module load (safe in Tauri webview).
// On reconnect we flip the `online` flag AND fire a window event so any
// consumer wired to drain the queue gets woken up. Before this the
// queue would sit indefinitely after a reconnect (audit P0 #4).
//
// The actual API replay lives outside this store so it stays
// dependency-free; the listener here is just the trigger surface. Wire
// the replay against the `rtylr-offline-queue-reconnected` event.
if (typeof window !== 'undefined') {
  window.addEventListener('online', () => {
    useOfflineQueueStore.getState().setOnline(true);
    const pending = useOfflineQueueStore.getState().queue.filter((q) => q.status === 'pending');
    if (pending.length > 0) {
      window.dispatchEvent(
        new CustomEvent<{ count: number }>('rtylr-offline-queue-reconnected', {
          detail: { count: pending.length },
        })
      );
    }
  });
  window.addEventListener('offline', () => useOfflineQueueStore.getState().setOnline(false));
}
