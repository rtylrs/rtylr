import { describe, expect, it } from 'vitest';
import { commaNumber, compactNumber } from './formatNumber';

describe('compactNumber', () => {
  it('uses K for thousands', () => {
    expect(compactNumber(12_500)).toBe('12.5K');
  });
  it('uses M for millions', () => {
    expect(compactNumber(3_400_000)).toBe('3.4M');
  });
});

describe('commaNumber', () => {
  it('uses thousands separators', () => {
    expect(commaNumber(1234567)).toBe('1,234,567');
  });
});
