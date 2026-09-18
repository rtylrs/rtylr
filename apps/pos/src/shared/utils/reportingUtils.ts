// Pure helpers for daily and period reporting calculations.

import type { DaySummary } from '../types/pos';

/** Calculate net sales (gross minus refunds). */
export function netSales(grossSales: number, refunds: number): number {
  if (!Number.isFinite(grossSales) || !Number.isFinite(refunds)) return 0;
  return Math.max(0, grossSales - refunds);
}

/** Calculate average ticket size given total sales and order count. */
export function averageTicket(totalSales: number, orderCount: number): number {
  if (!Number.isFinite(totalSales) || orderCount <= 0) return 0;
  return totalSales / orderCount;
}

/** Calculate the refund rate as a percentage of gross sales. */
export function refundRate(refunds: number, grossSales: number): number {
  if (!Number.isFinite(refunds) || !Number.isFinite(grossSales) || grossSales <= 0) return 0;
  return (refunds / grossSales) * 100;
}

/** Calculate the void rate as a percentage of gross sales. */
export function voidRate(voids: number, grossSales: number): number {
  if (!Number.isFinite(voids) || !Number.isFinite(grossSales) || grossSales <= 0) return 0;
  return (voids / grossSales) * 100;
}

/** Calculate the tip rate as a percentage of net sales. */
export function tipRate(tips: number, netSalesAmount: number): number {
  if (!Number.isFinite(tips) || !Number.isFinite(netSalesAmount) || netSalesAmount <= 0) return 0;
  return (tips / netSalesAmount) * 100;
}

/** Calculate the tax rate as a percentage of net sales. */
export function effectiveTaxRate(taxCollected: number, netSalesAmount: number): number {
  if (!Number.isFinite(taxCollected) || !Number.isFinite(netSalesAmount) || netSalesAmount <= 0) {
    return 0;
  }
  return (taxCollected / netSalesAmount) * 100;
}

/** Sum a numeric field across an array of day summaries. */
export function sumField(summaries: DaySummary[], field: keyof DaySummary): number {
  return summaries.reduce((acc, s) => {
    const val = s[field];
    return acc + (typeof val === 'number' && Number.isFinite(val) ? val : 0);
  }, 0);
}

/** Return the day summary with the highest gross sales. */
export function bestDay(summaries: DaySummary[]): DaySummary | undefined {
  if (!summaries.length) return undefined;
  return summaries.reduce((best, s) => (s.grossSales > best.grossSales ? s : best));
}

/** Return the day summary with the lowest gross sales. */
export function worstDay(summaries: DaySummary[]): DaySummary | undefined {
  if (!summaries.length) return undefined;
  return summaries.reduce((worst, s) => (s.grossSales < worst.grossSales ? s : worst));
}

/** Return a period total summary aggregated from an array of day summaries. */
export function aggregateSummaries(summaries: DaySummary[]): Omit<DaySummary, 'date'> {
  const grossSales = sumField(summaries, 'grossSales');
  const refunds = sumField(summaries, 'refunds');
  const orderCount = sumField(summaries, 'orderCount');
  return {
    grossSales,
    netSales: netSales(grossSales, refunds),
    taxCollected: sumField(summaries, 'taxCollected'),
    tips: sumField(summaries, 'tips'),
    refunds,
    voids: sumField(summaries, 'voids'),
    orderCount,
    averageTicket: averageTicket(grossSales, orderCount),
    cashDrops: sumField(summaries, 'cashDrops'),
    openingFloat: summaries.length > 0 ? summaries[0].openingFloat : 0,
    closingCash: summaries.length > 0 ? summaries[summaries.length - 1].closingCash : 0,
  };
}
