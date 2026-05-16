import { describe, expect, it } from 'vitest';
import { validatePrompt } from './promptValidate';
import { CARET_LIMITS } from './limits';

describe('validatePrompt', () => {
  it('returns null for normal input', () => {
    expect(validatePrompt('hi')).toBeNull();
  });
  it('blocks empty input', () => {
    expect(validatePrompt('   ')).toMatch(/something/i);
  });
  it('blocks oversized input', () => {
    const long = 'a'.repeat(CARET_LIMITS.maxMessageChars + 10);
    expect(validatePrompt(long)).toMatch(/too long/i);
  });
});
