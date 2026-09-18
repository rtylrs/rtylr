import { create } from 'zustand';
import { persist } from 'zustand/middleware';
import type { SessionClosePayload } from '../../features/close-session/closeSessionWorkflow';

export interface QueuedSessionClose {
  id: string;
  payload: SessionClosePayload;
  queuedAt: string;
  retries: number;
  lastError?: string;
  status: 'pending' | 'syncing' | 'failed';
}

interface SessionCloseQueueState {
  queue: QueuedSessionClose[];
}

interface SessionCloseQueueActions {
  enqueue: (payload: SessionClosePayload) => QueuedSessionClose;
  retry: (id: string) => void;
  markSynced: (id: string) => void;
  markFailed: (id: string, error: string) => void;
  /** Reset entries stranded in 'syncing' by a crash back to 'pending'. */
  requeueStalled: () => void;
  clear: () => void;
}

export const useSessionCloseQueueStore = create<
  SessionCloseQueueState & SessionCloseQueueActions
>()(
  persist(
    (set) => ({
      queue: [],
      enqueue: (payload) => {
        const queued: QueuedSessionClose = {
          id: 'sc_' + Math.random().toString(36).slice(2, 10),
          payload,
          queuedAt: new Date().toISOString(),
          retries: 0,
          status: 'pending',
        };
        set((state) => ({ queue: [queued, ...state.queue] }));
        return queued;
      },
      retry: (id) =>
        set((state) => ({
          queue: state.queue.map((item) =>
            item.id === id
              ? { ...item, retries: item.retries + 1, status: 'syncing' as const }
              : item
          ),
        })),
      markSynced: (id) =>
        set((state) => ({ queue: state.queue.filter((item) => item.id !== id) })),
      markFailed: (id, error) =>
        set((state) => ({
          queue: state.queue.map((item) =>
            item.id === id ? { ...item, status: 'failed' as const, lastError: error } : item
          ),
        })),
      requeueStalled: () =>
        set((state) => ({
          queue: state.queue.map((item) =>
            item.status === 'syncing' ? { ...item, status: 'pending' as const } : item
          ),
        })),
      clear: () => set({ queue: [] }),
    }),
    { name: 'pos-session-close-queue' }
  )
);
