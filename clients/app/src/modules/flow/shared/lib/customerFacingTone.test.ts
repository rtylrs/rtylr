import fs from 'node:fs';
import path from 'node:path';
import { describe, expect, it } from 'vitest';

/**
 * Customer copy does not name the implementation.
 *
 * `brand.md` puts the in-app tone at "direct + functional" and reserves
 * "engineering-direct" for internal docs. Three strings broke that, telling a
 * restaurant owner about endpoints:
 *
 *   "Grouping by project or type of work needs an endpoint that doesn't exist yet."
 *   "Logging against another day needs a date on the create endpoint."
 *   "They will move to the workspace once the time settings endpoint lands"
 *
 * A limitation is worth stating. Which service call is missing is not: the
 * reader cannot act on it and it reads as an excuse.
 *
 * This scan reads every locale file but is not about one, so it lives here
 * rather than in locales/ beside flowLocale and plurals. Anything listing
 * that directory to find the languages has to filter for directories, since
 * it already holds those two test files.
 */
const LOCALES = path.resolve(__dirname, '../../locales');

// Words that only make sense to someone who has the codebase open. Matched
// per language, since the Arabic and French strings translated them faithfully.
const IMPLEMENTATION_WORDS = [
  'endpoint',
  'API call',
  'backend',
  'stubbed',
  'not implemented',
  'نقطة الوصول',
  'نقطة الإنشاء',
  'نقطة إعدادات',
  "point d'accès",
  'point de création',
];

function localeFiles(): string[] {
  const found: string[] = [];
  for (const lang of fs.readdirSync(LOCALES, { withFileTypes: true })) {
    if (!lang.isDirectory()) continue;
    for (const file of fs.readdirSync(path.join(LOCALES, lang.name))) {
      if (file.endsWith('.json')) found.push(path.join(LOCALES, lang.name, file));
    }
  }
  return found;
}

function strings(value: unknown, keyPath: string, into: [string, string][]): void {
  if (typeof value === 'string') {
    into.push([keyPath, value]);
    return;
  }
  if (value !== null && typeof value === 'object') {
    for (const [key, child] of Object.entries(value as Record<string, unknown>)) {
      strings(child, keyPath ? `${keyPath}.${key}` : key, into);
    }
  }
}

describe('customer-facing copy', () => {
  const files = localeFiles();

  it('reads the locale files, so a passing run means something', () => {
    expect(files.length).toBeGreaterThan(20);
  });

  it('names no implementation detail', () => {
    const offenders: string[] = [];
    for (const file of files) {
      const entries: [string, string][] = [];
      strings(JSON.parse(fs.readFileSync(file, 'utf8')), '', entries);
      for (const [keyPath, text] of entries) {
        for (const word of IMPLEMENTATION_WORDS) {
          if (text.toLowerCase().includes(word.toLowerCase())) {
            offenders.push(
              `${path.basename(path.dirname(file))}/${path.basename(file)}  ${keyPath}`
            );
            break;
          }
        }
      }
    }
    expect([...new Set(offenders)]).toEqual([]);
  });
});
