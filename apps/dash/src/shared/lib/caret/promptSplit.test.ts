import { describe, expect, it } from 'vitest';
import { splitPrompt } from './promptSplit';

describe('splitPrompt', () => {
  it('returns the input as a single chunk when short', () => {
    expect(splitPrompt('hi')).toEqual(['hi']);
  });
  it('splits at paragraph boundaries when possible', () => {
    const a = 'a'.repeat(300);
    const b = 'b'.repeat(300);
    const out = splitPrompt(`${a}\n\n${b}`, 350);
    expect(out.length).toBe(2);
  });
  it('falls back to word boundaries', () => {
    const long = 'word '.repeat(200).trim();
    const out = splitPrompt(long, 100);
    expect(out.length).toBeGreaterThan(1);
    for (const c of out) expect(c.length).toBeLessThanOrEqual(110);
  });
});
