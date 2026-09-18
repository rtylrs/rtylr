import i18next from 'i18next';
import { initReactI18next } from 'react-i18next';

import ar from '../locales/ar.json';
import en from '../locales/en.json';
import fr from '../locales/fr.json';
import { DEFAULT_LOCALE, localeFromPath } from '../lib/locales';

export const resources = {
  en: { translation: en },
  fr: { translation: fr },
  ar: { translation: ar },
};

/**
 * The active language comes from the URL, never from a cookie.
 *
 * It used to be `lng: getCookie('language') || 'en'`, and that was a live
 * bug waiting for content. `_document.tsx` derives the `lang` and `dir`
 * attributes from the route, so the two disagreed by construction: a
 * visitor carrying `language=ar` would have been served Arabic copy at
 * `/pricing/` under `lang="en"`. That is the doorway case Google treats as
 * duplicate content, on a page whose own markup contradicts its content.
 *
 * It was harmless only because ar.json and fr.json were empty and every key
 * fell through to English. Writing the copy is what would have activated
 * it, which is exactly the kind of bug that appears to arrive from nowhere.
 *
 * Under `output: 'export'` there is a second reason: the server renders with
 * no cookie and the client hydrates with one, so a cookie-selected language
 * is a guaranteed hydration mismatch. The URL is the same on both sides.
 *
 * The URL is also simply the right answer. It is what was requested, it is
 * what gets shared, and it is what the canonical and hreflang point at.
 */
const initialLanguage = (): string => {
  if (typeof window === 'undefined') return DEFAULT_LOCALE;
  return localeFromPath(window.location.pathname).code;
};

i18next.use(initReactI18next).init({
  resources,
  lng: initialLanguage(),
  // en, not fr. This was 'fr', which meant any key missing from the
  // active locale resolved to French: on an English page, for an
  // English visitor. en.json is the source of truth and the only
  // complete locale, so it is the only safe fallback.
  fallbackLng: 'en',
  interpolation: {
    escapeValue: false,
  },
});

export default i18next;
