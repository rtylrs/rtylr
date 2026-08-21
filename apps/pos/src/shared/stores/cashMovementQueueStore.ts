import { create } from 'zustand';
import { persist } from 'zustand/middleware';

export type QueuedCashMovementAction = 'drop' | 'paid_out' | 'paid_in';

export interface CashMovementPayload {
  branch_id: number;
  terminal_id?: number | null;
  session_id: number;
  employee_id: number;
  amount: number;
  reason?: string | null;
}

export interface QueuedCashMovement {
  id: string;
  action: QueuedCashMovementAction;
  payload: CashMovementPayload;
  queuedAt: string;
  retries: number;
  lastError?: string;
  status: 'pending' | 'syncing' | 'failed';
}

interface CashMovementQueueState {
  queue: QueuedCashMovement[];
}

interface CashMovementQueueActions {
  enqueue: (movement: Pick<QueuedCashMovement, 'action' | 'payload'>) => QueuedCashMovement;
  retry: (id: string) => void;
  markSynced: (id: string) => void;
  markFailed: (id: string, error: string) => void;
  /** Reset entries stranded in 'syncing' by a crash back to 'pending'. */
  requeueStalled: () => void;
  clear: () => void;
}

export const useCashMovementQueueStore = create<
  CashMovementQueueState & CashMovementQueueActions
>()(
  persist(
    (set) => ({
      queue: [],
      enqueue: ({ action, payload }) => {
        const queued: QueuedCashMovement = {
          id: 'cm_' + Math.random().toString(36).slice(2, 10),
          action,
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
    { name: 'pos-cash-movement-queue' }
  )
);
