import { describe, it, expect } from 'vitest';
import {
  truncate,
  capitalise,
  nameFromEmail,
  maskEmail,
  isValidUrl,
  normaliseWhitespace,
} from '../format';

describe('truncate', () => {
  it('returns the original string when within limit', () => {
    expect(truncate('hello', 10)).toBe('hello');
  });

  it('truncates and appends ellipsis when over limit', () => {
    expect(truncate('hello world', 5)).toBe('hello...');
  });

  it('returns exact string when equal to limit', () => {
    expect(truncate('hello', 5)).toBe('hello');
  });
});

describe('capitalise', () => {
  it('capitalises the first character', () => {
    expect(capitalise('hello')).toBe('Hello');
  });

  it('returns empty string for empty input', () => {
    expect(capitalise('')).toBe('');
  });
});

describe('nameFromEmail', () => {
  it('derives a full name from dot-separated local part', () => {
    expect(nameFromEmail('john.doe@example.com')).toBe('John Doe');
  });

  it('handles underscore-separated local part', () => {
    expect(nameFromEmail('jane_smith@example.com')).toBe('Jane Smith');
  });

  it('handles a plain local part', () => {
    expect(nameFromEmail('alice@example.com')).toBe('Alice');
  });
});

describe('maskEmail', () => {
  it('masks the local part keeping only the first character', () => {
    expect(maskEmail('john@example.com')).toBe('j***@example.com');
  });

  it('returns the original value for strings with no at-sign', () => {
    expect(maskEmail('notanemail')).toBe('notanemail');
  });
});

describe('isValidUrl', () => {
  it('returns true for https URLs', () => {
    expect(isValidUrl('https://example.com')).toBe(true);
  });

  it('returns false for plain strings', () => {
    expect(isValidUrl('not a url')).toBe(false);
  });
});

describe('normaliseWhitespace', () => {
  it('trims and collapses whitespace', () => {
    expect(normaliseWhitespace('  hello   world  ')).toBe('hello world');
  });

  it('returns empty string unchanged', () => {
    expect(normaliseWhitespace('')).toBe('');
  });
});
