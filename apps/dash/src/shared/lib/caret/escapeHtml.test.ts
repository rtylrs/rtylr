import { describe, expect, it } from 'vitest';
import { escapeHtml } from './escapeHtml';

describe('escapeHtml', () => {
  it('escapes the five core entities', () => {
    expect(escapeHtml('<a href="x">&\'')).toBe('&lt;a href=&quot;x&quot;&gt;&amp;&#39;');
  });
  it('returns unchanged when nothing to escape', () => {
    expect(escapeHtml('hello world')).toBe('hello world');
  });
});
