/**
 * SKU validation and formatting utilities.
 */

/** Maximum allowed length for a SKU. */
export const SKU_MAX_LENGTH = 64;

/** Minimum allowed length for a SKU. */
export const SKU_MIN_LENGTH = 2;

/** Pattern: uppercase letters, digits, and hyphens only. */
const SKU_PATTERN = /^[A-Z0-9-]+$/;

/** Normalise a raw SKU string: trim and uppercase. */
export function normalizeSku(raw: string): string {
  return raw.trim().toUpperCase();
}

/** Return true when the (already normalised) SKU passes all validation rules. */
export function isValidSku(sku: string): boolean {
  if (!sku || sku.length < SKU_MIN_LENGTH || sku.length > SKU_MAX_LENGTH) return false;
  return SKU_PATTERN.test(sku);
}

/** Generate a simple SKU from a product name and an optional numeric suffix. */
export function generateSku(name: string, suffix?: number): string {
  const base = name
    .toUpperCase()
    .replace(/[^A-Z0-9\s]/g, '')
    .trim()
    .split(/\s+/)
    .slice(0, 3)
    .map((w) => w.slice(0, 4))
    .join('-');
  return suffix !== undefined ? `${base}-${String(suffix).padStart(3, '0')}` : base;
}

/** Return a human-readable validation error message or null when valid. */
export function skuValidationError(sku: string): string | null {
  const s = normalizeSku(sku);
  if (s.length < SKU_MIN_LENGTH) return `SKU must be at least ${SKU_MIN_LENGTH} characters`;
  if (s.length > SKU_MAX_LENGTH) return `SKU must be at most ${SKU_MAX_LENGTH} characters`;
  if (!SKU_PATTERN.test(s)) return 'SKU may only contain letters, digits, and hyphens';
  return null;
}
