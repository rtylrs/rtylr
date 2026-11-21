import path from 'node:path';

import { initTestI18nFrom } from '../../../../../test/i18nTestInstance';

/** Initialise i18next from this module's real locale files. */
export function initTestI18n(lang: string = 'en'): Promise<void> {
  return initTestI18nFrom(path.resolve(__dirname, '../../locales'), lang);
}
