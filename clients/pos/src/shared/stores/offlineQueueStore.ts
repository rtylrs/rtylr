import { create } from 'zustand';
import { persist } from 'zustand/middleware';
import type { OfflineQueuedOrder } from '../types/pos';
import type { CreateOrderRequest } from '@voxire/voxie/pos/orders/v1/types';

interface OfflineQueueState {
  queue: OfflineQueuedOrder[];
  online: boolean;
}

interface OfflineQueueActions {
  enqueue: (order: CreateOrderRequest) => OfflineQueuedOrder;
  setOnline: (online: boolean) => void;
  retry: (id: string) => void;
  markSynced: (id: string) => void;
  markFailed: (id: string, err: string) => void;
  /** Reset entries stranded in 'syncing' by a crash back to 'pending'. */
  requeueStalled: () => void;
  clear: () => void;
}

export const useOfflineQueueStore = create<OfflineQueueState & OfflineQueueActions>()(
  persist(
    (set) => ({
      queue: [],
      online: typeof navigator !== 'undefined' ? navigator.onLine : true,
      enqueue: (order) => {
        const queued: OfflineQueuedOrder = {
          id: 'oq_' + Math.random().toString(36).slice(2, 10),
          order,
          queuedAt: new Date().toISOString(),
          retries: 0,
          status: 'pending',
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
