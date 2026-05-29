import { describe, it, expect } from 'vitest';
import {
  totalInvoiceAmount,
  totalInvoiceCount,
  avgInvoiceAmount,
  invoiceBucketLabels,
  invoiceBucketAmounts,
} from './invoice';
import type { InvoiceAgingBucket } from '../hooks/queries/types';

function makeBucket(overrides: Partial<InvoiceAgingBucket> = {}): InvoiceAgingBucket {
  return {
    bucket: '0-30 days',
    invoiceCount: 5,
    amountMinorUnits: 100000,
    ...overrides,
  };
}

describe('totalInvoiceAmount', () => {
  it('sums amountMinorUnits across buckets', () => {
    const buckets = [
      makeBucket({ amountMinorUnits: 50000 }),
      makeBucket({ amountMinorUnits: 30000 }),
    ];
    expect(totalInvoiceAmount(buckets)).toBe(80000);
  });

  it('returns 0 for empty input', () => {
    expect(totalInvoiceAmount([])).toBe(0);
  });
});

describe('totalInvoiceCount', () => {
  it('sums invoiceCount across buckets', () => {
    const buckets = [makeBucket({ invoiceCount: 3 }), makeBucket({ invoiceCount: 7 })];
    expect(totalInvoiceCount(buckets)).toBe(10);
  });

  it('returns 0 for empty input', () => {
    expect(totalInvoiceCount([])).toBe(0);
  });
});

describe('avgInvoiceAmount', () => {
  it('returns average invoice amount in minor units', () => {
    const buckets = [
      makeBucket({ invoiceCount: 2, amountMinorUnits: 40000 }),
      makeBucket({ invoiceCount: 2, amountMinorUnits: 20000 }),
    ];
    expect(avgInvoiceAmount(buckets)).toBe(15000);
  });

  it('returns 0 when there are no invoices', () => {
    expect(avgInvoiceAmount([])).toBe(0);
    expect(avgInvoiceAmount([makeBucket({ invoiceCount: 0, amountMinorUnits: 0 })])).toBe(0);
  });
});

describe('invoiceBucketLabels', () => {
  it('extracts bucket labels', () => {
    const buckets = [makeBucket({ bucket: '0-30 days' }), makeBucket({ bucket: '31-60 days' })];
    expect(invoiceBucketLabels(buckets)).toEqual(['0-30 days', '31-60 days']);
  });
});

describe('invoiceBucketAmounts', () => {
  it('extracts amountMinorUnits', () => {
    const buckets = [
      makeBucket({ amountMinorUnits: 10000 }),
      makeBucket({ amountMinorUnits: 20000 }),
    ];
    expect(invoiceBucketAmounts(buckets)).toEqual([10000, 20000]);
  });
});
