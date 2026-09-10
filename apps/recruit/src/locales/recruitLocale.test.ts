import { describe, it, expect } from 'vitest';
import fs from 'node:fs';
import path from 'node:path';

import { STATUSES } from '../features/admin/SubmissionsAdmin';
import { EMPLOYMENT_TYPES, WORK_MODES } from '../features/admin/CareerEditor';
import { KIND_ORDER } from '../features/admin/FormFieldsEditor';

/**
 * Recruit ships en, ar and fr, so an English string that never reaches a
 * locale file is not latent: Arabic and French candidates read English on a
 * public careers page, and nothing fails, because English is what the key
 * would have returned anyway (rtylr#1266).
 *
 * Three separate things can go wrong, so each gets its own assertion:
 *
 *   1. a namespace exists in one language and not another
 *   2. a key exists in en and not in ar or fr
 *   3. a translation was copied from en and never translated
 *
 * A fourth, an unresolved key, is `keysResolve.test.ts`.
 */
const LOCALES = path.resolve(__dirname);
const SRC = path.resolve(__dirname, '..');
const LANGS = ['en', 'ar', 'fr'] as const;

/**
 * Plural categories are language specific: Arabic uses all six CLDR forms
 * while English and French use two. So ar holding `roleCount_few` and en not
 * is correct, and comparing suffixed keys directly would report it as a defect.
 */
const PLURAL_SUFFIX = /_(zero|one|two|few|many|other)$/;

function flatten(value: unknown, prefix: string, into: Map<string, string>): void {
  if (value && typeof value === 'object' && !Array.isArray(value)) {
    for (const [key, nested] of Object.entries(value as Record<string, unknown>)) {
      flatten(nested, prefix ? `${prefix}.${key}` : key, into);
    }
  } else if (typeof value === 'string') {
    into.set(prefix, value);
  }
}

function namespaces(lang: string): string[] {
  return fs
    .readdirSync(path.join(LOCALES, lang))
    .filter((f) => f.endsWith('.json'))
    .map((f) => f.replace(/\.json$/, ''))
    .sort();
}

function entries(lang: string): Map<string, string> {
  const out = new Map<string, string>();
  for (const ns of namespaces(lang)) {
    flatten(JSON.parse(fs.readFileSync(path.join(LOCALES, lang, `${ns}.json`), 'utf8')), ns, out);
  }
  return out;
}

