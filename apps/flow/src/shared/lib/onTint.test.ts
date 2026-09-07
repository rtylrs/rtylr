import { describe, it, expect } from 'vitest';
import { contrastRatio, compositeOver } from '@voxire/utils';
import fs from 'node:fs';
import path from 'node:path';
import { onTint, PROJECT_COLORS } from './colorUtils';

/**
 * Guards the one number in `onTint`.
 *
 * A chip's background is a tint of its own colour, so using that colour as the
 * foreground put 16 of 20 sampled pairings under AA at 10% alpha and 17 of 20 at
 * 13%, the worst being Flow's teal at 1.92:1 in light mode (rtylr#1246).
 *
 * This cannot be a rendering test: happy-dom resolves neither `color-mix` nor
 * custom properties, so a computed style would tell us nothing. It reproduces
 * the mix arithmetically and measures it with the repo's own colour maths.
 */
const COLOR_SHARE = 0.5;

/** `--flow-text` as it actually lands, composited over each mode's card. */
const THEME_TEXT = {
  dark: compositeOver('#f0f0ee', 0.92, '#161616'),
  light: '#09090b',
};
const CARD = { dark: '#161616', light: '#ffffff' };
const MODES = ['dark', 'light'] as const;

/** Every alpha the chips actually pass to withAlpha, plus the ends of the range. */
const ALPHAS = [0.1, 0.13, 0.15, 0.2];

const hexToRgb = (hex: string): number[] => [1, 3, 5].map((i) => parseInt(hex.slice(i, i + 2), 16));
const rgbToHex = (rgb: number[]): string =>
  '#' + rgb.map((v) => Math.round(v).toString(16).padStart(2, '0')).join('');
const mix = (a: string, b: string, shareOfA: number): string =>
  rgbToHex(hexToRgb(a).map((v, i) => v * shareOfA + hexToRgb(b)[i] * (1 - shareOfA)));

/** Flow's own palette, its accent, and the mid-tones a free-form picker allows. */
const COLORS = [...PROJECT_COLORS, '#64c5b9', '#4583c4', '#a855f7', '#f59e0b', '#7f7f7f'];

const pairings = (share: number) =>
  COLORS.flatMap((color) =>
    MODES.flatMap((mode) =>
      ALPHAS.map((alpha) => ({
        color,
        mode,
        alpha,
        tint: compositeOver(color, alpha, CARD[mode]),
        fg: mix(color, THEME_TEXT[mode], share),
      }))
    )
  );

describe('the foreground for a colour on its own tint', () => {
  it('mixes toward the theme text rather than using the colour itself', () => {
    // If this stops holding, the arithmetic below measures something other than
    // what the components render.
    expect(onTint('#64c5b9')).toBe('color-mix(in srgb, #64c5b9 50%, var(--flow-text))');
  });

  it('measures why: the raw colour fails on its own tint in light mode', () => {
    // The reason the helper exists, asserted rather than described.
    for (const color of COLORS) {
      for (const alpha of ALPHAS) {
        const tint = compositeOver(color, alpha, CARD.light);
        expect(
          contrastRatio(color, tint),
          `${color} raw on its own ${alpha * 100}% light tint`
        ).toBeLessThan(4.5);
      }
    }
  });

  it('clears AA for every colour, mode and alpha the chips use', () => {
    const failures = pairings(COLOR_SHARE)
      .filter((p) => contrastRatio(p.fg, p.tint) < 4.5)
      .map(
        (p) =>
          `${p.color} ${p.mode} @${p.alpha * 100}%: ${contrastRatio(p.fg, p.tint).toFixed(2)}:1`
      );

    expect(failures, 'a chip must be readable on its own tint').toEqual([]);
  });

  it('keeps real headroom, so the ratio is not sitting on the minimum', () => {
    // Worst pairing across all 120 is yellow #eab308 in light mode at 20% alpha,
    // at 5.38:1. Asserted as a margin rather than a bare pass, because a value
    // derived to its own minimum is brittle by construction.
    const worst = Math.min(...pairings(COLOR_SHARE).map((p) => contrastRatio(p.fg, p.tint)));
    expect(worst).toBeGreaterThan(5.3);
  });

  it('shows that a larger share of the colour would fail outright', () => {
    // Why 50% and not 60%: on the full palette 60% is not merely thin, it drops
    // the same yellow pairing to 4.14:1 and below AA. An earlier measurement put
    // 60% at 4.61:1, but that used ten colours and did not include this one.
    const at60 = pairings(0.6).map((p) => contrastRatio(p.fg, p.tint));
    expect(Math.min(...at60), '60% of the colour fails AA').toBeLessThan(4.5);
  });

  it('leaves no source painting a colour on a tint of itself', () => {
    // Written because sweeping for this by hand missed sites twice. The first
    // pass matched only `color: <x>.color` and walked past both calendars, which
    // hold the value in a `tone` variable; 7 sites were fixed and 6 were left.
    //
    // The shape is a withAlpha() background whose first argument is used, bare,
    // as the foreground within the same style object. A fallback beside it is
    // fine: `color: x ? onTint(x) : 'var(--accent-on-muted)'` is the fixed form,
    // and the token half was already correct from rtylr#1241.
    const SRC = path.resolve(__dirname, '../..');
    const offenders: string[] = [];

    const walk = (dir: string): string[] =>
      fs.readdirSync(dir, { withFileTypes: true }).flatMap((e) => {
        const full = path.join(dir, e.name);
        if (e.isDirectory()) {
          return e.name === 'node_modules' || e.name === 'locales' ? [] : walk(full);
        }
        return /\.tsx?$/.test(e.name) && !/\.test\.tsx?$/.test(e.name) ? [full] : [];
      });

    for (const file of walk(SRC)) {
      const lines = fs.readFileSync(file, 'utf8').split('\n');
      lines.forEach((line, i) => {
        const call = /withAlpha\(\s*([A-Za-z_$][\w$.?[\]' ]*?)\s*,/.exec(line);
        if (!call) return;
        const expr = call[1].trim();
        // Wide enough to reach a foreground set on a child element rather than
        // in the same style object: Timeline's bar is tinted on one line and its
        // label coloured five lines later, which an i+5 window missed.
        const window = lines.slice(Math.max(0, i - 4), i + 10).join('\n');
        if (window.includes('onTint(')) return;
        // `color: <expr>` with nothing between, i.e. the raw colour as text.
        const bare = new RegExp(
          `color:\\s*${expr.replace(/[.*+?^${}()|[\]\\]/g, '\\$&')}\\s*[,}?]`
        );
        if (bare.test(window)) {
          offenders.push(
            `${path.relative(SRC, file)}:${i + 1} (withAlpha(${expr}) with bare ${expr})`
          );
        }
      });
    }

    expect(offenders, 'wrap the foreground in onTint()').toEqual([]);
  });
});
