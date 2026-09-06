import { describe, it, expect } from 'vitest';
import fs from 'node:fs';
import path from 'node:path';
import { contrastRatio, compositeOver } from '@voxire/utils';

/**
 * `--accent` and `--accent-fg` are derived against the card surface. That
 * guarantee does not transfer to `--accent-muted`, which is a different
 * background, so `--accent-on-muted` exists and is derived against the tint
 * itself (rtylr#970).
 *
 * Flow had 16 places pairing the tint with the wrong foreground. It looked
 * correct in review because for the default teal the two tokens hold the same
 * value; the difference only appears once a user picks their own accent, and
 * `deriveAccentFgAcross` recomputes `--accent-on-muted` per accent while
 * `--accent` stays whatever they chose.
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
    if (!/\.tsx?$/.test(entry.name)) continue;
    if (/\.test\.tsx?$/.test(entry.name)) continue;
    into.push(full);
  }
  return into;
}

describe('the accent tint is paired with its own foreground', () => {
  it('measures why: --accent on --accent-muted fails AA in light mode', () => {
    // Not an assertion about the code, an assertion about the reason. If this
    // ever passes, the pairing stopped being a problem and the rule below can
    // be revisited rather than blindly kept.
    const accents = ['#64c5b9', '#3b82f6', '#f59e0b', '#ef4444', '#a855f7'];
    for (const accent of accents) {
      const lightTint = compositeOver(accent, 0.12, '#ffffff');
      expect(contrastRatio(accent, lightTint), `${accent} on its own light tint`).toBeLessThan(4.5);
    }
  });

  it('leaves no source pairing --accent with the muted tint', () => {
    const offenders: string[] = [];

    for (const file of sourceFiles(SRC)) {
      const lines = fs.readFileSync(file, 'utf8').split('\n');
      lines.forEach((line, i) => {
        const trimmed = line.trim();
        if (trimmed.startsWith('*') || trimmed.startsWith('//')) return;
        if (!/color:\s*[^,\n]*var\(--accent(-fg)?\)/.test(line)) return;
        const window = lines.slice(Math.max(0, i - 3), i + 4).join('\n');
        if (!window.includes('accent-muted')) return;
        if (window.includes('accent-on-muted')) return;
        offenders.push(`${path.relative(SRC, file)}:${i + 1}`);
      });
    }

    expect(offenders, 'use var(--accent-on-muted) on an --accent-muted background').toEqual([]);
  });
});
