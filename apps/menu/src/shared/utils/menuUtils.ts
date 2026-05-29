/**
 * Menu-specific utility functions.
 */

export interface PriceRange {
  min: number;
  max: number;
}

/**
 * Return a human-readable price string.
 * When min equals max, return a single price; otherwise return a range.
 */
export function formatPriceRange(range: PriceRange, currency = 'USD', locale = 'en-US'): string {
  const fmt = new Intl.NumberFormat(locale, {
    style: 'currency',
    currency,
    minimumFractionDigits: 2,
  });
  if (range.min === range.max) return fmt.format(range.min);
  return `${fmt.format(range.min)} - ${fmt.format(range.max)}`;
}

/** Return true if the item is available (not sold out and within the given hour). */
export function isAvailableNow(
  soldOut: boolean,
  availableHours?: { open: number; close: number }
): boolean {
  if (soldOut) return false;
  if (!availableHours) return true;
  const hour = new Date().getHours();
  return hour >= availableHours.open && hour < availableHours.close;
}

/** Return the display label for an item quantity in the cart (e.g. "x2"). */
export function formatCartQuantity(quantity: number): string {
  return `x${quantity}`;
}

/** Return a shortened description capped at maxLength characters. */
export function excerptDescription(description: string, maxLength = 120): string {
  if (description.length <= maxLength) return description;
  return `${description.slice(0, maxLength - 3)}...`;
}

/** Return the total price for a cart line (unit price multiplied by quantity). */
export function lineTotal(unitPrice: number, quantity: number): number {
  return parseFloat((unitPrice * quantity).toFixed(2));
}

/** Return true if a category slug is valid (lowercase letters, digits, hyphens). */
export function isValidCategorySlug(slug: string): boolean {
  return /^[a-z0-9-]+$/.test(slug) && slug.length > 0;
}
