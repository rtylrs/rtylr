// Formatting utilities for the CRM app.

import { toFiniteNumber } from '@voxire/api';
import { parseCrmDate } from './dateUtils';

export function formatCurrency(amount: number, currency = 'USD', locale = 'en-US'): string {
  return new Intl.NumberFormat(locale, {
    style: 'currency',
    currency,
    minimumFractionDigits: 2,
    maximumFractionDigits: 2,
  }).format(toFiniteNumber(amount));
}

export function formatPercent(value: number, decimals = 1): string {
  return toFiniteNumber(value).toFixed(decimals) + '%';
}

export function formatNumber(value: number, locale = 'en-US'): string {
  return new Intl.NumberFormat(locale).format(toFiniteNumber(value));
}

export function formatDate(value: string | Date | null | undefined, locale = 'en-US'): string {
  if (!value) return '';
  const date = typeof value === 'string' ? parseCrmDate(value) : value;
  if (isNaN(date.getTime())) return '';
  return date.toLocaleDateString(locale, { year: 'numeric', month: 'short', day: 'numeric' });
}

export function getInitials(name: string): string {
  const parts = name.trim().split(/ +/).filter(Boolean);
  if (parts.length === 0) return '';
  if (parts.length === 1) return parts[0].charAt(0).toUpperCase();
  return (parts[0].charAt(0) + parts[parts.length - 1].charAt(0)).toUpperCase();
}

export function truncate(str: string, maxLength: number, suffix = '...'): string {
  if (str.length <= maxLength) return str;
  return str.slice(0, maxLength - suffix.length) + suffix;
}
