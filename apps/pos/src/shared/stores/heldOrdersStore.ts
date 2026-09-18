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
  hold: (partial: Omit<HeldOrder, 'id' | 'createdAt'>) => HeldOrder | undefined;
  release: (id: string) => HeldOrder | undefined;
  remove: (id: string) => void;
  updateNote: (id: string, note: string) => void;
}

function textOrUndefined(value: string | null | undefined): string | undefined {
  const trimmed = value?.trim() ?? '';
  return trimmed || undefined;
}

function requiredText(value: string): string | null {
  const trimmed = value.trim();
  return trimmed || null;
}

function nonNegativeNumber(value: number): number {
  if (!Number.isFinite(value) || value < 0) return 0;
  return value;
}

function positiveInteger(value: number): number {
  if (!Number.isFinite(value) || value <= 0) return 0;
  return Math.floor(value);
}

function normalizeItems(items: OrderLineItem[]): OrderLineItem[] {
  return items.map((item) => ({
    ...item,
    id: item.id.trim(),
    name: item.name.trim(),
    price: nonNegativeNumber(item.price),
    quantity: positiveInteger(item.quantity),
  }));
}

export const useHeldOrdersStore = create<HeldOrdersState & HeldOrdersActions>()(
  persist(
    (set, get) => ({
      orders: [],
      hold: (partial) => {
        const label = requiredText(partial.label);
        if (!label) return undefined;

        const order: HeldOrder = {
          ...partial,
          label,
          items: normalizeItems(partial.items),
          tableNumber: textOrUndefined(partial.tableNumber),
          note: textOrUndefined(partial.note),
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
          orders: s.orders.map((o) => (o.id === id ? { ...o, note: textOrUndefined(note) } : o)),
        })),
    }),
    { name: 'pos-held-orders' }
  )
);
