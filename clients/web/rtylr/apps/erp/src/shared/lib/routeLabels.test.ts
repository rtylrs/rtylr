import fs from 'node:fs';
import path from 'node:path';
import { describe, expect, it } from 'vitest';
import { FALLBACK_TITLE_KEY, pageTitleKey, routeLabels } from './routeLabels';

const LOCALES = path.join(__dirname, '..', '..', 'locales');
const LANGUAGES = ['en', 'ar', 'fr'] as const;

/** The whole locale tree for a language, the way the app's loader returns it. */
function tree(lang: string): Record<string, unknown> {
  const dir = path.join(LOCALES, lang);
  const out: Record<string, unknown> = {};
  for (const file of fs.readdirSync(dir)) {
    if (!file.endsWith('.json')) continue;
    out[file.replace(/\.json$/, '')] = JSON.parse(fs.readFileSync(path.join(dir, file), 'utf8'));
  }
  return out;
}

function resolve(root: Record<string, unknown>, key: string): unknown {
  return key.split('.').reduce<unknown>((node, part) => {
    if (node === null || typeof node !== 'object') return undefined;
    return (node as Record<string, unknown>)[part];
  }, root);
}

describe('pageTitleKey', () => {
  it('takes an exact route', () => {
    expect(pageTitleKey('/brands')).toBe('nav.brands');
  });

  it('gives a route its own title where that differs from its sidebar label', () => {
    expect(routeLabels['/variants'].nav).toBe('nav.variants');
    expect(pageTitleKey('/variants')).toBe('nav.titles.variants');
  });

  it('falls back to the parent route for a nested path', () => {
    expect(pageTitleKey('/products/42/edit')).toBe('nav.inventory');
  });

  it('prefers the longest matching parent', () => {
    // '/' prefixes everything, so a naive scan would title every nested path
    // "Dashboard".
    expect(pageTitleKey('/price-lists/7')).toBe('nav.priceLists');
  });

  it('falls back to the app name for an unknown path', () => {
    expect(pageTitleKey('/nothing-here')).toBe(FALLBACK_TITLE_KEY);
  });
});

/**
 * These keys live as values in a table rather than inside a literal `t('…')`
 * call, so the repo's key-resolution guard cannot see them. Without this, a
 * typo here would paint a raw key into the sidebar and nothing would fail.
 */
describe('every route key resolves', () => {
  const keys = [
    ...new Set([
      ...Object.values(routeLabels).flatMap((l) => [l.nav, l.title]),
      FALLBACK_TITLE_KEY,
      'nav.groups.workspace',
      'nav.groups.catalog',
      'nav.groups.warehouseOps',
      'nav.groups.insights',
      'nav.groups.purchasing',
    ]),
  ].sort();

  it.each(LANGUAGES)('in %s', (lang) => {
    const root = tree(lang);
    const missing = keys.filter((key) => typeof resolve(root, key) !== 'string');
    expect(missing).toEqual([]);
  });

  it('checks a real number of keys', () => {
    // A floor, so a table that silently empties cannot pass by checking none.
    expect(keys.length).toBeGreaterThanOrEqual(25);
  });
});
