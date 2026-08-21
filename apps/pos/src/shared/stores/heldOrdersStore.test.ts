import { describe, it, expect, beforeEach } from 'vitest';
import { useHeldOrdersStore } from './heldOrdersStore';

const { getState, setState } = useHeldOrdersStore;
const act = (fn: () => void) => {
  fn();
};

beforeEach(() => {
  setState({ orders: [] });
});

describe('heldOrdersStore', () => {
  // ── Initial state ──────────────────────────────────────────────

  it('starts with no held orders', () => {
    expect(getState().orders).toEqual([]);
  });

  // ── hold ───────────────────────────────────────────────────────

  it('holds an order and returns it', () => {
    const order = getState().hold({
      label: 'Table 5',
      items: [{ id: '1', name: 'Burger', price: 10, quantity: 1 }],
    })!;

    expect(order.id).toMatch(/^held_/);
    expect(order.label).toBe('Table 5');
    expect(order.createdAt).toBeTruthy();
    expect(getState().orders).toHaveLength(1);
  });

  it('rejects held orders without a usable label', () => {
    const order = getState().hold({
      label: '   ',
      items: [{ id: '1', name: 'Burger', price: 10, quantity: 1 }],
    });

    expect(order).toBeUndefined();
    expect(getState().orders).toEqual([]);
  });

  it('normalizes held order labels, notes, table numbers, and line items', () => {
    const order = getState().hold({
      label: ' Table 5 ',
      tableNumber: ' T5 ',
      note: ' Rush ',
      items: [
        {
          id: ' 1 ',
          name: ' Burger ',
          price: Number.NaN,
          quantity: 2.8,
        },
      ],
    });

    expect(order).toMatchObject({
      label: 'Table 5',
      tableNumber: 'T5',
      note: 'Rush',
      items: [{ id: '1', name: 'Burger', price: 0, quantity: 2 }],
    });
  });

  it('prepends new held orders (newest first)', () => {
    getState().hold({ label: 'Order A', items: [] });
    getState().hold({ label: 'Order B', items: [] });

    expect(getState().orders[0].label).toBe('Order B');
    expect(getState().orders[1].label).toBe('Order A');
  });

  it('holds order with optional fields', () => {
    const order = getState().hold({
      label: 'Table 3',
      items: [],
      tableNumber: '3',
      note: 'Rush order',
    })!;

    expect(order.tableNumber).toBe('3');
    expect(order.note).toBe('Rush order');
  });

  // ── release ────────────────────────────────────────────────────

  it('releases a held order and returns it', () => {
    const order = getState().hold({ label: 'Table 1', items: [] })!;
    const released = getState().release(order.id);

    expect(released).toBeDefined();
    expect(released?.label).toBe('Table 1');
    expect(getState().orders).toHaveLength(0);
  });

  it('returns undefined when releasing a non-existent id', () => {
    const released = getState().release('does-not-exist');
    expect(released).toBeUndefined();
  });

  it('only removes the targeted order on release', () => {
    const a = getState().hold({ label: 'A', items: [] })!;
    getState().hold({ label: 'B', items: [] });

    getState().release(a.id);
    expect(getState().orders).toHaveLength(1);
    expect(getState().orders[0].label).toBe('B');
  });

  // ── remove ─────────────────────────────────────────────────────

  it('removes an order by id', () => {
    const order = getState().hold({ label: 'Table 2', items: [] })!;
    act(() => getState().remove(order.id));
    expect(getState().orders).toHaveLength(0);
  });

  it('remove is a no-op for non-existent id', () => {
    getState().hold({ label: 'A', items: [] });
    act(() => getState().remove('nope'));
    expect(getState().orders).toHaveLength(1);
  });

  // ── updateNote ─────────────────────────────────────────────────

  it('updates the note of a held order', () => {
    const order = getState().hold({ label: 'Table 4', items: [] })!;
    act(() => getState().updateNote(order.id, 'No onions'));
    expect(getState().orders[0].note).toBe('No onions');
  });

  it('normalizes updated notes', () => {
    const order = getState().hold({ label: 'Table 4', items: [] })!;
    act(() => getState().updateNote(order.id, '   '));
    expect(getState().orders[0].note).toBeUndefined();
  });

  it('updateNote does not affect other orders', () => {
    const a = getState().hold({ label: 'A', items: [], note: 'Note A' })!;
    const b = getState().hold({ label: 'B', items: [], note: 'Note B' })!;

    act(() => getState().updateNote(a.id, 'Updated A'));

    const orderA = getState().orders.find((o) => o.id === a.id);
    const orderB = getState().orders.find((o) => o.id === b.id);
    expect(orderA?.note).toBe('Updated A');
    expect(orderB?.note).toBe('Note B');
  });

  // ── Edge cases ─────────────────────────────────────────────────

  it('can hold multiple orders and release them independently', () => {
    const o1 = getState().hold({ label: '1', items: [] });
    const o2 = getState().hold({ label: '2', items: [] })!;
    const o3 = getState().hold({ label: '3', items: [] });

    getState().release(o2.id);
    expect(getState().orders).toHaveLength(2);
    expect(getState().orders.find((o) => o.id === o2.id)).toBeUndefined();
  });

  it('holds order with empty items array', () => {
    const order = getState().hold({ label: 'Empty', items: [] })!;
    expect(order.items).toEqual([]);
  });
});
