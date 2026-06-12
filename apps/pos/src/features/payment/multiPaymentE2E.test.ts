/**
 * End-to-end: split-tender (multi-payment) → charge → receipt → drawer.
 *
 * Customer pays $30 cash + $20 card on a $50 order.
 * The backend takes a single {payment_method, amount_paid} so the dominant
 * method (cash, $30) is selected and the full amount including tip is summed.
 */
import { describe, expect, it, vi, beforeEach } from 'vitest';
import { finalizeMultiPayment } from './multiPaymentWorkflow';
import { printReceiptAndOpenDrawer } from './paymentHardware';

vi.mock('../../shared/services/hardware/receiptPrinter', () => ({
  printReceipt: vi.fn(),
}));
vi.mock('../../shared/services/hardware/cashDrawerHardware', () => ({
  openConfiguredCashDrawer: vi.fn(),
}));
vi.mock('../../shared/services/hardware/crossPlatformDrawer', () => ({
  openDrawerSerial: vi.fn(),
  openDrawerNetwork: vi.fn(),
}));

import { printReceipt } from '../../shared/services/hardware/receiptPrinter';
import { openConfiguredCashDrawer } from '../../shared/services/hardware/cashDrawerHardware';
import type { CartItem } from '../../shared/stores/cartStore';

const item: CartItem = {
  lineId: '1::',
  id: '1::',
  productId: 1,
  name: 'Steak',
  unitPrice: 50,
  price: 50,
  quantity: 1,
  modifiers: [],
  discountPct: 0,
  discountAmount: 0,
  taxRate: 0,
  status: 'pending',
  notes: '',
};

const printer = {
  name: 'EPSON TM-T88',
  copies: 1,
  cashDrawerKick: true,
  drawerPin: 2 as const,
  transport: 'windows' as const,
  serialPort: '',
  serialBaudRate: 9600,
  networkHost: '',
  networkPort: 9100,
};

beforeEach(() => {
  vi.clearAllMocks();
});

describe('split-tender checkout: $30 cash + $20 card', () => {
  const payments = [
    { id: 'p1', method: 'cash' as const, amount: 30, createdAt: '2026-05-06T00:00:00.000Z' },
    { id: 'p2', method: 'card' as const, amount: 20, createdAt: '2026-05-06T00:00:00.000Z' },
  ];

  it('finalizeMultiPayment calls chargeOrder with collapsed dominant-method payload', async () => {
    const chargeOrder = vi.fn(async () => undefined);

    await finalizeMultiPayment({
      orderId: '99',
      payments,
      tip: 5,
      total: 55,
      chargeOrder,
    });

    // cash ($30) dominates card ($20); total = 30+20+5 tip = 55
    expect(chargeOrder).toHaveBeenCalledWith(99, {
      payment_method: 'cash',
      amount_paid: 55,
    });
  });

  it('rejects an invalid order id before calling chargeOrder', async () => {
    const chargeOrder = vi.fn();

    await expect(
      finalizeMultiPayment({
        orderId: undefined,
        payments,
        tip: 0,
        total: 50,
        chargeOrder,
      }),
    ).rejects.toThrow('Select a valid order before charging.');

    expect(chargeOrder).not.toHaveBeenCalled();
  });

  it('receipt print + drawer kick succeed after multi-payment charge', async () => {
    const chargeOrder = vi.fn(async () => undefined);
    vi.mocked(printReceipt).mockResolvedValue({ status: 'ok' });
    vi.mocked(openConfiguredCashDrawer).mockResolvedValue({ status: 'ok' });

    await finalizeMultiPayment({ orderId: '99', payments, tip: 5, total: 55, chargeOrder });

    const hardwareResult = await printReceiptAndOpenDrawer({
      printer,
      items: [item],
      subtotal: 50,
      taxAmount: 0,
      discountAmount: 0,
      tipAmount: 5,
      total: 55,
      paymentMethod: 'split',
      amountPaid: 55,
      changeDue: 0,
      orderNumber: 'ORD-99',
      businessName: 'Test Grill',
      kickDrawer: true,
    });

    expect(hardwareResult).toEqual({ kind: 'ok' });
    expect(printReceipt).toHaveBeenCalledOnce();
    expect(openConfiguredCashDrawer).toHaveBeenCalledOnce();
  });

  it('receipt reflects tip in total displayed to customer', async () => {
    vi.mocked(printReceipt).mockResolvedValue({ status: 'ok' });
    vi.mocked(openConfiguredCashDrawer).mockResolvedValue({ status: 'ok' });

    await printReceiptAndOpenDrawer({
      printer,
      items: [item],
      subtotal: 50,
      taxAmount: 0,
      discountAmount: 0,
      tipAmount: 5,
      total: 55,
      paymentMethod: 'split',
      amountPaid: 55,
      changeDue: 0,
      orderNumber: 'ORD-99',
      businessName: 'Test Grill',
      kickDrawer: true,
    });

    const receiptData = vi.mocked(printReceipt).mock.calls[0][1];
    expect(receiptData.tipAmount).toBe(5);
    expect(receiptData.total).toBe(55);
  });
});
