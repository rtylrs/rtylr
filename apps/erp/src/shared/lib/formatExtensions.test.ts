import { describe, it, expect } from 'vitest';
import {
  formatPercent,
  formatQuantity,
  formatCompact,
  formatDays,
  truncate,
  formatSku,
  formatPriceRange,
} from './formatExtensions';

describe('formatPercent', () => {
  it('formats with 1 decimal by default', () => {
    expect(formatPercent(50)).toBe('50.0%');
  });

  it('formats with custom decimal places', () => {
    expect(formatPercent(33.333, 2)).toBe('33.33%');
  });

  it('formats 0%', () => {
    expect(formatPercent(0)).toBe('0.0%');
  });
});

describe('formatQuantity', () => {
  it('formats without unit', () => {
    expect(formatQuantity(1000)).toBe('1,000');
  });

  it('formats with unit', () => {
    expect(formatQuantity(50, 'pcs')).toBe('50 pcs');
  });

  it('handles zero quantity', () => {
    expect(formatQuantity(0, 'kg')).toBe('0 kg');
  });
});

describe('formatCompact', () => {
  it('formats thousands with k', () => {
    expect(formatCompact(1200)).toBe('1.2k');
  });

  it('formats millions with M', () => {
    expect(formatCompact(1_500_000)).toBe('1.5M');
  });

  it('returns small numbers as-is', () => {
    expect(formatCompact(999)).toBe('999');
  });

  it('handles negative values', () => {
    expect(formatCompact(-1200)).toBe('-1.2k');
  });
});

describe('formatDays', () => {
  it('returns "Today" for 0', () => {
    expect(formatDays(0)).toBe('Today');
  });

  it('returns "Tomorrow" for 1', () => {
    expect(formatDays(1)).toBe('Tomorrow');
  });

  it('returns "Yesterday" for -1', () => {
    expect(formatDays(-1)).toBe('Yesterday');
  });

  it('returns relative past for negative', () => {
    expect(formatDays(-5)).toBe('5d ago');
  });

  it('returns relative future for positive', () => {
    expect(formatDays(7)).toBe('In 7d');
  });
});

describe('truncate', () => {
  it('returns text unchanged when within limit', () => {
    expect(truncate('hello', 10)).toBe('hello');
  });

  it('truncates with ellipsis when over limit', () => {
    const result = truncate('Hello World', 8);
    expect(result).toHaveLength(8);
    expect(result.endsWith('...')).toBe(true);
  });

  it('handles exact length', () => {
    expect(truncate('hello', 5)).toBe('hello');
  });
});

describe('formatSku', () => {
  it('converts to uppercase', () => {
    expect(formatSku('wgt-001')).toBe('WGT-001');
  });

  it('trims whitespace', () => {
    expect(formatSku('  WGT-001  ')).toBe('WGT-001');
  });
});

describe('formatPriceRange', () => {
  it('formats a range', () => {
    const result = formatPriceRange(5, 10);
    expect(result).toContain('$5.00');
    expect(result).toContain('$10.00');
  });

  it('returns single price when min equals max', () => {
    expect(formatPriceRange(10, 10)).toBe('$10.00');
  });
});
