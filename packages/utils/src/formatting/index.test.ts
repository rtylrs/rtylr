import { describe, it, expect } from 'vitest';
import { formatting } from './index';

describe('formatting', () => {
  // ── Date formatting ──────────────────────────────────────────

  describe('formatDate', () => {
    it('formats Date object with default format', () => {
      const date = new Date(2024, 0, 15); // Jan 15, 2024
      expect(formatting.formatDate(date)).toBe('Jan 15, 2024');
    });

    it('formats string date with default format', () => {
      const result = formatting.formatDate('2024-06-15');
      expect(result).toContain('Jun');
      expect(result).toContain('15');
      expect(result).toContain('2024');
    });

    it('formats with custom format string', () => {
      const date = new Date(2024, 5, 15);
      expect(formatting.formatDate(date, 'yyyy-MM-dd')).toBe('2024-06-15');
    });

    it('formats with another custom format', () => {
      const date = new Date(2024, 0, 1);
      expect(formatting.formatDate(date, 'dd/MM/yyyy')).toBe('01/01/2024');
    });
  });

  describe('formatRelativeDate', () => {
    it('formats a recent date', () => {
      const result = formatting.formatRelativeDate(new Date());
      expect(typeof result).toBe('string');
      expect(result.length).toBeGreaterThan(0);
    });

    it('handles string input', () => {
      const result = formatting.formatRelativeDate('2024-01-01');
      expect(typeof result).toBe('string');
    });
  });

  describe('formatDistance', () => {
    it('returns a distance string with suffix', () => {
      const pastDate = new Date(Date.now() - 1000 * 60 * 60 * 24); // 1 day ago
      const result = formatting.formatDistance(pastDate);
      expect(result).toContain('ago');
    });

    it('handles string date input', () => {
      const result = formatting.formatDistance('2020-01-01');
      expect(result).toContain('ago');
    });
  });

  // ── Number formatting ──────────────────────────────────────────

  describe('formatCurrency', () => {
    it('formats USD by default', () => {
      const result = formatting.formatCurrency(1234.56);
      expect(result).toContain('1,234.56');
      expect(result).toContain('$');
    });

    it('formats zero', () => {
      const result = formatting.formatCurrency(0);
      expect(result).toContain('0');
    });

    it('formats negative amounts', () => {
      const result = formatting.formatCurrency(-50);
      expect(result).toContain('50');
    });

    it('formats with different currency', () => {
      const result = formatting.formatCurrency(100, 'EUR', 'de-DE');
      expect(result).toContain('100');
    });
  });

  describe('formatNumber', () => {
    it('formats number with commas', () => {
      expect(formatting.formatNumber(1234567)).toBe('1,234,567');
    });

    it('formats zero', () => {
      expect(formatting.formatNumber(0)).toBe('0');
    });

    it('formats decimal numbers', () => {
      const result = formatting.formatNumber(1234.56);
      expect(result).toContain('1,234');
    });

    it('formats negative numbers', () => {
      const result = formatting.formatNumber(-1000);
      expect(result).toContain('1,000');
    });
  });

  describe('formatPercentage', () => {
    it('formats 0.5 as 50%', () => {
      expect(formatting.formatPercentage(0.5)).toBe('50.00%');
    });

    it('formats 1 as 100%', () => {
      expect(formatting.formatPercentage(1)).toBe('100.00%');
    });

    it('formats 0 as 0%', () => {
      expect(formatting.formatPercentage(0)).toBe('0.00%');
    });

    it('respects custom decimals', () => {
      expect(formatting.formatPercentage(0.3333, 1)).toBe('33.3%');
    });

    it('formats values over 1', () => {
      expect(formatting.formatPercentage(1.5, 0)).toBe('150%');
    });
  });

  // ── Text formatting ──────────────────────────────────────────

  describe('capitalize', () => {
    it('capitalizes first letter', () => {
      expect(formatting.capitalize('hello')).toBe('Hello');
    });

    it('lowercases remaining letters', () => {
      expect(formatting.capitalize('hELLO')).toBe('Hello');
    });

    it('handles single character', () => {
      expect(formatting.capitalize('a')).toBe('A');
    });

    it('handles empty string', () => {
      expect(formatting.capitalize('')).toBe('');
    });
  });

  describe('capitalizeWords', () => {
    it('capitalizes each word', () => {
      expect(formatting.capitalizeWords('hello world')).toBe('Hello World');
    });

    it('handles single word', () => {
      expect(formatting.capitalizeWords('hello')).toBe('Hello');
    });

    it('handles all caps', () => {
      expect(formatting.capitalizeWords('HELLO WORLD')).toBe('Hello World');
    });
  });

  describe('truncate', () => {
    it('returns text unchanged if shorter than maxLength', () => {
      expect(formatting.truncate('short', 10)).toBe('short');
    });

    it('truncates long text with default suffix', () => {
      expect(formatting.truncate('this is a long string', 10)).toBe('this is...');
    });

    it('truncates with custom suffix', () => {
      expect(formatting.truncate('this is a long string', 10, '~')).toBe('this is a~');
    });

    it('handles exact length', () => {
      expect(formatting.truncate('exact', 5)).toBe('exact');
    });
  });

  describe('slugify', () => {
    it('converts to lowercase slug', () => {
      expect(formatting.slugify('Hello World')).toBe('hello-world');
    });

    it('replaces special characters', () => {
      expect(formatting.slugify('Hello! World?')).toBe('hello-world');
    });

    it('handles multiple spaces', () => {
      expect(formatting.slugify('hello   world')).toBe('hello-world');
    });

    it('removes leading/trailing dashes', () => {
      expect(formatting.slugify('-hello-world-')).toBe('hello-world');
    });

    it('handles underscores', () => {
      expect(formatting.slugify('hello_world_test')).toBe('hello-world-test');
    });

    it('handles empty string', () => {
      expect(formatting.slugify('')).toBe('');
    });
  });
});
