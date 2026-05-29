import { describe, it, expect } from 'vitest';
import { getInitials, toSlug, isAscii, padRight, countOccurrences, camelToLabel } from '../strings';

describe('getInitials', () => {
  it('returns two uppercase initials for a full name', () => {
    expect(getInitials('John Doe')).toBe('JD');
  });

  it('returns a single initial for a one-word name', () => {
    expect(getInitials('Alice')).toBe('A');
  });

  it('returns initials from first and last word for multi-word names', () => {
    expect(getInitials('Mary Jane Watson')).toBe('MW');
  });

  it('returns empty string for empty input', () => {
    expect(getInitials('')).toBe('');
  });
});

describe('toSlug', () => {
  it('lowercases and replaces spaces with hyphens', () => {
    expect(toSlug('Hello World')).toBe('hello-world');
  });

  it('strips non-alphanumeric characters', () => {
    expect(toSlug('Hello, World!')).toBe('hello-world');
  });

  it('returns empty string for empty input', () => {
    expect(toSlug('')).toBe('');
  });
});

describe('isAscii', () => {
  it('returns true for ASCII-only strings', () => {
    expect(isAscii('hello world')).toBe(true);
  });

  it('returns false for non-ASCII characters', () => {
    expect(isAscii('café')).toBe(false);
  });
});

describe('padRight', () => {
  it('pads a string to the specified length', () => {
    expect(padRight('hi', 5)).toBe('hi   ');
  });

  it('uses a custom fill character', () => {
    expect(padRight('hi', 5, '-')).toBe('hi---');
  });
});

describe('countOccurrences', () => {
  it('counts non-overlapping occurrences', () => {
    expect(countOccurrences('abcabc', 'abc')).toBe(2);
  });

  it('returns 0 for empty needle', () => {
    expect(countOccurrences('hello', '')).toBe(0);
  });
});

describe('camelToLabel', () => {
  it('converts camelCase to a readable label', () => {
    expect(camelToLabel('firstName')).toBe('First Name');
  });

  it('leaves a plain word unchanged', () => {
    expect(camelToLabel('name')).toBe('Name');
  });
});
