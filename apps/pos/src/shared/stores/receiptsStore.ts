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

function textOrNull(value: string | undefined): string | undefined {
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

// Full receipt payloads persisted to one localStorage value: without a cap a
// long-lived terminal accumulates thousands of receipts and every write
// re-serializes the whole history.
const MAX_PERSISTED_RECEIPTS = 200;

function normalizeReceipt(receipt: Receipt): Receipt | null {
  const id = requiredText(receipt.id);
  const orderId = requiredText(receipt.orderId);
  if (!id || !orderId) return null;

  return {
    ...receipt,
    id,
    orderId,
    total: nonNegativeNumber(receipt.total),
    customerName: textOrNull(receipt.customerName),
    items: receipt.items.map((item) => ({
      name: item.name.trim(),
      qty: positiveInteger(item.qty),
      price: nonNegativeNumber(item.price),
    })),
    payments: receipt.payments.map((payment) => ({
      ...payment,
      id: payment.id.trim(),
      amount: nonNegativeNumber(payment.amount),
      tipAmount: payment.tipAmount === undefined ? undefined : nonNegativeNumber(payment.tipAmount),
    })),
  };
}

export const useReceiptsStore = create<ReceiptsState & ReceiptsActions>()(
  persist(
    (set, get) => ({
      receipts: [],
      add: (receipt) => {
        const normalized = normalizeReceipt(receipt);
        if (!normalized) return;
        set((s) => ({
          receipts: [normalized, ...s.receipts].slice(0, MAX_PERSISTED_RECEIPTS),
        }));
      },
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
