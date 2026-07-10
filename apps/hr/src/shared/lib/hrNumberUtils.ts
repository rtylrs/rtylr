export function clamp(value: number, min: number, max: number): number {
  return Math.min(max, Math.max(min, value));
}

export function roundTo(value: number, decimals: number): number {
  const factor = Math.pow(10, decimals);
  return Math.round(value * factor) / factor;
}

export function percentOf(value: number, total: number): number {
  if (total === 0) return 0;
  return (value / total) * 100;
}

export function sum(values: number[]): number {
  return values.reduce((acc, v) => acc + v, 0);
}

export function mean(values: number[]): number {
  if (values.length === 0) return 0;
  return sum(values) / values.length;
}

export function minOf(values: number[]): number {
  if (values.length === 0) return 0;
  return Math.min(...values);
}

export function maxOf(values: number[]): number {
  if (values.length === 0) return 0;
  return Math.max(...values);
}

export function formatCurrency(amount: number, currency = 'USD', locale = 'en-US'): string {
  return new Intl.NumberFormat(locale, {
    style: 'currency',
    currency,
  }).format(toFiniteNumber(amount));
}

export function formatPercent(ratio: number, decimals = 1): string {
  return `${(toFiniteNumber(ratio) * 100).toFixed(decimals)}%`;
}

export function isInRange(value: number, min: number, max: number): boolean {
  return value >= min && value <= max;
}
import { toFiniteNumber } from '@voxire/api';
