import { describe, it, expect } from 'vitest';
import { truncate, truncateHard, isTruncated } from './truncate';

describe('truncate', () => {
  it('returns the original string when it is shorter than maxLength', () => {
    expect(truncate('hello', 10)).toBe('hello');
  });

  it('returns the original string when it equals maxLength', () => {
    expect(truncate('hello', 5)).toBe('hello');
  });

  it('appends ellipsis when the string exceeds maxLength', () => {
    const result = truncate('Hello World', 8);
    expect(result).toContain('...');
    expect(result.length).toBeLessThanOrEqual(8);
  });

  it('does not leave trailing spaces before the ellipsis', () => {
    const result = truncate('Hello World', 9);
    expect(result).not.toMatch(/\s\.\.\./);
  });

  it('uses a custom ellipsis when provided', () => {
    const result = truncate('Hello World', 8, '-');
    expect(result).toContain('-');
  });

  it('handles an empty string', () => {
    expect(truncate('', 5)).toBe('');
  });
});

describe('truncateHard', () => {
  it('returns the original string when shorter than maxLength', () => {
    expect(truncateHard('hi', 10)).toBe('hi');
  });

  it('cuts at exactly maxLength - ellipsis.length characters', () => {
    const result = truncateHard('Hello World', 8);
    expect(result).toBe('Hello...');
    expect(result.length).toBe(8);
  });

  it('handles a custom ellipsis', () => {
    const result = truncateHard('Hello World', 7, '!');
    expect(result).toBe('Hello !');
  });
});

describe('isTruncated', () => {
  it('returns false when string length equals maxLength', () => {
    expect(isTruncated('hello', 5)).toBe(false);
  });

  it('returns false when string length is less than maxLength', () => {
    expect(isTruncated('hi', 10)).toBe(false);
  });

  it('returns true when string length exceeds maxLength', () => {
    expect(isTruncated('hello world', 5)).toBe(true);
  });
});
