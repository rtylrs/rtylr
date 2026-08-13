// Central money helper. All money fields on the wire are integer minor units
// (cents). Hooks NEVER divide by 100 inline - they pass the raw minor value
// through, and display components call `formatMoney` which handles the
// conversion via `minorToMajor`.
//
// JPY has zero fractional digits, so its "minor" unit is the whole yen. All
// other currencies in this app have two fractional digits.

export type Currency = 'USD' | 'EUR' | 'GBP' | 'JPY' | 'CAD' | 'AUD' | 'CHF' | 'CNY';

const ZERO_DECIMAL_CURRENCIES: ReadonlySet<Currency> = new Set<Currency>(['JPY']);

function fractionDigitsForCurrency(currency: Currency): number {
  return ZERO_DECIMAL_CURRENCIES.has(currency) ? 0 : 2;
}

export function minorToMajor(amount: number, currency: Currency): number {
  const digits = fractionDigitsForCurrency(currency);
  const divisor = Math.pow(10, digits);
  return amount / divisor;
}

export function majorToMinor(amount: number, currency: Currency): number {
  const digits = fractionDigitsForCurrency(currency);
  return Math.round(amount * Math.pow(10, digits));
}

export function formatMoney(amount: number, currency: Currency, locale = 'en-US'): string {
  const major = minorToMajor(amount, currency);
  const digits = fractionDigitsForCurrency(currency);
  try {
    return major.toLocaleString(locale, {
      style: 'currency',
      currency,
      minimumFractionDigits: digits,
      maximumFractionDigits: digits,
    });
  } catch {
    return `${currency} ${major.toFixed(digits)}`;
  }
}
