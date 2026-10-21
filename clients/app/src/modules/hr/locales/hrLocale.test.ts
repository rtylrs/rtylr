import { describe, expect, it } from 'vitest';
import fs from 'node:fs';
import path from 'node:path';

/**
 * HR had no locale test, and five components used **no `t()` at all**.
 *
 * `TimeOffNew`, `ReviewForm`, `LeavePolicySettings`, `PayStubCard` and
 * `AnniversaryWidget` were entirely English, in an app with 31 translated
 * namespaces around them (rtylr#1266). Two of them are the forms an employee
 * uses to request leave and a manager uses to file a review.
 *
 * An English default hides this: only Arabic and French readers see it, and
 * nothing fails. So all three halves are checked, the keys, the values and the
 * source.
 *
 * The namespaces are read from disk rather than imported, because there are 31
 * of them and a hand-written import list is one more thing to forget when
 * someone adds the 32nd.
 */
const LOCALES = path.resolve(__dirname);
const LANGS = ['en', 'ar', 'fr'] as const;

type Flat = Record<string, string>;

function flatten(value: unknown, prefix = '', into: Flat = {}): Flat {
  if (value && typeof value === 'object' && !Array.isArray(value)) {
    for (const [key, nested] of Object.entries(value as Record<string, unknown>)) {
      flatten(nested, prefix ? `${prefix}.${key}` : key, into);
    }
  } else if (typeof value === 'string') {
    into[prefix] = value;
  }
  return into;
}

function read(lang: string): Record<string, Flat> {
  const dir = path.join(LOCALES, lang);
  const out: Record<string, Flat> = {};
  for (const file of fs.readdirSync(dir)) {
    if (!file.endsWith('.json')) continue;
    out[file.replace(/\.json$/, '')] = flatten(
      JSON.parse(fs.readFileSync(path.join(dir, file), 'utf8'))
    );
  }
  return out;
}

const byLang = Object.fromEntries(LANGS.map((lang) => [lang, read(lang)])) as Record<
  string,
  Record<string, Flat>
>;

describe('HR translations', () => {
  it('has the same namespaces in every language', () => {
    const en = Object.keys(byLang.en).sort();
    expect(Object.keys(byLang.ar).sort()).toEqual(en);
    expect(Object.keys(byLang.fr).sort()).toEqual(en);
  });

  it('has the same keys in every language', () => {
    const missing: string[] = [];

    for (const [namespace, keys] of Object.entries(byLang.en)) {
      for (const key of Object.keys(keys)) {
        for (const lang of ['ar', 'fr'] as const) {
          if (!(key in (byLang[lang][namespace] ?? {}))) {
            missing.push(`${lang}: ${namespace}.${key}`);
          }
        }
      }
    }

    expect(missing, 'add the key to every locale, not just en').toEqual([]);
  });

  it('leaves no user-facing English literal in the source', () => {
    const SRC = path.resolve(__dirname, '..');

    const walk = (dir: string): string[] =>
      fs.readdirSync(dir, { withFileTypes: true }).flatMap((entry) => {
        const full = path.join(dir, entry.name);
        if (entry.isDirectory()) {
          return entry.name === 'locales' || entry.name === 'node_modules' ? [] : walk(full);
        }
        return /\.tsx$/.test(entry.name) && !/\.test\.tsx$/.test(entry.name) ? [full] : [];
      });

    const offenders: string[] = [];

    for (const file of walk(SRC)) {
      const lines = fs.readFileSync(file, 'utf8').split('\n');
      lines.forEach((line, i) => {
        const trimmed = line.trim();
        if (trimmed.startsWith('*') || trimmed.startsWith('/hr/') || trimmed.startsWith('{/*')) {
          return;
        }
        // Text between JSX tags: a capitalised word plus at least one more.
        // One word alone is too often a component name or an enum value.
        if (!/>\s*[A-Z][a-z]+(\s+[a-z'()/]+){1,8}[.?!]?\s*</.test(line)) return;
        if (line.includes('{t(')) return;
        offenders.push(`${path.relative(SRC, file)}:${i + 1}`);
      });
    }

    expect(offenders, 'put the string through t()').toEqual([]);
  });
});
