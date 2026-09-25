import fs from 'node:fs';
import path from 'node:path';
import { describe, expect, it } from 'vitest';

/**
 * The card lift comes from a token.
 *
 * `TaskCard` hardcoded `0 1px 3px rgba(0,0,0,0.15)`, the only shadow in the
 * app from outside the token set, and it could not change with the theme:
 * that weight reads as a heavy drop on white and as almost nothing on the
 * dark ground, where the card surface is already darker than the page.
 *
 * `--flow-shadow` is not the same thing. It is the modal and popover shadow
 * at 8px/32px, so reusing it would have put a dialog's lift on every card.
 */
const SRC = path.resolve(__dirname, '../..');
const GLOBALS = path.join(SRC, 'globals.css');

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

describe('card shadow', () => {
  const css = fs.readFileSync(GLOBALS, 'utf8');

  it('declares the card shadow in both modes', () => {
    expect([...css.matchAll(/--flow-shadow-card:/g)]).toHaveLength(2);
  });

  it('is not the same value as the modal shadow', () => {
    const card = [...css.matchAll(/--flow-shadow-card:\s*([^;]+);/g)].map((m) => m[1].trim());
    const modal = [...css.matchAll(/--flow-shadow:\s*([^;]+);/g)].map((m) => m[1].trim());
    expect(card).toHaveLength(2);
    for (const value of card) expect(modal).not.toContain(value);
  });

  it('carries a lighter weight in light mode than in dark', () => {
    // A card on white needs a hint of lift; the same alpha that reads as a
    // hint on a dark ground reads as a heavy drop on white.
    const alphas = [...css.matchAll(/--flow-shadow-card:[^;]*rgba\([^)]*,\s*([0-9.]+)\)/g)].map(
      (m) => Number(m[1])
    );
    expect(alphas).toHaveLength(2);
    const [dark, light] = alphas;
    expect(light).toBeLessThan(dark);
  });

  it('the card reads the token rather than a literal', () => {
    const card = fs.readFileSync(path.join(SRC, 'shared/components/TaskCard.tsx'), 'utf8');
    expect(card).toContain('var(--flow-shadow-card)');
  });

  it('no component hardcodes a drop shadow', () => {
    // An inset ring such as `0 0 0 3px var(--accent-muted)` is a focus or
    // selection treatment rather than a lift, and it already reads a token.
    const offenders: string[] = [];
    for (const file of sourceFiles(SRC)) {
      const lines = fs.readFileSync(file, 'utf8').split('\n');
      lines.forEach((line, index) => {
        const match = line.match(/boxShadow:\s*'([^']+)'/);
        if (!match) return;
        const value = match[1];
        if (value.includes('var(--')) return;
        offenders.push(`${path.relative(SRC, file)}:${index + 1}  ${value}`);
      });
    }
    expect(offenders).toEqual([]);
  });
});
