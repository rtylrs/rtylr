import { describe, it, expect } from 'vitest';
import {
  isEan13,
  isUpcA,
  isValidSku,
  normalizeBarcode,
  isSupportedBarcode,
  skuNumericPart,
  maskedBarcode,
  normalizeSku,
} from './barcodeUtils';

describe('isEan13', () => {
  it('returns true for a 13-digit numeric string', () => {
    expect(isEan13('5901234123457')).toBe(true);
  });

  it('returns false for 12 digits', () => {
    expect(isEan13('590123412345')).toBe(false);
  });

  it('returns false for non-numeric characters', () => {
    expect(isEan13('590123412345X')).toBe(false);
  });

  it('strips leading and trailing whitespace before checking', () => {
    expect(isEan13(' 5901234123457 ')).toBe(true);
  });
});

describe('isUpcA', () => {
  it('returns true for a 12-digit numeric string', () => {
    expect(isUpcA('012345678905')).toBe(true);
  });

  it('returns false for 13 digits', () => {
    expect(isUpcA('0123456789050')).toBe(false);
  });

  it('returns false for non-numeric characters', () => {
    expect(isUpcA('01234567890X')).toBe(false);
  });
});

describe('isValidSku', () => {
  it('returns true for a valid SKU', () => {
    expect(isValidSku('ABC-123')).toBe(true);
  });

  it('returns false for a SKU that is too short', () => {
    expect(isValidSku('AB')).toBe(false);
  });

  it('returns false for a SKU that is too long', () => {
    expect(isValidSku('A'.repeat(21))).toBe(false);
  });

  it('returns false for a SKU with spaces', () => {
    expect(isValidSku('ABC 123')).toBe(false);
  });
});

describe('normalizeBarcode', () => {
  it('strips leading and trailing whitespace', () => {
    expect(normalizeBarcode('  012345  ')).toBe('012345');
  });

  it('strips internal whitespace', () => {
    expect(normalizeBarcode('012 345')).toBe('012345');
  });
});

describe('isSupportedBarcode', () => {
  it('returns true for a valid EAN-13', () => {
    expect(isSupportedBarcode('5901234123457')).toBe(true);
  });

  it('returns true for a valid UPC-A', () => {
    expect(isSupportedBarcode('012345678905')).toBe(true);
  });

  it('returns false for an unsupported format', () => {
    expect(isSupportedBarcode('123')).toBe(false);
  });
});

describe('skuNumericPart', () => {
  it('extracts the numeric portion from a SKU', () => {
    expect(skuNumericPart('ITEM-042')).toBe(42);
  });

  it('returns null when there are no digits', () => {
    expect(skuNumericPart('ABC')).toBeNull();
  });
});

describe('maskedBarcode', () => {
  it('masks all but the last 4 characters', () => {
    expect(maskedBarcode('5901234123457')).toBe('*********3457');
  });

  it('returns the full barcode when it is 4 or fewer characters', () => {
    expect(maskedBarcode('3457')).toBe('3457');
  });

  it('respects a custom visibleCount', () => {
    expect(maskedBarcode('5901234123457', 6)).toBe('*******123457');
  });
});

describe('normalizeSku', () => {
  it('converts to uppercase and replaces underscores with hyphens', () => {
    expect(normalizeSku('item_042')).toBe('ITEM-042');
  });

  it('trims whitespace', () => {
    expect(normalizeSku('  abc_123  ')).toBe('ABC-123');
  });
});
