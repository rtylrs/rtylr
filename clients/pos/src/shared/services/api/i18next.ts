import { initI18n } from '@voxire/i18n';

// Locale selection remains available for localized number formatting.
const modules = import.meta.glob<{ default: Record<string, unknown> }>(
  '../../../locales/**/*.json'
);

function loadNamespace(lang: string, ns: string): Promise<Record<string, unknown>> {
  const loader = modules[`../../../locales/${lang}/${ns}.json`];
  if (!loader) return Promise.resolve({});
  return loader().then((mod) => mod.default);
}

const i18n = initI18n({
  defaultNS: 'translation',
  loadResource: loadNamespace,
});

export { selectLanguage } from '@voxire/i18n';

export default i18n;
