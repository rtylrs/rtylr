import { describe, expect, it } from 'vitest';
import { escapeRegex } from './regexUtils';

describe('escapeRegex', () => {
  it('escapes regex meta characters', () => {
    expect(escapeRegex('a.b+c*')).toBe('a\\.b\\+c\\*');
  });
  it('leaves plain text alone', () => {
    expect(escapeRegex('hello')).toBe('hello');
  });
});
