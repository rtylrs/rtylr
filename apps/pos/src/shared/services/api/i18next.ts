import { initI18n } from '@voxire/i18n';

// Each file under locales/{lang}/*.json is auto-discovered.
// The filename (without .json) becomes the i18next namespace key, so
// `useTranslation('server')` + `t('title')` maps to locales/{lang}/server.json.
// Non-eager glob: locale files load on demand instead of bundling every
// language into the entry chunk.
const modules = import.meta.glob<{ default: Record<string, unknown> }>(
  '../../../locales/**/*.json'
);

function loadNamespace(lang: string, ns: string): Promise<Record<string, unknown>> {
  const loader = modules[`../../../locales/${lang}/${ns}.json`];
  if (!loader) return Promise.resolve({});
  return loader().then((mod) => mod.default);
}

const i18n = initI18n({
  defaultNS: 'server',
  ns: ['server'],
  loadResource: loadNamespace,
});

export { selectLanguage } from '@voxire/i18n';

export default i18n;
