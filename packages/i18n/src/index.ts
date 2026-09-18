export { initI18n, selectLanguage } from './init';
export type { I18nConfig } from './init';

export { SUPPORTED_LANGUAGES, RTL_LANGUAGES, isRTL, getLanguage } from './languages';
export type { LanguageDefinition } from './languages';

// Re-export the hooks and components that apps will use most often
export { useTranslation, Trans, Translation } from 'react-i18next';
export { default as i18next } from 'i18next';
