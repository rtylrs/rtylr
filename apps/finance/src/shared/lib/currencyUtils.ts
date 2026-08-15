import { formatMoney, getCurrencySymbol as getMoneyCurrencySymbol } from './money';
import type { Currency } from './money';

export type CurrencyCode = Currency | 'AED' | 'SAR';

export function getCurrencySymbol(code: string): string {
  return getMoneyCurrencySymbol(code as Currency);
}

export function formatCurrency(amount: number, currency = 'USD', _locale = 'en-US'): string {
  return formatMoney(amount, currency as Currency);
}

export function formatCurrencyCompact(amount: number, currency = 'USD', locale = 'en-US'): string {
  const abs = Math.abs(amount);
  if (abs >= 1_000_000) {
    return `${getCurrencySymbol(currency)}${(amount / 1_000_000).toFixed(1)}M`;
  }
  if (abs >= 1_000) {
    return `${getCurrencySymbol(currency)}${(amount / 1_000).toFixed(1)}K`;
  }
  return formatCurrency(amount, currency, locale);
}

export function convertCurrency(amount: number, rate: number): number {
  if (rate <= 0) return 0;
  return Math.round(amount * rate * 100) / 100;
}

export function parseCurrencyString(value: string): number {
  if (!value || !value.trim()) return NaN;
  const cleaned = value.replace(/[^0-9.\-]/g, '');
  if (!cleaned) return NaN;
  return parseFloat(cleaned);
}
