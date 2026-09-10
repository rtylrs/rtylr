import { describe, it, expect } from 'vitest';
import fs from 'node:fs';
import path from 'node:path';

/**
 * A `var()` must never be concatenated with a hex alpha suffix.
 *
 * `${color}40` is how this app builds a tint, and it needs a real hex: CSS has
 * no way to read `var(--accent)40`, so the whole declaration is invalid and the
 * browser drops it. Silently, which is the problem.
 *
 * It bit during rtylr#1242. Turning `const ACCENT = '#8b5cf6'` into
 * `'var(--accent)'` fixed five style uses and broke a sixth,
 * `1px solid ${ACCENT}40`, whose border simply stopped existing. Nothing
 * failed: not the types, not the tests, not the build.
 *
 * The fix at that site was `--accent-border`, a tint this app already declares.
 * Where no token exists, `color-mix(in srgb, var(--x) N%, transparent)` does the
 * same job and does accept a `var()`.
 *
 * This is a source check because nothing at runtime here can see it. happy-dom
 * does not resolve custom properties, so the invalid declaration and the valid
 * one look identical to a test that renders the component.
 */
const SRC = path.resolve(__dirname, '../..');

function sourceFiles(dir: string, into: string[] = []): string[] {
  for (const entry of fs.readdirSync(dir, { withFileTypes: true })) {
    const full = path.join(dir, entry.name);
    if (entry.isDirectory()) {
      if (entry.name === 'node_modules' || entry.name === 'locales') continue;
      sourceFiles(full, into);
      continue;
    }
    if (/\.tsx?$/.test(entry.name) && !/\.test\.tsx?$/.test(entry.name)) into.push(full);
  }
  return into;
}

describe('a var() never carries a hex alpha suffix', () => {
  it('leaves no template literal appending hex digits to a var()', () => {
    const offenders: string[] = [];

    for (const file of sourceFiles(SRC)) {
      const lines = fs.readFileSync(file, 'utf8').split('\n');
      lines.forEach((line, i) => {
        const trimmed = line.trim();
        if (trimmed.startsWith('*') || trimmed.startsWith('//')) return;
        // `${something}` immediately followed by two hex digits, where the
        // something resolves to a var(). Checked by name rather than by value,
        // since the constant is declared elsewhere in the file.
        for (const match of line.matchAll(/\$\{([A-Za-z_$][\w$.?]*)\}([0-9a-fA-F]{2})\b/g)) {
          const name = match[1];
          const declaresVar = new RegExp(
            `(?:const|let)\\s+${name.split(/[.?]/)[0]}\\s*=\\s*['"\`]var\\(`
          );
          if (declaresVar.test(lines.join('\n'))) {
            offenders.push(`${path.relative(SRC, file)}:${i + 1} (\${${name}}${match[2]})`);
          }
        }
      });
    }

    expect(offenders, 'use a tint token or color-mix, not a hex suffix').toEqual([]);
  });
});
