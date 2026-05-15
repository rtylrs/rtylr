import { describe, expect, it } from 'vitest';
import { highlight } from './highlight';

describe('highlight', () => {
  it('returns the text unchanged when the query is empty', () => {
    expect(highlight('hello', '')).toBe('hello');
  });
  it('wraps matches in <mark>', () => {
    expect(highlight('the cat sat', 'cat')).toBe('the <mark>cat</mark> sat');
  });
  it('is case-insensitive', () => {
    expect(highlight('Foo bar Foo', 'foo')).toBe('<mark>Foo</mark> bar <mark>Foo</mark>');
  });
  it('escapes regex metacharacters in the query', () => {
    expect(highlight('1+1=2', '1+1')).toBe('<mark>1+1</mark>=2');
  });
});
