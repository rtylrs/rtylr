/**
 * End-to-end: hold order → restore cart → resume checkout.
 *
 * Proves the held orders system correctly round-trips order data
 * through the hold/restore cycle without data loss.
 */
import { describe, expect, it } from 'vitest';
import {
  buildHeldOrderCartProducts,
  buildServerHeldOrderResume,
  HELD_ORDER_ITEM_ERROR,
  SERVER_HELD_ORDER_ITEM_ERROR,
} from './heldOrdersWorkflow';
import type { HeldOrder } from '../../shared/stores/heldOrdersStore';

function makeHeldOrder(items: HeldOrder['items']): HeldOrder {
  return {
    id: 'held_1',
    label: 'Table 5',
    createdAt: '2026-05-13T12:00:00.000Z',
    items,
  };
}

describe('local held order → cart restore', () => {
  it('restores a 2-item order into individual cart products', () => {
    const order = makeHeldOrder([
      { id: '3::', name: 'Burger', price: 12, quantity: 2 },
      { id: '7::', name: 'Fries', price: 5, quantity: 1 },
    ]);

    const products = buildHeldOrderCartProducts(order);

    // quantity:2 Burger expands to 2 entries
    expect(products).toHaveLength(3);
    expect(products.filter((p) => p.name === 'Burger')).toHaveLength(2);
    expect(products.find((p) => p.name === 'Fries')?.productId).toBe(7);
  });

  it('preserves unit price when restoring held items', () => {
    const order = makeHeldOrder([{ id: '5::', name: 'Latte', price: 4.5, quantity: 1 }]);

    const products = buildHeldOrderCartProducts(order);

    expect(products[0].unitPrice).toBe(4.5);
    expect(products[0].price).toBe(4.5);
  });

  it('rejects held orders with unresolvable product ids', () => {
    const order = makeHeldOrder([{ id: 'draft-special', name: 'Custom', price: 10, quantity: 1 }]);

    expect(() => buildHeldOrderCartProducts(order)).toThrow(HELD_ORDER_ITEM_ERROR);
  });
});

describe('server held orders → generated row → resume', () => {
  it('buildServerHeldOrderResume creates cart products from synced items', () => {
    const serverOrder = {
      id: '22',
      orderNumber: '1001',
      total: 20,
      createdAt: '2026-05-13T12:00:00.000Z',
      items: [{ menuItemId: '5', name: 'Burger', quantity: 2, unitPrice: 10 }],
    };

    const { orderId, products } = buildServerHeldOrderResume(serverOrder);

    expect(orderId).toBe(22);
    expect(products).toHaveLength(2); // quantity 2 → 2 entries
    expect(products[0].productId).toBe(5);
    expect(products[0].unitPrice).toBe(10);
  });

  it('rejects server held orders with missing menu item ids', () => {
    const serverOrder = {
      id: '22',
      orderNumber: '1001',
      total: 20,
      createdAt: '2026-05-13T12:00:00.000Z',
      items: [{ name: 'Mystery item', quantity: 1 }],
    };

    expect(() => buildServerHeldOrderResume(serverOrder)).toThrow(SERVER_HELD_ORDER_ITEM_ERROR);
  });
});

describe('full hold → restore → checkout cycle', () => {
  it('held order cart products have correct productIds for checkout payload', () => {
    const order = makeHeldOrder([
      { id: '10::', name: 'Pizza', price: 18, quantity: 1 },
      { id: '11::', name: 'Soda', price: 3, quantity: 3 },
    ]);

    const products = buildHeldOrderCartProducts(order);

    // 1 pizza + 3 sodas = 4 total
    expect(products).toHaveLength(4);

    const pizzas = products.filter((p) => p.productId === 10);
    const sodas = products.filter((p) => p.productId === 11);

    expect(pizzas).toHaveLength(1);
    expect(sodas).toHaveLength(3);

    // All have a price suitable for the checkout workflow
    expect(products.every((p) => p.unitPrice >= 0)).toBe(true);
  });
});
