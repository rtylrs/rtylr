import { describe, expect, it } from 'vitest';
import { extractCitations, stripCitations } from './citations';

describe('extractCitations', () => {
  it('finds orders citations', () => {
    const c = extractCitations('Refund issued. [source: orders/123]');
    expect(c).toHaveLength(1);
    expect(c[0].href).toBe('/orders/123');
  });

  it('ignores unknown prefixes', () => {
    expect(extractCitations('[source: secret/1]')).toEqual([]);
  });

  it('handles multiple citations', () => {
    const c = extractCitations('foo [source: orders/1] bar [source: tasks/2]');
    expect(c).toHaveLength(2);
  });
});

describe('stripCitations', () => {
  it('removes the citation markers and trims whitespace', () => {
    const cleaned = stripCitations('Hello [source: orders/1] world');
    expect(cleaned).toBe('Hello world');
  });
});
