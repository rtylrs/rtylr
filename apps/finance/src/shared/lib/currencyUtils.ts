/**
 * Currency formatting and conversion utilities for the finance app.
 */

import { toFiniteNumber } from '@voxire/api';

export type CurrencyCode = 'USD' | 'EUR' | 'GBP' | 'AED' | 'SAR' | 'CAD' | 'AUD';

const CURRENCY_SYMBOLS: Record<CurrencyCode, string> = {
  USD: '$',
  EUR: '€',
  GBP: '£',
  AED: 'د.إ',
  SAR: 'ر.س',
  CAD: 'CA$',
  AUD: 'A$',
};

/**
 * Returns the symbol for a currency code, or the code itself if unknown.
 */
export function getCurrencySymbol(code: string): string {
  return CURRENCY_SYMBOLS[code as CurrencyCode] ?? code;
}

/**
 * Formats a number as a currency string with the given currency code.
 * Uses Intl.NumberFormat for locale-aware formatting.
 */
export function formatCurrency(amount: number, currency = 'USD', locale = 'en-US'): string {
  const safeAmount = toFiniteNumber(amount);
  return new Intl.NumberFormat(locale, {
    style: 'currency',
    currency,
    minimumFractionDigits: 2,
    maximumFractionDigits: 2,
  }).format(safeAmount);
}

/**
 * Formats a number as a compact currency string (e.g. $1.2K, $3.5M).
 */
export function formatCurrencyCompact(amount: number, currency = 'USD', locale = 'en-US'): string {
  const safeAmount = toFiniteNumber(amount);
  if (Math.abs(safeAmount) >= 1_000_000) {
    return `${getCurrencySymbol(currency)}${(safeAmount / 1_000_000).toFixed(1)}M`;
  }
  if (Math.abs(safeAmount) >= 1_000) {
    return `${getCurrencySymbol(currency)}${(safeAmount / 1_000).toFixed(1)}K`;
  }
  return formatCurrency(safeAmount, currency, locale);
}

/**
 * Converts an amount from one currency to another using a flat exchange rate.
 */
export function convertCurrency(amount: number, rate: number): number {
  if (rate <= 0) return 0;
  return Math.round(amount * rate * 100) / 100;
}

/**
 * Parses a currency string to a number, stripping symbols and commas.
 * Returns NaN if the string cannot be parsed.
 */
export function parseCurrencyString(value: string): number {
  const cleaned = value.replace(/[^0-9.\-]/g, '');
  return parseFloat(cleaned);
}
