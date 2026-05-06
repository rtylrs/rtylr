import { describe, expect, it } from 'vitest';
import { buildKdsStatusUpdate, mapKdsOrders } from './kdsWorkflow';

describe('kdsWorkflow', () => {
  it('maps KDS orders with schema defaults and nullable item fields normalized', () => {
    const orders = mapKdsOrders([
      {
        id: 12,
        station: 'Kitchen',
        items: [{ name: 'Burger', qty: 2, course: null, notes: null }],
        status: 'invalid',
      },
    ]);

    expect(orders).toEqual([
      expect.objectContaining({
        id: '12',
        number: '12',
        station: 'Kitchen',
        status: 'new',
        items: [{ name: 'Burger', qty: 2, course: undefined, notes: undefined }],
      }),
    ]);
    expect(orders[0].createdAt).toEqual(expect.any(String));
  });

  it('builds status updates with positive integer item IDs', () => {
    expect(buildKdsStatusUpdate('42', 'ready')).toEqual({ itemId: 42, status: 'ready' });
  });

  it('rejects invalid item IDs before status updates hit the API', () => {
    expect(() => buildKdsStatusUpdate('local-42', 'ready')).toThrow(
      'Select a valid KDS item before updating status.',
    );
    expect(() => buildKdsStatusUpdate('0', 'ready')).toThrow(
      'Select a valid KDS item before updating status.',
    );
  });
});
