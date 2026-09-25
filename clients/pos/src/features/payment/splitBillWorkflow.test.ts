import { describe, expect, it } from 'vitest';
import { buildSplitBillRequest } from './splitBillWorkflow';
import type { CartItem } from '../../shared/stores/cartStore';

const item = (id: string): CartItem =>
  ({
    id,
    lineId: id,
    productId: 10,
    name: `Item ${id}`,
    unitPrice: 5,
    price: 5,
    quantity: 1,
    modifiers: [],
    discountPct: 0,
    discountAmount: 0,
    taxRate: 0,
    status: 'pending',
  }) satisfies CartItem;

describe('splitBillWorkflow', () => {
  it('builds a server split request from valid synced order item IDs', () => {
    expect(
      buildSplitBillRequest({
        routeOrderId: '7',
        serverOrderId: null,
        seats: 2,
        items: [item('11'), item('12')],
        assignments: { '11': 0, '12': 1 },
      }),
    ).toEqual({
      localOnly: false,
      payload: {
        orderId: '7',
        splits: [{ itemIds: ['11'] }, { itemIds: ['12'] }],
      },
    });
  });

  it('uses the synced server order ID before the route order ID', () => {
    expect(
      buildSplitBillRequest({
        routeOrderId: '7',
        serverOrderId: 99,
        seats: 2,
        items: [item('11'), item('12')],
        assignments: { '11': 0, '12': 1 },
      }),
    ).toMatchObject({ payload: { orderId: '99' } });
  });

  it('returns local-only when no valid server order ID exists', () => {
    expect(
      buildSplitBillRequest({
        routeOrderId: 'current',
        serverOrderId: null,
        seats: 2,
        items: [item('local-1'), item('local-2')],
        assignments: { 'local-1': 0, 'local-2': 1 },
      }),
    ).toEqual({ localOnly: true });
  });

  it('rejects server splits when item IDs are not synced order item IDs', () => {
    expect(() =>
      buildSplitBillRequest({
        routeOrderId: '7',
        serverOrderId: null,
        seats: 2,
        items: [item('local-1'), item('12')],
        assignments: { 'local-1': 0, '12': 1 },
      }),
    ).toThrow('Sync this order before splitting by item.');
  });

  it('rejects server splits without at least two non-empty split groups', () => {
    expect(() =>
      buildSplitBillRequest({
        routeOrderId: '7',
        serverOrderId: null,
        seats: 2,
        items: [item('11'), item('12')],
        assignments: { '11': 0, '12': 0 },
      }),
    ).toThrow('Assign items to at least two splits before confirming.');
  });
});
