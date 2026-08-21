import { describe, it, expect, beforeEach } from 'vitest';
import { useOfflineQueueStore } from './offlineQueueStore';
import type { Order } from '../types/pos';

const { getState, setState } = useOfflineQueueStore;
const act = (fn: () => void) => {
  fn();
};

const makeOrder = (overrides: Partial<Order> = {}): Order => ({
  id: 'ord_' + Math.random().toString(36).slice(2, 8),
  status: 'paid',
  items: [],
  subtotal: 50,
  tax: 5,
  tip: 0,
  discount: 0,
  total: 55,
  createdAt: new Date().toISOString(),
  ...overrides,
});

beforeEach(() => {
  setState({ queue: [], online: true });
});

describe('offlineQueueStore', () => {
  // ── Initial state ──────────────────────────────────────────────

  it('starts with empty queue and online', () => {
    expect(getState().queue).toEqual([]);
    expect(getState().online).toBe(true);
  });

  // ── enqueue ────────────────────────────────────────────────────

  it('enqueues an order', () => {
    const order = makeOrder();
    const queued = getState().enqueue(order);

    expect(queued.id).toMatch(/^oq_/);
    expect(queued.order).toEqual(order);
    expect(queued.retries).toBe(0);
    expect(queued.status).toBe('pending');
    expect(queued.queuedAt).toBeTruthy();
    expect(getState().queue).toHaveLength(1);
  });

  it('prepends new items (newest first)', () => {
    getState().enqueue(makeOrder({ customerId: 'first' }));
    getState().enqueue(makeOrder({ customerId: 'second' }));

    expect(getState().queue[0].order.customerId).toBe('second');
    expect(getState().queue[1].order.customerId).toBe('first');
  });

  // ── setOnline ──────────────────────────────────────────────────

  it('sets online status to false', () => {
    act(() => getState().setOnline(false));
    expect(getState().online).toBe(false);
  });

  it('sets online status to true', () => {
    setState({ online: false });
    act(() => getState().setOnline(true));
    expect(getState().online).toBe(true);
  });

  // ── retry ──────────────────────────────────────────────────────

  it('increments retries and sets status to syncing', () => {
    const queued = getState().enqueue(makeOrder());
    act(() => getState().retry(queued.id));

    const item = getState().queue.find((q) => q.id === queued.id);
    expect(item?.retries).toBe(1);
    expect(item?.status).toBe('syncing');
  });

  it('increments retries multiple times', () => {
    const queued = getState().enqueue(makeOrder());
    act(() => {
      getState().retry(queued.id);
      getState().retry(queued.id);
      getState().retry(queued.id);
    });

    const item = getState().queue.find((q) => q.id === queued.id);
    expect(item?.retries).toBe(3);
  });

  // ── markSynced ─────────────────────────────────────────────────

  it('removes item from queue when synced', () => {
    const queued = getState().enqueue(makeOrder());
    act(() => getState().markSynced(queued.id));
    expect(getState().queue).toHaveLength(0);
  });

  it('only removes the targeted item', () => {
    const a = getState().enqueue(makeOrder());
    const b = getState().enqueue(makeOrder());

    act(() => getState().markSynced(a.id));
    expect(getState().queue).toHaveLength(1);
    expect(getState().queue[0].id).toBe(b.id);
  });

  // ── markFailed ─────────────────────────────────────────────────

  it('marks an item as failed with error message', () => {
    const queued = getState().enqueue(makeOrder());
    act(() => getState().markFailed(queued.id, 'Server error'));

    const item = getState().queue.find((q) => q.id === queued.id);
    expect(item?.status).toBe('failed');
    expect(item?.lastError).toBe('Server error');
  });

  it('markFailed preserves retries count', () => {
    const queued = getState().enqueue(makeOrder());
    act(() => {
      getState().retry(queued.id);
      getState().retry(queued.id);
      getState().markFailed(queued.id, 'timeout');
    });

    const item = getState().queue.find((q) => q.id === queued.id);
    expect(item?.retries).toBe(2);
    expect(item?.status).toBe('failed');
  });

  // ── clear ──────────────────────────────────────────────────────

  it('clears the entire queue', () => {
    getState().enqueue(makeOrder());
    getState().enqueue(makeOrder());
    act(() => getState().clear());
    expect(getState().queue).toEqual([]);
  });

  // ── Edge cases ─────────────────────────────────────────────────

  it('clear is safe on empty queue', () => {
    act(() => getState().clear());
    expect(getState().queue).toEqual([]);
  });

  it('retry on non-existent id does not throw', () => {
    act(() => getState().retry('nope'));
    expect(getState().queue).toEqual([]);
  });
});
