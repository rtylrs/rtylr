// Invoice helpers for the insights financial feature.

import type { InvoiceAgingBucket } from '../hooks/queries/types';

/**
 * totalInvoiceAmount sums amountMinorUnits across all aging buckets.
 */
export function totalInvoiceAmount(buckets: InvoiceAgingBucket[]): number {
  return buckets.reduce((sum, b) => sum + b.amountMinorUnits, 0);
}

/**
 * totalInvoiceCount sums invoiceCount across all aging buckets.
 */
export function totalInvoiceCount(buckets: InvoiceAgingBucket[]): number {
  return buckets.reduce((sum, b) => sum + b.invoiceCount, 0);
}

/**
 * avgInvoiceAmount returns the average invoice amount in minor units.
 * Returns 0 when there are no invoices.
 */
export function avgInvoiceAmount(buckets: InvoiceAgingBucket[]): number {
  const count = totalInvoiceCount(buckets);
  if (count === 0) return 0;
  return Math.round(totalInvoiceAmount(buckets) / count);
}

/**
 * invoiceBucketLabels extracts the bucket label from each entry.
 */
export function invoiceBucketLabels(buckets: InvoiceAgingBucket[]): string[] {
  return buckets.map((b) => b.bucket);
}

/**
 * invoiceBucketAmounts extracts amountMinorUnits from each bucket.
 */
export function invoiceBucketAmounts(buckets: InvoiceAgingBucket[]): number[] {
  return buckets.map((b) => b.amountMinorUnits);
}
