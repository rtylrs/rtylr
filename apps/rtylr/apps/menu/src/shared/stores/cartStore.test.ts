import { describe, it, expect, beforeEach } from 'vitest';
import { useCartStore } from './cartStore';

const { getState, setState } = useCartStore;

beforeEach(() => {
  setState({ items: [], businessId: 0 });
});

describe('cartStore', () => {
  it('has empty items and businessId 0 by default', () => {
    const s = getState();
    expect(s.items).toEqual([]);
    expect(s.businessId).toBe(0);
  });

  it('addItem adds a new item with quantity 1', () => {
    getState().addItem({ menuItemId: 1, name: 'Burger', unitPrice: 9.99 });
    const { items } = getState();
    expect(items).toHaveLength(1);
    expect(items[0].quantity).toBe(1);
    expect(items[0].notes).toBe('');
  });

  it('addItem increments quantity for an existing item', () => {
    getState().addItem({ menuItemId: 1, name: 'Burger', unitPrice: 9.99 });
    getState().addItem({ menuItemId: 1, name: 'Burger', unitPrice: 9.99 });
    expect(getState().items[0].quantity).toBe(2);
  });

  it('removeItem removes the item from the cart', () => {
    getState().addItem({ menuItemId: 1, name: 'Burger', unitPrice: 9.99 });
    getState().removeItem(1);
    expect(getState().items).toHaveLength(0);
  });

  it('updateQuantity sets the quantity of an item', () => {
    getState().addItem({ menuItemId: 1, name: 'Burger', unitPrice: 9.99 });
    getState().updateQuantity(1, 5);
    expect(getState().items[0].quantity).toBe(5);
  });

  it('updateQuantity removes item when quantity is set to 0', () => {
    getState().addItem({ menuItemId: 1, name: 'Burger', unitPrice: 9.99 });
    getState().updateQuantity(1, 0);
    expect(getState().items).toHaveLength(0);
  });

  it('updateQuantity removes item when quantity is negative', () => {
    getState().addItem({ menuItemId: 1, name: 'Burger', unitPrice: 9.99 });
    getState().updateQuantity(1, -1);
    expect(getState().items).toHaveLength(0);
  });

  it('setItemNote sets the note for a cart item', () => {
    getState().addItem({ menuItemId: 1, name: 'Burger', unitPrice: 9.99 });
    getState().setItemNote(1, 'no pickles');
    expect(getState().items[0].notes).toBe('no pickles');
  });

  it('setBusinessId updates businessId', () => {
    getState().setBusinessId(42);
    expect(getState().businessId).toBe(42);
  });

  it('clearCart empties items and resets businessId', () => {
    getState().addItem({ menuItemId: 1, name: 'Burger', unitPrice: 9.99 });
    getState().setBusinessId(7);
    getState().clearCart();
    expect(getState().items).toHaveLength(0);
    expect(getState().businessId).toBe(0);
  });

  it('total returns the sum of unit prices times quantities', () => {
    getState().addItem({ menuItemId: 1, name: 'Burger', unitPrice: 10 });
    getState().addItem({ menuItemId: 2, name: 'Fries', unitPrice: 5 });
    getState().updateQuantity(1, 2);
    expect(getState().total()).toBe(25);
  });

  it('total returns 0 for an empty cart', () => {
    expect(getState().total()).toBe(0);
  });

  it('itemCount returns the total number of units', () => {
    getState().addItem({ menuItemId: 1, name: 'Burger', unitPrice: 10 });
    getState().addItem({ menuItemId: 2, name: 'Fries', unitPrice: 5 });
    getState().updateQuantity(1, 3);
    expect(getState().itemCount()).toBe(4);
  });

  it('itemCount returns 0 for an empty cart', () => {
    expect(getState().itemCount()).toBe(0);
  });
});
