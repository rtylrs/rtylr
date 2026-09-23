import { readFileSync } from 'node:fs';
import path from 'node:path';
import { fileURLToPath } from 'node:url';
import i18next from 'i18next';
import { describe, expect, it } from 'vitest';
import { initTestI18nFrom } from '../../../../test/i18nTestInstance';

const APPS = ['auth', 'crm', 'dash', 'erp', 'finance', 'flow', 'hr', 'insights'] as const;
const LANGUAGES = ['en', 'fr', 'ar'] as const;
const LICENSE_KEYS = [
  'license.expired',
  'license.expiring',
  'license.unreachable',
  'license.unlicensed',
  'license.dismiss',
] as const;

const appsDir = path.resolve(path.dirname(fileURLToPath(import.meta.url)), '../../..');

describe('license banner translations', () => {
  it('keeps the CRM trial banner on the flat dismiss key', () => {
    const source = readFileSync(
      path.join(appsDir, 'crm', 'src', 'shared', 'components', 'TrialBanner.tsx'),
      'utf8'
    );

    expect(source).toContain("t('license.dismiss')");
    expect(source).not.toContain("t('license.license.dismiss')");
  });

  it('resolves the component keys in every app and supported language', async () => {
    for (const app of APPS) {
      for (const language of LANGUAGES) {
        const localeDir = path.join(appsDir, app, 'src', 'locales', language);
        const resource = JSON.parse(
          readFileSync(path.join(localeDir, 'license.json'), 'utf8')
        ) as Record<string, unknown>;
        expect(resource, `${app}/${language} license shape`).not.toHaveProperty('license');
        for (const key of LICENSE_KEYS) {
          expect(resource[key.slice('license.'.length)], `${app}/${language}: ${key}`).toEqual(
            expect.any(String)
          );
        }

        await initTestI18nFrom(path.join(appsDir, app, 'src', 'locales'), language);

        for (const key of LICENSE_KEYS) {
          const value = i18next.t(key);
          expect(value, `${app}/${language}: ${key}`).not.toBe(key);
          expect(String(value).trim(), `${app}/${language}: ${key}`).not.toBe('');
        }
      }
    }
  });
});
