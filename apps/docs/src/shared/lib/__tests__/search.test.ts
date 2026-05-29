import { describe, it, expect } from 'vitest';
import { stripHtml, extractSnippet, searchDocs } from '../search';
import type { DocContent } from '../../content';

describe('stripHtml', () => {
  it('removes simple tags', () => {
    expect(stripHtml('<p>Hello</p>')).toBe('Hello');
  });

  it('removes nested tags', () => {
    expect(stripHtml('<div><strong>World</strong></div>')).toBe('World');
  });

  it('returns plain text unchanged', () => {
    expect(stripHtml('no tags here')).toBe('no tags here');
  });

  it('handles self-closing tags', () => {
    expect(stripHtml('line1<br/>line2')).toBe('line1line2');
  });
});

describe('extractSnippet', () => {
  const body = 'The quick brown fox jumps over the lazy dog and runs away quickly.';

  it('returns a snippet around the query term', () => {
    const snippet = extractSnippet(body, 'fox');
    expect(snippet).toContain('fox');
  });

  it('adds leading ellipsis when the match is not at the start', () => {
    const long = 'a'.repeat(100) + 'TARGET' + 'b'.repeat(100);
    const snippet = extractSnippet(long, 'TARGET', 20);
    expect(snippet.startsWith('...')).toBe(true);
  });

  it('adds trailing ellipsis when match is not at the end', () => {
    const long = 'TARGET' + 'b'.repeat(200);
    const snippet = extractSnippet(long, 'TARGET', 10);
    expect(snippet.endsWith('...')).toBe(true);
  });

  it('returns truncated start when query is not found', () => {
    const snippet = extractSnippet('hello world', 'xyz');
    expect(snippet).toBe('hello world');
  });

  it('is case-insensitive for matching position', () => {
    const snippet = extractSnippet('The FOX jumped', 'fox');
    expect(snippet).toContain('FOX');
  });
});

describe('searchDocs', () => {
  const contentMap: Record<string, DocContent> = {
    intro: { title: 'Introduction', body: '<p>Welcome to rtylr documentation.</p>' },
    pos: { title: 'Point of Sale', body: '<p>Manage orders and payments at the counter.</p>' },
    hr: { title: 'HR', body: '<p>Employee management and payroll features.</p>' },
    erp: {
      title: 'ERP',
      body: '<p>Inventory and supplier management for your business.</p>',
    },
  };

  it('returns empty array for empty query', () => {
    expect(searchDocs(contentMap, '')).toHaveLength(0);
  });

  it('returns empty array for whitespace-only query', () => {
    expect(searchDocs(contentMap, '   ')).toHaveLength(0);
  });

  it('finds a result matching the title', () => {
    const results = searchDocs(contentMap, 'Point of Sale');
    expect(results.some((r) => r.slug === 'pos')).toBe(true);
  });

  it('finds a result matching the body text', () => {
    const results = searchDocs(contentMap, 'payroll');
    expect(results.some((r) => r.slug === 'hr')).toBe(true);
  });

  it('is case-insensitive', () => {
    const results = searchDocs(contentMap, 'INVENTORY');
    expect(results.some((r) => r.slug === 'erp')).toBe(true);
  });

  it('returns no results for an unmatched query', () => {
    expect(searchDocs(contentMap, 'nonexistent_xyz_term')).toHaveLength(0);
  });

  it('returns slug, title and snippet in each result', () => {
    const [result] = searchDocs(contentMap, 'payroll');
    expect(result).toHaveProperty('slug');
    expect(result).toHaveProperty('title');
    expect(result).toHaveProperty('snippet');
  });

  it('places title matches before body-only matches', () => {
    const results = searchDocs(contentMap, 'ERP');
    const erpIdx = results.findIndex((r) => r.slug === 'erp');
    const bodyIdx = results.findIndex((r) => r.slug !== 'erp');
    if (bodyIdx !== -1) {
      expect(erpIdx).toBeLessThan(bodyIdx);
    }
  });
});
