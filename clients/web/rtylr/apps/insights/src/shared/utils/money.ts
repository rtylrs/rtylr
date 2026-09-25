// Central money helper. All money fields on the /insights/* wire are integer
// minor units (cents). Hooks pass the raw minor value through unchanged, and
// display components call `formatMoney` which handles the conversion via
// `minorToMajor`. Mirrors apps/finance/src/shared/utils/money.ts.
//
// JPY has zero fractional digits, so its "minor" unit is the whole yen. All
// other currencies in this app have two fractional digits.

import { toFiniteNumber } from '@voxire/api';

export type Currency = 'USD' | 'EUR' | 'GBP' | 'JPY' | 'CAD' | 'AUD' | 'CHF' | 'CNY';

const ZERO_DECIMAL_CURRENCIES: ReadonlySet<Currency> = new Set<Currency>(['JPY']);

function fractionDigitsForCurrency(currency: Currency): number {
  return ZERO_DECIMAL_CURRENCIES.has(currency) ? 0 : 2;
}

export function minorToMajor(amount: number, currency: Currency = 'USD'): number {
  const digits = fractionDigitsForCurrency(currency);
  const divisor = Math.pow(10, digits);
  return toFiniteNumber(amount) / divisor;
}

export function formatMoney(amount: number, currency: Currency = 'USD', locale = 'en-US'): string {
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
