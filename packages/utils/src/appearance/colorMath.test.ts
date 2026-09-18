/**
 * The suite is the deliverable as much as the code. Every number asserted here
 * was measured by running it, not copied from the design document: where the two
 * disagreed the measurement won and the disagreement is noted in the test.
 *
 * Mutation probe (repo convention, and required by the design's section 8):
 * forcing the `>= CONTRAST_TARGET` gate in deriveAccentFg to a constant must
 * fail the suite in both directions. Measured results:
 *   - `true`  : the walk never runs and --accent-fg is the raw accent. 56 of 112
 *               combinations fall below the floor, worst 1.0625:1. Four tests
 *               fail, including the sweep's `failures` assertion.
 *   - `false` : the walk always exhausts and returns pure black or white. Every
 *               floor assertion still passes, because black on a dark card is
 *               16.19:1. The hue-preservation test below is what fails, at 172.6
 *               degrees of drift, along with the two shipped-pair oracles.
 * Neither predicate survives. Note which test does the catching in each case:
 * the floor sweep is blind to a permanently-false gate, so the hue test is
 * load-bearing rather than cosmetic.
 */

import { describe, it, expect } from 'vitest';
import {
  contrastRatio,
  deriveAccentFg,
  deriveOnAccent,
  hexLuminance,
  hexToHsl,
  hexToRgb,
  hslToHex,
  isValidHex,
  NEAR_BLACK,
  NEAR_WHITE,
  normalizeHex,
  rgbaFromHex,
  rgbaFromTriple,
  rgbToHex,
  safeHex,
} from './colorMath';
import { ACCENT_SWATCHES, PALETTES, STRESS_ACCENTS } from './palettes';
import { CONTRAST_FLOOR, CONTRAST_TARGET } from './types';

/** The 14 fixtures: ten curated swatches plus the four deliberately dark hexes. */
const ALL_ACCENTS: readonly string[] = [
  ...ACCENT_SWATCHES.map((swatch) => swatch.hex),
  ...STRESS_ACCENTS,
];

/** Smallest angular distance between two hues, in degrees. */
function hueDelta(a: number, b: number): number {
  const raw = Math.abs(a - b);
  return Math.min(raw, 360 - raw);
}

// ── Conversion and measurement ──────────────────────────────────────

describe('hex parsing', () => {
  it('accepts six digits with or without a leading hash, and nothing else', () => {
    expect(isValidHex('#64c5b9')).toBe(true);
    expect(isValidHex('64c5b9')).toBe(true);
    expect(isValidHex('#FFF')).toBe(false);
    expect(isValidHex('#ggghhh')).toBe(false);
    expect(isValidHex('')).toBe(false);
  });

  it('normalizes to lowercase with a hash', () => {
    expect(normalizeHex('64C5B9')).toBe('#64c5b9');
    expect(normalizeHex('#64C5B9')).toBe('#64c5b9');
  });

  it('falls back rather than propagating an unparseable colour', () => {
    expect(safeHex('nonsense', '#64c5b9')).toBe('#64c5b9');
    expect(safeHex(null, '#64c5b9')).toBe('#64c5b9');
    expect(safeHex(undefined, '#64c5b9')).toBe('#64c5b9');
    expect(safeHex('AABBCC', '#64c5b9')).toBe('#aabbcc');
  });

  it('returns black channels for an unparseable hex instead of NaN', () => {
    // NaN channels would poison every ratio downstream, turning one bad input
    // into an unreadable screen rather than a wrong-but-legible one.
    expect(hexToRgb('#ggg')).toEqual({ r: 0, g: 0, b: 0 });
    expect(Number.isFinite(hexLuminance('nonsense'))).toBe(true);
  });
});

describe('hsl round trip', () => {
  it('survives a round trip for every accent fixture', () => {
    for (const accent of ALL_ACCENTS) {
      const round = hslToHex(hexToHsl(accent));
      const before = hexToRgb(accent);
      const after = hexToRgb(round);
      // One 8-bit unit of slack: HSL is a lossy intermediate representation.
      expect(Math.abs(after.r - before.r)).toBeLessThanOrEqual(1);
      expect(Math.abs(after.g - before.g)).toBeLessThanOrEqual(1);
      expect(Math.abs(after.b - before.b)).toBeLessThanOrEqual(1);
    }
  });

  it('treats a grey as achromatic rather than guessing a hue', () => {
    expect(hexToHsl('#808080').s).toBe(0);
    expect(rgbToHex({ r: 128, g: 128, b: 128 })).toBe('#808080');
  });
});

