import { describe, expect, it, vi } from 'vitest';
import { syncQueuedOrders } from './syncOfflineQueue';
import type { OfflineQueuedOrder, Order } from '../../shared/types/pos';

const makeOrder = (id: string): Order => ({
  id,
  status: 'paid',
  items: [],
  subtotal: 10,
  tax: 1,
  tip: 0,
  discount: 0,
  total: 11,
  createdAt: '2026-05-06T00:00:00.000Z',
});

const makeEntry = (
  id: string,
  status: OfflineQueuedOrder['status'] = 'pending'
): OfflineQueuedOrder => ({
  id,
  order: makeOrder(`order-${id}`),
  queuedAt: '2026-05-06T00:00:00.000Z',
  retries: 0,
  status,
});

describe('syncQueuedOrders', () => {
  it('marks an order synced after successful replay', async () => {
    const entry = makeEntry('q1');
    const retry = vi.fn();
    const markSynced = vi.fn();

    const result = await syncQueuedOrders({
      entries: [entry],
      syncOrder: vi.fn(() => Promise.resolve()),
      retry,
      markSynced,
      markFailed: vi.fn(),
    });

    expect(result).toEqual({ attempted: 1, synced: 1, failed: 0 });
    expect(retry).toHaveBeenCalledWith('q1');
    expect(markSynced).toHaveBeenCalledWith('q1');
  });

  it('marks an order failed with a useful message after replay failure', async () => {
    const entry = makeEntry('q1');
    const markFailed = vi.fn();
    // oneKit ApiError shape: the backend message lives in a JSON body string.
    const rejection = Object.assign(new Error('Request failed with status 422'), {
      body: JSON.stringify({ message: 'Out of stock' }),
    });

    const result = await syncQueuedOrders({
      entries: [entry],
      syncOrder: vi.fn(() => Promise.reject(rejection)),
      retry: vi.fn(),
      markSynced: vi.fn(),
      markFailed,
    });

    expect(result).toEqual({ attempted: 1, synced: 0, failed: 1 });
    expect(markFailed).toHaveBeenCalledWith('q1', 'Out of stock');
  });

  it('skips entries already marked syncing', async () => {
    const result = await syncQueuedOrders({
      entries: [makeEntry('q1', 'syncing')],
      syncOrder: vi.fn(() => Promise.resolve()),
      retry: vi.fn(),
      markSynced: vi.fn(),
      markFailed: vi.fn(),
    });

    expect(result).toEqual({ attempted: 0, synced: 0, failed: 0 });
  });
});
