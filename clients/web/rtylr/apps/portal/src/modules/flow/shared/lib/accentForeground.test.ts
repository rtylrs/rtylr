import fs from 'node:fs';
import path from 'node:path';
import { describe, expect, it } from 'vitest';
import {
  ACCENT_SWATCHES,
  STRESS_ACCENTS,
  appearanceTokens,
  contrastRatio,
  deriveOnAccent,
  isValidHex,
} from '@voxire/utils';

/**
 * Text on an accent fill must stay readable at every accent the picker offers.
 *
 * The accent is user-settable, so a hardcoded foreground is not merely
 * inconsistent: with `#ffe680`, a shade the picker itself offers, white text
 * on the accent measures 1.24:1 and the app's main action becomes invisible.
 *
 * `AppearanceProvider` already derives `--flow-on-accent` per accent, and
 * measures both candidates against the accent in play rather than assuming.
 * So the app's job is only to consume it, which is what the source scan below
 * enforces. The contrast cases above it re-check the guarantee Flow is now
 * relying on, at the concrete accents it can actually be set to.
 */
const SRC = path.resolve(__dirname, '../..');
const AA_NORMAL_TEXT = 4.5;

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

// ACCENT_SWATCHES holds { id, labelKey, hex }, not bare strings. Passing the
// objects through made deriveOnAccent fall back to its default and every case
// pass on a value no user can select, so the hex is unwrapped here and
// asserted to be one below.
//
// #ffe680 is the value the checklist names for this test. It is not one of the
// ten swatches, so it is reachable through the custom hex field rather than
// the picker grid, and it is included explicitly because it is the shade that
// exposed the bug.
const everyAccent = [
  ...new Set([...ACCENT_SWATCHES.map((swatch) => swatch.hex), ...STRESS_ACCENTS, '#ffe680']),
];

describe('foreground on an accent fill', () => {
  it('has real hex accents to check, so a passing run means something', () => {
    expect(ACCENT_SWATCHES.length).toBeGreaterThanOrEqual(10);
    expect(everyAccent.length).toBeGreaterThan(ACCENT_SWATCHES.length);
    // Without this, an accent list of the wrong shape makes every case below
    // pass on a fallback colour instead of the one under test.
    for (const accent of everyAccent) {
      expect(isValidHex(accent), accent).toBe(true);
    }
  });

  it.each(everyAccent)('the derived foreground clears AA on %s', (accent) => {
    expect(contrastRatio(deriveOnAccent(accent), accent)).toBeGreaterThanOrEqual(AA_NORMAL_TEXT);
  });

  it.each(['dark', 'light'] as const)(
    'the token the provider writes clears AA on every accent in %s mode',
    (mode) => {
      for (const accent of everyAccent) {
        const tokens = appearanceTokens(
          { accent },
          { resolvedMode: mode, workspaceAccent: accent }
        ) as unknown as Record<string, string>;
        const onAccent = tokens['--flow-on-accent'];
        expect(onAccent, `--flow-on-accent for ${accent}`).toBeTruthy();
        expect(
          contrastRatio(onAccent, accent),
          `${onAccent} on ${accent} in ${mode}`
        ).toBeGreaterThanOrEqual(AA_NORMAL_TEXT);
      }
    }
  );

  it('never hardcodes a foreground on an accent fill', () => {
    // The failure mode this exists for: `background: 'var(--accent)', color:
    // '#ffffff'`. A literal here cannot adapt when the user changes the
    // accent, whether it is white or the near-black that suits today's teal.
    const offenders: string[] = [];
    const pattern = /background:\s*'var\(--accent\)',\s*(?:\n\s*)?color:\s*'#[0-9a-fA-F]{3,8}'/g;
    for (const file of sourceFiles(SRC)) {
      const source = fs.readFileSync(file, 'utf8');
      for (const match of source.matchAll(pattern)) {
        offenders.push(`${path.relative(SRC, file)}  ${match[0].replace(/\s+/g, ' ')}`);
      }
    }
    expect(offenders).toEqual([]);
  });

  it('never puts a utility-class white on an accent fill', () => {
    const offenders: string[] = [];
    for (const file of sourceFiles(SRC)) {
      const lines = fs.readFileSync(file, 'utf8').split('\n');
      lines.forEach((line, index) => {
        if (!line.includes('text-white')) return;
        if (!line.includes('var(--accent)')) return;
        offenders.push(`${path.relative(SRC, file)}:${index + 1}`);
      });
    }
    expect(offenders).toEqual([]);
  });
});