describe('WCAG measurement', () => {
  it('reproduces the published anchor values', () => {
    expect(hexLuminance('#000000')).toBeCloseTo(0, 10);
    expect(hexLuminance('#ffffff')).toBeCloseTo(1, 10);
    expect(contrastRatio('#000000', '#ffffff')).toBeCloseTo(21, 6);
    expect(contrastRatio('#ffffff', '#ffffff')).toBeCloseTo(1, 10);
  });

  it('is order-free', () => {
    expect(contrastRatio('#64c5b9', '#161616')).toBeCloseTo(
      contrastRatio('#161616', '#64c5b9'),
      10
    );
  });
});

// ── deriveAccentFg ──────────────────────────────────────────────────

describe('deriveAccentFg', () => {
  it("leaves Flow's shipped teal on the dark card untouched, at the measured 8.81:1", () => {
    // Section 2.1 of the design: the hardcoded teal holds 8.81:1 on --flow-card
    // in dark mode. Measured here at 8.8145, so the design's figure is right.
    // An accent that already clears the target must come back unchanged, which
    // is what keeps dark mode pixel-identical to what ships today.
    expect(contrastRatio('#64c5b9', '#161616')).toBeCloseTo(8.8145, 3);
    expect(deriveAccentFg('#64c5b9', '#161616')).toBe('#64c5b9');
  });

  it("reproduces Flow's shipped light foreground closely enough to be a regression oracle", () => {
    // The shipped light value is #1f7d72 at 4.9590:1 on white. The walk stops at
    // the first value that clears 4.6, so it lands lighter: #308177 at 4.6336.
    // Same hue, one stop brighter. Both clear AA, which is what the oracle is
    // for. Asserting equality with #1f7d72 would be asserting that a hand-picked
    // value and a first-crossing walk agree exactly, which they cannot.
    const derived = deriveAccentFg('#64c5b9', '#ffffff');
    expect(contrastRatio('#1f7d72', '#ffffff')).toBeCloseTo(4.959, 2);
    expect(contrastRatio(derived, '#ffffff')).toBeGreaterThanOrEqual(CONTRAST_TARGET);
    expect(hueDelta(hexToHsl(derived).h, hexToHsl('#1f7d72').h)).toBeLessThan(6);
    // The raw accent on white is 2.05:1, so the walk is doing real work here.
    expect(contrastRatio('#64c5b9', '#ffffff')).toBeLessThan(CONTRAST_FLOOR);
  });

  it('preserves hue across all 112 combinations, so the fix is not "make everything grey"', () => {
    // This is the assertion that distinguishes a derived accent from a safe
    // grey. It is also the one that catches a permanently-false contrast gate,
    // which would return pure black or white and destroy the hue entirely.
    let maxDelta = 0;
    for (const palette of PALETTES) {
      for (const accent of ALL_ACCENTS) {
        const fg = deriveAccentFg(accent, palette.card);
        const delta = hueDelta(hexToHsl(accent).h, hexToHsl(fg).h);
        maxDelta = Math.max(maxDelta, delta);
        expect(delta).toBeLessThan(3);
        // Saturation is held too: a desaturated result reads as grey even at
        // the right hue.
        expect(hexToHsl(fg).s).toBeGreaterThan(hexToHsl(accent).s * 0.8);
      }
    }
    // Measured worst case: 1.008 degrees, which is 8-bit rounding, not drift.
    expect(maxDelta).toBeLessThan(1.5);
  });

  it('falls back to the best extreme when no tint of the hue can clear the target', () => {
    // #006cff is inside the narrow band where neither pure black (4.589:1) nor
    // pure white (4.576:1) reaches 4.6, so the walk must exhaust. It is not a
    // palette value; it exists to prove the bound is real rather than dead code.
    expect(contrastRatio('#000000', '#006cff')).toBeLessThan(CONTRAST_TARGET);
    expect(contrastRatio('#ffffff', '#006cff')).toBeLessThan(CONTRAST_TARGET);
    const fg = deriveAccentFg('#a3d84b', '#006cff');
    expect([NEAR_WHITE, '#000000']).toContain(fg);
    // Black wins that comparison, and still clears the 4.5 floor.
    expect(fg).toBe('#000000');
    expect(contrastRatio(fg, '#006cff')).toBeGreaterThanOrEqual(CONTRAST_FLOOR);
  });

  it('never throws on malformed input', () => {
    for (const bad of ['', 'nonsense', '#ggg', '#abc', '64c5b9']) {
      expect(() => deriveAccentFg(bad, '#161616')).not.toThrow();
      expect(() => deriveAccentFg('#64c5b9', bad)).not.toThrow();
      expect(isValidHex(deriveAccentFg(bad, '#161616'))).toBe(true);
    }
  });
});

