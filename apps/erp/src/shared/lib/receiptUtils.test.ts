import { describe, it, expect } from 'vitest';
import {
  lineReceiveStatus,
  receiveSummary,
  pendingLines,
  overReceivedLines,
  canMarkReceived,
  canSendOrder,
  poStatusLabel,
} from './receiptUtils';
import type { ReceiveLine } from './receiptUtils';
import type { PurchaseOrder } from '../services/erpApi';

const makeLine = (ordered: number, received: number, id = 1): ReceiveLine => ({
  productId: id,
  sku: `SKU-${id}`,
  name: `Product ${id}`,
  ordered,
  received,
});

const makePO = (status: PurchaseOrder['status']): PurchaseOrder => ({
  id: 1,
  supplier: 'ACME',
  supplierId: 10,
  date: '2026-05-01',
  expectedDate: '2026-05-10',
  items: 3,
  total: '150.00',
  status,
});

describe('lineReceiveStatus', () => {
  it('returns not-started when received is 0', () => {
    expect(lineReceiveStatus(makeLine(10, 0))).toBe('not-started');
  });

  it('returns complete when fully received', () => {
    expect(lineReceiveStatus(makeLine(10, 10))).toBe('complete');
  });

  it('returns partial when partially received', () => {
    expect(lineReceiveStatus(makeLine(10, 5))).toBe('partial');
  });

  it('returns over-received when received exceeds ordered', () => {
    expect(lineReceiveStatus(makeLine(10, 11))).toBe('over-received');
  });
});

describe('receiveSummary', () => {
  it('computes correct summary for complete receive', () => {
    const lines = [makeLine(10, 10, 1), makeLine(5, 5, 2)];
    const summary = receiveSummary(lines);
    expect(summary.status).toBe('complete');
    expect(summary.completionPct).toBe(100);
    expect(summary.totalPending).toBe(0);
  });

  it('computes partial summary', () => {
    const lines = [makeLine(10, 5, 1), makeLine(10, 0, 2)];
    const summary = receiveSummary(lines);
    expect(summary.status).toBe('partial');
    expect(summary.completionPct).toBe(25);
    expect(summary.totalPending).toBe(15);
  });

  it('returns not-started when nothing received', () => {
    const lines = [makeLine(10, 0)];
    expect(receiveSummary(lines).status).toBe('not-started');
  });
});

describe('pendingLines', () => {
  it('returns lines with pending items', () => {
    const lines = [makeLine(10, 10, 1), makeLine(10, 5, 2)];
    const pending = pendingLines(lines);
    expect(pending).toHaveLength(1);
    expect(pending[0].productId).toBe(2);
  });
});

describe('overReceivedLines', () => {
  it('returns lines where received exceeds ordered', () => {
    const lines = [makeLine(10, 10, 1), makeLine(10, 12, 2)];
    const over = overReceivedLines(lines);
    expect(over).toHaveLength(1);
    expect(over[0].productId).toBe(2);
  });
});

describe('canMarkReceived', () => {
  it('allows marking received for Sent and Approved orders', () => {
    expect(canMarkReceived(makePO('Sent'))).toBe(true);
    expect(canMarkReceived(makePO('Approved'))).toBe(true);
  });

  it('disallows marking received for Draft or already Received orders', () => {
    expect(canMarkReceived(makePO('Draft'))).toBe(false);
    expect(canMarkReceived(makePO('Received'))).toBe(false);
  });
});

describe('canSendOrder', () => {
  it('allows sending Draft and Pending Approval orders', () => {
    expect(canSendOrder(makePO('Draft'))).toBe(true);
    expect(canSendOrder(makePO('Pending Approval'))).toBe(true);
  });

  it('disallows sending already Sent orders', () => {
    expect(canSendOrder(makePO('Sent'))).toBe(false);
    expect(canSendOrder(makePO('Received'))).toBe(false);
  });
});

describe('poStatusLabel', () => {
  it('returns correct labels for all statuses', () => {
    expect(poStatusLabel('Draft')).toBe('Draft');
    expect(poStatusLabel('Sent')).toBe('Sent to Supplier');
    expect(poStatusLabel('Received')).toBe('Received');
    expect(poStatusLabel('Pending Approval')).toBe('Awaiting Approval');
    expect(poStatusLabel('Approved')).toBe('Approved');
  });
});
