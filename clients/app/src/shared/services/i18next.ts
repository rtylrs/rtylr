import { initI18n } from '@voxire/i18n';

// Each file under locales/{lang}/*.json is auto-discovered.
// The filename (without .json) becomes the i18next namespace key,
// so t('common.save') maps to locales/{lang}/common.json → save.
const modules = import.meta.glob<{ default: Record<string, unknown> }>(
  [
    '../../locales/**/*.json',
    '../../modules/flow/locales/**/*.json',
    '../../modules/crm/locales/**/*.json',
    '../../modules/erp/locales/**/*.json',
    '../../modules/finance/locales/**/*.json',
    '../../modules/hr/locales/**/*.json',
    '../../modules/recruit/locales/**/*.json',
  ],
  {
    eager: true,
  }
);

function buildLocale(lang: string): Record<string, unknown> {
  const result: Record<string, unknown> = {};
  for (const [path, mod] of Object.entries(modules)) {
    const parts = path.split('/');
    const langDir = parts[parts.length - 2];
    const ns = parts[parts.length - 1].replace('.json', '');
    if (langDir === lang) {
      result[ns] = mod.default;
    }
  }
  return result;
}

const i18n = initI18n({
  loadResource: (lang) => Promise.resolve(buildLocale(lang)),
});

export { selectLanguage } from '@voxire/i18n';

export default i18n;
