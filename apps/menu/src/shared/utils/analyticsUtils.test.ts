import { describe, it, expect } from 'vitest';
import {
  menuViewEvent,
  addToCartEvent,
  removeFromCartEvent,
  checkoutInitiatedEvent,
  orderPlacedEvent,
  searchEvent,
} from './analyticsUtils';

describe('menuViewEvent', () => {
  it('returns the correct event name', () => {
    expect(menuViewEvent('brunch').event).toBe('menu_viewed');
  });

  it('includes the menu slug in properties', () => {
    expect(menuViewEvent('lunch').properties.menu_slug).toBe('lunch');
  });
});

describe('addToCartEvent', () => {
  it('returns the correct event name', () => {
    expect(addToCartEvent(1, 'Burger', 9.99, 2).event).toBe('add_to_cart');
  });

  it('includes all item properties', () => {
    const { properties } = addToCartEvent(1, 'Burger', 9.99, 2);
    expect(properties.item_id).toBe(1);
    expect(properties.item_name).toBe('Burger');
    expect(properties.price).toBe(9.99);
    expect(properties.quantity).toBe(2);
  });
});

describe('removeFromCartEvent', () => {
  it('returns the correct event name', () => {
    expect(removeFromCartEvent(1, 'Burger').event).toBe('remove_from_cart');
  });

  it('includes item_id and item_name', () => {
    const { properties } = removeFromCartEvent(5, 'Fries');
    expect(properties.item_id).toBe(5);
    expect(properties.item_name).toBe('Fries');
  });
});

describe('checkoutInitiatedEvent', () => {
  it('returns the correct event name', () => {
    expect(checkoutInitiatedEvent(29.99, 3).event).toBe('checkout_initiated');
  });

  it('includes subtotal and item_count', () => {
    const { properties } = checkoutInitiatedEvent(29.99, 3);
    expect(properties.subtotal).toBe(29.99);
    expect(properties.item_count).toBe(3);
  });
});

describe('orderPlacedEvent', () => {
  it('returns the correct event name', () => {
    expect(orderPlacedEvent('order-123', 49.99).event).toBe('order_placed');
  });

  it('includes order_id and total', () => {
    const { properties } = orderPlacedEvent('abc', 15);
    expect(properties.order_id).toBe('abc');
    expect(properties.total).toBe(15);
  });
});

describe('searchEvent', () => {
  it('returns the correct event name', () => {
    expect(searchEvent('burger', 5).event).toBe('menu_searched');
  });

  it('includes query and result_count', () => {
    const { properties } = searchEvent('pizza', 3);
    expect(properties.query).toBe('pizza');
    expect(properties.result_count).toBe(3);
  });

  it('truncates very long queries to 200 characters', () => {
    const longQuery = 'a'.repeat(300);
    const { properties } = searchEvent(longQuery, 0);
    expect((properties.query as string).length).toBe(200);
  });
});
