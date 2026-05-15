import { describe, expect, it } from 'vitest';
import { formatBytes } from './bytes';

describe('formatBytes', () => {
  it('formats bytes under 1KB', () => {
    expect(formatBytes(512)).toBe('512 B');
  });
  it('formats KB with no decimals once above 10', () => {
    expect(formatBytes(10 * 1024)).toBe('10 KB');
  });
  it('formats small KB with one decimal', () => {
    expect(formatBytes(1.5 * 1024)).toBe('1.5 KB');
  });
  it('formats MB', () => {
    expect(formatBytes(5 * 1024 * 1024)).toBe('5.0 MB');
  });
  it('returns 0 B for invalid input', () => {
    expect(formatBytes(NaN)).toBe('0 B');
  });
});
