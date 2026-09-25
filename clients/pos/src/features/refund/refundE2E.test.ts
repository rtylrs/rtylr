/**
 * End-to-end: processRefund chains API call → cash drawer adjustment.
 * Cash drawer movement is driven by the tender the cashier confirms in the
 * Refund UI, and the full-refund total comes from server order data first.
 */
import { describe, expect, it, vi } from 'vitest';
import { processRefund } from './refundWorkflow';

describe('processRefund - full cash refund', () => {
  it('calls API with the confirmed order total and triggers cash drawer adjustment', async () => {
    const refundOrder = vi.fn(async () => undefined);
    const recordCashRefund = vi.fn();

    await processRefund({
      orderId: '1',
      fullRefund: true,
      amountInput: '',
      reason: 'Customer changed mind',
      orderTotal: 50,
      refundMethod: 'cash',
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
  it('calls API with the confirmed total but does not trigger cash drawer', async () => {
    const refundOrder = vi.fn(async () => undefined);
    const recordCashRefund = vi.fn();

    await processRefund({
      orderId: '2',
      fullRefund: true,
      amountInput: '',
      reason: 'Wrong item',
      orderTotal: 50,
      refundMethod: 'card',
      refundOrder,
      recordCashRefund,
    });

    expect(refundOrder).toHaveBeenCalledWith({ orderId: '2', amount: 50, reason: 'Wrong item' });
    // Card payment → no cash in drawer to return
    expect(recordCashRefund).not.toHaveBeenCalled();
  });
});

describe('processRefund - partial refund', () => {
  it('sends only the partial amount and adjusts drawer for a confirmed cash refund', async () => {
    const refundOrder = vi.fn(async () => undefined);
    const recordCashRefund = vi.fn();

    await processRefund({
      orderId: '1',
      fullRefund: false,
      amountInput: '20',
      reason: 'Partial complaint',
      refundMethod: 'cash',
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
        refundMethod: 'cash',
        refundOrder,
        recordCashRefund,
      }),
    ).rejects.toThrow('Enter a refund amount greater than zero.');

    expect(refundOrder).not.toHaveBeenCalled();
  });

  it('does not adjust the drawer when no method was confirmed', async () => {
    const refundOrder = vi.fn(async () => undefined);
    const recordCashRefund = vi.fn();

    await processRefund({
      orderId: '1',
      fullRefund: false,
      amountInput: '20',
      reason: 'Partial complaint',
      refundOrder,
      recordCashRefund,
    });

    expect(recordCashRefund).not.toHaveBeenCalled();
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
        refundMethod: 'cash',
        refundOrder,
        recordCashRefund,
      }),
    ).rejects.toThrow('Network error');

    expect(recordCashRefund).not.toHaveBeenCalled();
  });
});
