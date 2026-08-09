/**
 * Multi-currency conversion helpers for the ERP.
 * All functions are pure with no side-effects.
 */

export interface ExchangeRate {
  from: string;
  to: string;
  rate: number;
}

/** Convert an amount from one currency to another using the provided rates. */
export function convertCurrency(
  amount: number,
  from: string,
  to: string,
  rates: ExchangeRate[]
): number | null {
  if (from === to) return amount;
  const direct = rates.find((r) => r.from === from && r.to === to);
  if (direct) return amount * direct.rate;
  const inverse = rates.find((r) => r.from === to && r.to === from);
  if (inverse) return amount / inverse.rate;
  return null;
}

/** Format a number as a currency string with symbol. */
export function formatCurrency(amount: number, currency: string, locale = 'en-US'): string {
  return new Intl.NumberFormat(locale, { style: 'currency', currency }).format(amount);
}

/** Round a currency amount to the given decimal places (default 2). */
export function roundCurrency(amount: number, decimals = 2): number {
  const factor = Math.pow(10, decimals);
  return Math.round(amount * factor) / factor;
}

/** Return the currency symbol for a currency code. */
export function currencySymbol(currency: string, locale = 'en-US'): string {
  return (
    new Intl.NumberFormat(locale, { style: 'currency', currency, minimumFractionDigits: 0 })
      .formatToParts(0)
      .find((p) => p.type === 'currency')?.value ?? currency
  );
}

/** Build a lookup map from an array of exchange rates for O(1) access. */
export function buildRateMap(rates: ExchangeRate[]): Map<string, number> {
  const map = new Map<string, number>();
  for (const r of rates) {
    map.set(`${r.from}:${r.to}`, r.rate);
  }
  return map;
}

/** Convert an amount using a pre-built rate map. */
export function convertWithMap(
  amount: number,
  from: string,
  to: string,
  map: Map<string, number>
): number | null {
  if (from === to) return amount;
  const rate = map.get(`${from}:${to}`);
  if (rate !== undefined) return amount * rate;
  const inverseRate = map.get(`${to}:${from}`);
  if (inverseRate !== undefined) return amount / inverseRate;
  return null;
}
