import { create } from 'zustand';
import { persist } from 'zustand/middleware';

export interface CashMovement {
  id: string;
  type: 'float' | 'drop' | 'paid_out' | 'paid_in';
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
  expectedCash: () => number;
}

export const useCashDrawerStore = create<CashDrawerState & CashDrawerActions>()(
  persist(
    (set, get) => ({
      openingFloat: 0,
      movements: [],
      opened: false,
      openDrawer: (openingFloat) => set({ openingFloat, opened: true, movements: [] }),
      closeDrawer: () => set({ opened: false }),
      addMovement: (partial) => {
        const movement: CashMovement = {
          ...partial,
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
      expectedCash: () => {
        const s = get();
        const delta = s.movements.reduce((sum, m) => {
          if (m.type === 'paid_in') return sum + m.amount;
          if (m.type === 'paid_out' || m.type === 'drop') return sum - m.amount;
          return sum;
        }, 0);
        return s.openingFloat + delta;
      },
    }),
    { name: 'pos-cash-drawer' },
  ),
);
