import { describe, expect, it } from 'vitest';
import { capitalize } from './capitalize';

describe('capitalize', () => {
  it('capitalizes the first letter', () => {
    expect(capitalize('hello')).toBe('Hello');
  });
  it('returns empty for empty/null', () => {
    expect(capitalize('')).toBe('');
    expect(capitalize(null)).toBe('');
  });
});
