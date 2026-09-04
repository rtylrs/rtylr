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
 * Enforced apps are listed explicitly. The other apps have not been measured
 * against this rule, and each needs its own measurement and issue before it
 * is added here: turning the guard on for an unmeasured app would fail the
 * suite rather than report a finding.
 */
const ENFORCED_APPS = ['dash', 'flow'] as const;

// A floor per app, so a regex that silently stops matching cannot make this
// suite pass by finding nothing to check.
const MINIMUM_KEYS: Record<(typeof ENFORCED_APPS)[number], number> = {
  dash: 300,
  flow: 500,
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
