import { describe, expect, it } from 'vitest';
import { safeHighlight } from './searchHighlight';

describe('safeHighlight', () => {
  it('escapes the content', () => {
    expect(safeHighlight('<a>', 'a')).toBe('&lt;<mark>a</mark>&gt;');
  });
  it('returns the text unchanged for empty query', () => {
    expect(safeHighlight('hello', '')).toBe('hello');
  });
});
