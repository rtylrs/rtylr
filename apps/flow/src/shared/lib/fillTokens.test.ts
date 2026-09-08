import { describe, it, expect } from 'vitest';
import fs from 'node:fs';
import path from 'node:path';
import { contrastRatio, deriveOnAccent } from '@voxire/utils';

/**
 * The danger and warn tones are used as fills, and a fill needs two things a
 * foreground tone does not: text on it that clears 4.5:1, and enough separation
 * from the surface behind it that the control is distinguishable at 3:1.
 *
 * Four destructive buttons painted white on the danger tone at 3.76:1 and the
 * trial banner painted white on the warn tone at 2.29:1 (rtylr#1234). White was
 * assumed rather than measured, and which extreme actually wins flips with the
 * mode.
 */
const CSS = fs.readFileSync(path.resolve(__dirname, '../../globals.css'), 'utf8');

/** Declarations from the dark block (or base) and the light block separately. */
function tone(token: string): { dark: string; light: string } {
  const all = [...CSS.matchAll(new RegExp(`^\\s*${token}:\\s*(#[0-9a-f]{6})\\s*;`, 'gim'))].map(
    (m) => m[1]
  );
  expect(all.length, `${token} should be declared for both modes`).toBe(2);
  return { dark: all[0], light: all[1] };
}

const CARD = { dark: '#161616', light: '#ffffff' };
const MODES = ['dark', 'light'] as const;

describe('the danger and warn fills', () => {
  it('carries a foreground that clears AA in both modes', () => {
    const failures: string[] = [];

    for (const family of ['danger', 'warn'] as const) {
      const fill = tone(`--flow-${family}-fg`);
      const fg = tone(`--flow-${family}-on-fill`);

      for (const mode of MODES) {
        const ratio = contrastRatio(fg[mode], fill[mode]);
        if (ratio < 4.5) {
          failures.push(`${family} ${mode}: ${fg[mode]} on ${fill[mode]} is ${ratio.toFixed(2)}:1`);
        }
      }
    }

    expect(failures, 'text on a filled control must clear 4.5:1').toEqual([]);
  });

  it('measures why white could not simply be kept', () => {
    // The assumption that broke: white reads correctly on one mode's tone and
    // not the other, in opposite directions for the two families.
    expect(contrastRatio('#ffffff', tone('--flow-danger-fg').dark)).toBeLessThan(4.5);
    expect(contrastRatio('#ffffff', tone('--flow-warn-fg').dark)).toBeLessThan(4.5);
    // And near-black is no better as a blanket answer, which is why these are
    // per-mode rather than one value.
    expect(contrastRatio('#0a0a0a', tone('--flow-danger-fg').light)).toBeLessThan(4.5);
    expect(contrastRatio('#0a0a0a', tone('--flow-warn-fg').light)).toBeLessThan(4.5);
  });

  it('agrees with the derivation the app uses for accent fills', () => {
    // Not a second opinion, the same one. If these ever diverge, one of the two
    // is being maintained by hand and should not be.
    for (const family of ['danger', 'warn'] as const) {
      const fill = tone(`--flow-${family}-fg`);
      const fg = tone(`--flow-${family}-on-fill`);
      for (const mode of MODES) {
        expect(fg[mode], `--flow-${family}-on-fill ${mode}`).toBe(deriveOnAccent(fill[mode]));
      }
    }
  });

  it('keeps each fill distinguishable from the surface behind it', () => {
    // A filled control needs 3:1 against its own background, or it stops reading
    // as a control. This is the axis that rules out simply darkening the tone
    // until white passes: the two requirements pull opposite ways.
    const failures: string[] = [];

    for (const family of ['danger', 'warn'] as const) {
      const fill = tone(`--flow-${family}-fg`);
      for (const mode of MODES) {
        const ratio = contrastRatio(fill[mode], CARD[mode]);
        if (ratio < 3) {
          failures.push(
            `${family} ${mode}: ${fill[mode]} on ${CARD[mode]} is ${ratio.toFixed(2)}:1`
          );
        }
      }
    }

    expect(failures, 'a filled control must clear 3:1 against its surface').toEqual([]);
  });

  it('leaves no filled control painting its own foreground literally', () => {
    const SRC = path.resolve(__dirname, '../..');
    const walk = (dir: string): string[] =>
      fs.readdirSync(dir, { withFileTypes: true }).flatMap((e) => {
        const full = path.join(dir, e.name);
        if (e.isDirectory()) {
          return e.name === 'node_modules' || e.name === 'locales' ? [] : walk(full);
        }
        return /\.tsx?$/.test(e.name) && !/\.test\.tsx?$/.test(e.name) ? [full] : [];
      });

    const offenders: string[] = [];
    for (const file of walk(SRC)) {
      const lines = fs.readFileSync(file, 'utf8').split('\n');
      lines.forEach((line, i) => {
        const trimmed = line.trim();
        if (trimmed.startsWith('*') || trimmed.startsWith('//')) return;
        if (!/background:\s*'?var\(--flow-(danger|warn)-(fg|fill)\)/.test(line)) return;
        const window = lines.slice(i, i + 3).join('\n');
        if (/color:\s*'?#[0-9a-fA-F]{3,6}'?/.test(window)) {
          offenders.push(`${path.relative(SRC, file)}:${i + 1}`);
        }
      });
    }

    expect(offenders, 'use var(--flow-<family>-on-fill) on a danger or warn fill').toEqual([]);
  });
});
