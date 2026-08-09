/**
 * Barcode validation and formatting helpers for the ERP app.
 * All functions are pure with no side-effects.
 */

export type BarcodeFormat = 'EAN-13' | 'EAN-8' | 'UPC-A' | 'UPC-E' | 'CODE-128' | 'QR' | 'unknown';

/** Detect the barcode format from a raw barcode string. */
export function detectBarcodeFormat(barcode: string): BarcodeFormat {
  const digits = barcode.replace(/\s/g, '');
  if (/^\d{13}$/.test(digits)) return 'EAN-13';
  if (/^\d{8}$/.test(digits)) return 'EAN-8';
  if (/^\d{12}$/.test(digits)) return 'UPC-A';
  if (/^\d{6,8}$/.test(digits) && digits.startsWith('0')) return 'UPC-E';
  if (/^[A-Za-z0-9\-. $/+%]{1,107}$/.test(digits)) return 'CODE-128';
  return 'unknown';
}

/** Validate an EAN-13 barcode checksum. */
export function validateEan13(barcode: string): boolean {
  if (!/^\d{13}$/.test(barcode)) return false;
  const digits = barcode.split('').map(Number);
  const sum = digits.slice(0, 12).reduce((acc, d, i) => acc + d * (i % 2 === 0 ? 1 : 3), 0);
  const checkDigit = (10 - (sum % 10)) % 10;
  return checkDigit === digits[12];
}

/** Validate a UPC-A barcode checksum. */
export function validateUpcA(barcode: string): boolean {
  if (!/^\d{12}$/.test(barcode)) return false;
  const digits = barcode.split('').map(Number);
  const sum = digits.slice(0, 11).reduce((acc, d, i) => acc + d * (i % 2 === 0 ? 3 : 1), 0);
  const checkDigit = (10 - (sum % 10)) % 10;
  return checkDigit === digits[11];
}

/** Normalize a barcode by removing whitespace and converting to uppercase. */
export function normalizeBarcode(barcode: string): string {
  return barcode.replace(/\s/g, '').toUpperCase();
}

/** Convert a UPC-A code to EAN-13 by prepending a leading zero. */
export function upcAToEan13(upcA: string): string {
  if (!/^\d{12}$/.test(upcA)) return upcA;
  return `0${upcA}`;
}

/** Generate the check digit for an EAN-13 barcode (first 12 digits). */
export function ean13CheckDigit(first12: string): number {
  if (!/^\d{12}$/.test(first12)) return -1;
  const digits = first12.split('').map(Number);
  const sum = digits.reduce((acc, d, i) => acc + d * (i % 2 === 0 ? 1 : 3), 0);
  return (10 - (sum % 10)) % 10;
}

/** Return a display-friendly formatted barcode string. */
export function formatBarcode(barcode: string, format: BarcodeFormat): string {
  const raw = normalizeBarcode(barcode);
  switch (format) {
    case 'EAN-13':
      return `${raw.slice(0, 1)} ${raw.slice(1, 7)} ${raw.slice(7, 13)}`;
    case 'EAN-8':
      return `${raw.slice(0, 4)} ${raw.slice(4, 8)}`;
    case 'UPC-A':
      return `${raw.slice(0, 1)} ${raw.slice(1, 6)} ${raw.slice(6, 11)} ${raw.slice(11, 12)}`;
    default:
      return raw;
  }
}

/** Check if a string looks like a valid barcode (any known format). */
export function isValidBarcode(barcode: string): boolean {
  const normalized = normalizeBarcode(barcode);
  const format = detectBarcodeFormat(normalized);
  if (format === 'EAN-13') return validateEan13(normalized);
  if (format === 'UPC-A') return validateUpcA(normalized);
  return format !== 'unknown';
}
