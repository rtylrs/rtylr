/**
 * Complete sale chain: checkout → receipt print → drawer kick → KDS ticket → inventory adjustment.
 *
 * This is the critical path for every real-world sale. Each stage must succeed,
 * and a failure at any hardware stage must not roll back the payment.
 */
import { describe, expect, it, vi, beforeEach } from 'vitest';
import { buildCheckoutOrder, buildCreateOrderPayload, processCheckout } from '../payment/checkoutWorkflow';
import { printReceiptAndOpenDrawer } from '../payment/paymentHardware';
import { sendKdsTicket } from '../../shared/services/hardware/kdsHardware';
import { buildStockAction } from '../inventory/inventoryWorkflow';
import type { CartItem } from '../../shared/stores/cartStore';
import type { KdsHardwareSettings, PrinterSettings } from '../../shared/stores/settingsStore';

// ── Mocks ─────────────────────────────────────────────────────────────────────

vi.mock('@tauri-apps/api/core', () => ({ invoke: vi.fn() }));
vi.mock('../../shared/services/hardware/receiptPrinter', () => ({ printReceipt: vi.fn() }));
vi.mock('../../shared/services/hardware/cashDrawerHardware', () => ({ openConfiguredCashDrawer: vi.fn() }));
vi.mock('../../shared/services/hardware/crossPlatformDrawer', () => ({
  openDrawerSerial: vi.fn(),
  openDrawerNetwork: vi.fn(),
}));

import { invoke } from '@tauri-apps/api/core';
import { printReceipt } from '../../shared/services/hardware/receiptPrinter';
import { openConfiguredCashDrawer } from '../../shared/services/hardware/cashDrawerHardware';

// ── Fixtures ──────────────────────────────────────────────────────────────────

const item: CartItem = {
  lineId: '5::',
  id: '5::',
  productId: 5,
  name: 'Margherita Pizza',
  unitPrice: 14.5,
  price: 14.5,
  quantity: 2,
  modifiers: [{ id: 1, name: 'Extra cheese', priceDelta: 0 }],
  discountPct: 0,
  discountAmount: 0,
  taxRate: 0.08,
  status: 'pending',
  notes: 'thin crust',
};

const printer: PrinterSettings = {
  name: 'EPSON TM-T88VI',
  copies: 1,
  cashDrawerKick: true,
  drawerPin: 2,
  transport: 'windows',
  serialPort: '',
  serialBaudRate: 9600,
  networkHost: '',
  networkPort: 9100,
};

const kds: KdsHardwareSettings = {
  transport: 'network',
  serialPort: '',
  serialBaudRate: 9600,
  networkHost: '192.168.1.50',
  networkPort: 9100,
};

beforeEach(() => {
  vi.clearAllMocks();
});

// ── Helper: run the complete sale chain ───────────────────────────────────────

async function runCompleteSale() {
  const order = buildCheckoutOrder({
    id: 'local-chain-1',
    items: [item],
    subtotal: 29,
    tax: 2.32,
    total: 31.32,
    customerId: 3,
    createdAt: '2026-05-13T12:00:00.000Z',
  });

  const payload = buildCreateOrderPayload({ items: [item], method: 'cash' });
  const createOrder = vi.fn(() => Promise.resolve({ orderId: '999' }));
  const chargeOrder = vi.fn(() => Promise.resolve());

  const checkoutResult = await processCheckout({
    order,
    payload,
    method: 'cash',
    createOrder,
    chargeOrder,
    enqueueOffline: vi.fn(),
    recordCashSale: vi.fn(),
  });

  const hardwareResult = await printReceiptAndOpenDrawer({
    printer,
    items: [item],
    subtotal: 29,
    taxAmount: 2.32,
    discountAmount: 0,
    tipAmount: 0,
    total: 31.32,
    paymentMethod: 'cash',
    amountPaid: 40,
    changeDue: 8.68,
    orderNumber: `ORD-${(checkoutResult as { orderId: number }).orderId}`,
    businessName: 'rtylr POS',
    kickDrawer: true,
  });

  const kdsResult = await sendKdsTicket(kds, {
    orderNumber: `ORD-${(checkoutResult as { orderId: number }).orderId}`,
    station: 'KITCHEN',
    orderType: 'Dine In',
    tableNumber: '4',
    items: [
      {
        name: item.name,
        quantity: item.quantity,
        notes: item.notes,
        modifiers: item.modifiers.map((m) => m.name),
      },
    ],
    createdAt: '2026-05-13T12:00:00.000Z',
  });

  const stockAction = buildStockAction({
    itemId: item.productId,
    action: 'adjust',
    quantityInput: String(item.quantity),
    branchId: 1,
    reason: `Sold in ORD-${(checkoutResult as { orderId: number }).orderId}`,
  });

  return { checkoutResult, hardwareResult, kdsResult, stockAction };
}

// ── Tests ─────────────────────────────────────────────────────────────────────

