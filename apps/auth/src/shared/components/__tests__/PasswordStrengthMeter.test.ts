import { describe, it, expect } from 'vitest';
import { scorePassword } from '../PasswordStrengthMeter';

describe('scorePassword', () => {
  it('returns 0 for empty string', () => {
    expect(scorePassword('')).toBe(0);
  });

  it('returns 1 for a short lowercase-only password', () => {
    // length >= 8 gives +1, but no uppercase/lowercase combo, no digit, no special char
    expect(scorePassword('abcdefgh')).toBe(1);
  });

  it('returns 2 for 12+ char lowercase-only password', () => {
    // length >= 8 (+1) + length >= 12 (+1)
    expect(scorePassword('abcdefghijkl')).toBe(2);
  });

  it('returns higher score with uppercase + lowercase', () => {
    // length >= 8 (+1), uppercase+lowercase (+1)
    const score = scorePassword('Abcdefgh');
    expect(score).toBeGreaterThanOrEqual(2);
  });

  it('returns higher score with digit', () => {
    // length >= 8 (+1), uppercase+lowercase (+1), digit (+1)
    const score = scorePassword('Abcdefg1');
    expect(score).toBeGreaterThanOrEqual(3);
  });

  it('returns max score 4 for strong password', () => {
    // length >= 8, length >= 12, upper+lower, digit, special
    expect(scorePassword('MyStr0ng!Pass')).toBe(4);
  });

  it('never exceeds 4', () => {
    expect(scorePassword('Sup3r$trong_P@ssw0rd_ExtraLong!!')).toBe(4);
  });

  it('returns 0 for a very short string', () => {
    expect(scorePassword('hi')).toBe(0);
  });
});
