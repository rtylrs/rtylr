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

  it('normalizes invalid product pricing before adding an item', () => {
    act(() =>
      getState().addItem({
        productId: 1,
        name: 'Broken item',
        unitPrice: Number.NaN,
        taxRate: Number.POSITIVE_INFINITY,
        modifiers: [{ id: 1, name: 'Bad modifier', priceDelta: -5 }],
      })
    );

    expect(getState().items[0]).toMatchObject({
      unitPrice: 0,
      price: 0,
      taxRate: 0,
      modifiers: [{ id: 1, name: 'Bad modifier', priceDelta: 0 }],
    });
    expect(getState().total()).toBe(0);
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

  it('normalizes quantities to positive integers', () => {
    act(() => {
      getState().addItem({ productId: 1, name: 'Burger', unitPrice: 10 });
      const lineId = getState().items[0].lineId;
      getState().updateQuantity(lineId, 2.8);
    });

    expect(getState().items[0].quantity).toBe(2);

    act(() => {
      const lineId = getState().items[0].lineId;
      getState().updateQuantity(lineId, Number.NaN);
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

  it('clamps invalid per-item discounts', () => {
    act(() => {
      getState().addItem({ productId: 1, name: 'Burger', unitPrice: 100 });
      const lineId = getState().items[0].lineId;
      getState().setItemDiscount(lineId, 250, Number.NaN);
    });

    expect(getState().items[0].discountPct).toBe(100);
    expect(getState().items[0].discountAmount).toBe(0);
    expect(getState().subtotal()).toBe(0);
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

  it('normalizes invalid order-level discounts', () => {
    act(() => {
      getState().addItem({ productId: 1, name: 'Burger', unitPrice: 100 });
      getState().setOrderDiscount({ type: 'percentage', value: 120 });
    });

    expect(getState().orderDiscount).toMatchObject({ type: 'percentage', value: 100 });
    expect(getState().total()).toBe(0);

    act(() => getState().setOrderDiscount({ type: 'fixed', value: Number.NaN }));

    expect(getState().orderDiscount).toMatchObject({ type: 'fixed', value: 0 });
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

  it('normalizes invalid tips to zero', () => {
    act(() => {
      getState().addItem({ productId: 1, name: 'Burger', unitPrice: 100 });
      getState().setTip(Number.NaN);
    });

    expect(getState().tipAmount).toBe(0);
    expect(getState().total()).toBe(100);
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

  it('normalizes invalid payment amounts to zero', () => {
    act(() => {
      getState().addPayment({ method: 'cash', amount: Number.POSITIVE_INFINITY });
      getState().addPayment({ method: 'card', amount: -10 });
    });

    expect(getState().payments.map((payment) => payment.amount)).toEqual([0, 0]);
    expect(getState().amountPaid()).toBe(0);
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

  it('clears branch/terminal/employee context when any id is invalid', () => {
    act(() => getState().setContext(1, 0, 3));

    expect(getState().branchId).toBeNull();
    expect(getState().terminalId).toBeNull();
    expect(getState().employeeId).toBeNull();

    act(() => getState().setContext(1, 2, Number.NaN));

    expect(getState().branchId).toBeNull();
    expect(getState().terminalId).toBeNull();
    expect(getState().employeeId).toBeNull();
  });

  it('rounds valid branch/terminal/employee context ids', () => {
    act(() => getState().setContext(1.9, 2.8, 3.7));

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

  it('normalizes invalid customer ids and blank names', () => {
    act(() => getState().setCustomer(Number.NaN, ' Alice '));

    expect(getState().customerId).toBeNull();
    expect(getState().customerName).toBeNull();

    act(() => getState().setCustomer(42.9, '   '));

    expect(getState().customerId).toBe(42);
    expect(getState().customerName).toBeNull();
  });

  // ── Table ─────────────────────────────────────────────────────

  it('normalizes table and floor ids', () => {
    act(() => getState().setTable(12.8, 3.4));

    expect(getState().tableId).toBe(12);
    expect(getState().floorId).toBe(3);

    act(() => getState().setTable(0, Number.NaN));

    expect(getState().tableId).toBeNull();
    expect(getState().floorId).toBeNull();
  });

  // ── Order tracking ────────────────────────────────────────────

  it('normalizes server order id and order number', () => {
    act(() => {
      getState().setServerOrderId(99.8);
      getState().setOrderNumber('  ORD-1  ');
      getState().setTableNumber(' T7 ');
    });

    expect(getState().serverOrderId).toBe(99);
    expect(getState().orderNumber).toBe('ORD-1');
    expect(getState().tableNumber).toBe('T7');

    act(() => {
      getState().setServerOrderId(Number.NaN);
      getState().setOrderNumber('   ');
      getState().setTableNumber('   ');
    });

    expect(getState().serverOrderId).toBeNull();
    expect(getState().orderNumber).toBeNull();
    expect(getState().tableNumber).toBeNull();
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
