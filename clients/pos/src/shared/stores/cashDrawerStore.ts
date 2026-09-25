import { create } from 'zustand';
import { persist } from 'zustand/middleware';

export interface CashMovement {
  id: string;
  type: 'float' | 'drop' | 'paid_out' | 'paid_in' | 'sale' | 'refund';
  amount: number;
  reason?: string;
  employeeName?: string;
  at: string;
}

interface CashDrawerState {
  openingFloat: number;
  movements: CashMovement[];
  opened: boolean;
}

interface CashDrawerActions {
  openDrawer: (openingFloat: number) => void;
  closeDrawer: () => void;
  addMovement: (partial: Omit<CashMovement, 'id' | 'at'>) => CashMovement;
  cashDrop: (amount: number, reason?: string) => void;
  paidOut: (amount: number, reason: string) => void;
  paidIn: (amount: number, reason: string) => void;
  recordCashSale: (amount: number, orderId: string) => void;
  recordCashRefund: (amount: number, orderId: string) => void;
  expectedCash: () => number;
}

function normalizeCashAmount(amount: number): number {
  if (!Number.isFinite(amount) || amount < 0) return 0;
  return amount;
}

export const useCashDrawerStore = create<CashDrawerState & CashDrawerActions>()(
  persist(
    (set, get) => ({
      openingFloat: 0,
      movements: [],
      opened: false,
      openDrawer: (openingFloat) =>
        set({ openingFloat: normalizeCashAmount(openingFloat), opened: true, movements: [] }),
      closeDrawer: () => set({ opened: false }),
      addMovement: (partial) => {
        const movement: CashMovement = {
          ...partial,
          amount: normalizeCashAmount(partial.amount),
          id: 'mv_' + Math.random().toString(36).slice(2, 10),
          at: new Date().toISOString(),
        };
        set((s) => ({ movements: [movement, ...s.movements] }));
        return movement;
      },
      cashDrop: (amount, reason) => {
        get().addMovement({ type: 'drop', amount, reason });
      },
      paidOut: (amount, reason) => {
        get().addMovement({ type: 'paid_out', amount, reason });
      },
      paidIn: (amount, reason) => {
        get().addMovement({ type: 'paid_in', amount, reason });
      },
      recordCashSale: (amount, orderId) => {
        get().addMovement({ type: 'sale', amount, reason: orderId });
      },
      recordCashRefund: (amount, orderId) => {
        get().addMovement({ type: 'refund', amount, reason: orderId });
      },
      expectedCash: () => {
        const s = get();
        const delta = s.movements.reduce((sum, m) => {
          if (m.type === 'paid_in' || m.type === 'sale') return sum + m.amount;
          if (m.type === 'paid_out' || m.type === 'drop' || m.type === 'refund') {
            return sum - m.amount;
          }
          return sum;
        }, 0);
        return s.openingFloat + delta;
      },
    }),
    { name: 'pos-cash-drawer' }
  )
);
