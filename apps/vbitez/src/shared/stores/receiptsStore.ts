import { create } from 'zustand';
import { persist } from 'zustand/middleware';
import type { Receipt } from '../types/pos';

interface ReceiptsState {
  receipts: Receipt[];
}

interface ReceiptsActions {
  add: (receipt: Receipt) => void;
  markEmailed: (id: string) => void;
  clear: () => void;
  byOrder: (orderId: string) => Receipt | undefined;
}

export const useReceiptsStore = create<ReceiptsState & ReceiptsActions>()(
  persist(
    (set, get) => ({
      receipts: [],
      add: (receipt) => set((s) => ({ receipts: [receipt, ...s.receipts] })),
      markEmailed: (id) =>
        set((s) => ({
          receipts: s.receipts.map((r) => (r.id === id ? { ...r, emailed: true } : r)),
        })),
      clear: () => set({ receipts: [] }),
      byOrder: (orderId) => get().receipts.find((r) => r.orderId === orderId),
    }),
    { name: 'pos-receipts' }
  )
);
