import fs from 'node:fs';
import path from 'node:path';
import { describe, expect, it } from 'vitest';

/**
 * Every interactive element gets Flow's focus ring, without opting in.
 *
 * `globals.css` called `.flow-focus` "the one focus treatment" and it was on
 * 21 of the app's 317 interactive elements. An opt-in class meant the ring was
 * missing far more often than present, and a new component started out
 * missing it, so the rule now selects the elements themselves.
 *
 * 53 `outline-none` classes were the other half. Tailwind's utility beats an
 * element selector on specificity, and most of those sites paired it with no
 * focus style at all, so those controls had no focus indicator whatsoever.
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

describe('focus ring', () => {
  const css = fs.readFileSync(GLOBALS, 'utf8');

  it('applies to the interactive elements themselves, not only to a class', () => {
    expect(css).toMatch(/:is\(button,\s*a,\s*input,\s*select,\s*textarea/);
    expect(css).toMatch(/:focus-visible\s*\{[^}]*outline:\s*2px solid var\(--flow-focus-ring\)/);
  });

  it('declares each focus rule exactly once', () => {
    // The block was duplicated verbatim in this file, comment included.
    expect([...css.matchAll(/\.flow-focus:focus-visible\s*\{/g)]).toHaveLength(1);
    expect([...css.matchAll(/\.flow-focus-inset:focus-visible\s*\{/g)]).toHaveLength(1);
  });

  it('offers an inset variant for clipped containers', () => {
    // A ring at a positive offset is cut off inside overflow:hidden, which the
    // sidebar drawer needs for its slide transition.
    const inset = css.match(/\.flow-focus-inset:focus-visible\s*\{([^}]*)\}/);
    expect(inset).not.toBeNull();
    expect(inset?.[1]).toContain('outline-offset: -2px');
  });

  it('puts the inset variant on the sidebar rows that are clipped', () => {
    const nav = fs.readFileSync(path.join(SRC, 'shared/components/SidebarNavigation.tsx'), 'utf8');
    expect(nav).toContain('flow-focus-inset');
  });

  it('suppresses the outline nowhere in the app', () => {
    const offenders: string[] = [];
    for (const file of sourceFiles(SRC)) {
      const lines = fs.readFileSync(file, 'utf8').split('\n');
      lines.forEach((line, index) => {
        if (!/\boutline-none\b|outline:\s*'?none'?/.test(line)) return;
        offenders.push(`${path.relative(SRC, file)}:${index + 1}`);
      });
    }
    expect(offenders).toEqual([]);
  });

  it('has interactive elements to protect, so a passing run means something', () => {
    let interactive = 0;
    for (const file of sourceFiles(SRC)) {
      const source = fs.readFileSync(file, 'utf8');
      interactive += [...source.matchAll(/<(button|a|input|select|textarea)\b/g)].length;
    }
    expect(interactive).toBeGreaterThan(250);
  });
});
