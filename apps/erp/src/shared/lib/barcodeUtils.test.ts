import { describe, it, expect } from 'vitest';
import {
  detectBarcodeFormat,
  validateEan13,
  validateUpcA,
  normalizeBarcode,
  upcAToEan13,
  ean13CheckDigit,
  formatBarcode,
  isValidBarcode,
} from './barcodeUtils';

describe('detectBarcodeFormat', () => {
  it('detects EAN-13', () => {
    expect(detectBarcodeFormat('5901234123457')).toBe('EAN-13');
  });

  it('detects EAN-8', () => {
    expect(detectBarcodeFormat('96385074')).toBe('EAN-8');
  });

  it('detects UPC-A', () => {
    expect(detectBarcodeFormat('036000291452')).toBe('UPC-A');
  });

  it('detects CODE-128 for alphanumeric', () => {
    expect(detectBarcodeFormat('ABC-123')).toBe('CODE-128');
  });

  it('returns unknown for unsupported format', () => {
    expect(detectBarcodeFormat('!@#$%')).toBe('unknown');
  });
});

describe('validateEan13', () => {
  it('validates a correct EAN-13', () => {
    expect(validateEan13('5901234123457')).toBe(true);
  });

  it('rejects wrong length', () => {
    expect(validateEan13('590123412345')).toBe(false);
  });

  it('rejects bad checksum', () => {
    expect(validateEan13('5901234123458')).toBe(false);
  });
});

describe('validateUpcA', () => {
  it('validates a correct UPC-A', () => {
    expect(validateUpcA('036000291452')).toBe(true);
  });

  it('rejects wrong length', () => {
    expect(validateUpcA('03600029145')).toBe(false);
  });
});

describe('normalizeBarcode', () => {
  it('removes whitespace', () => {
    expect(normalizeBarcode('123 456 789')).toBe('123456789');
  });

  it('converts to uppercase', () => {
    expect(normalizeBarcode('abc-123')).toBe('ABC-123');
  });
});

describe('upcAToEan13', () => {
  it('prepends 0 to a 12-digit UPC-A', () => {
    expect(upcAToEan13('036000291452')).toBe('0036000291452');
  });

  it('returns original string for non-UPC-A', () => {
    expect(upcAToEan13('12345')).toBe('12345');
  });
});

describe('ean13CheckDigit', () => {
  it('computes correct check digit', () => {
    expect(ean13CheckDigit('590123412345')).toBe(7);
  });

  it('returns -1 for invalid input', () => {
    expect(ean13CheckDigit('12345')).toBe(-1);
  });
});

describe('formatBarcode', () => {
  it('formats EAN-13 with spaces', () => {
    const formatted = formatBarcode('5901234123457', 'EAN-13');
    expect(formatted).toBe('5 901234 123457');
  });

  it('formats EAN-8 with spaces', () => {
    expect(formatBarcode('96385074', 'EAN-8')).toBe('9638 5074');
  });

  it('returns raw for CODE-128', () => {
    expect(formatBarcode('abc-123', 'CODE-128')).toBe('ABC-123');
  });
});

describe('isValidBarcode', () => {
  it('returns true for valid EAN-13', () => {
    expect(isValidBarcode('5901234123457')).toBe(true);
  });

  it('returns false for invalid EAN-13', () => {
    expect(isValidBarcode('5901234123458')).toBe(false);
  });

  it('returns true for CODE-128 format', () => {
    expect(isValidBarcode('ABC-123')).toBe(true);
  });
});
