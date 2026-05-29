import { describe, it, expect } from 'vitest';
import { highlightTerm, excerptAround, containsAll, stripTags } from '../highlight';

describe('highlightTerm', () => {
  it('wraps the term with mark tags', () => {
    expect(highlightTerm('hello world', 'world')).toBe('hello <mark>world</mark>');
  });

  it('is case-insensitive', () => {
    expect(highlightTerm('Hello World', 'world')).toBe('Hello <mark>World</mark>');
  });

  it('returns original text when term is empty', () => {
    expect(highlightTerm('hello', '')).toBe('hello');
  });

  it('supports custom tag', () => {
    expect(highlightTerm('foo bar', 'bar', 'strong')).toBe('foo <strong>bar</strong>');
  });
});

describe('excerptAround', () => {
  it('returns a window around the term', () => {
    const text = 'The quick brown fox jumps over the lazy dog';
    const result = excerptAround(text, 'fox', 5);
    expect(result).toContain('fox');
  });

  it('falls back to start when term not found', () => {
    const result = excerptAround('hello world', 'missing', 5);
    expect(result).toBe('hello world');
  });

  it('returns empty string for empty text', () => {
    expect(excerptAround('', 'term', 10)).toBe('');
  });
});

describe('containsAll', () => {
  it('returns true when all terms are present', () => {
    expect(containsAll('hello world foo', ['hello', 'foo'])).toBe(true);
  });

  it('returns false when any term is missing', () => {
    expect(containsAll('hello world', ['hello', 'missing'])).toBe(false);
  });

  it('is case-insensitive', () => {
    expect(containsAll('Hello World', ['hello', 'WORLD'])).toBe(true);
  });
});

describe('stripTags', () => {
  it('removes HTML tags', () => {
    expect(stripTags('<p>hello <b>world</b></p>')).toBe('hello world');
  });

  it('returns plain text unchanged', () => {
    expect(stripTags('plain text')).toBe('plain text');
  });
});
