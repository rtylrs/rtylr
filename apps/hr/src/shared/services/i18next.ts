import i18next from 'i18next';
import { initReactI18next } from 'react-i18next';

import en from '../../locales/en.json';
import fr from '../../locales/fr.json';
import ar from '../../locales/ar.json';

const resources = {
  en: { translation: en },
  fr: { translation: fr },
  ar: { translation: ar },
};

function getCookie(name: string): string | null {
  if (typeof document === 'undefined') return null;
  const nameEQ = `${name}=`;
  for (const part of document.cookie.split('; ')) {
    if (part.startsWith(nameEQ)) return decodeURIComponent(part.slice(nameEQ.length));
  }
  return null;
}

i18next.use(initReactI18next).init({
  resources,
  lng: getCookie('language') || 'en',
  fallbackLng: 'en',
  compatibilityJSON: 'v3',
  interpolation: {
    escapeValue: false,
  },
});

export default i18next;
