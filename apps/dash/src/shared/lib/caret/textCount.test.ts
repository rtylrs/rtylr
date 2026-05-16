import { describe, expect, it } from 'vitest';
import { countText } from './textCount';

describe('countText', () => {
  it('counts chars and words', () => {
    expect(countText('hello world')).toEqual({ chars: 11, words: 2 });
  });
  it('handles empty input', () => {
    expect(countText('')).toEqual({ chars: 0, words: 0 });
  });
});
