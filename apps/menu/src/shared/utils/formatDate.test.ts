import { describe, it, expect } from 'vitest';
import { formatDate } from './formatDate';

describe('formatDate', () => {
  it('formats a valid ISO date string', () => {
    const result = formatDate('2024-01-15');
    expect(result).toMatch(/2024/);
  });

  it('returns "Invalid Date" for an empty string', () => {
    expect(formatDate('')).toBe('Invalid Date');
  });

  it('returns "Invalid Date" for a non-date string', () => {
    expect(formatDate('not-a-date')).toBe('Invalid Date');
  });

  it('formats a date object passed as string', () => {
    const dateStr = new Date(2023, 5, 15).toISOString();
    const result = formatDate(dateStr);
    expect(result).toMatch(/2023/);
  });

  it('handles a Unix timestamp string gracefully', () => {
    const result = formatDate('2000-12-31');
    expect(result).toMatch(/2000/);
  });
});
