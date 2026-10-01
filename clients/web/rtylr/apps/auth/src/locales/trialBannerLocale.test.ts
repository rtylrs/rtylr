import { readFileSync } from 'node:fs';
import path from 'node:path';
import { fileURLToPath } from 'node:url';
import i18next from 'i18next';
import { describe, expect, it } from 'vitest';
import { initTestI18nFrom } from '../../../../test/i18nTestInstance';

const LANGUAGES = ['en', 'fr', 'ar'] as const;
const appsDir = path.resolve(path.dirname(fileURLToPath(import.meta.url)), '../../..');

describe('CRM trial banner translations', () => {
  it('uses the localized dismiss action in every supported language', async () => {
    const source = readFileSync(
      path.join(appsDir, 'crm', 'src', 'shared', 'components', 'TrialBanner.tsx'),
      'utf8'
    );
    expect(source).toContain("t('license.dismiss')");
    expect(source).not.toContain("t('license.license.dismiss')");

    for (const language of LANGUAGES) {
      const resource = JSON.parse(
        readFileSync(path.join(appsDir, 'crm', 'src', 'locales', language, 'license.json'), 'utf8')
      ) as Record<string, unknown>;
      expect(resource.dismiss, `crm/${language}: license.dismiss`).toEqual(expect.any(String));

      await initTestI18nFrom(path.join(appsDir, 'crm', 'src', 'locales'), language);
      expect(i18next.t('license.dismiss')).not.toBe('license.dismiss');
    }
  });
});
