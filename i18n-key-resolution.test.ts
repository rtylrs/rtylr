import fs from 'node:fs';
import os from 'node:os';
import path from 'node:path';
import { describe, expect, it } from 'vitest';
import {
  availableKeys,
  collectLiteralKeys,
  localeLanguages,
  localeNamespaces,
  unresolvedKeys,
} from './test/i18nKeyResolution';

/**
 * Every literal `t('…')` key in an enforced app must resolve in every locale.
 *
 * The apps' `loadResource(lang)` ignores the namespace i18next asks for and
 * returns the whole locale tree, so the namespace is cosmetic and a key only
 * resolves when it is qualified from the tree root. `t('copyFailed')` under
 * `useTranslation('common')` cannot resolve; `t('common.copyFailed')` does.
 *
 * The failure is quiet in both directions. Where a call site passes an English
 * default, as most do, English renders correctly and only Arabic and French
 * are wrong. Where it passes none, the raw key paints to the user.
 *
 * Enforced apps are listed explicitly, and every app is enforced except one.
 *
 * `menu` is the exception, and not because it is unmeasured: its keys all
 * resolve since rtylr#1277. It fails the namespace-file parity assertion
 * below, because `en` declares 15 locale files, `ar` 11 and `fr` 9, and ten of
 * those 15 namespaces are read by nothing at all. Translating 126 strings into
 * namespaces no call site reaches would satisfy this suite and deliver
 * nothing, so the parity gap is being resolved first, in rtylr#1278.
 */
const ENFORCED_APPS = [
  'auth',
  'crm',
  'dash',
  'erp',
  'finance',
  'flow',
  'hr',
  'insights',
  'recruit',
] as const;

// A floor per app, so a regex that silently stops matching cannot make this
// suite pass by finding nothing to check.
const MINIMUM_KEYS: Record<(typeof ENFORCED_APPS)[number], number> = {
  auth: 40,
  crm: 90,
  dash: 300,
  erp: 150,
  finance: 90,
  flow: 500,
  hr: 700,
  insights: 110,
  recruit: 250,
};

describe.each(ENFORCED_APPS)('%s literal i18n keys resolve', (app) => {
  const srcDir = path.resolve(__dirname, 'apps', app, 'src');
  const localesDir = path.join(srcDir, 'locales');
  const used = collectLiteralKeys(srcDir);
  const languages = localeLanguages(localesDir);

  it('finds translation calls to check, so a passing run means something', () => {
    expect(used.size).toBeGreaterThan(MINIMUM_KEYS[app]);
  });

  it('ships more than one language, so cross-locale checks are meaningful', () => {
    expect(languages.length).toBeGreaterThan(1);
    expect(languages).toContain('en');
  });

  it.each(languages)('every literal key used in the app resolves in %s', (lang) => {
    expect(unresolvedKeys(used, availableKeys(localesDir, lang))).toEqual([]);
  });

  it.each(languages)('%s declares the same namespace files as en', (lang) => {
    expect(localeNamespaces(localesDir, lang)).toEqual(localeNamespaces(localesDir, 'en'));
  });
});

/**
 * Unit coverage for the machinery behind `i18n-key-resolution.test.ts`.
 *
 * That suite runs the machinery over the real apps, which cannot exercise the
 * cases that matter most: a plural key, where the bare form deliberately does
 * not exist, and a key quoted in a comment, where a match would be a false
 * positive. Both are silent failures in opposite directions, so both get a
 * fixture rather than waiting for an app to happen to contain one.
 */
function fixture(files: Record<string, string>): string {
  const dir = fs.mkdtempSync(path.join(os.tmpdir(), 'i18nkeys-'));
  for (const [rel, body] of Object.entries(files)) {
    const full = path.join(dir, rel);
    fs.mkdirSync(path.dirname(full), { recursive: true });
    fs.writeFileSync(full, body);
  }
  return dir;
}

describe('collectLiteralKeys', () => {
  it('finds a call whatever precedes it, including inside an object literal', () => {
    const dir = fixture({
      'a.tsx': [
        "const x = t('one.alpha');",
        "const o = { label: t('one.beta') };",
        "const y = cond ? t('one.gamma') : t('one.delta');",
        "const z = f(t('one.epsilon'));",
      ].join('\n'),
    });

    expect([...collectLiteralKeys(dir).keys()].sort()).toEqual([
      'one.alpha',
      'one.beta',
      'one.delta',
      'one.epsilon',
      'one.gamma',
    ]);
  });

  it('does not treat an identifier ending in t as a translation call', () => {
    const dir = fixture({
      'a.tsx': ["const parts = 'a/b'.split('/');", "const s = format('x');"].join('\n'),
    });

    expect([...collectLiteralKeys(dir).keys()]).toEqual([]);
  });

  it('ignores a key quoted in a comment, which documents rather than calls', () => {
    const dir = fixture({
      'a.tsx': [
        "// so t('common.save') maps to locales/{lang}/common.json",
        ' * measured: t(\'emptyStates.rolesNone\') -> "emptyStates.rolesNone"',
        "const real = t('one.alpha');",
      ].join('\n'),
    });

    expect([...collectLiteralKeys(dir).keys()]).toEqual(['one.alpha']);
  });

  it('records the first place a key is used', () => {
    const dir = fixture({ 'a.tsx': ["const a = t('one.alpha');"].join('\n') });

    expect(collectLiteralKeys(dir).get('one.alpha')).toBe('a.tsx:1');
  });
});

describe('unresolvedKeys', () => {
  const available = (): Set<string> => {
    const dir = fixture({
      'en/one.json': JSON.stringify({ alpha: 'A', count_one: 'one', count_other: 'many' }),
    });
    return availableKeys(dir, 'en');
  };

  it('accepts a plural key by its suffixed forms, since the bare key never exists', () => {
    const used = new Map([['one.count', 'a.tsx:1']]);

    expect(unresolvedKeys(used, available())).toEqual([]);
  });

  it('reports a key that resolves in no form, with where it is used', () => {
    const used = new Map([['one.missing', 'a.tsx:7']]);

    expect(unresolvedKeys(used, available())).toEqual(['one.missing  (a.tsx:7)']);
  });

  it('reports an unqualified key, which is the defect the guard exists for', () => {
    const used = new Map([['alpha', 'a.tsx:3']]);

    expect(unresolvedKeys(used, available())).toEqual(['alpha  (a.tsx:3)']);
  });
});
