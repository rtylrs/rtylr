import { useMemo } from 'react';
import { LANGUAGES } from '../components/LanguageSwitcher';

/**
 * Detects the best-matching UI language for the user based on navigator.language.
 * Returns the matched `code` from the LANGUAGES list or `'en'` as a fallback.
 */
export const useBrowserLanguage = (): string => {
  return useMemo(() => {
    if (typeof navigator === 'undefined') return 'en';
    const nav = navigator.language?.slice(0, 2).toLowerCase();
    const match = LANGUAGES.find((l) => l.code === nav);
    return match?.code ?? 'en';
  }, []);
};
