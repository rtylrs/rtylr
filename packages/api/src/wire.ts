/** Convert API scalars to a finite number without allowing NaN or Infinity into app state. */
export function toFiniteNumber(value: unknown, fallback = 0): number {
  if (value === null || value === undefined || value === '') return fallback;
  const parsed = Number(value);
  return Number.isFinite(parsed) ? parsed : fallback;
}

/** Convert an optional API scalar while preserving the absence of a value. */
export function toOptionalFiniteNumber(value: unknown): number | undefined {
  if (value === null || value === undefined || value === '') return undefined;
  const parsed = Number(value);
  return Number.isFinite(parsed) ? parsed : undefined;
}

/** Calculate a finite ratio, returning the fallback for invalid values or a zero denominator. */
export function safeRatio(numerator: unknown, denominator: unknown, fallback = 0): number {
  const top = toFiniteNumber(numerator, fallback);
  const bottom = toFiniteNumber(denominator, 0);
  if (bottom === 0) return fallback;
  const ratio = top / bottom;
  return Number.isFinite(ratio) ? ratio : fallback;
}
