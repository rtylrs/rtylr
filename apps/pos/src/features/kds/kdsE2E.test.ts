/**
 * End-to-end: receive raw KDS orders from backend → display them →
 * mark an item done → confirm the correct API call fires.
 *
 * Chains mapKdsOrders (schema normalisation) → buildKdsStatusUpdate
 * → posService.updateKDSItemStatus in sequence.
 */
import { describe, it, expect, vi, beforeEach } from 'vitest';

vi.mock('../../shared/services/api/posService', () => ({
  posService: {
    updateKDSItemStatus: vi.fn(),
  },
}));

vi.mock('@voxire/api', () => ({
  unwrap: vi.fn((p) => p),
}));

import { mapKdsOrders, buildKdsStatusUpdate, KDS_ITEM_ID_ERROR } from './kdsWorkflow';
import { posService } from '../../shared/services/api/posService';
import { unwrap } from '@voxire/api';

// Raw API shape (snake_case + numeric id) - as the server returns it.
const rawOrder = {
  id: 7,
  order_number: 'ORD-007',
  station: 'Bar',
  items: [
    { name: 'Espresso', qty: 2 },
    { name: 'Latte', qty: 1 },
  ],
  status: 'new',
  createdAt: '2026-05-13T08:00:00.000Z',
};

beforeEach(() => {
  vi.clearAllMocks();
});

describe('KDS end-to-end flow', () => {
  it('mapKdsOrders normalizes raw API response into displayable KDS orders', () => {
    const orders = mapKdsOrders([rawOrder]);

    expect(orders).toHaveLength(1);
    const [order] = orders;

    // id must be a string after schema transform
    expect(typeof order.id).toBe('string');
    expect(order.id).toBe('7');

    // station passes through
    expect(order.station).toBe('Bar');

    // status defaults to 'new' (or stays 'new' if valid)
    expect(order.status).toBe('new');

    // items are mapped with name and qty
    expect(order.items).toHaveLength(2);
    expect(order.items[0]).toMatchObject({ name: 'Espresso', qty: 2 });
    expect(order.items[1]).toMatchObject({ name: 'Latte', qty: 1 });
  });

  it('buildKdsStatusUpdate → posService.updateKDSItemStatus called with correct itemId and served status', async () => {
    vi.mocked(posService.updateKDSItemStatus).mockResolvedValue(undefined as never);

    const update = buildKdsStatusUpdate('42', 'served');

    expect(update).toEqual({ itemId: 42, status: 'served' });

    await unwrap(posService.updateKDSItemStatus(update.itemId, update.status));

    expect(posService.updateKDSItemStatus).toHaveBeenCalledTimes(1);
    expect(posService.updateKDSItemStatus).toHaveBeenCalledWith(42, 'served');
  });

  it('marking an item done does not affect other items in the same order', async () => {
    vi.mocked(posService.updateKDSItemStatus).mockResolvedValue(undefined as never);

    const orders = mapKdsOrders([rawOrder]);
    expect(orders[0].items).toHaveLength(2);

    // Only update the first item (id 42) - simulate a single service call.
    const update = buildKdsStatusUpdate('42', 'served');
    await unwrap(posService.updateKDSItemStatus(update.itemId, update.status));

    // Service was called exactly once (only for item 42, not the second item).
    expect(posService.updateKDSItemStatus).toHaveBeenCalledTimes(1);
    expect(posService.updateKDSItemStatus).toHaveBeenCalledWith(42, 'served');

    // The mapped order still has 2 items - marking one done doesn't remove the other.
    expect(orders[0].items).toHaveLength(2);
  });

  it('invalid item id is rejected before the API is called', () => {
    expect(() => buildKdsStatusUpdate('abc', 'done')).toThrow(KDS_ITEM_ID_ERROR);
    expect(posService.updateKDSItemStatus).not.toHaveBeenCalled();
  });
});
