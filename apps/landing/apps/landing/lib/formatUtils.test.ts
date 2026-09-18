import { describe, expect, it } from 'vitest';

import {
  abbreviateNumber,
  formatBytes,
  formatCurrency,
  formatDuration,
  formatNumber,
  formatOrdinal,
  formatPercent,
  ordinalSuffix,
} from './formatUtils';

describe('formatCurrency', () => {
  it('formats a whole dollar amount', () => {
    expect(formatCurrency(5)).toBe('$5.00');
  });

  it('formats a fractional amount', () => {
    expect(formatCurrency(1.5)).toBe('$1.50');
  });

  it('supports non-USD currencies', () => {
    const result = formatCurrency(100, 'EUR', 'en-US');
    expect(result).toContain('100');
    expect(result).toContain('100');
  });
});

describe('formatNumber', () => {
  it('formats an integer without decimals', () => {
    expect(formatNumber(1234)).toBe('1,234');
  });

  it('formats with specified decimal places', () => {
    expect(formatNumber(1234.567, 2)).toBe('1,234.57');
  });

  it('returns zero correctly', () => {
    expect(formatNumber(0)).toBe('0');
  });
});

describe('abbreviateNumber', () => {
  it('returns small numbers unchanged', () => {
    expect(abbreviateNumber(999)).toBe('999');
  });

  it('abbreviates thousands', () => {
    expect(abbreviateNumber(1000)).toBe('1K');
    expect(abbreviateNumber(1500)).toBe('1.5K');
  });

  it('abbreviates millions', () => {
    expect(abbreviateNumber(2_000_000)).toBe('2M');
    expect(abbreviateNumber(2_500_000)).toBe('2.5M');
  });

  it('abbreviates billions', () => {
    expect(abbreviateNumber(1_000_000_000)).toBe('1B');
  });

  it('handles negative values', () => {
    expect(abbreviateNumber(-1500)).toBe('-1.5K');
  });
});

describe('formatDuration', () => {
  it('formats seconds only', () => {
    expect(formatDuration(42)).toBe('42s');
  });

  it('formats minutes and seconds', () => {
    expect(formatDuration(90)).toBe('1m 30s');
  });

  it('formats exact minutes without seconds', () => {
    expect(formatDuration(120)).toBe('2m');
  });

  it('formats hours and minutes', () => {
    expect(formatDuration(3900)).toBe('1h 5m');
  });

  it('formats exact hours without minutes', () => {
    expect(formatDuration(7200)).toBe('2h');
  });

  it('returns 0s for negative input', () => {
    expect(formatDuration(-5)).toBe('0s');
  });
});

describe('formatBytes', () => {
  it('formats bytes', () => {
    expect(formatBytes(512)).toBe('512 B');
  });

  it('formats kilobytes', () => {
    expect(formatBytes(1536)).toBe('1.5 KB');
  });

  it('formats megabytes', () => {
    expect(formatBytes(1.5 * 1024 * 1024)).toBe('1.5 MB');
  });

  it('formats gigabytes', () => {
    expect(formatBytes(2 * 1024 * 1024 * 1024)).toBe('2.0 GB');
  });

  it('returns 0 B for negative input', () => {
    expect(formatBytes(-1)).toBe('0 B');
  });
});

describe('formatPercent', () => {
  it('converts a fraction to percent with one decimal', () => {
    expect(formatPercent(0.1234)).toBe('12.3%');
  });

  it('handles exact values', () => {
    expect(formatPercent(1)).toBe('100.0%');
  });

  it('supports custom decimal places', () => {
    expect(formatPercent(0.5, 0)).toBe('50%');
  });
});

describe('ordinalSuffix', () => {
  it('returns st for 1', () => {
    expect(ordinalSuffix(1)).toBe('st');
  });

  it('returns nd for 2', () => {
    expect(ordinalSuffix(2)).toBe('nd');
  });

  it('returns rd for 3', () => {
    expect(ordinalSuffix(3)).toBe('rd');
  });

  it('returns th for 4-20', () => {
    expect(ordinalSuffix(11)).toBe('th');
    expect(ordinalSuffix(12)).toBe('th');
    expect(ordinalSuffix(13)).toBe('th');
  });

  it('returns st for 21', () => {
    expect(ordinalSuffix(21)).toBe('st');
  });
});

describe('formatOrdinal', () => {
  it('formats 1st', () => {
    expect(formatOrdinal(1)).toBe('1st');
  });

  it('formats 2nd', () => {
    expect(formatOrdinal(2)).toBe('2nd');
  });

  it('formats 11th', () => {
    expect(formatOrdinal(11)).toBe('11th');
  });

  it('formats 22nd', () => {
    expect(formatOrdinal(22)).toBe('22nd');
  });
});
