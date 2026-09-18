import fs from 'node:fs';
import path from 'node:path';
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
  it('prefers the most specific nested route label', () => {
    expect(pageTitleKey('/sales/leads/42')).toBe('nav.leads');
    expect(pageTitleKey('/sales/contracts/42')).toBe('nav.contracts');
  });

  it('does not treat similarly named paths as nested routes', () => {
    // '/salesforce' starts with '/sales', so matching on the bare prefix would
    // title it "Pipeline".
    expect(pageTitleKey('/salesforce')).toBe(FALLBACK_TITLE_KEY);
  });

  it('gives a route its own title where that differs from its sidebar label', () => {
    expect(routeLabels['/feedback'].nav).toBe('nav.feedback');
    expect(pageTitleKey('/feedback')).toBe('nav.titles.feedback');
  });
});

/**
 * These keys live as values in a table rather than inside a literal `t('…')`
 * call, so the repo's key-resolution guard cannot see them. Without this, a
 * typo would paint a raw key into the sidebar and nothing would fail.
 */
describe('every route key resolves', () => {
  const keys = [
    ...new Set([
      ...Object.values(routeLabels).flatMap((l) => [l.nav, l.title]),
      FALLBACK_TITLE_KEY,
      'nav.groups.navigate',
      'nav.groups.create',
      'nav.groups.customers',
      'nav.groups.loyalty',
      'nav.actions.importCustomersCsv',
      'nav.actions.loyaltyRewardsCatalog',
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
