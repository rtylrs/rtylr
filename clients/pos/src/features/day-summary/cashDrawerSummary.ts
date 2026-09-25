import type { DaySummaryResponse } from '@voxire/voxie/pos/day_summary/v1/types';
import type { DaySummary } from '../../shared/types/pos';
import type { CashMovement } from '../../shared/stores/cashDrawerStore';

interface BuildLocalDaySummaryParams {
  date: string;
  openingFloat: number;
  closingCash: number;
  movements: CashMovement[];
}

interface DaySummaryFallbacks {
  date: string;
  openingFloat: number;
  closingCash: number;
}

export function normalizeDaySummaryResponse(
  response: DaySummaryResponse,
  fallbacks: DaySummaryFallbacks,
): DaySummary {
  const totalSales = response.totalSales ?? 0;
  const totalRefunds = response.totalRefunds ?? 0;
  const totalOrders = response.totalOrders ?? 0;

  return {
    date: fallbacks.date,
    grossSales: totalSales,
    netSales: totalSales - totalRefunds,
    taxCollected: response.totalTax ?? 0,
    tips: response.totalTips ?? 0,
    refunds: totalRefunds,
    voids: 0,
    orderCount: totalOrders,
    averageTicket: totalOrders > 0 ? totalSales / totalOrders : 0,
    cashDrops: 0,
    openingFloat: fallbacks.openingFloat,
    closingCash: fallbacks.closingCash,
  };
}

export function buildLocalDaySummary({
  date,
  openingFloat,
  closingCash,
  movements,
}: BuildLocalDaySummaryParams): DaySummary {
  const sales = movements.filter((movement) => movement.type === 'sale');
  const grossSales = sales.reduce((sum, movement) => sum + movement.amount, 0);
  const refunds = movements
    .filter((movement) => movement.type === 'refund')
    .reduce((sum, movement) => sum + movement.amount, 0);
  const cashDrops = movements
    .filter((movement) => movement.type === 'drop')
    .reduce((sum, movement) => sum + movement.amount, 0);

  return {
    date,
    grossSales,
    netSales: grossSales - refunds,
    taxCollected: 0,
    tips: 0,
    refunds,
    voids: 0,
    orderCount: sales.length,
    averageTicket: sales.length > 0 ? grossSales / sales.length : 0,
    cashDrops,
    openingFloat,
    closingCash,
  };
}
