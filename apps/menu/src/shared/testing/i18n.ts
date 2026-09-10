import i18next from 'i18next';
import { initReactI18next } from 'react-i18next';
import resourcesToBackend from 'i18next-resources-to-backend';
import fs from 'node:fs';
import path from 'node:path';

/**
 * Initialise i18next for a test the way the app does at runtime.
 *
 * The alternative, mocking `useTranslation` with an identity `t`, is what this
 * app's suites did, and it is why they passed while **every** key on screen was
 * a dotted path (rtylr#1269): the assertion and the component agree on the key,
 * so neither notices the key never resolves.
 *
 * Being faithful matters here, because the app's resource loading is odd in a
 * way that changes what resolves. `shared/services/i18next.ts` ignores the
 * namespace i18next asks for and returns the **whole** locale tree:
 *
 *   loadResource: (lang) => Promise.resolve(buildLocale(lang))
 *
 * So i18next asks for `('en', 'checkout')` and stores the entire tree under
 * the namespace `checkout`, and again under every other namespace it asks for.
 * That is why `useTranslation('checkout')` plus `t('checkout.emptyCart')`
 * resolves, and why `t('emptyCart')` cannot: the namespace is cosmetic and the
 * root-qualified path is the only thing that works.
 *
 * Registering the tree under a single namespace instead would make every
 * component that names one, which in this app is all of them, read back its
 * own key and fail for the wrong reason.
 */
const LOCALES = path.resolve(__dirname, '../../locales');

function buildLocale(lang: string): Record<string, unknown> {
  const dir = path.join(LOCALES, lang);
  const out: Record<string, unknown> = {};
  for (const file of fs.readdirSync(dir)) {
    if (!file.endsWith('.json')) continue;
    out[file.replace(/\.json$/, '')] = JSON.parse(fs.readFileSync(path.join(dir, file), 'utf8'));
  }
  return out;
}

function namespaces(lang: string): string[] {
  return fs
    .readdirSync(path.join(LOCALES, lang))
    .filter((file) => file.endsWith('.json'))
    .map((file) => file.replace(/\.json$/, ''));
}

export async function initTestI18n(lang: string = 'en'): Promise<void> {
  // Every namespace is preloaded rather than fetched on demand, because
  // `useSuspense: false` lets a component render before a lazy namespace
  // arrives, and it would render the key.
  const ns = namespaces(lang);
  await i18next
    .use(initReactI18next)
    .use(resourcesToBackend(() => Promise.resolve(buildLocale(lang))))
    .init({
      lng: lang,
      fallbackLng: lang,
      defaultNS: 'translation',
      ns: ['translation', ...ns],
      interpolation: { escapeValue: false },
      react: { useSuspense: false },
    });
}
