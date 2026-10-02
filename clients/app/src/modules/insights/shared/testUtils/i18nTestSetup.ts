import i18next from 'i18next';
import { initReactI18next } from 'react-i18next';
import enInsights from '../../../../locales/en/insights.json';
import enCommon from '../../../../locales/en/common.json';

let initialized = false;

/**
 * Initializes i18next once with the real, shipped English copy so component
 * tests assert against actual production strings instead of raw i18n keys.
 *
 * This app's i18n bootstrap (`shared/services/i18next.ts`) merges every
 * locales/en/*.json file into a single 'translation' namespace keyed by
 * filename, so `t('insights.reports.exportRevenue')` resolves as a dot-path
 * within that namespace. Reproduce that shape here.
 */
export function initTestI18n(): void {
  if (initialized) return;
  initialized = true;
  void i18next.use(initReactI18next).init({
    lng: 'en',
    fallbackLng: 'en',
    resources: { en: { translation: { insights: enInsights, common: enCommon } } },
    interpolation: { escapeValue: false },
  });
}
