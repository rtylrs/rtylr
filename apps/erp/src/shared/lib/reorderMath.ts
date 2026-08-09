/**
 * Calculate an automatic reorder point.
 *
 * Reorder point = (average daily sales × lead time) + safety stock
 * Safety stock is ~ half a standard deviation of demand during lead time.
 */
export function calculateReorderPoint(
  avgDailySales: number,
  leadTimeDays: number,
  safetyFactor = 0.5
): number {
  const demandDuringLeadTime = avgDailySales * leadTimeDays;
  const safetyStock = Math.ceil(demandDuringLeadTime * safetyFactor);
  return Math.max(0, Math.ceil(demandDuringLeadTime + safetyStock));
}

/**
 * Economic Order Quantity (EOQ) = sqrt((2 × D × S) / H)
 * D = annual demand, S = order cost, H = holding cost per unit per year
 */
export function calculateEOQ(annualDemand: number, orderCost: number, holdingCost: number): number {
  if (annualDemand <= 0 || orderCost <= 0 || holdingCost <= 0) return 0;
  return Math.ceil(Math.sqrt((2 * annualDemand * orderCost) / holdingCost));
}

export function calculateDaysCoverage(currentStock: number, avgDailySales: number): number {
  if (avgDailySales <= 0) return Infinity;
  return Math.floor(currentStock / avgDailySales);
}
