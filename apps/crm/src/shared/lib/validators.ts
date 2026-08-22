// Validation helpers for CRM forms.

import { parseCrmDate } from './dateUtils';

export function isValidEmail(value: string): boolean {
  return /^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(value.trim());
}

export function isValidPhone(value: string): boolean {
  return /^\+?[\d\s\-().]{7,20}$/.test(value.trim());
}

export function isNonEmpty(value: string): boolean {
  return value.trim().length > 0;
}

export function isInRange(value: number, min: number, max: number): boolean {
  return Number.isFinite(value) && value >= min && value <= max;
}

export function isIsoDate(value: string): boolean {
  if (!/^\d{4}-\d{2}-\d{2}$/.test(value)) return false;
  return Number.isFinite(parseCrmDate(value).getTime());
}

export function isValidPoints(points: number): boolean {
  return Number.isInteger(points) && points >= 0;
}

export function isOptionalPositiveInteger(value: string): boolean {
  const trimmed = value.trim();
  if (!trimmed) return true;
  const parsed = Number(trimmed);
  return Number.isInteger(parsed) && parsed > 0;
}

export function isValidUrl(value: string): boolean {
  try {
    const url = new URL(value.trim());
    return url.protocol === 'http:' || url.protocol === 'https:';
  } catch {
    return false;
  }
}
