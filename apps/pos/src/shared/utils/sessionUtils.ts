// Pure helpers for POS session and cash drawer management.

/** Calculate the expected cash in the drawer:
 *  opening float + paid_in - paid_out - drops + sales - refunds */
export function expectedCash(params: {
  openingFloat: number;
  sales: number;
  refunds: number;
  paidIn: number;
  paidOut: number;
  drops: number;
}): number {
  const { openingFloat, sales, refunds, paidIn, paidOut, drops } = params;
  const safe = (n: number) => (Number.isFinite(n) && n >= 0 ? n : 0);
  return Math.max(
    0,
    safe(openingFloat) +
      safe(paidIn) +
      safe(sales) -
      safe(refunds) -
      safe(paidOut) -
      safe(drops),
  );
}

/** Calculate the cash variance (actual - expected). */
export function cashVariance(actual: number, expected: number): number {
  const safeActual = Number.isFinite(actual) ? actual : 0;
  const safeExpected = Number.isFinite(expected) ? expected : 0;
  return Math.round((safeActual - safeExpected) * 100) / 100;
}

/** Describe the variance as over, under, or balanced. */
export function varianceLabel(variance: number): 'over' | 'under' | 'balanced' {
  if (!Number.isFinite(variance)) return 'balanced';
  if (variance > 0) return 'over';
  if (variance < 0) return 'under';
  return 'balanced';
}

/** Return true if the session variance is within an acceptable tolerance. */
export function isVarianceAcceptable(variance: number, toleranceAmount = 1): boolean {
  if (!Number.isFinite(variance)) return false;
  return Math.abs(variance) <= toleranceAmount;
}
