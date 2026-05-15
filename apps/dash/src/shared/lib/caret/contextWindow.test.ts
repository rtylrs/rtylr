import { describe, expect, it } from 'vitest';
import { approxTokens, contextWindowPercent } from './contextWindow';

describe('approxTokens', () => {
  it('returns 0 for empty input', () => {
    expect(approxTokens('')).toBe(0);
  });
  it('rounds up to the nearest token', () => {
    expect(approxTokens('hi')).toBe(1);
    expect(approxTokens('hellos'.repeat(2))).toBe(3);
  });
});

describe('contextWindowPercent', () => {
  it('returns 0 when the model has an unknown context size', () => {
    expect(contextWindowPercent('auto', 'a'.repeat(1000))).toBe(0);
  });
  it('caps at 100', () => {
    expect(contextWindowPercent('gpt-4o-mini', 'a'.repeat(10_000_000))).toBeLessThanOrEqual(100);
  });
});
