import fs from 'node:fs';
import path from 'node:path';
import { describe, expect, it } from 'vitest';
import { compositeOver, contrastRatio } from '@voxire/utils';

/**
 * A control's boundary has to be visible; a card's edge does not.
 *
 * WCAG 2.1 SC 1.4.11 asks 3:1 for the visual boundary of a user interface
 * component. `--flow-border` is 1.20:1 and stays that way: it is the
 * decorative hairline for dividers and card edges, where no rule applies.
 * `--flow-border-control` is the line for anything whose shape is its only
 * affordance, and this checks it against every ground it lands on rather than
 * against one.
 */
const GLOBALS = path.resolve(__dirname, '../../globals.css');
const SRC = path.resolve(__dirname, '../..');
const NON_TEXT_CONTRAST = 3;

// The grounds a control border is actually painted on, from globals.css.
const GROUNDS = {
  dark: { ink: '#f0f0ee', on: ['#161616', '#111111', '#0a0a0a'] },
  light: { ink: '#09090b', on: ['#ffffff', '#f4f4f5'] },
} as const;

function alphaOf(css: string, mode: 'dark' | 'light'): number {
  // The dark declaration comes first in the file, light second.
  const all = [...css.matchAll(/--flow-border-control:\s*rgba\([^)]*?,\s*([0-9.]+)\)/g)];
  expect(all.length, 'two --flow-border-control declarations').toBe(2);
  return Number(all[mode === 'dark' ? 0 : 1][1]);
}

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

describe('control border contrast', () => {
  const css = fs.readFileSync(GLOBALS, 'utf8');

  it('declares the control border in both modes', () => {
    expect(css).toContain('--flow-border-control:');
    expect([...css.matchAll(/--flow-border-control:/g)]).toHaveLength(2);
  });

  it.each(['dark', 'light'] as const)('clears 3:1 on every %s ground', (mode) => {
    const alpha = alphaOf(css, mode);
    const { ink, on } = GROUNDS[mode];
    for (const ground of on) {
      const solid = compositeOver(ink, alpha, ground);
      expect(
        contrastRatio(solid, ground),
        `${ink} at ${alpha} on ${ground} resolves to ${solid}`
      ).toBeGreaterThanOrEqual(NON_TEXT_CONTRAST);
    }
  });

  it('leaves the decorative border decorative, and below 3:1', () => {
    // Not an oversight to be fixed later. Raising this token would redraw
    // every card edge in the product, which is why the control line is a
    // second token. If someone "fixes" this one, this test says why not to.
    expect(contrastRatio(compositeOver('#ffffff', 0.07, '#161616'), '#161616')).toBeLessThan(
      NON_TEXT_CONTRAST
    );
    expect(css).toContain('--flow-border: rgba(255, 255, 255, 0.07)');
    expect(css).toContain('--flow-border: rgba(0, 0, 0, 0.08)');
  });

  it('never borders an input-background field with the decorative token', () => {
    // A style object that sets the input background is a field by
    // construction, so its border is a component boundary. This is the one
    // mechanical rule available: the rest of the call sites need judgement
    // about whether a border is a control edge or a card edge.
    const offenders: string[] = [];
    const styleObject = /\{[^{}]*var\(--flow-input-bg\)[^{}]*\}/gs;
    for (const file of sourceFiles(SRC)) {
      const source = fs.readFileSync(file, 'utf8');
      for (const match of source.matchAll(styleObject)) {
        if (!/var\(--flow-border\)/.test(match[0])) continue;
        offenders.push(path.relative(SRC, file));
      }
    }
    expect([...new Set(offenders)]).toEqual([]);
  });

  it('actually uses the control token, so a passing run means something', () => {
    let used = 0;
    for (const file of sourceFiles(SRC)) {
      used += fs.readFileSync(file, 'utf8').split('var(--flow-border-control)').length - 1;
    }
    expect(used).toBeGreaterThan(40);
  });
});
