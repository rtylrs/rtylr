import path from 'node:path';

import { initTestI18nFrom } from '../../../../../../../test/i18nTestInstance';

/**
 * Initialise i18next from this app's real locale files. See the shared
 * implementation for why an identity `t` mock is not the alternative.
 */
export function initTestI18n(lang: string = 'en'): Promise<void> {
  return initTestI18nFrom(path.resolve(__dirname, '../../locales'), lang);
}
