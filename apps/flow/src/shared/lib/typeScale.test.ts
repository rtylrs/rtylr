import fs from 'node:fs';
import path from 'node:path';
import { describe, expect, it } from 'vitest';

/**
 * Flow's type sizes come from the scale, not from arbitrary values.
 *
 * globals.css declares five steps in an `@theme` block, named for the roles
 * design-review-checklist.md G-T2 uses: screen 20, title 16, section 14,
 * body 13, meta 12. There is deliberately no token below 12px, so the only
 * way to put a 10px label back is an arbitrary value, and that is what this
 * test refuses.
 *
 * It scans source rather than computed styles, so it cannot see a size set
 * through an inline `style` prop or a raw CSS rule. Those are checked
 * separately below, for font-size only.
 */
const SRC = path.resolve(__dirname, '../..');
const GLOBALS = path.join(SRC, 'globals.css');

const SCALE = ['screen', 'title', 'section', 'body', 'meta'] as const;

// `text-[13px]`, with or without a variant prefix such as `md:`.
const ARBITRARY_TEXT_SIZE = /\btext-\[[0-9.]+(px|rem|em|pt)\]/g;
// An inline font-size declaration, value included. Whether the value is a
// scale token is decided in JS below rather than by a lookahead: a lookahead
// after `\s*` can backtrack to a position where it trivially succeeds, so it
// matched the compliant sites too.
const INLINE_FONT_SIZE = /fontSize\s*:[^,\n}]*/g;

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

function offenders(pattern: RegExp): string[] {
  const found: string[] = [];
  for (const file of sourceFiles(SRC)) {
    const lines = fs.readFileSync(file, 'utf8').split('\n');
    lines.forEach((line, index) => {
      for (const match of line.matchAll(pattern)) {
        found.push(`${path.relative(SRC, file)}:${index + 1}  ${match[0]}`);
      }
    });
  }
  return found;
}

describe('Flow type scale', () => {
  const css = fs.readFileSync(GLOBALS, 'utf8');

  it('declares every step of the scale as a theme token', () => {
    for (const step of SCALE) {
      expect(css, `--text-${step}`).toContain(`--text-${step}:`);
      expect(css, `--text-${step}--line-height`).toContain(`--text-${step}--line-height:`);
    }
  });

  it('declares no step below 12px', () => {
    const sizes = [...css.matchAll(/--text-(screen|title|section|body|meta):\s*([0-9.]+)px/g)].map(
      (match) => Number(match[2])
    );
    expect(sizes).toHaveLength(SCALE.length);
    expect(Math.min(...sizes)).toBe(12);
  });

  it('uses no arbitrary font size anywhere in the app', () => {
    expect(offenders(ARBITRARY_TEXT_SIZE)).toEqual([]);
  });

  it('sets no font size through an inline style prop, bar one documented case', () => {
    // An inline fontSize bypasses the scale. The one exception is the avatar
    // monogram, which is sized from the circle's diameter: at the sizes in
    // use a 12px floor overflows two characters, and reaching 12px
    // proportionally needs a 30px avatar, which changes row height. Listing
    // it here rather than dropping the rule means a second one cannot appear
    // without this test failing and someone justifying it.
    const ALLOWED = ['shared/components/AssigneeAvatar.tsx'];
    const found = offenders(INLINE_FONT_SIZE)
      .filter((entry) => !entry.includes('var(--text-'))
      .filter((entry) => !ALLOWED.some((allowed) => entry.startsWith(allowed)));
    expect(found).toEqual([]);
  });

  it('keeps the allowed inline font size above the old nine pixel floor', () => {
    const source = fs.readFileSync(path.join(SRC, 'shared/components/AssigneeAvatar.tsx'), 'utf8');
    const floor = source.match(/fontSize:\s*Math\.max\((\d+),/);
    expect(floor, 'the avatar monogram should still floor its size').not.toBeNull();
    expect(Number(floor?.[1])).toBeGreaterThanOrEqual(11);
  });

  it('actually uses the scale, so a passing run means something', () => {
    let used = 0;
    for (const file of sourceFiles(SRC)) {
      const source = fs.readFileSync(file, 'utf8');
      for (const step of SCALE) {
        used += source.split(`text-${step}`).length - 1;
      }
    }
    expect(used).toBeGreaterThan(500);
  });
});
