import { describe, expect, it } from 'vitest';
import { totalTokens } from './tokensEstimate';

describe('totalTokens', () => {
  it('returns 0 for empty input', () => {
    expect(totalTokens([])).toBe(0);
  });
  it('sums tokens across multiple strings', () => {
    expect(totalTokens(['abcd', 'efgh'])).toBe(2);
  });
});
