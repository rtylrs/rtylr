// Sorting utilities for CRM list views.

import { parseCrmDate } from './dateUtils';

export function compareStrings(a: string | undefined | null, b: string | undefined | null): number {
  return (a ?? '').toLowerCase().localeCompare((b ?? '').toLowerCase());
}

export function compareNumbers(a: number, b: number): number {
  return a - b;
}

function dateTimestamp(value: string | undefined | null): number {
  return value ? parseCrmDate(value).getTime() : NaN;
}

export function compareDates(a: string | undefined | null, b: string | undefined | null): number {
  const ta = dateTimestamp(a);
  const tb = dateTimestamp(b);
  if (!Number.isFinite(ta) && !Number.isFinite(tb)) return 0;
  if (!Number.isFinite(ta)) return 1;
  if (!Number.isFinite(tb)) return -1;
  return ta - tb;
}

/** Sorts valid dates newest first while keeping missing or malformed dates last. */
export function compareDatesDescending(
  a: string | undefined | null,
  b: string | undefined | null
): number {
  const ta = dateTimestamp(a);
  const tb = dateTimestamp(b);
  if (!Number.isFinite(ta) && !Number.isFinite(tb)) return 0;
  if (!Number.isFinite(ta)) return 1;
  if (!Number.isFinite(tb)) return -1;
  return tb - ta;
}

export function sortByKey<T>(
  key: keyof T,
  direction: 'asc' | 'desc' = 'asc'
): (a: T, b: T) => number {
  return (a: T, b: T): number => {
    const av = a[key];
    const bv = b[key];
    let result: number;
    if (typeof av === 'string' && typeof bv === 'string') result = compareStrings(av, bv);
    else if (typeof av === 'number' && typeof bv === 'number') result = compareNumbers(av, bv);
    else result = String(av).localeCompare(String(bv));
    return direction === 'desc' ? -result : result;
  };
}
