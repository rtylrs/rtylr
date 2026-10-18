import { describe, expect, it } from 'vitest';
import {
  getPurchaseOrderStatusCounts,
  purchaseOrderStatusFilter,
} from './purchaseOrderStatusUtils';

describe('purchase-order status filters and counts', () => {
  it('counts statuses across the complete fetched collection', () => {
    const counts = getPurchaseOrderStatusCounts([
      { status: 'draft' },
      { status: 'sent' },
      { status: 'approved' },
      { status: 'received' },
      { status: 'sent' },
      { status: 'partial' },
    ]);

    expect(counts).toEqual({ Draft: 1, Sent: 2, Approved: 1, Partial: 1, Received: 1 });
  });

  it('converts a selected status card into the API filter', () => {
    expect(purchaseOrderStatusFilter('All')).toBe('');
    expect(purchaseOrderStatusFilter('Approved')).toBe('approved');
    expect(purchaseOrderStatusFilter('Partial')).toBe('partial');
  });
});
