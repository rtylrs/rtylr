import { create } from 'zustand';
import { persist } from 'zustand/middleware';
import type { OfflineQueuedOrder, Order } from '../types/pos';

interface OfflineQueueState {
  queue: OfflineQueuedOrder[];
  online: boolean;
}

interface OfflineQueueActions {
  enqueue: (order: Order) => OfflineQueuedOrder;
  setOnline: (online: boolean) => void;
  retry: (id: string) => void;
  markSynced: (id: string) => void;
  markFailed: (id: string, err: string) => void;
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
            q.id === id ? { ...q, retries: q.retries + 1, status: 'syncing' as const } : q,
          ),
        })),
      markSynced: (id) => set((s) => ({ queue: s.queue.filter((q) => q.id !== id) })),
      markFailed: (id, err) =>
        set((s) => ({
          queue: s.queue.map((q) =>
            q.id === id ? { ...q, status: 'failed' as const, lastError: err } : q,
          ),
        })),
      clear: () => set({ queue: [] }),
    }),
    { name: 'pos-offline-queue' },
  ),
);

// Initialize online listeners once at module load (safe in Tauri webview)
if (typeof window !== 'undefined') {
  window.addEventListener('online', () => useOfflineQueueStore.getState().setOnline(true));
  window.addEventListener('offline', () => useOfflineQueueStore.getState().setOnline(false));
}
