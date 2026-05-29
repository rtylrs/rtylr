/**
 * Stock movement tracking and analysis helpers for the ERP app.
 * All functions are pure with no side-effects.
 */

export type MovementType = 'in' | 'out' | 'adjustment' | 'transfer';

export interface StockMovement {
  id: number;
  productId: number;
  productName: string;
  sku: string;
  type: MovementType;
  quantity: number;
  date: string;
  reference?: string;
  warehouseId?: number;
}

export interface MovementSummary {
  totalIn: number;
  totalOut: number;
  netChange: number;
  adjustments: number;
  transfers: number;
}

/** Compute a net quantity change from a movement. */
export function netQuantity(movement: StockMovement): number {
  switch (movement.type) {
    case 'in':
      return movement.quantity;
    case 'out':
      return -movement.quantity;
    case 'adjustment':
      return movement.quantity;
    case 'transfer':
      return 0;
  }
}

/** Summarize a list of movements. */
export function summarizeMovements(movements: StockMovement[]): MovementSummary {
  let totalIn = 0;
  let totalOut = 0;
  let adjustments = 0;
  let transfers = 0;

  for (const m of movements) {
    switch (m.type) {
      case 'in':
        totalIn += m.quantity;
        break;
      case 'out':
        totalOut += m.quantity;
        break;
      case 'adjustment':
        adjustments += m.quantity;
        break;
      case 'transfer':
        transfers += m.quantity;
        break;
    }
  }

  return { totalIn, totalOut, netChange: totalIn - totalOut + adjustments, adjustments, transfers };
}

/** Filter movements by product id. */
export function movementsForProduct(
  movements: StockMovement[],
  productId: number
): StockMovement[] {
  return movements.filter((m) => m.productId === productId);
}

/** Filter movements by type. */
export function movementsOfType(movements: StockMovement[], type: MovementType): StockMovement[] {
  return movements.filter((m) => m.type === type);
}

/** Filter movements within a date range (inclusive, ISO date strings). */
export function movementsInDateRange(
  movements: StockMovement[],
  from: string,
  to: string
): StockMovement[] {
  return movements.filter((m) => m.date >= from && m.date <= to);
}

/** Sort movements by date descending. */
export function sortByDateDesc(movements: StockMovement[]): StockMovement[] {
  return [...movements].sort((a, b) => b.date.localeCompare(a.date));
}

/** Return the most recent N movements. */
export function recentMovements(movements: StockMovement[], n = 10): StockMovement[] {
  return sortByDateDesc(movements).slice(0, n);
}

/** Compute running stock balance from an initial value and a list of movements. */
export function runningBalance(
  initial: number,
  movements: StockMovement[]
): Array<{ date: string; balance: number }> {
  const sorted = [...movements].sort((a, b) => a.date.localeCompare(b.date));
  let balance = initial;
  return sorted.map((m) => {
    balance += netQuantity(m);
    return { date: m.date, balance };
  });
}
