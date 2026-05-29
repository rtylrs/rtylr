/**
 * Tax rate and tax calculation helpers for the ERP app.
 * All functions are pure with no side-effects.
 */

export interface TaxRate {
  id: number;
  name: string;
  rate: number;
  country?: string;
  region?: string;
  isDefault: boolean;
}

export interface TaxBreakdown {
  subtotal: number;
  taxAmount: number;
  total: number;
  effectiveRate: number;
}

/** Apply a tax rate to a subtotal amount (rate as a percentage 0-100). */
export function applyTax(subtotal: number, ratePct: number): number {
  return subtotal * (ratePct / 100);
}

/** Compute the full tax breakdown for a subtotal and rate. */
export function taxBreakdown(subtotal: number, ratePct: number): TaxBreakdown {
  const taxAmount = applyTax(subtotal, ratePct);
  const total = subtotal + taxAmount;
  return {
    subtotal,
    taxAmount: Math.round(taxAmount * 100) / 100,
    total: Math.round(total * 100) / 100,
    effectiveRate: ratePct,
  };
}

/** Extract the subtotal from a tax-inclusive total. */
export function backCalculateSubtotal(totalInclTax: number, ratePct: number): number {
  return totalInclTax / (1 + ratePct / 100);
}

/** Extract the tax amount from a tax-inclusive total. */
export function extractTaxAmount(totalInclTax: number, ratePct: number): number {
  const subtotal = backCalculateSubtotal(totalInclTax, ratePct);
  return totalInclTax - subtotal;
}

/** Find the default tax rate from a list. */
export function defaultTaxRate(rates: TaxRate[]): TaxRate | undefined {
  return rates.find((r) => r.isDefault);
}

/** Format a tax rate for display (e.g. "GST 10%"). */
export function formatTaxRate(rate: TaxRate): string {
  return `${rate.name} ${rate.rate}%`;
}

/** Apply a combined tax (multiple rates stacked). */
export function combinedTaxAmount(subtotal: number, rates: TaxRate[]): number {
  return rates.reduce((total, r) => total + applyTax(subtotal, r.rate), 0);
}

/** Return the effective combined rate for multiple tax rates. */
export function effectiveCombinedRate(rates: TaxRate[]): number {
  return rates.reduce((sum, r) => sum + r.rate, 0);
}

/** Round a currency amount to 2 decimal places. */
export function roundCurrency(amount: number): number {
  return Math.round(amount * 100) / 100;
}
