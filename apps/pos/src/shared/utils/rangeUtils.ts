// Pure helpers for numeric ranges, pagination, and bounded values.

/** Clamp a number between min and max (inclusive). */
export function clamp(value: number, min: number, max: number): number {
  if (!Number.isFinite(value)) return min;
  return Math.min(Math.max(value, min), max);
}

/** Return true if value is strictly within the range [min, max]. */
export function inRange(value: number, min: number, max: number): boolean {
  return Number.isFinite(value) && value >= min && value <= max;
}

/** Generate an array of integers from start (inclusive) to end (exclusive). */
export function range(start: number, end: number): number[] {
  if (!Number.isFinite(start) || !Number.isFinite(end) || end <= start) return [];
  const result: number[] = [];
  for (let i = start; i < end; i++) result.push(i);
  return result;
}

/** Return the total number of pages given a total count and a page size. */
export function pageCount(total: number, pageSize: number): number {
  if (!Number.isFinite(total) || !Number.isFinite(pageSize) || pageSize <= 0) return 0;
  return Math.ceil(total / pageSize);
}

/** Return the zero-based slice bounds for a page. */
export function pageBounds(
  page: number,
  pageSize: number,
): { start: number; end: number } {
  const safeSize = Math.max(1, Math.floor(pageSize));
  const safePage = Math.max(0, Math.floor(page));
  return { start: safePage * safeSize, end: (safePage + 1) * safeSize };
}

/** Linearly interpolate between two values. */
export function lerp(a: number, b: number, t: number): number {
  if (!Number.isFinite(a) || !Number.isFinite(b) || !Number.isFinite(t)) return a;
  return a + (b - a) * clamp(t, 0, 1);
}

/** Normalize a value to a 0-1 range given min and max. */
export function normalize(value: number, min: number, max: number): number {
  if (!Number.isFinite(value) || max === min) return 0;
  return clamp((value - min) / (max - min), 0, 1);
}
