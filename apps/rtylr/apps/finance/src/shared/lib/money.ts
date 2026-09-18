// Shared money formatting helpers.

import { toFiniteNumber } from '@voxire/api';

export type Currency = 'USD' | 'EUR' | 'GBP' | 'JPY' | 'CAD' | 'AUD' | 'CHF' | 'CNY';

const CURRENCY_SYMBOL: Record<Currency, string> = {
  USD: '$',
  EUR: '\u20AC',
  GBP: '\u00A3',
  JPY: '\u00A5',
  CAD: 'CA$',
  AUD: 'A$',
  CHF: 'CHF',
  CNY: '\u00A5',
};

export const CURRENCY_OPTIONS: Currency[] = [
  'USD',
  'EUR',
  'GBP',
  'JPY',
  'CAD',
  'AUD',
  'CHF',
  'CNY',
];

export function getCurrencySymbol(c: Currency): string {
  return CURRENCY_SYMBOL[c] ?? c;
}

export function formatMoney(
  value: number | string | null | undefined,
  currency: Currency = 'USD',
  fractionDigits = 2
): string {
  const n = typeof value === 'string' ? parseFloat(value) : (value ?? 0);
  if (!isFinite(n as number)) return '-';
  try {
    return (n as number).toLocaleString('en-US', {
      style: 'currency',
      currency,
      minimumFractionDigits: fractionDigits,
      maximumFractionDigits: fractionDigits,
    });
  } catch {
    return `${getCurrencySymbol(currency)}${(n as number).toFixed(fractionDigits)}`;
  }
}

export function formatCompact(n: number, currency: Currency = 'USD'): string {
  const safeNumber = toFiniteNumber(n);
  const abs = Math.abs(safeNumber);
  if (abs >= 1_000_000_000)
    return `${getCurrencySymbol(currency)}${(safeNumber / 1_000_000_000).toFixed(1)}B`;
  if (abs >= 1_000_000)
    return `${getCurrencySymbol(currency)}${(safeNumber / 1_000_000).toFixed(1)}M`;
  if (abs >= 1_000) return `${getCurrencySymbol(currency)}${(safeNumber / 1_000).toFixed(1)}K`;
  return `${getCurrencySymbol(currency)}${safeNumber.toFixed(0)}`;
}

export const FX_RATES: Record<Currency, number> = {
  USD: 1.0,
  EUR: 0.92,
  GBP: 0.79,
  JPY: 155.0,
  CAD: 1.36,
  AUD: 1.51,
  CHF: 0.91,
  CNY: 7.23,
};

export function convert(amount: number, from: Currency, to: Currency): number {
  if (from === to) return amount;
  const fromRate = FX_RATES[from] ?? 1.0;
  const toRate = FX_RATES[to] ?? 1.0;
  const amountInUsd = amount / fromRate;
  return amountInUsd * toRate;
}

const ZERO_DECIMAL_CURRENCIES: ReadonlySet<Currency> = new Set<Currency>(['JPY']);

export function fractionDigitsForCurrency(currency: Currency): number {
  return ZERO_DECIMAL_CURRENCIES.has(currency) ? 0 : 2;
}

export function minorToMajor(amount: number, currency: Currency = 'USD'): number {
  const digits = fractionDigitsForCurrency(currency);
  const divisor = Math.pow(10, digits);
  return amount / divisor;
}

export function majorToMinor(amount: number, currency: Currency = 'USD'): number {
  const digits = fractionDigitsForCurrency(currency);
  return Math.round(amount * Math.pow(10, digits));
}

export function parseMoney(input: string): number {
  const cleaned = input.replace(/[^0-9.\-]/g, '');
  const n = parseFloat(cleaned);
  return isFinite(n) ? n : 0;
}