describe('complete sale chain: checkout → receipt → drawer → KDS → inventory', () => {
  it('all 4 stages succeed on the happy path', async () => {
    vi.mocked(printReceipt).mockResolvedValue({ status: 'printed', message: 'ok' });
    vi.mocked(openConfiguredCashDrawer).mockResolvedValue({ status: 'opened' });
    vi.mocked(invoke).mockResolvedValue('KDS ticket sent to 192.168.1.50:9100');

    const { checkoutResult, hardwareResult, kdsResult, stockAction } = await runCompleteSale();

    expect(checkoutResult).toEqual({ status: 'online', orderId: 999 });

    expect(hardwareResult).toEqual({ kind: 'ok' });
    expect(printReceipt).toHaveBeenCalledOnce();
    expect(openConfiguredCashDrawer).toHaveBeenCalledOnce();

    expect(kdsResult.status).toBe('sent');

    expect(stockAction.action).toBe('adjust');
    expect(stockAction.payload.productId).toBe('5');
    expect(stockAction.payload.quantity).toBe(2);
    expect(stockAction.payload.note).toContain('ORD-999');
  });

  it('server order id propagates through all stages', async () => {
    vi.mocked(printReceipt).mockResolvedValue({ status: 'printed', message: 'ok' });
    vi.mocked(openConfiguredCashDrawer).mockResolvedValue({ status: 'opened' });
    vi.mocked(invoke).mockResolvedValue('ok');

    const { checkoutResult } = await runCompleteSale();

    const orderId = (checkoutResult as { orderId: number }).orderId;
    expect(orderId).toBe(999);

    const receiptArg = vi.mocked(printReceipt).mock.calls[0][1];
    expect(receiptArg.orderNumber).toBe(`ORD-${orderId}`);
  });

  it('payment is preserved even when receipt printer is offline', async () => {
    vi.mocked(printReceipt).mockResolvedValue({ status: 'failed', error: 'printer offline' });
    vi.mocked(invoke).mockResolvedValue('ok');

    const { checkoutResult, hardwareResult } = await runCompleteSale();

    expect(checkoutResult).toEqual({ status: 'online', orderId: 999 });
    expect(hardwareResult).toEqual({ kind: 'print_failed', error: 'printer offline' });
  });

  it('KDS ticket is skipped without affecting payment when transport is none', async () => {
    vi.mocked(printReceipt).mockResolvedValue({ status: 'printed', message: 'ok' });
    vi.mocked(openConfiguredCashDrawer).mockResolvedValue({ status: 'opened' });

    const order = buildCheckoutOrder({
      id: 'local-chain-2',
      items: [item],
      subtotal: 29,
      tax: 2.32,
      total: 31.32,
      customerId: 1,
      createdAt: '2026-05-13T12:00:00.000Z',
    });

    const payload = buildCreateOrderPayload({ items: [item], method: 'card' });

    const checkoutResult = await processCheckout({
      order,
      payload,
      method: 'card',
      createOrder: vi.fn(() => Promise.resolve({ orderId: '1000' })),
      chargeOrder: vi.fn(() => Promise.resolve()),
      enqueueOffline: vi.fn(),
      recordCashSale: vi.fn(),
    });

    const kdsResult = await sendKdsTicket(
      { ...kds, transport: 'none' },
      { orderNumber: 'ORD-1000', items: [{ name: 'Pizza', quantity: 1, modifiers: [] }] },
    );

    expect(checkoutResult).toEqual({ status: 'online', orderId: 1000 });
    expect(kdsResult).toEqual({ status: 'skipped' });
    expect(invoke).not.toHaveBeenCalled();
  });

  it('offline checkout still prints receipt and sends KDS ticket', async () => {
    vi.mocked(printReceipt).mockResolvedValue({ status: 'printed', message: 'ok' });
    vi.mocked(openConfiguredCashDrawer).mockResolvedValue({ status: 'opened' });
    vi.mocked(invoke).mockResolvedValue('sent');

    const order = buildCheckoutOrder({
      id: 'local-offline-chain',
      items: [item],
      subtotal: 29,
      tax: 2.32,
      total: 31.32,
      customerId: 1,
      createdAt: '2026-05-13T12:00:00.000Z',
    });

    const payload = buildCreateOrderPayload({ items: [item], method: 'cash' });
    const enqueueOffline = vi.fn();

    const checkoutResult = await processCheckout({
      order,
      payload,
      method: 'cash',
      createOrder: vi.fn(async () => { throw new Error('Network Error'); }),
      chargeOrder: vi.fn(),
      enqueueOffline,
      recordCashSale: vi.fn(),
    });

    expect(checkoutResult).toEqual({ status: 'queued' });
    expect(enqueueOffline).toHaveBeenCalledWith(payload);

    const hardwareResult = await printReceiptAndOpenDrawer({
      printer,
      items: [item],
      subtotal: 29,
      taxAmount: 2.32,
      discountAmount: 0,
      tipAmount: 0,
      total: 31.32,
      paymentMethod: 'cash',
      amountPaid: 35,
      changeDue: 3.68,
      orderNumber: 'OFFLINE',
      businessName: 'rtylr POS',
      kickDrawer: true,
    });

    const kdsResult = await sendKdsTicket(kds, {
      orderNumber: 'OFFLINE',
      items: [{ name: item.name, quantity: item.quantity, modifiers: [] }],
    });

    expect(hardwareResult).toEqual({ kind: 'ok' });
    expect(kdsResult.status).toBe('sent');
  });
});
