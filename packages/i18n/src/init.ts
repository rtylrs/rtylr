import i18next from 'i18next';
import { initReactI18next } from 'react-i18next';
import resourcesToBackend from 'i18next-resources-to-backend';
import { getCookie, setCookie, getDefaultCookieDomain } from '@voxire/auth';
import { isRTL } from './languages';

export interface I18nConfig {
  /**
   * Async resource loader called by i18next when it needs a (language, namespace) pair.
   *
   * For Vite apps use `import.meta.glob` to lazy-load JSON files:
   *
   * ```ts
   * const modules = import.meta.glob('../locales/**\/*.json');
   *
   * initI18n({
   *   loadResource: (lang, ns) => {
   *     const loader = modules[`../locales/${lang}/${ns}.json`];
   *     if (!loader) throw new Error(`Missing: ${lang}/${ns}`);
   *     return loader().then((m: any) => m.default);
   *   },
   * });
   * ```
   *
   * For apps that still use a flat single-file locale:
   *
   * ```ts
   * import en from '../locales/en.json';
   * import fr from '../locales/fr.json';
   * import ar from '../locales/ar.json';
   *
   * const flat: Record<string, Record<string, unknown>> = { en, fr, ar };
   *
   * initI18n({
   *   loadResource: (lang) => Promise.resolve(flat[lang] ?? flat['en']),
   * });
   * ```
   */
  loadResource: (lang: string, ns: string) => Promise<Record<string, unknown>>;

  /**
   * Default (and initial) i18next namespace.
   * @default 'translation'
   */
  defaultNS?: string;

  /**
   * Fallback language when the user's language doesn't have a translation.
   * @default 'en'
   */
  fallbackLng?: string;

  /**
   * Additional namespaces to preload (besides defaultNS).
   */
  ns?: string[];
}

/**
 * Initialise i18next for a Vite app.
 *
 * Call once at app startup, before rendering. Returns the i18next instance
 * so you can chain `.on('languageChanged', ...)` or inspect state.
 */
export function initI18n(config: I18nConfig): typeof i18next {
  const { loadResource, defaultNS = 'translation', fallbackLng = 'en', ns = [] } = config;

  const savedLang = getCookie('language') || 'en';

  void i18next
    .use(initReactI18next)
    .use(resourcesToBackend(loadResource))
    .init({
      lng: savedLang,
      fallbackLng,
      defaultNS,
      ns: [defaultNS, ...ns],
      interpolation: { escapeValue: false },
      react: { useSuspense: false },
      // Avoids console warnings for missing keys during development
      saveMissing: false,
    });

  // Keep <html dir> in sync when the language changes
  i18next.on('languageChanged', (lang: string) => {
    if (typeof document !== 'undefined') {
      document.documentElement.setAttribute('dir', isRTL(lang) ? 'rtl' : 'ltr');
      document.documentElement.setAttribute('lang', lang);
    }
  });

  // Apply dir immediately on init
  if (typeof document !== 'undefined') {
    document.documentElement.setAttribute('dir', isRTL(savedLang) ? 'rtl' : 'ltr');
    document.documentElement.setAttribute('lang', savedLang);
  }

  return i18next;
}

/**
 * Change the active language, persist in a cookie, and update the page direction.
 */
export function selectLanguage(language: string = 'en'): void {
  void i18next.changeLanguage(language);
  setCookie('language', language, {
    days: 365,
    sameSite: 'lax',
    domain: getDefaultCookieDomain(),
  });
}
