import { initI18n } from '@voxire/i18n';

// Each file under locales/{lang}/*.json is auto-discovered.
// The filename (without .json) becomes the i18next namespace key, so
// `useTranslation('server')` + `t('title')` maps to locales/{lang}/server.json.
const modules = import.meta.glob<{ default: Record<string, unknown> }>(
  '../../../locales/**/*.json',
  {
    eager: true,
  }
);

function loadNamespace(lang: string, ns: string): Record<string, unknown> {
  for (const [path, mod] of Object.entries(modules)) {
    const parts = path.split('/');
    const langDir = parts[parts.length - 2];
    const fileNs = parts[parts.length - 1].replace('.json', '');
    if (langDir === lang && fileNs === ns) {
      return mod.default;
    }
  }
  return {};
}

const i18n = initI18n({
  ns: ['server'],
  loadResource: (lang, ns) => Promise.resolve(loadNamespace(lang, ns)),
});

export { selectLanguage } from '@voxire/i18n';

export default i18n;
