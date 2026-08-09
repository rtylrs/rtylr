/**
 * Extended formatting helpers for the ERP application.
 * Supplements the core format.ts with ERP-specific patterns.
 */

/** Format a percentage value with optional decimal places. */
export function formatPercent(value: number, decimals = 1): string {
  return `${value.toFixed(decimals)}%`;
}

/** Format a quantity with an optional unit label. */
export function formatQuantity(quantity: number, unit?: string): string {
  const formatted = new Intl.NumberFormat('en-US').format(quantity);
  return unit ? `${formatted} ${unit}` : formatted;
}

/** Format a compact number (e.g. 1,200 becomes "1.2k"). */
export function formatCompact(value: number): string {
  if (Math.abs(value) >= 1_000_000) return `${(value / 1_000_000).toFixed(1)}M`;
  if (Math.abs(value) >= 1_000) return `${(value / 1_000).toFixed(1)}k`;
  return String(value);
}

/** Format a duration in days as a human-readable string. */
export function formatDays(days: number): string {
  if (days === 0) return 'Today';
  if (days === 1) return 'Tomorrow';
  if (days === -1) return 'Yesterday';
  if (days < 0) return `${Math.abs(days)}d ago`;
  return `In ${days}d`;
}

/** Truncate a string to a maximum length with an ellipsis. */
export function truncate(text: string, maxLength: number): string {
  if (text.length <= maxLength) return text;
  return `${text.substring(0, maxLength - 3)}...`;
}

/** Format a SKU for display, ensuring uppercase. */
export function formatSku(sku: string): string {
  return sku.toUpperCase().trim();
}

/** Format a price range string (e.g. "$5.00 - $10.00"). */
export function formatPriceRange(min: number, max: number, currency = 'USD'): string {
  const fmt = new Intl.NumberFormat('en-US', {
    style: 'currency',
    currency,
    minimumFractionDigits: 2,
  });
  if (min === max) return fmt.format(min);
  return `${fmt.format(min)} - ${fmt.format(max)}`;
}
