import { describe, it, expect } from 'vitest';
import {
  isRetryable,
  isPermanentlyFailed,
  isPending,
  isSyncing,
  queueStatusCounts,
  sortQueueOldestFirst,
  retryableEntries,
  totalQueuedItemCount,
} from './offlineQueueUtils';
import type { OfflineQueuedOrder } from '../types/pos';

const makeEntry = (overrides: Partial<OfflineQueuedOrder> = {}): OfflineQueuedOrder => ({
  id: '1',
  order: {
    sessionId: 'session-1',
    type: 'dine_in',
    paymentMethod: 'cash',
    items: [{ menuItemId: 'i1', name: 'Burger', unitPrice: 10, quantity: 1 }],
  },
  queuedAt: '2024-06-01T10:00:00Z',
  retries: 0,
  status: 'pending',
  ...overrides,
});

describe('isRetryable', () => {
  it('returns true for a failed entry within retry limit', () => {
    expect(isRetryable(makeEntry({ status: 'failed', retries: 2 }))).toBe(true);
  });

  it('returns false when retries equal maxRetries', () => {
    expect(isRetryable(makeEntry({ status: 'failed', retries: 3 }))).toBe(false);
  });

  it('returns false for a pending entry', () => {
    expect(isRetryable(makeEntry({ status: 'pending' }))).toBe(false);
  });

  it('respects a custom maxRetries', () => {
    expect(isRetryable(makeEntry({ status: 'failed', retries: 5 }), 5)).toBe(false);
    expect(isRetryable(makeEntry({ status: 'failed', retries: 4 }), 5)).toBe(true);
  });
});

describe('isPermanentlyFailed', () => {
  it('returns true when retries have exceeded the limit', () => {
    expect(isPermanentlyFailed(makeEntry({ status: 'failed', retries: 3 }))).toBe(true);
  });

  it('returns false when still retryable', () => {
    expect(isPermanentlyFailed(makeEntry({ status: 'failed', retries: 2 }))).toBe(false);
  });
});

describe('isPending', () => {
  it('returns true for pending entries', () => {
    expect(isPending(makeEntry({ status: 'pending' }))).toBe(true);
  });

  it('returns false for syncing entries', () => {
    expect(isPending(makeEntry({ status: 'syncing' }))).toBe(false);
  });
});

describe('isSyncing', () => {
  it('returns true for syncing entries', () => {
    expect(isSyncing(makeEntry({ status: 'syncing' }))).toBe(true);
  });

  it('returns false for pending entries', () => {
    expect(isSyncing(makeEntry({ status: 'pending' }))).toBe(false);
  });
});

describe('queueStatusCounts', () => {
  it('counts entries by status', () => {
    const entries = [
      makeEntry({ status: 'pending' }),
      makeEntry({ status: 'syncing' }),
      makeEntry({ status: 'failed' }),
      makeEntry({ status: 'failed' }),
    ];
    expect(queueStatusCounts(entries)).toEqual({ pending: 1, syncing: 1, failed: 2 });
  });

  it('returns zeros for an empty queue', () => {
    expect(queueStatusCounts([])).toEqual({ pending: 0, syncing: 0, failed: 0 });
  });
});

describe('sortQueueOldestFirst', () => {
  it('sorts entries oldest-first by queuedAt', () => {
    const entries = [
      makeEntry({ id: '2', queuedAt: '2024-06-01T10:05:00Z' }),
      makeEntry({ id: '1', queuedAt: '2024-06-01T10:00:00Z' }),
    ];
    const sorted = sortQueueOldestFirst(entries);
    expect(sorted[0].id).toBe('1');
    expect(sorted[1].id).toBe('2');
  });

  it('does not mutate the original array', () => {
    const entries = [
      makeEntry({ id: '2', queuedAt: '2024-06-01T10:05:00Z' }),
      makeEntry({ id: '1', queuedAt: '2024-06-01T10:00:00Z' }),
    ];
    sortQueueOldestFirst(entries);
    expect(entries[0].id).toBe('2');
  });
});

describe('retryableEntries', () => {
  it('returns only retryable entries', () => {
    const entries = [
      makeEntry({ id: '1', status: 'failed', retries: 1 }),
      makeEntry({ id: '2', status: 'failed', retries: 3 }),
      makeEntry({ id: '3', status: 'pending' }),
    ];
    const result = retryableEntries(entries);
    expect(result).toHaveLength(1);
    expect(result[0].id).toBe('1');
  });
});

describe('totalQueuedItemCount', () => {
  it('sums items across all queued orders', () => {
    const entry1 = makeEntry({
      order: {
        ...makeEntry().order,
        items: [
          { menuItemId: 'a', name: 'Burger', unitPrice: 10, quantity: 1 },
          { menuItemId: 'b', name: 'Fries', unitPrice: 3, quantity: 1 },
        ],
      },
    });
    const entry2 = makeEntry({
      order: { ...makeEntry().order, items: [{ menuItemId: 'c', name: 'Soda', unitPrice: 2, quantity: 1 }] },
    });
    expect(totalQueuedItemCount([entry1, entry2])).toBe(3);
  });

  it('returns 0 for an empty queue', () => {
    expect(totalQueuedItemCount([])).toBe(0);
  });
});
