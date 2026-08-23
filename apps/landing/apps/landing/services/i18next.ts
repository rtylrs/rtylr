import i18next from 'i18next';
import { initReactI18next } from 'react-i18next';

import en from '../locales/en.json';
import fr from '../locales/fr.json';
import ar from '../locales/ar.json';

export const resources = {
  en: {
    translation: en,
  },
  fr: {
    translation: fr,
  },
  ar: {
    translation: ar,
  },
};

import { getCookie, setCookie, getDefaultCookieDomain } from '../lib/cookies';

i18next.use(initReactI18next).init({
  resources,
  lng: getCookie('language') || 'en',
  // en, not fr. This was 'fr', which meant any key missing from the
  // active locale resolved to French: on an English page, for an
  // English visitor. en.json is the source of truth and the only
  // complete locale, so it is the only safe fallback.
  fallbackLng: 'en',
  interpolation: {
    escapeValue: false,
  },
});

export const selectLanguage = (language?: string) => {
  if (language) {
    i18next.changeLanguage(language);
    setCookie('language', language, {
      days: 365,
      sameSite: 'lax',
      domain: getDefaultCookieDomain(),
    });
  } else {
    language = 'en';
    i18next.changeLanguage(language);
    setCookie('language', language, {
      days: 365,
      sameSite: 'lax',
      domain: getDefaultCookieDomain(),
    });
  }
};

export default i18next;
