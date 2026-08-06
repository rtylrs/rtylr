/**
 * End-to-end: processRefund chains API call → cash drawer adjustment.
 * Covers full refund, partial refund, and cash-vs-card split behaviour.
 */
import { describe, expect, it, vi } from 'vitest';
import { processRefund } from './refundWorkflow';
import type { Receipt } from '../../shared/types/pos';

const cashReceipt: Receipt = {
  id: 'receipt-1',
  orderId: 'order-1',
  total: 50,
  printedAt: '2026-05-13T10:00:00.000Z',
  items: [],
  payments: [{ id: 'p-1', method: 'cash', amount: 50, createdAt: '2026-05-13T10:00:00.000Z' }],
};

const cardReceipt: Receipt = {
  id: 'receipt-2',
  orderId: 'order-2',
  total: 50,
  printedAt: '2026-05-13T10:00:00.000Z',
  items: [],
  payments: [{ id: 'p-2', method: 'card', amount: 50, createdAt: '2026-05-13T10:00:00.000Z' }],
};

const splitReceipt: Receipt = {
  id: 'receipt-3',
  orderId: 'order-3',
  total: 80,
  printedAt: '2026-05-13T10:00:00.000Z',
  items: [],
  payments: [
    { id: 'p-3a', method: 'cash', amount: 30, createdAt: '2026-05-13T10:00:00.000Z' },
    { id: 'p-3b', method: 'card', amount: 50, createdAt: '2026-05-13T10:00:00.000Z' },
  ],
};

describe('processRefund - full cash refund', () => {
  it('calls API with receipt total and triggers cash drawer adjustment', async () => {
    const refundOrder = vi.fn(async () => undefined);
    const recordCashRefund = vi.fn();

    await processRefund({
      orderId: '1',
      fullRefund: true,
      amountInput: '',
      reason: 'Customer changed mind',
      receipt: cashReceipt,
      refundOrder,
      recordCashRefund,
    });

    expect(refundOrder).toHaveBeenCalledWith({
      orderId: '1',
      amount: 50,
      reason: 'Customer changed mind',
    });
    expect(recordCashRefund).toHaveBeenCalledWith(50, '1');
  });
});

describe('processRefund - full card refund', () => {
  it('calls API with receipt total but does not trigger cash drawer', async () => {
    const refundOrder = vi.fn(async () => undefined);
    const recordCashRefund = vi.fn();

    await processRefund({
      orderId: '2',
      fullRefund: true,
      amountInput: '',
      reason: 'Wrong item',
      receipt: cardReceipt,
      refundOrder,
      recordCashRefund,
    });

    expect(refundOrder).toHaveBeenCalledWith({ orderId: '2', amount: 50, reason: 'Wrong item' });
    // Card payment → no cash in drawer to return
    expect(recordCashRefund).not.toHaveBeenCalled();
  });
});

describe('processRefund - partial refund', () => {
  it('sends only the partial amount and adjusts drawer accordingly', async () => {
    const refundOrder = vi.fn(async () => undefined);
    const recordCashRefund = vi.fn();

    await processRefund({
      orderId: '1',
      fullRefund: false,
      amountInput: '20',
      reason: 'Partial complaint',
      receipt: cashReceipt,
      refundOrder,
      recordCashRefund,
    });

    expect(refundOrder).toHaveBeenCalledWith({ orderId: '1', amount: 20, reason: 'Partial complaint' });
    expect(recordCashRefund).toHaveBeenCalledWith(20, '1');
  });

  it('rejects partial refund with zero amount', async () => {
    const refundOrder = vi.fn();
    const recordCashRefund = vi.fn();

    await expect(
      processRefund({
        orderId: '1',
        fullRefund: false,
        amountInput: '0',
        reason: 'Test',
        receipt: cashReceipt,
        refundOrder,
        recordCashRefund,
      }),
    ).rejects.toThrow('Enter a refund amount greater than zero.');

    expect(refundOrder).not.toHaveBeenCalled();
  });
});

describe('processRefund - split payment receipt', () => {
  it('caps cash drawer refund to cash portion when partial refund exceeds it', async () => {
    const refundOrder = vi.fn(async () => undefined);
    const recordCashRefund = vi.fn();

    // $40 refund on an order where only $30 was paid cash
    await processRefund({
      orderId: '3',
      fullRefund: false,
      amountInput: '40',
      reason: 'Partial',
      receipt: splitReceipt,
      refundOrder,
      recordCashRefund,
    });

    expect(refundOrder).toHaveBeenCalledWith({ orderId: '3', amount: 40, reason: 'Partial' });
    // Cash drawer only gets back the $30 that was cash - not the card portion
    expect(recordCashRefund).toHaveBeenCalledWith(30, '3');
  });
});

describe('processRefund - API failure', () => {
  it('does not record cash refund if API call throws', async () => {
    const refundOrder = vi.fn(async () => {
      throw new Error('Network error');
    });
    const recordCashRefund = vi.fn();

    await expect(
      processRefund({
        orderId: '1',
        fullRefund: false,
        amountInput: '15',
        reason: 'Test failure',
        receipt: cashReceipt,
        refundOrder,
        recordCashRefund,
      }),
    ).rejects.toThrow('Network error');

    expect(recordCashRefund).not.toHaveBeenCalled();
  });
});