// ── deriveOnAccent ──────────────────────────────────────────────────

describe('deriveOnAccent', () => {
  it('keeps near-black where near-black is right, for every curated swatch', () => {
    // All ten curated swatches are light enough that near-black is the better
    // choice. This is the design's section 2.2 point restated as a test: a
    // curated-only product would have no contrast bug to fix.
    for (const swatch of ACCENT_SWATCHES) {
      expect(deriveOnAccent(swatch.hex)).toBe(NEAR_BLACK);
      expect(contrastRatio(NEAR_BLACK, swatch.hex)).toBeGreaterThanOrEqual(CONTRAST_FLOOR);
    }
  });

  it('fixes the unreadable-text bug that hardcoded near-black ships on a dark accent', () => {
    // This is the bug the whole feature exists to prevent. Hardcoded #0a0a0a on
    // these four measures 1.299:1 to 1.628:1, which is unreadable rather than
    // marginal. The design says 1.30 to 1.63; measured, it is 1.299 to 1.628.
    const measured: Record<string, number> = {};
    for (const accent of STRESS_ACCENTS) {
      const hardcoded = contrastRatio(NEAR_BLACK, accent);
      measured[accent] = hardcoded;

      // What ships today, and why it is a bug and not a rounding complaint.
      expect(hardcoded).toBeGreaterThan(1.25);
      expect(hardcoded).toBeLessThan(1.7);
      expect(hardcoded).toBeLessThan(CONTRAST_FLOOR);

      // What the derivation does about it.
      const derived = deriveOnAccent(accent);
      expect(derived).toBe(NEAR_WHITE);
      expect(contrastRatio(derived, accent)).toBeGreaterThanOrEqual(CONTRAST_FLOOR);
      expect(contrastRatio(derived, accent)).toBeGreaterThan(hardcoded);
    }
    expect(Math.min(...Object.values(measured))).toBeCloseTo(1.299, 2);
    expect(Math.max(...Object.values(measured))).toBeCloseTo(1.628, 2);
  });

  it('clears the floor for every fixture, curated and stress alike', () => {
    for (const accent of ALL_ACCENTS) {
      expect(contrastRatio(deriveOnAccent(accent), accent)).toBeGreaterThanOrEqual(CONTRAST_FLOOR);
    }
  });

  it('never throws on malformed input', () => {
    for (const bad of ['', 'nonsense', '#ggg', '#abc', '64c5b9']) {
      expect(() => deriveOnAccent(bad)).not.toThrow();
      expect([NEAR_BLACK, NEAR_WHITE]).toContain(deriveOnAccent(bad));
    }
  });
});

// ── rgba emitters ───────────────────────────────────────────────────

describe('rgba emitters', () => {
  it('matches the string shape the existing accentTints produces', () => {
    // Format parity matters: these values land in inline custom properties that
    // a snapshot or a diff against the current app will be compared to.
    expect(rgbaFromHex('#64c5b9', 0.12)).toBe('rgba(100, 197, 185, 0.12)');
    expect(rgbaFromTriple('240,240,238', 0.58)).toBe('rgba(240, 240, 238, 0.58)');
  });

  it('degrades to black rather than emitting NaN', () => {
    expect(rgbaFromTriple('broken', 0.5)).toBe('rgba(0, 0, 0, 0.5)');
    expect(rgbaFromHex('#ggg', 0.5)).toBe('rgba(0, 0, 0, 0.5)');
  });
});

// ── The floor, at function level ────────────────────────────────────

describe('the target leaves a margin over the floor', () => {
  it('is aimed above the bar it is tested against', () => {
    // Section 2.4: stopping at 4.5 produced exactly 4.50:1 on every light
    // palette, leaving nothing for rounding or a later surface-token change.
    expect(CONTRAST_TARGET).toBeGreaterThan(CONTRAST_FLOOR);
  });
});
