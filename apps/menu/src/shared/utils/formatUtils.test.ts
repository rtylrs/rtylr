import { describe, it, expect } from 'vitest';
import {
  formatCompact,
  formatOrdinal,
  pluralise,
  padStart,
  padEnd,
  formatBytes,
} from './formatUtils';

describe('formatCompact', () => {
  it('returns the number as-is when below 1000', () => {
    expect(formatCompact(999)).toBe('999');
  });

  it('formats thousands with k suffix', () => {
    expect(formatCompact(1200)).toBe('1.2k');
  });

  it('formats millions with M suffix', () => {
    expect(formatCompact(2_500_000)).toBe('2.5M');
  });
});

describe('formatOrdinal', () => {
  it('formats 1 as 1st', () => {
    expect(formatOrdinal(1)).toBe('1st');
  });

  it('formats 2 as 2nd', () => {
    expect(formatOrdinal(2)).toBe('2nd');
  });

  it('formats 3 as 3rd', () => {
    expect(formatOrdinal(3)).toBe('3rd');
  });

  it('formats 4 as 4th', () => {
    expect(formatOrdinal(4)).toBe('4th');
  });

  it('formats 11 as 11th', () => {
    expect(formatOrdinal(11)).toBe('11th');
  });

  it('formats 21 as 21st', () => {
    expect(formatOrdinal(21)).toBe('21st');
  });
});

describe('pluralise', () => {
  it('uses singular for count of 1', () => {
    expect(pluralise(1, 'item')).toBe('1 item');
  });

  it('adds s for count other than 1', () => {
    expect(pluralise(2, 'item')).toBe('2 items');
  });

  it('uses custom plural when provided', () => {
    expect(pluralise(2, 'child', 'children')).toBe('2 children');
  });

  it('handles 0', () => {
    expect(pluralise(0, 'result')).toBe('0 results');
  });
});

describe('padStart', () => {
  it('pads a string with spaces by default', () => {
    expect(padStart('5', 3)).toBe('  5');
  });

  it('pads with a custom fill character', () => {
    expect(padStart('5', 3, '0')).toBe('005');
  });

  it('does not truncate when already long enough', () => {
    expect(padStart('hello', 3)).toBe('hello');
  });
});

describe('padEnd', () => {
  it('pads a string on the right', () => {
    expect(padEnd('hi', 5, '.')).toBe('hi...');
  });
});

describe('formatBytes', () => {
  it('formats bytes under 1024 as B', () => {
    expect(formatBytes(500)).toBe('500 B');
  });

  it('formats kilobytes', () => {
    expect(formatBytes(1024)).toBe('1.0 KB');
  });

  it('formats megabytes', () => {
    expect(formatBytes(1024 * 1024)).toBe('1.0 MB');
  });

  it('formats gigabytes', () => {
    expect(formatBytes(1024 * 1024 * 1024)).toBe('1.0 GB');
  });
});
