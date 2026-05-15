import { describe, expect, it } from 'vitest';
import { deriveChatTitle } from './titles';

describe('deriveChatTitle', () => {
  it('returns short messages unchanged', () => {
    expect(deriveChatTitle('Hello')).toBe('Hello');
  });

  it('collapses internal whitespace', () => {
    expect(deriveChatTitle('  hi    there ')).toBe('hi there');
  });

  it('truncates at the configured length with an ellipsis', () => {
    const t = deriveChatTitle('a'.repeat(80), 10);
    expect(t).toHaveLength(10);
    expect(t.endsWith('…')).toBe(true);
  });

  it('does not leave a trailing space before the ellipsis', () => {
    const t = deriveChatTitle('hello world ' + 'a'.repeat(50), 13);
    expect(t.endsWith(' …')).toBe(false);
  });
});
