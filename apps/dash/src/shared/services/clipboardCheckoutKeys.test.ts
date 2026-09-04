import { readFileSync } from 'node:fs';
import { describe, expect, it } from 'vitest';
import i18n from './i18next';

// Dash's loadResource ignores the namespace i18next asks for and returns the
// whole locale tree, so the namespace is cosmetic and every key has to be
// fully qualified from the tree root: `common.copyFailed`, not `copyFailed`
// under useTranslation('common'). An unqualified key resolves to nothing, and
// because these call sites pass an English default value the miss is silent -
// ar and fr render English rather than showing a missing-key marker.
//
// These are the clipboard and checkout failure paths from #1120. Asserting the
// keys exist in the JSON is not enough: the keys already existed and were
// still unreachable. Resolution through the real i18n instance is the check.
const LANGUAGES = ['en', 'ar', 'fr'] as const;

const FAILURE_PATH_KEYS = [
  'common.copyFailed',
  'billing.checkoutFailed',
  'developer.apiKeys.copied',
  'developer.apiKeys.copyFailed',
  'developer.apiKeys.previewCopied',
] as const;

// file -> the qualified keys its clipboard/checkout handlers must use.
const CALL_SITES: Record<string, readonly string[]> = {
  'apps/dash/src/features/security/Security.tsx': ['common.copyFailed'],
  'apps/dash/src/shared/components/dash/ApiKeyRow.tsx': [
    'common.copyFailed',
    'developer.apiKeys.previewCopied',
  ],
  'apps/dash/src/features/developer/ApiKeys.tsx': [
    'developer.apiKeys.copied',
    'developer.apiKeys.copyFailed',
  ],
  'apps/dash/src/features/billing/Billing.tsx': ['billing.checkoutFailed'],
};

describe('clipboard and checkout failure copy is localized', () => {
  it('resolves every failure-path key in each language', async () => {
    await i18n.init();
    await i18n.loadNamespaces(['translation', 'common']);

    for (const lng of LANGUAGES) {
      await i18n.changeLanguage(lng);
      for (const key of FAILURE_PATH_KEYS) {
        const value = i18n.t(key);
        // A miss returns the key itself, which is the exact failure #1120 hit.
        expect(value, `${key} in ${lng}`).not.toBe(key);
        expect(value.trim().length, `${key} in ${lng}`).toBeGreaterThan(0);
      }
    }
  });

  it('gives ar and fr their own wording rather than falling back to English', async () => {
    await i18n.init();
    await i18n.loadNamespaces(['translation', 'common']);
    const byLanguage: Record<string, string[]> = {};

    for (const lng of LANGUAGES) {
      await i18n.changeLanguage(lng);
      byLanguage[lng] = FAILURE_PATH_KEYS.map((key) => i18n.t(key));
    }

    expect(byLanguage.ar).not.toEqual(byLanguage.en);
    expect(byLanguage.fr).not.toEqual(byLanguage.en);
    for (const [index, key] of FAILURE_PATH_KEYS.entries()) {
      expect(byLanguage.ar[index], `ar ${key}`).not.toBe(byLanguage.en[index]);
      expect(byLanguage.fr[index], `fr ${key}`).not.toBe(byLanguage.en[index]);
    }
  });

  it('keeps the call sites on the qualified key form', () => {
    for (const [file, keys] of Object.entries(CALL_SITES)) {
      const source = readFileSync(file, 'utf8');
      for (const key of keys) {
        expect(source, `${file} should call t('${key}')`).toContain(`'${key}'`);
      }
      // The unqualified tail of each key must not appear as a key on its own,
      // which is how the miss was originally written.
      for (const key of keys) {
        const tail = key.split('.').pop() as string;
        expect(source, `${file} still calls t('${tail}')`).not.toContain(`t('${tail}'`);
      }
    }
  });
});
