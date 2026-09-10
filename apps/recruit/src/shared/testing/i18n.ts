import i18next from 'i18next';
import { initReactI18next } from 'react-i18next';
import fs from 'node:fs';
import path from 'node:path';

/**
 * Initialise i18next for a test the way the app does at runtime, reading the
 * real locale files off disk.
 *
 * The alternative, mocking `useTranslation` with an identity `t`, is what most
 * of this app's suites do, and it is why five of them passed while every key
 * on screen was unresolved (rtylr#1270): the assertion and the component agree
 * on the key, so neither notices the key never resolves. Loading the real
 * files means an assertion written in English also proves the lookup works.
 *
 * `buildLocale` mirrors `shared/services/i18next.ts`: the loader ignores the
 * namespace argument and returns every file keyed by filename, so the filename
 * is part of the key.
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

export async function initTestI18n(lang: string = 'en'): Promise<void> {
  await i18next.use(initReactI18next).init({
    lng: lang,
    fallbackLng: 'en',
    defaultNS: 'translation',
    ns: ['translation'],
    interpolation: { escapeValue: false },
    react: { useSuspense: false },
    resources: { [lang]: { translation: buildLocale(lang) } },
  });
}
