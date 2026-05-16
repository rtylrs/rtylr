/**
 * Format a number as a localized currency string. Defaults to USD because
 * most plans use USD for billing; the per-business currency comes from
 * the config store at the call site.
 */
export function formatCurrency(value: number, currency = 'USD', locale = 'en-US'): string {
  if (!Number.isFinite(value)) return '—';
  return new Intl.NumberFormat(locale, {
    style: 'currency',
    currency,
    maximumFractionDigits: 2,
  }).format(value);
}