describe('recruit locales', () => {
  it('has the same namespaces in every language', () => {
    const base = namespaces('en');
    for (const lang of LANGS) {
      expect(namespaces(lang), `${lang} namespaces`).toEqual(base);
    }
  });

  it('has the same keys in every language, allowing per-language plural forms', () => {
    const base = [...entries('en').keys()].map((k) => k.replace(PLURAL_SUFFIX, '_*')).sort();
    const unique = [...new Set(base)];
    for (const lang of LANGS) {
      const got = [
        ...new Set([...entries(lang).keys()].map((k) => k.replace(PLURAL_SUFFIX, '_*'))),
      ].sort();
      expect(got, `${lang} keys`).toEqual(unique);
    }
  });

  it('leaves no value untranslated', () => {
    const en = entries('en');
    // A value may legitimately match English: a brand name, a URL, or a
    // string that is the same word in that language.
    const SHARED = new Set([
      // Carry the brand name, which is not translated.
      'public.notFound.metaTitle',
      'public.landing.metaTitle',
      // Genuinely the same word in French, not an untranslated copy.
      'admin.careers.colActions',
      'admin.submissions.page',
      'admin.questions.optionPlaceholder',
      'admin.questions.seedOption',
    ]);
    const copied: string[] = [];
    for (const lang of ['ar', 'fr'] as const) {
      for (const [key, value] of entries(lang)) {
        const source = en.get(key.replace(PLURAL_SUFFIX, '_one')) ?? en.get(key);
        if (source === undefined || SHARED.has(key)) continue;
        if (value === source && /[A-Za-z]{4}/.test(value)) copied.push(`${lang} ${key}`);
      }
    }
    expect(copied, 'these read as English to an ar or fr candidate').toEqual([]);
  });

  it('labels every submission stage in every language', () => {
    // These labels are read with a runtime built key, `admin.submissions.
    // status.${status}`, so keysResolve.test.ts skips them by design and a
    // missing one would render as the dotted path on the stage filter.
    const missing: string[] = [];
    for (const lang of LANGS) {
      const keys = entries(lang);
      for (const status of STATUSES) {
        if (!keys.has(`admin.submissions.status.${status}`)) missing.push(`${lang} ${status}`);
      }
    }
    expect(missing, 'a stage with no label renders as its key').toEqual([]);
  });

  it('labels every option read with a runtime built key', () => {
    // Each of these is looked up as `${prefix}.${value}`, which
    // keysResolve.test.ts skips by design, so a missing one would render as
    // its dotted path in a dropdown with nothing failing.
    const expected = [
      ...EMPLOYMENT_TYPES.map((o) => `admin.careerEditor.employmentType.${o.key}`),
      ...WORK_MODES.map((o) => `admin.careerEditor.workMode.${o.key}`),
      ...KIND_ORDER.map((k) => `admin.questions.kind.${k}`),
      // A seed label is stored and then read by the candidate, so a missing
      // one would persist a dotted path into the database.
      ...KIND_ORDER.map((k) => `admin.questions.defaultLabel.${k}`),
    ];
    const missing: string[] = [];
    for (const lang of LANGS) {
      const keys = entries(lang);
      for (const key of expected) if (!keys.has(key)) missing.push(`${lang} ${key}`);
    }
    expect(missing, 'an option with no label renders as its key').toEqual([]);
  });

  it('leaves no user-facing English literal on the public careers site', () => {
    // Scoped to what candidates see. The admin screens still carry literals,
    // tracked separately in rtylr#1266; widen this list as each one lands.
    const DIRS = ['features/public', 'templates'];
    const TEXT = /^[A-Z][A-Za-z0-9 ,.'’!?&:/-]*$/;
    const literals: string[] = [];

    for (const dir of DIRS) {
      const walk = (d: string): void => {
        for (const entry of fs.readdirSync(d, { withFileTypes: true })) {
          const full = path.join(d, entry.name);
          if (entry.isDirectory()) {
            walk(full);
            continue;
          }
          if (!/\.tsx$/.test(entry.name) || /\.test\.tsx$/.test(entry.name)) continue;
          const lines = fs.readFileSync(full, 'utf8').split('\n');
          lines.forEach((line, i) => {
            const s = line.trim();
            if (s.startsWith('//') || s.startsWith('*') || s.startsWith('/*')) return;
            if (line.includes('t(')) return;
            // A string handed to a toast or an error setter reaches the
            // candidate just as directly as one in JSX, and the JSX patterns
            // below cannot see it.
            const call =
              /(?:toast\.(?:error|success|info|warning|message)|setUploadError)\(\s*['"]([A-Z][^'"]{3,})['"]/.exec(
                line
              );
            const inline = />\s*([A-Z][A-Za-z][^<>{}]*?)\s*</.exec(line);
            const prev =
              lines
                .slice(0, i)
                .reverse()
                .find((l) => l.trim() !== '')
                ?.trim() ?? '';
            const next =
              lines
                .slice(i + 1)
                .find((l) => l.trim() !== '')
                ?.trim() ?? '';
            const standalone =
              TEXT.test(s) && s.length > 2 && /[>{}]$/.test(prev) && next.startsWith('<');
            const found = call?.[1] ?? inline?.[1] ?? (standalone ? s : undefined);
            if (found !== undefined) literals.push(`${path.relative(SRC, full)}:${i + 1} ${found}`);
          });
        }
      };
      walk(path.join(SRC, dir));
    }

    expect(literals, 'an ar or fr candidate reads these in English').toEqual([]);
  });
});
