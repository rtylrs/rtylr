/**
 * End-to-end integration: order creation → receipt print → cash drawer kick.
 *
 * This is the critical path executed on every successful sale.
 * Tests prove all three stages chain correctly and that a failure at any
 * stage does not silently swallow the error.
 */
import { describe, expect, it, vi, beforeEach } from 'vitest';
import { buildCheckoutOrder, buildCreateOrderPayload, processCheckout } from './checkoutWorkflow';
import { printReceiptAndOpenDrawer } from './paymentHardware';
import type { CartItem } from '../../shared/stores/cartStore';

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

const item: CartItem = {
  lineId: '1::',
  id: '1::',
  productId: 1,
  name: 'Espresso',
  unitPrice: 4.5,
  price: 4.5,
  quantity: 2,
  modifiers: [],
  discountPct: 0,
  discountAmount: 0,
  taxRate: 0,
  status: 'pending',
  notes: 'oat milk',
};

const printerWindows = {
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

async function runFullSale() {
  const order = buildCheckoutOrder({
    id: 'local-42',
    items: [item],
    subtotal: 9,
    tax: 0.72,
    total: 9.72,
    customerId: 3,
    createdAt: '2026-05-13T10:00:00.000Z',
  });

  const payload = buildCreateOrderPayload({ items: [item], method: 'cash' });

  const checkoutResult = await processCheckout({
    order,
    payload,
    method: 'cash',
    createOrder: vi.fn(() => Promise.resolve({ orderId: '500' })),
    chargeOrder: vi.fn(() => Promise.resolve()),
    enqueueOffline: vi.fn(),
    recordCashSale: vi.fn(),
  });

  const hardwareResult = await printReceiptAndOpenDrawer({
    printer: printerWindows,
    items: [item],
    subtotal: 9,
    taxAmount: 0.72,
    discountAmount: 0,
    tipAmount: 0,
    total: 9.72,
    paymentMethod: 'cash',
    amountPaid: 10,
    changeDue: 0.28,
    orderNumber: `ORD-${(checkoutResult as { orderId: number }).orderId}`,
    businessName: 'Test Cafe',
    kickDrawer: true,
  });

  return { checkoutResult, hardwareResult };
}

describe('full sale flow: order → receipt → drawer', () => {
  it('completes end-to-end with all stages succeeding', async () => {
    vi.mocked(printReceipt).mockResolvedValue({ status: 'printed', message: 'receipt printed' });
    vi.mocked(openConfiguredCashDrawer).mockResolvedValue({ status: 'opened' });

    const { checkoutResult, hardwareResult } = await runFullSale();

    expect(checkoutResult).toEqual({ status: 'online', orderId: 500 });
    expect(hardwareResult).toEqual({ kind: 'ok' });
    expect(printReceipt).toHaveBeenCalledOnce();
    expect(openConfiguredCashDrawer).toHaveBeenCalledOnce();
  });

  it('receipt includes the server order id in the order number', async () => {
    vi.mocked(printReceipt).mockResolvedValue({ status: 'printed', message: 'receipt printed' });
    vi.mocked(openConfiguredCashDrawer).mockResolvedValue({ status: 'opened' });

    await runFullSale();

    const receiptData = vi.mocked(printReceipt).mock.calls[0][1];
    expect(receiptData.orderNumber).toBe('ORD-500');
    expect(receiptData.businessName).toBe('Test Cafe');
    expect(receiptData.items[0].name).toBe('Espresso');
  });

  it('checkout succeeds even when printer is offline - hardware failure is independent', async () => {
    vi.mocked(printReceipt).mockResolvedValue({ status: 'failed', error: 'printer offline' });

    const { checkoutResult, hardwareResult } = await runFullSale();

    expect(checkoutResult).toEqual({ status: 'online', orderId: 500 });
    expect(hardwareResult).toEqual({ kind: 'print_failed', error: 'printer offline' });
  });

  it('checkout falls back to offline queue when server is unreachable, hardware still fires', async () => {
    vi.mocked(printReceipt).mockResolvedValue({ status: 'printed', message: 'receipt printed' });
    vi.mocked(openConfiguredCashDrawer).mockResolvedValue({ status: 'opened' });

    const order = buildCheckoutOrder({
      id: 'local-99',
      items: [item],
      subtotal: 9,
      tax: 0.72,
      total: 9.72,
      customerId: 3,
      createdAt: '2026-05-13T10:00:00.000Z',
    });

    const payload = buildCreateOrderPayload({ items: [item], method: 'cash' });
    const enqueueOffline = vi.fn();

    const checkoutResult = await processCheckout({
      order,
      payload,
      method: 'cash',
      createOrder: vi.fn(() => Promise.reject(new Error('Network Error'))),
      chargeOrder: vi.fn(),
      enqueueOffline,
      recordCashSale: vi.fn(),
    });

    expect(checkoutResult).toEqual({ status: 'queued' });
    expect(enqueueOffline).toHaveBeenCalledWith(payload);

    const hardwareResult = await printReceiptAndOpenDrawer({
      printer: printerWindows,
      items: [item],
      subtotal: 9,
      taxAmount: 0.72,
      discountAmount: 0,
      tipAmount: 0,
      total: 9.72,
      paymentMethod: 'cash',
      amountPaid: 10,
      changeDue: 0.28,
      orderNumber: 'OFFLINE',
      businessName: 'Test Cafe',
      kickDrawer: true,
    });

    expect(hardwareResult).toEqual({ kind: 'ok' });
  });
});
