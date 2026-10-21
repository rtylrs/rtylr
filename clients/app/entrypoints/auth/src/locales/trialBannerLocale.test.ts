import { readFileSync } from 'node:fs';
import path from 'node:path';
import { fileURLToPath } from 'node:url';
import i18next from 'i18next';
import { describe, expect, it } from 'vitest';
import { initTestI18nFrom } from '../../../../test/i18nTestInstance';

const LANGUAGES = ['en', 'fr', 'ar'] as const;
const crmDir = path.resolve(
  path.dirname(fileURLToPath(import.meta.url)),
  '../../../../src/modules/crm'
);

describe('CRM trial banner translations', () => {
  it('uses the localized dismiss action in every supported language', async () => {
    const source = readFileSync(
      path.join(crmDir, 'shared', 'components', 'TrialBanner.tsx'),
      'utf8'
    );
    expect(source).toContain("t('crmLicense.dismiss')");
    expect(source).not.toContain("t('crmLicense.license.dismiss')");

    for (const language of LANGUAGES) {
      const resource = JSON.parse(
        readFileSync(path.join(crmDir, 'locales', language, 'crmLicense.json'), 'utf8')
      ) as Record<string, unknown>;
      expect(resource.dismiss, `crm/${language}: crmLicense.dismiss`).toEqual(expect.any(String));

      await initTestI18nFrom(path.join(crmDir, 'locales'), language);
      expect(i18next.t('crmLicense.dismiss')).not.toBe('license.dismiss');
    }
  });
});
