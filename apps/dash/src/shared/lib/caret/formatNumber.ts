/** Compact-format a large integer (e.g. 12.4K, 3.2M). */
export function compactNumber(n: number, locale = 'en-US'): string {
  return new Intl.NumberFormat(locale, { notation: 'compact', maximumFractionDigits: 1 }).format(n);
}

/** Strict integer formatter with thousands separators. */
export function commaNumber(n: number, locale = 'en-US'): string {
  return new Intl.NumberFormat(locale).format(n);
}
