// Pure helpers for barcode and SKU validation and formatting.

/** Return true if the string is a valid EAN-13 barcode (13 numeric digits). */
export function isEan13(input: string): boolean {
  return /^\d{13}$/.test(input.trim());
}

/** Return true if the string is a valid UPC-A barcode (12 numeric digits). */
export function isUpcA(input: string): boolean {
  return /^\d{12}$/.test(input.trim());
}

/** Return true if the string looks like a valid SKU (alphanumeric, 3-20 chars). */
export function isValidSku(input: string): boolean {
  return /^[A-Za-z0-9\-_]{3,20}$/.test(input.trim());
}

/** Normalize a barcode by stripping whitespace and leading zeros for UPC-E. */
export function normalizeBarcode(input: string): string {
  return input.trim().replace(/\s/g, '');
}

/** Return true if the input could be any supported barcode format (EAN-13 or UPC-A). */
export function isSupportedBarcode(input: string): boolean {
  const normalized = normalizeBarcode(input);
  return isEan13(normalized) || isUpcA(normalized);
}

/** Extract the numeric portion of a SKU (for display or sorting). */
export function skuNumericPart(sku: string): number | null {
  const digits = sku.replace(/\D/g, '');
  if (!digits) return null;
  const num = parseInt(digits, 10);
  return Number.isFinite(num) ? num : null;
}

/** Mask a barcode for display, showing only the last 4 characters. */
export function maskedBarcode(barcode: string, visibleCount = 4): string {
  const normalized = normalizeBarcode(barcode);
  if (normalized.length <= visibleCount) return normalized;
  return '*'.repeat(normalized.length - visibleCount) + normalized.slice(-visibleCount);
}

/** Convert a SKU to uppercase with hyphens replacing underscores. */
export function normalizeSku(sku: string): string {
  return sku.trim().toUpperCase().replace(/_/g, '-');
}
