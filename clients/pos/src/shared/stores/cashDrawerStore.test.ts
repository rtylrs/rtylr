import { describe, it, expect, beforeEach } from 'vitest';
import { useCashDrawerStore } from './cashDrawerStore';

const { getState, setState } = useCashDrawerStore;
const act = (fn: () => void) => {
  fn();
};

beforeEach(() => {
  setState({ openingFloat: 0, movements: [], opened: false });
});

describe('cashDrawerStore', () => {
  // ── Initial state ──────────────────────────────────────────────

  it('starts closed with zero float and no movements', () => {
    expect(getState().opened).toBe(false);
    expect(getState().openingFloat).toBe(0);
    expect(getState().movements).toEqual([]);
  });

  // ── openDrawer ─────────────────────────────────────────────────

  it('opens the drawer with a float', () => {
    act(() => getState().openDrawer(200));
    expect(getState().opened).toBe(true);
    expect(getState().openingFloat).toBe(200);
    expect(getState().movements).toEqual([]);
  });

  it('normalizes invalid opening floats to zero', () => {
    act(() => getState().openDrawer(-50));
    expect(getState().openingFloat).toBe(0);

    act(() => getState().openDrawer(Number.NaN));
    expect(getState().openingFloat).toBe(0);
  });

  it('opening the drawer clears previous movements', () => {
    act(() => {
      getState().openDrawer(100);
      getState().paidIn(50, 'change');
      getState().openDrawer(200);
    });
    expect(getState().movements).toEqual([]);
    expect(getState().openingFloat).toBe(200);
  });

  // ── closeDrawer ────────────────────────────────────────────────

  it('closes the drawer', () => {
    act(() => {
      getState().openDrawer(100);
      getState().closeDrawer();
    });
    expect(getState().opened).toBe(false);
  });

  // ── addMovement ────────────────────────────────────────────────

  it('adds a movement with generated id and timestamp', () => {
    act(() => getState().openDrawer(100));
    const movement = getState().addMovement({ type: 'float', amount: 50 });

    expect(movement.id).toMatch(/^mv_/);
    expect(movement.at).toBeTruthy();
    expect(movement.type).toBe('float');
    expect(movement.amount).toBe(50);
    expect(getState().movements).toHaveLength(1);
  });

  it('normalizes invalid movement amounts to zero', () => {
    act(() => getState().openDrawer(100));
    const negative = getState().addMovement({ type: 'paid_in', amount: -25 });
    const nonFinite = getState().addMovement({ type: 'paid_out', amount: Number.POSITIVE_INFINITY });

    expect(negative.amount).toBe(0);
    expect(nonFinite.amount).toBe(0);
    expect(getState().expectedCash()).toBe(100);
  });

  it('prepends new movements (newest first)', () => {
    act(() => getState().openDrawer(100));
    const first = getState().addMovement({ type: 'paid_in', amount: 10, reason: 'first' });
    const second = getState().addMovement({ type: 'paid_in', amount: 20, reason: 'second' });

    expect(getState().movements[0].reason).toBe('second');
    expect(getState().movements[1].reason).toBe('first');
  });

  // ── cashDrop ───────────────────────────────────────────────────

  it('records a cash drop', () => {
    act(() => {
      getState().openDrawer(500);
      getState().cashDrop(200, 'Safe deposit');
    });

    expect(getState().movements).toHaveLength(1);
    expect(getState().movements[0].type).toBe('drop');
    expect(getState().movements[0].amount).toBe(200);
    expect(getState().movements[0].reason).toBe('Safe deposit');
  });

  // ── paidOut ────────────────────────────────────────────────────

  it('records a paid out', () => {
    act(() => {
      getState().openDrawer(500);
      getState().paidOut(30, 'Supplier refund');
    });

    expect(getState().movements[0].type).toBe('paid_out');
    expect(getState().movements[0].amount).toBe(30);
  });

  // ── paidIn ─────────────────────────────────────────────────────

  it('records a paid in', () => {
    act(() => {
      getState().openDrawer(100);
      getState().paidIn(25, 'Extra change');
    });

    expect(getState().movements[0].type).toBe('paid_in');
    expect(getState().movements[0].amount).toBe(25);
  });

  it('records a cash sale separately from manual paid in', () => {
    act(() => {
      getState().openDrawer(100);
      getState().recordCashSale(42.5, 'order-123');
    });

    expect(getState().movements[0].type).toBe('sale');
    expect(getState().movements[0].amount).toBe(42.5);
    expect(getState().movements[0].reason).toBe('order-123');
  });

  it('records a cash refund separately from manual paid out', () => {
    act(() => {
      getState().openDrawer(100);
      getState().recordCashRefund(12, 'order-123');
    });

    expect(getState().movements[0].type).toBe('refund');
    expect(getState().movements[0].amount).toBe(12);
    expect(getState().movements[0].reason).toBe('order-123');
  });

  // ── expectedCash ───────────────────────────────────────────────

  it('calculates expected cash with no movements', () => {
    act(() => getState().openDrawer(300));
    expect(getState().expectedCash()).toBe(300);
  });

  it('calculates expected cash with paid_in', () => {
    act(() => {
      getState().openDrawer(100);
      getState().paidIn(50, 'cash sale');
    });
    expect(getState().expectedCash()).toBe(150);
  });

  it('calculates expected cash with paid_out and drop', () => {
    act(() => {
      getState().openDrawer(500);
      getState().paidOut(100, 'vendor');
      getState().cashDrop(200, 'safe');
    });
    // 500 - 100 - 200 = 200
    expect(getState().expectedCash()).toBe(200);
  });

  it('calculates expected cash with mixed movements', () => {
    act(() => {
      getState().openDrawer(300);
      getState().recordCashSale(100, 'sale-1');
      getState().paidOut(50, 'petty cash');
      getState().cashDrop(150, 'safe');
      getState().recordCashSale(75, 'sale-2');
    });
    // 300 + 100 - 50 - 150 + 75 = 275
    expect(getState().expectedCash()).toBe(275);
  });

  it('subtracts cash refunds from expected cash', () => {
    act(() => {
      getState().openDrawer(300);
      getState().recordCashSale(100, 'sale-1');
      getState().recordCashRefund(25, 'sale-1');
    });

    expect(getState().expectedCash()).toBe(375);
  });

  // ── Edge cases ─────────────────────────────────────────────────

  it('expectedCash returns 0 when drawer not opened', () => {
    expect(getState().expectedCash()).toBe(0);
  });

  it('handles zero-amount movements', () => {
    act(() => {
      getState().openDrawer(100);
      getState().paidIn(0, 'no-op');
    });
    expect(getState().expectedCash()).toBe(100);
  });
});
