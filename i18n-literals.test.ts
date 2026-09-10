import fs from 'node:fs';
import os from 'node:os';
import path from 'node:path';
import { describe, expect, it } from 'vitest';

import { collectLiterals, formatLiterals } from './test/i18nLiterals';

/**
 * No hardcoded English in an enforced app's source.
 *
 * Every app ships en, ar and fr, so a literal in JSX is not latent: an Arabic
 * or French reader sees English and nothing fails (rtylr#1266).
 *
 * Enforced apps are listed explicitly, and the list is short on purpose. The
 * measured counts elsewhere are large, 449 in crm and 420 in dash, so turning
 * this on for an unmeasured app would fail the suite rather than report a
 * finding. Each app joins as it is cleared, the way `ENFORCED_APPS` in
 * `i18n-key-resolution.test.ts` grew from two to nine.
 */
const ENFORCED_APPS = ['flow', 'insights', 'recruit'] as const;

/**
 * Text that is deliberately not translated, per app.
 *
 * Three kinds belong here and nothing else: a format example, a physical
 * label such as a key cap, and a specimen that must stay in the script it
 * demonstrates. Anything else in this list is a defect being hidden.
 */
const ALLOWED: Record<(typeof ENFORCED_APPS)[number], readonly string[]> = {
  // A key cap. The user's keyboard says ESC whatever language they read.
  flow: ['ESC'],
  // A date format the user types into, not prose.
  insights: ['MM-DD'],
  // A pangram rendered in each pairing's own Latin faces, so translated
  // sample text would go through a fallback and misrepresent the typeface.
  recruit: ['The quick brown fox', 'jumps over the lazy dog.'],
};

/**
 * A floor per app, so a pattern that silently stops matching cannot make this
 * suite pass by finding nothing to check. Counted from the strings each app
 * already puts through `t()`.
 */
const MINIMUM_KEYS: Record<(typeof ENFORCED_APPS)[number], number> = {
  flow: 500,
  insights: 110,
  recruit: 250,
};

const KEY_CALL = /(?:^|[^A-Za-z0-9_$])t\(\s*(['"])([A-Za-z][A-Za-z0-9_.]*)\1/g;

function translatedKeys(srcDir: string): Set<string> {
  const keys = new Set<string>();
  const walk = (dir: string): void => {
    for (const entry of fs.readdirSync(dir, { withFileTypes: true })) {
      const full = path.join(dir, entry.name);
      if (entry.isDirectory()) {
        if (entry.name === 'node_modules' || entry.name === 'locales') continue;
        walk(full);
        continue;
      }
      if (!/\.tsx?$/.test(entry.name) || /\.test\.tsx?$/.test(entry.name)) continue;
      for (const match of fs.readFileSync(full, 'utf8').matchAll(KEY_CALL)) keys.add(match[2]);
    }
  };
  walk(srcDir);
  return keys;
}

describe.each(ENFORCED_APPS)('%s puts its user-facing copy through i18n', (app) => {
  const srcDir = path.resolve(__dirname, 'apps', app, 'src');

  it('translates enough to make a passing run mean something', () => {
    expect(translatedKeys(srcDir).size).toBeGreaterThan(MINIMUM_KEYS[app]);
  });

  it('leaves no hardcoded English string in the source', () => {
    expect(formatLiterals(collectLiterals(srcDir, ALLOWED[app]))).toEqual([]);
  });
});

/**
 * The collector's own cases. Running it over the real apps cannot exercise
 * what it must not match, and a guard that quietly matches nothing passes.
 */
function fixture(files: Record<string, string>): string {
  const dir = fs.mkdtempSync(path.join(os.tmpdir(), 'i18nlit-'));
  for (const [rel, body] of Object.entries(files)) {
    const full = path.join(dir, rel);
    fs.mkdirSync(path.dirname(full), { recursive: true });
    fs.writeFileSync(full, body);
  }
  return dir;
}

describe('collectLiterals', () => {
  it('finds text on its own line, which is where prettier puts most of it', () => {
    const dir = fixture({
      'a.tsx': ['<button', '  type="button"', '>', '  Save changes', '</button>'].join('\n'),
    });

    expect(formatLiterals(collectLiterals(dir))).toEqual(['a.tsx:4  Save changes']);
  });

  it('finds a toast argument and a confirmation label, which no markup sweep sees', () => {
    const dir = fixture({
      'a.tsx': ["toast.error('Could not save');", "toast.warning('Delete this section?');"].join(
        '\n'
      ),
    });

    expect(formatLiterals(collectLiterals(dir))).toEqual([
      'a.tsx:1  Could not save',
      'a.tsx:2  Delete this section?',
    ]);
  });

  it('finds a prop that is the only text an icon button has', () => {
    const dir = fixture({ 'a.tsx': ['<IconButton title="Remove" />'].join('\n') });

    expect(formatLiterals(collectLiterals(dir))).toEqual(['a.tsx:1  Remove']);
  });

  it('leaves a line that already calls t alone', () => {
    const dir = fixture({
      'a.tsx': ["<span>{t('a.b')}</span>", "toast.error(t('a.c'));"].join('\n'),
    });

    expect(formatLiterals(collectLiterals(dir))).toEqual([]);
  });

  it('leaves a comment alone, since documentation quotes copy as an example', () => {
    const dir = fixture({
      'a.tsx': ['// Save changes', ' * Delete this section?', '{/* Remove */}'].join('\n'),
    });

    expect(formatLiterals(collectLiterals(dir))).toEqual([]);
  });

  it('honours the allowed list, which is how a format example stays literal', () => {
    const dir = fixture({ 'a.tsx': ['<Input placeholder="MM-DD" />'].join('\n') });

    expect(formatLiterals(collectLiterals(dir, ['MM-DD']))).toEqual([]);
    expect(formatLiterals(collectLiterals(dir))).toEqual(['a.tsx:1  MM-DD']);
  });

  it('does not read a test file, which asserts on copy by design', () => {
    const dir = fixture({ 'a.test.tsx': ['<span>Save changes</span>'].join('\n') });

    expect(formatLiterals(collectLiterals(dir))).toEqual([]);
  });
});
