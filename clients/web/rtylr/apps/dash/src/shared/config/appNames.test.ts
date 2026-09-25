import fs from 'node:fs';
import path from 'node:path';
import { APP_NAMES, GRANTABLE_APPS } from './appNames';
import { APP_RESOURCES } from './appResources';

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

/**
 * These keys live as values in tables rather than inside a literal `t('…')`
 * call, so the repo's key-resolution guard cannot see them. Without this, a
 * typo would paint a raw key onto a badge and nothing would fail.
 */
describe('every app name key resolves', () => {
  const keys = [
    ...new Set([
      ...Object.values(APP_NAMES).flatMap((a) => [a.short, a.full]),
      ...APP_RESOURCES.flatMap((app) => [
        app.labelKey,
        ...app.sections.map((section) => section.labelKey),
      ]),
      ...GRANTABLE_APPS.map((id) => `appNames.desc.${id}`),
      ...(['pos', 'erp', 'hr', 'crm', 'finance', 'flow'] as const).map(
        (id) => `appNames.blurb.${id}`
      ),
    ]),
  ].sort();

  it.each(LANGUAGES)('in %s', (lang) => {
    const root = tree(lang);
    const missing = keys.filter((key) => typeof resolve(root, key) !== 'string');
    expect(missing).toEqual([]);
  });

  it('checks a real number of keys', () => {
    // A floor, so a table that silently empties cannot pass by checking none.
    expect(keys.length).toBeGreaterThanOrEqual(55);
  });
});

/**
 * The badge falls back to the raw app id in grey when it has no entry, so a
 * grantable app missing from `APP_META` renders as lowercase "insights" rather
 * than a real badge. That is what this pins: insights, menu and recruit were
 * all grantable and all missing.
 */
describe('every grantable app has a badge', () => {
  const badgeSource = fs.readFileSync(
    path.join(__dirname, '..', 'components', 'dash', 'AppBadge.tsx'),
    'utf8'
  );

  it.each(GRANTABLE_APPS)('%s', (id) => {
    expect(badgeSource).toContain(`${id}: {`);
  });
});

describe('the permission table and the app names agree', () => {
  it('names every app in the resource table from the shared registry', () => {
    const unknown = APP_RESOURCES.filter((app) => !app.labelKey.startsWith('appNames.')).map(
      (app) => app.id
    );
    expect(unknown).toEqual([]);
  });

  it('keeps the stored permission identifiers, which are values and not labels', () => {
    // Renaming one of these silently revokes a saved permission.
    const keys = APP_RESOURCES.flatMap((app) => app.sections.map((s) => s.key));
    expect(keys).toContain('erp.products');
    expect(keys).toContain('finance.reports');
    expect(keys).toContain('hr.payroll');
    expect(new Set(keys).size).toBe(keys.length);
  });
});
