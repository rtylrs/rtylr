import { describe, expect, it } from 'vitest';
import { asciiQuotes } from './stripUnicodeQuotes';

describe('asciiQuotes', () => {
  it('converts curly singles', () => {
    expect(asciiQuotes('‘hi’')).toBe("'hi'");
  });
  it('converts curly doubles', () => {
    expect(asciiQuotes('“hi”')).toBe('"hi"');
  });
  it('leaves ascii alone', () => {
    expect(asciiQuotes('"hi"')).toBe('"hi"');
  });
});
