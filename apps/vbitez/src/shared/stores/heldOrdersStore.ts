import { create } from 'zustand';
import { persist } from 'zustand/middleware';
import type { OrderLineItem } from '../types/pos';

export interface HeldOrder {
  id: string;
  label: string;
  items: OrderLineItem[];
  tableNumber?: string | null;
  note?: string;
  createdAt: string;
}

interface HeldOrdersState {
  orders: HeldOrder[];
}

interface HeldOrdersActions {
  hold: (partial: Omit<HeldOrder, 'id' | 'createdAt'>) => HeldOrder;
  release: (id: string) => HeldOrder | undefined;
  remove: (id: string) => void;
  updateNote: (id: string, note: string) => void;
}

export const useHeldOrdersStore = create<HeldOrdersState & HeldOrdersActions>()(
  persist(
    (set, get) => ({
      orders: [],
      hold: (partial) => {
        const order: HeldOrder = {
          ...partial,
          id: 'held_' + Math.random().toString(36).slice(2, 10),
          createdAt: new Date().toISOString(),
        };
        set((s) => ({ orders: [order, ...s.orders] }));
        return order;
      },
      release: (id) => {
        const order = get().orders.find((o) => o.id === id);
        if (order) {
          set((s) => ({ orders: s.orders.filter((o) => o.id !== id) }));
        }
        return order;
      },
      remove: (id) => set((s) => ({ orders: s.orders.filter((o) => o.id !== id) })),
      updateNote: (id, note) =>
        set((s) => ({
          orders: s.orders.map((o) => (o.id === id ? { ...o, note } : o)),
        })),
    }),
    { name: 'pos-held-orders' }
  )
);
