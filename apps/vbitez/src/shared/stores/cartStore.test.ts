import { describe, it, expect, beforeEach } from 'vitest';
import { useCartStore } from './cartStore';

const { getState } = useCartStore;
const act = (fn: () => void) => {
  fn();
};

beforeEach(() => {
  act(() => getState().clearCart());
});

describe('cartStore', () => {
  // ── Add/Remove items ──────────────────────────────────────────

  it('starts with empty cart', () => {
    expect(getState().items).toEqual([]);
    expect(getState().itemCount()).toBe(0);
  });

  it('adds an item', () => {
    act(() => getState().addItem({ productId: 1, name: 'Burger', unitPrice: 12.99 }));

    expect(getState().items).toHaveLength(1);
    expect(getState().items[0].name).toBe('Burger');
    expect(getState().items[0].quantity).toBe(1);
    expect(getState().itemCount()).toBe(1);
  });

  it('increments quantity for same product', () => {
    act(() => {
      getState().addItem({ productId: 1, name: 'Burger', unitPrice: 12.99 });
      getState().addItem({ productId: 1, name: 'Burger', unitPrice: 12.99 });
    });

    expect(getState().items).toHaveLength(1);
    expect(getState().items[0].quantity).toBe(2);
    expect(getState().itemCount()).toBe(2);
  });

  it('treats different modifiers as different line items', () => {
    act(() => {
      getState().addItem({
        productId: 1,
        name: 'Burger',
        unitPrice: 12.99,
        modifiers: [{ id: 1, name: 'Cheese', priceDelta: 1.5 }],
      });
      getState().addItem({
        productId: 1,
        name: 'Burger',
        unitPrice: 12.99,
        modifiers: [{ id: 2, name: 'Bacon', priceDelta: 2.0 }],
      });
    });

    expect(getState().items).toHaveLength(2);
  });

  it('removes an item', () => {
    act(() => {
      getState().addItem({ productId: 1, name: 'Burger', unitPrice: 12.99 });
      const lineId = getState().items[0].lineId;
      getState().removeItem(lineId);
    });

    expect(getState().items).toHaveLength(0);
  });

  it('removes item when quantity set to 0', () => {
    act(() => {
      getState().addItem({ productId: 1, name: 'Burger', unitPrice: 12.99 });
      const lineId = getState().items[0].lineId;
      getState().updateQuantity(lineId, 0);
    });

    expect(getState().items).toHaveLength(0);
  });

  // ── Subtotal ──────────────────────────────────────────────────

  it('calculates subtotal', () => {
    act(() => {
      getState().addItem({ productId: 1, name: 'Burger', unitPrice: 10 });
      getState().addItem({ productId: 2, name: 'Fries', unitPrice: 5 });
    });

    expect(getState().subtotal()).toBe(15);
  });

  it('subtotal includes modifiers', () => {
    act(() =>
      getState().addItem({
        productId: 1,
        name: 'Burger',
        unitPrice: 10,
        modifiers: [
          { id: 1, name: 'Cheese', priceDelta: 1.5 },
          { id: 2, name: 'Bacon', priceDelta: 2.0 },
        ],
      })
    );

    // 10 + 1.5 + 2.0 = 13.5
    expect(getState().subtotal()).toBe(13.5);
  });

  it('subtotal accounts for quantity', () => {
    act(() => {
      getState().addItem({ productId: 1, name: 'Burger', unitPrice: 10 });
      getState().addItem({ productId: 1, name: 'Burger', unitPrice: 10 });
      getState().addItem({ productId: 1, name: 'Burger', unitPrice: 10 });
    });

    expect(getState().subtotal()).toBe(30);
  });

  // ── Per-item discounts ────────────────────────────────────────

  it('applies percentage discount to item', () => {
    act(() => {
      getState().addItem({ productId: 1, name: 'Burger', unitPrice: 100 });
      const lineId = getState().items[0].lineId;
      getState().setItemDiscount(lineId, 20, 0); // 20% off
    });

    expect(getState().subtotal()).toBe(80);
  });

  it('applies flat discount to item', () => {
    act(() => {
      getState().addItem({ productId: 1, name: 'Burger', unitPrice: 100 });
      const lineId = getState().items[0].lineId;
      getState().setItemDiscount(lineId, 0, 15); // $15 off
    });

    expect(getState().subtotal()).toBe(85);
  });

  // ── Order-level discount ──────────────────────────────────────

  it('applies order-level percentage discount', () => {
    act(() => {
      getState().addItem({ productId: 1, name: 'Burger', unitPrice: 100 });
      getState().setOrderDiscount({ type: 'percentage', value: 10 });
    });

    // subtotal = 100, order discount = 10% = 10, tax = 0, total = 90
    expect(getState().total()).toBe(90);
  });

  it('applies order-level fixed discount', () => {
    act(() => {
      getState().addItem({ productId: 1, name: 'Burger', unitPrice: 100 });
      getState().setOrderDiscount({ type: 'fixed', value: 25 });
    });

    expect(getState().total()).toBe(75);
  });

  // ── Tax ────────────────────────────────────────────────────────

  it('calculates tax per item', () => {
    act(() =>
      getState().addItem({
        productId: 1,
        name: 'Burger',
        unitPrice: 100,
        taxRate: 10,
      })
    );

    expect(getState().taxTotal()).toBe(10);
    expect(getState().total()).toBe(110);
  });

  it('calculates tax with discount applied', () => {
    act(() => {
      getState().addItem({
        productId: 1,
        name: 'Burger',
        unitPrice: 100,
        taxRate: 10,
      });
      getState().setOrderDiscount({ type: 'percentage', value: 50 });
    });

    // subtotal=100, discount=50, taxable=50, tax=5, total=55
    expect(getState().taxTotal()).toBe(5);
    expect(getState().total()).toBe(55);
  });

  // ── Tips ──────────────────────────────────────────────────────

  it('adds tip to total', () => {
    act(() => {
      getState().addItem({ productId: 1, name: 'Burger', unitPrice: 100 });
      getState().setTip(15);
    });

    expect(getState().total()).toBe(115);
  });

  // ── Payments ──────────────────────────────────────────────────

  it('tracks amount paid', () => {
    act(() => {
      getState().addItem({ productId: 1, name: 'Burger', unitPrice: 100 });
      getState().addPayment({ method: 'cash', amount: 60 });
    });

    expect(getState().amountPaid()).toBe(60);
    expect(getState().amountDue()).toBe(40);
  });

  it('supports split payments', () => {
    act(() => {
      getState().addItem({ productId: 1, name: 'Burger', unitPrice: 100 });
      getState().addPayment({ method: 'cash', amount: 50 });
      getState().addPayment({ method: 'card', amount: 30 });
      getState().addPayment({ method: 'gift_card', amount: 20, giftCardCode: 'ABC' });
    });

    expect(getState().payments).toHaveLength(3);
    expect(getState().amountPaid()).toBe(100);
    expect(getState().amountDue()).toBe(0);
  });

  it('removes a payment', () => {
    act(() => {
      getState().addItem({ productId: 1, name: 'Burger', unitPrice: 100 });
      getState().addPayment({ method: 'cash', amount: 50 });
      getState().addPayment({ method: 'card', amount: 50 });
      getState().removePayment(0);
    });

    expect(getState().payments).toHaveLength(1);
    expect(getState().amountPaid()).toBe(50);
  });

  // ── Hold/Park ─────────────────────────────────────────────────

  it('holds and resumes order', () => {
    act(() => {
      getState().addItem({ productId: 1, name: 'Burger', unitPrice: 10 });
      getState().holdOrder('Table 5');
    });

    expect(getState().isHeld).toBe(true);
    expect(getState().holdLabel).toBe('Table 5');

    act(() => getState().resumeOrder());

    expect(getState().isHeld).toBe(false);
    expect(getState().items).toHaveLength(1);
  });

  // ── Context ───────────────────────────────────────────────────

  it('sets branch/terminal/employee context', () => {
    act(() => getState().setContext(1, 2, 3));

    expect(getState().branchId).toBe(1);
    expect(getState().terminalId).toBe(2);
    expect(getState().employeeId).toBe(3);
  });

  it('newOrder preserves context', () => {
    act(() => {
      getState().setContext(1, 2, 3);
      getState().addItem({ productId: 1, name: 'Burger', unitPrice: 10 });
      getState().newOrder();
    });

    expect(getState().items).toHaveLength(0);
    expect(getState().branchId).toBe(1);
    expect(getState().terminalId).toBe(2);
    expect(getState().employeeId).toBe(3);
    expect(getState().localOrderId).toBeTruthy();
  });

  // ── Customer ──────────────────────────────────────────────────

  it('sets and clears customer', () => {
    act(() => getState().setCustomer(42, 'John Doe'));

    expect(getState().customerId).toBe(42);
    expect(getState().customerName).toBe('John Doe');

    act(() => getState().setCustomer(null, null));
    expect(getState().customerId).toBeNull();
  });

  // ── Order type ────────────────────────────────────────────────

  it('changes order type', () => {
    act(() => getState().setOrderType('takeaway'));
    expect(getState().orderType).toBe('takeaway');

    act(() => getState().setOrderType('delivery'));
    expect(getState().orderType).toBe('delivery');
  });

  // ── Clear ─────────────────────────────────────────────────────

  it('clears everything', () => {
    act(() => {
      getState().addItem({ productId: 1, name: 'Burger', unitPrice: 10 });
      getState().setTip(5);
      getState().setCustomer(1, 'Test');
      getState().addPayment({ method: 'cash', amount: 10 });
      getState().clearCart();
    });

    expect(getState().items).toHaveLength(0);
    expect(getState().tipAmount).toBe(0);
    expect(getState().customerId).toBeNull();
    expect(getState().payments).toHaveLength(0);
  });
});
