import { initI18n } from '@voxire/i18n';

import en from '../../locales/en.json';
import fr from '../../locales/fr.json';
import ar from '../../locales/ar.json';

const flat: Record<string, Record<string, unknown>> = {
  en: en as Record<string, unknown>,
  fr: fr as Record<string, unknown>,
  ar: ar as Record<string, unknown>,
};

const i18n = initI18n({
  loadResource: (lang) => Promise.resolve(flat[lang] ?? flat['en']),
});

export { selectLanguage } from '@voxire/i18n';

export default i18n;
