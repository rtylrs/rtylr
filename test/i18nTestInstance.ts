import i18next from 'i18next';
import { initReactI18next } from 'react-i18next';
import resourcesToBackend from 'i18next-resources-to-backend';
import fs from 'node:fs';
import path from 'node:path';

/**
 * Initialise i18next for a test the way the apps do at runtime.
 *
 * The alternative, mocking `useTranslation` with an identity `t`, is what most
 * suites in this repo do, and it is the blind spot behind rtylr#1269: the
 * assertion and the component agree on a key, so neither notices the key never
 * resolves. Menu's entire i18n was dotted paths on screen with a green suite.
 * Reading the real files means an assertion written in English also proves the
 * lookup.
 *
 * Being faithful matters, because the apps' resource loading is odd in a way
 * that changes what resolves. Every `shared/services/i18next.ts` ignores the
 * namespace i18next asks for and returns the **whole** locale tree:
 *
 *   loadResource: (lang) => Promise.resolve(buildLocale(lang))
 *
 * So i18next asks for `('en', 'erp')` and stores the entire tree under the
 * namespace `erp`, and again under every other namespace it asks for. The
 * namespace is cosmetic and the root-qualified path is the only thing that
 * works, which is why `useTranslation('erp')` plus `t('warehouses.title')`
 * cannot resolve while `t('erp.warehouses.title')` can.
 *
 * Registering the tree under one namespace instead would make every component
 * that names one read back its own key and fail for the wrong reason.
 */
function buildLocale(localesDir: string, lang: string): Record<string, unknown> {
  const dir = path.join(localesDir, lang);
  const out: Record<string, unknown> = {};
  for (const file of fs.readdirSync(dir)) {
    if (!file.endsWith('.json')) continue;
    out[file.replace(/\.json$/, '')] = JSON.parse(fs.readFileSync(path.join(dir, file), 'utf8'));
  }
  return out;
}

export async function initTestI18nFrom(localesDir: string, lang: string = 'en'): Promise<void> {
  const namespaces = fs
    .readdirSync(path.join(localesDir, lang))
    .filter((file) => file.endsWith('.json'))
    .map((file) => file.replace(/\.json$/, ''));

  // Every namespace is preloaded rather than fetched on demand, because
  // `useSuspense: false` lets a component render before a lazy namespace
  // arrives, and it would render the key.
  await i18next
    .use(initReactI18next)
    .use(resourcesToBackend(() => Promise.resolve(buildLocale(localesDir, lang))))
    .init({
      lng: lang,
      fallbackLng: lang,
      defaultNS: 'translation',
      ns: ['translation', ...namespaces],
      interpolation: { escapeValue: false },
      react: { useSuspense: false },
    });
}
