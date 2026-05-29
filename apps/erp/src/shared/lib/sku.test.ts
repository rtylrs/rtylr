import { describe, it, expect } from 'vitest';
import {
  normalizeSku,
  isValidSku,
  generateSku,
  skuValidationError,
  SKU_MAX_LENGTH,
  SKU_MIN_LENGTH,
} from './sku';

describe('normalizeSku', () => {
  it('trims surrounding whitespace', () => {
    expect(normalizeSku('  AB-123  ')).toBe('AB-123');
  });

  it('uppercases lowercase letters', () => {
    expect(normalizeSku('ab-cd')).toBe('AB-CD');
  });

  it('returns an empty string for blank input', () => {
    expect(normalizeSku('')).toBe('');
  });
});

describe('isValidSku', () => {
  it('accepts a basic uppercase alphanumeric SKU', () => {
    expect(isValidSku('AB-123')).toBe(true);
  });

  it('accepts a SKU using digits and hyphens only', () => {
    expect(isValidSku('100-200')).toBe(true);
  });

  it('rejects an empty string', () => {
    expect(isValidSku('')).toBe(false);
  });

  it('rejects a single-character SKU (below minimum length)', () => {
    expect(isValidSku('A')).toBe(false);
  });

  it('rejects a SKU that is too long', () => {
    expect(isValidSku('A'.repeat(SKU_MAX_LENGTH + 1))).toBe(false);
  });

  it('accepts a SKU at the maximum allowed length', () => {
    expect(isValidSku('A'.repeat(SKU_MAX_LENGTH))).toBe(true);
  });

  it('accepts a SKU at the minimum allowed length', () => {
    expect(isValidSku('AB')).toBe(true);
  });

  it('rejects lowercase letters', () => {
    expect(isValidSku('ab-cd')).toBe(false);
  });

  it('rejects spaces inside the SKU', () => {
    expect(isValidSku('AB CD')).toBe(false);
  });

  it('rejects special characters like underscores', () => {
    expect(isValidSku('AB_CD')).toBe(false);
  });
});

describe('generateSku', () => {
  it('generates an uppercase hyphenated SKU from a simple name', () => {
    const sku = generateSku('Widget Alpha');
    expect(sku).toBe('WIDG-ALPH');
  });

  it('appends a zero-padded suffix when provided', () => {
    const sku = generateSku('Widget Alpha', 1);
    expect(sku).toBe('WIDG-ALPH-001');
  });

  it('uses at most 3 words from the name', () => {
    const sku = generateSku('One Two Three Four');
    expect(sku).toBe('ONE-TWO-THRE');
  });

  it('strips non-alphanumeric characters from the name', () => {
    const sku = generateSku('Widget (v2)');
    expect(sku).toBe('WIDG-V2');
  });

  it('pads the suffix to 3 digits', () => {
    const sku = generateSku('Item', 5);
    expect(sku).toBe('ITEM-005');
  });
});

describe('skuValidationError', () => {
  it('returns null for a valid SKU', () => {
    expect(skuValidationError('AB-123')).toBeNull();
  });

  it('returns a length error when shorter than minimum', () => {
    const err = skuValidationError('A');
    expect(err).toMatch(/at least/i);
    expect(err).toContain(String(SKU_MIN_LENGTH));
  });

  it('returns a length error when longer than maximum', () => {
    const err = skuValidationError('A'.repeat(SKU_MAX_LENGTH + 1));
    expect(err).toMatch(/at most/i);
    expect(err).toContain(String(SKU_MAX_LENGTH));
  });

  it('returns a character error for invalid characters', () => {
    const err = skuValidationError('ab_cd');
    expect(err).toMatch(/only contain/i);
  });

  it('normalises the input before validating', () => {
    expect(skuValidationError('  ab-cd  ')).toBeNull();
  });
});
