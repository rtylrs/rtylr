import { describe, it, expect } from 'vitest';
import {
  isTransferComplete,
  isTransferInTransit,
  isTransferCancelled,
  transferStatusVariant,
  sortTransfersByDateDesc,
  filterTransfersByStatus,
  totalItemsInTransfer,
} from './transferUtils';
import type { StockTransfer } from '../services/erpApi';

function makeTransfer(overrides: Partial<StockTransfer> = {}): StockTransfer {
  return {
    id: 1,
    fromWarehouseId: 1,
    toWarehouseId: 2,
    fromName: 'Main',
    toName: 'Secondary',
    items: 5,
    status: 'Draft',
    createdAt: '2024-01-01T00:00:00Z',
    ...overrides,
  };
}

describe('isTransferComplete', () => {
  it('returns true for Received status', () => {
    expect(isTransferComplete(makeTransfer({ status: 'Received' }))).toBe(true);
  });

  it('returns true for Cancelled status', () => {
    expect(isTransferComplete(makeTransfer({ status: 'Cancelled' }))).toBe(true);
  });

  it('returns false for active statuses', () => {
    expect(isTransferComplete(makeTransfer({ status: 'Draft' }))).toBe(false);
    expect(isTransferComplete(makeTransfer({ status: 'In Transit' }))).toBe(false);
  });
});

describe('isTransferInTransit', () => {
  it('returns true for In Transit status', () => {
    expect(isTransferInTransit(makeTransfer({ status: 'In Transit' }))).toBe(true);
  });

  it('returns false for other statuses', () => {
    expect(isTransferInTransit(makeTransfer({ status: 'Draft' }))).toBe(false);
  });
});

describe('isTransferCancelled', () => {
  it('returns true for Cancelled status', () => {
    expect(isTransferCancelled(makeTransfer({ status: 'Cancelled' }))).toBe(true);
  });

  it('returns false for other statuses', () => {
    expect(isTransferCancelled(makeTransfer({ status: 'Draft' }))).toBe(false);
  });
});

describe('transferStatusVariant', () => {
  it('returns default for Received', () => {
    expect(transferStatusVariant(makeTransfer({ status: 'Received' }))).toBe('default');
  });

  it('returns secondary for In Transit', () => {
    expect(transferStatusVariant(makeTransfer({ status: 'In Transit' }))).toBe('secondary');
  });

  it('returns destructive for Cancelled', () => {
    expect(transferStatusVariant(makeTransfer({ status: 'Cancelled' }))).toBe('destructive');
  });

  it('returns outline for Draft', () => {
    expect(transferStatusVariant(makeTransfer({ status: 'Draft' }))).toBe('outline');
  });
});

describe('sortTransfersByDateDesc', () => {
  it('sorts newest first', () => {
    const transfers = [
      makeTransfer({ id: 1, createdAt: '2024-01-01T00:00:00Z' }),
      makeTransfer({ id: 2, createdAt: '2024-03-01T00:00:00Z' }),
    ];
    const sorted = sortTransfersByDateDesc(transfers);
    expect(sorted[0].id).toBe(2);
  });

  it('does not mutate original', () => {
    const transfers = [
      makeTransfer({ id: 1, createdAt: '2024-01-01T00:00:00Z' }),
      makeTransfer({ id: 2, createdAt: '2024-03-01T00:00:00Z' }),
    ];
    sortTransfersByDateDesc(transfers);
    expect(transfers[0].id).toBe(1);
  });
});

describe('filterTransfersByStatus', () => {
  const transfers = [
    makeTransfer({ id: 1, status: 'Draft' }),
    makeTransfer({ id: 2, status: 'In Transit' }),
    makeTransfer({ id: 3, status: 'Draft' }),
  ];

  it('filters by status', () => {
    expect(filterTransfersByStatus(transfers, 'Draft')).toHaveLength(2);
  });

  it('returns empty for no matches', () => {
    expect(filterTransfersByStatus(transfers, 'Received')).toHaveLength(0);
  });
});

describe('totalItemsInTransfer', () => {
  it('sums items across transfers', () => {
    const transfers = [makeTransfer({ items: 3 }), makeTransfer({ items: 7 })];
    expect(totalItemsInTransfer(transfers)).toBe(10);
  });

  it('returns 0 for empty array', () => {
    expect(totalItemsInTransfer([])).toBe(0);
  });
});
