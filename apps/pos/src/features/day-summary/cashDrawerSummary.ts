import type { DaySummary } from '../../shared/types/pos';
import type { CashMovement } from '../../shared/stores/cashDrawerStore';
import { daySummarySchema } from '../../shared/utils/schemas/daySummary';

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

function record(value: unknown): Record<string, unknown> | null {
  return typeof value === 'object' && value !== null && !Array.isArray(value)
    ? (value as Record<string, unknown>)
    : null;
}

function numberOrUndefined(value: unknown): number | undefined {
  if (value === undefined || value === null || value === '') return undefined;
  const numeric = Number(value);
  return Number.isFinite(numeric) ? numeric : undefined;
}

function textOrUndefined(value: unknown): string | undefined {
  const normalized = String(value ?? '').trim();
  return normalized || undefined;
}

function daySummaryPayload(response: unknown): Record<string, unknown> {
  const root = record(response);
  if (!root) return {};
  return record(root.data) ?? root;
}

export function normalizeDaySummaryResponse(
  response: unknown,
  fallbacks: DaySummaryFallbacks,
): DaySummary {
  const raw = daySummaryPayload(response);

  return daySummarySchema.parse({
    date: textOrUndefined(raw.date) ?? fallbacks.date,
    grossSales: numberOrUndefined(raw.grossSales ?? raw.gross_sales),
    netSales: numberOrUndefined(raw.netSales ?? raw.net_sales),
    taxCollected: numberOrUndefined(raw.taxCollected ?? raw.tax_collected),
    tips: numberOrUndefined(raw.tips),
    refunds: numberOrUndefined(raw.refunds),
    voids: numberOrUndefined(raw.voids),
    orderCount: numberOrUndefined(raw.orderCount ?? raw.order_count),
    averageTicket: numberOrUndefined(raw.averageTicket ?? raw.average_ticket),
    cashDrops: numberOrUndefined(raw.cashDrops ?? raw.cash_drops),
    openingFloat: numberOrUndefined(raw.openingFloat ?? raw.opening_float) ?? fallbacks.openingFloat,
    closingCash: numberOrUndefined(raw.closingCash ?? raw.closing_cash) ?? fallbacks.closingCash,
  });
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
