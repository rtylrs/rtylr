import { describe, it, expect, beforeEach } from 'vitest';
import { useReceiptsStore } from './receiptsStore';
import type { Receipt } from '../types/pos';

const { getState, setState } = useReceiptsStore;
const act = (fn: () => void) => {
  fn();
};

const makeReceipt = (overrides: Partial<Receipt> = {}): Receipt => ({
  id: 'r_' + Math.random().toString(36).slice(2, 8),
  orderId: 'ord_' + Math.random().toString(36).slice(2, 8),
  total: 50,
  printedAt: new Date().toISOString(),
  items: [{ name: 'Burger', qty: 1, price: 10 }],
  payments: [],
  ...overrides,
});

beforeEach(() => {
  act(() => getState().clear());
});

describe('receiptsStore', () => {
  // ── Initial state ──────────────────────────────────────────────

  it('starts with no receipts', () => {
    expect(getState().receipts).toEqual([]);
  });

  // ── add ────────────────────────────────────────────────────────

  it('adds a receipt', () => {
    const receipt = makeReceipt({ id: 'r1', orderId: 'o1' });
    act(() => getState().add(receipt));
    expect(getState().receipts).toHaveLength(1);
    expect(getState().receipts[0].id).toBe('r1');
  });

  it('rejects receipts without usable identity', () => {
    act(() => getState().add(makeReceipt({ id: '   ', orderId: 'o1' })));
    act(() => getState().add(makeReceipt({ id: 'r1', orderId: '   ' })));

    expect(getState().receipts).toEqual([]);
  });

  it('normalizes receipt totals, lines, payments, and text fields', () => {
    act(() =>
      getState().add(
        makeReceipt({
          id: ' r1 ',
          orderId: ' o1 ',
          total: Number.NaN,
          customerName: ' Alice ',
          items: [{ name: ' Burger ', qty: 2.8, price: Number.POSITIVE_INFINITY }],
          payments: [
            {
              id: ' p1 ',
              method: 'cash',
              amount: -5,
              createdAt: '2026-05-06T00:00:00.000Z',
            },
          ],
        })
      )
    );

    expect(getState().receipts[0]).toMatchObject({
      id: 'r1',
      orderId: 'o1',
      total: 0,
      customerName: 'Alice',
      items: [{ name: 'Burger', qty: 2, price: 0 }],
      payments: [{ id: 'p1', amount: 0 }],
    });
  });

  it('prepends new receipts (newest first)', () => {
    act(() => {
      getState().add(makeReceipt({ id: 'r1' }));
      getState().add(makeReceipt({ id: 'r2' }));
    });

    expect(getState().receipts[0].id).toBe('r2');
    expect(getState().receipts[1].id).toBe('r1');
  });

  it('adds multiple receipts', () => {
    for (let i = 0; i < 5; i++) {
      act(() => getState().add(makeReceipt()));
    }
    expect(getState().receipts).toHaveLength(5);
  });

  // ── markEmailed ────────────────────────────────────────────────

  it('marks a receipt as emailed', () => {
    const receipt = makeReceipt({ id: 'r_email' });
    act(() => {
      getState().add(receipt);
      getState().markEmailed('r_email');
    });

    expect(getState().receipts[0].emailed).toBe(true);
  });

  it('markEmailed does not affect other receipts', () => {
    act(() => {
      getState().add(makeReceipt({ id: 'r1', emailed: false }));
      getState().add(makeReceipt({ id: 'r2', emailed: false }));
      getState().markEmailed('r1');
    });

    const r1 = getState().receipts.find((r) => r.id === 'r1');
    const r2 = getState().receipts.find((r) => r.id === 'r2');
    expect(r1?.emailed).toBe(true);
    expect(r2?.emailed).toBeFalsy();
  });

  it('markEmailed is a no-op for non-existent id', () => {
    act(() => {
      getState().add(makeReceipt({ id: 'r1' }));
      getState().markEmailed('nope');
    });
    expect(getState().receipts).toHaveLength(1);
  });

  // ── clear ──────────────────────────────────────────────────────

  it('clears all receipts', () => {
    act(() => {
      getState().add(makeReceipt());
      getState().add(makeReceipt());
      getState().clear();
    });
    expect(getState().receipts).toEqual([]);
  });

  // ── byOrder ────────────────────────────────────────────────────

  it('finds a receipt by order id', () => {
    const receipt = makeReceipt({ orderId: 'order-123' });
    act(() => getState().add(receipt));
    const found = getState().byOrder('order-123');
    expect(found).toBeDefined();
    expect(found?.orderId).toBe('order-123');
  });

  it('returns undefined for unknown order id', () => {
    expect(getState().byOrder('unknown')).toBeUndefined();
  });

  it('finds the first match when multiple receipts exist for same order', () => {
    act(() => {
      getState().add(makeReceipt({ id: 'r1', orderId: 'ord-dup' }));
      getState().add(makeReceipt({ id: 'r2', orderId: 'ord-dup' }));
    });
    // newest is first, so byOrder returns r2
    const found = getState().byOrder('ord-dup');
    expect(found?.id).toBe('r2');
  });

  // ── Edge cases ─────────────────────────────────────────────────

  it('receipt with customer name', () => {
    const receipt = makeReceipt({ customerName: 'John' });
    act(() => getState().add(receipt));
    expect(getState().receipts[0].customerName).toBe('John');
  });

  it('clear is safe when already empty', () => {
    act(() => getState().clear());
    expect(getState().receipts).toEqual([]);
  });
});
