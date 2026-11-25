import React from 'react';
import { useTranslation } from 'react-i18next';
import { selectRenniliLanguage } from '../lib/renniliLanguage';

// The three locales this module ships. A customer arrives from a QR code with
// no language chosen, so the choice has to be on the page itself.
const LANGUAGES = [
  { code: 'ar', short: 'ع', name: 'العربية' },
  { code: 'en', short: 'EN', name: 'English' },
  { code: 'fr', short: 'FR', name: 'Français' },
] as const;

const LanguageSwitch = (): React.JSX.Element => {
  const { t, i18n } = useTranslation('pager');
  const current = i18n.resolvedLanguage ?? i18n.language;

  return (
    <div className="pager-languages" role="group" aria-label={t('pager.language.label')}>
      {LANGUAGES.map((language) => (
        <button
          key={language.code}
          type="button"
          className="pager-language"
          lang={language.code}
          aria-label={language.name}
          aria-pressed={current === language.code}
          onClick={() => selectRenniliLanguage(language.code)}
        >
          {language.short}
        </button>
      ))}
    </div>
  );
};

export default LanguageSwitch;
