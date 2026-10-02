import { describe, it, expect } from 'vitest';
import fs from 'node:fs';
import path from 'node:path';

/**
 * A motion element must not set `opacity` in `style` while also animating it.
 *
 * The two fight, and the animated value loses. On `TaskCard` the style wrote
 * `opacity: dimmed ? 0.5 : undefined` beside an animated opacity, and the value
 * stayed pinned at whatever `initial` set: **every card on the board rendered at
 * opacity 0**, so the columns, the counts and the empty states all painted
 * around invisible cards until some later re-render shook them loose
 * (rtylr#1078).
 *
 * Measured in headless and headed Chromium, on a dev server and on a production
 * build. Not a headless artefact, which is what `flow.spec.ts` had assumed when
 * it weakened its own assertion to `toBeAttached`.
 *
 * `BulkEditBar` had the same conflict with `opacity: isPending ? 0.7 : 1`. That
 * one always resolved to a number, so the bar stayed visible and only its
 * entrance animation was dead, which is why nobody noticed.
 *
 * This is a source check because no runtime here can see it: happy-dom runs no
 * animations at all, and Playwright's own visibility check ignores opacity.
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
    if (/\.tsx$/.test(entry.name) && !/\.test\.tsx$/.test(entry.name)) into.push(full);
  }
  return into;
}

/** Strips block and line comments, so a comment about opacity is not a finding. */
function stripComments(text: string): string {
  return text.replace(/\/\*[\s\S]*?\*\//g, '').replace(/^[ \t]*\/\/.*$/gm, '');
}

/**
 * The text inside a `style={{ ... }}` prop, matched by brace depth.
 *
 * A regex cannot do this. `style=\{\{[\s\S]*?\bopacity:` ran straight past the
 * style object's own closing braces and matched the `opacity` in the
 * `initial={{ ... }}` that followed, reporting three elements whose style holds
 * only a background and a border. Written down because that was the first
 * version of this guard.
 */
function styleObjects(text: string): string[] {
  const found: string[] = [];
  const marker = /style=\{\{/g;
  let match: RegExpExecArray | null;

  while ((match = marker.exec(text)) !== null) {
    let depth = 2;
    let i = match.index + match[0].length;
    const from = i;
    while (i < text.length && depth > 0) {
      if (text[i] === '{') depth += 1;
      else if (text[i] === '}') depth -= 1;
      i += 1;
    }
    found.push(text.slice(from, i));
  }
  return found;
}

describe('motion elements and opacity', () => {
  it('never animates opacity and sets it in style at the same time', () => {
    const offenders: string[] = [];

    for (const file of sourceFiles(SRC)) {
      // Scanned on the raw text so the reported line number is the real one.
      // Stripping the whole file first shifted every position by however many
      // comment lines came before, and the first version of this guard reported
      // a calendar cell twelve lines above where it actually is. Comments are
      // stripped per tag instead, below.
      const text = fs.readFileSync(file, 'utf8');
      if (!text.includes('motion.')) continue;

      for (const match of text.matchAll(/<motion\.\w+\b/g)) {
        // The opening tag, taken by brace and bracket depth rather than by
        // looking for the next `>`: an `onClick={(e) => ...}` contains one.
        let i = match.index + match[0].length;
        let curly = 0;
        let tag = '';
        while (i < text.length) {
          const ch = text[i];
          if (ch === '{') curly += 1;
          else if (ch === '}') curly -= 1;
          else if (ch === '>' && curly === 0) break;
          tag += ch;
          i += 1;
        }

        // Per tag, so a comment *about* opacity is not a finding. The first
        // version matched the word inside this very file's explanation of the
        // bug, on the element it had just been fixed on.
        const clean = stripComments(tag);
        const animates = /(?:initial|animate|exit|whileHover|whileTap)=\{\{[^}]*opacity/.test(
          clean
        );
        const styled = styleObjects(clean).some((style) => /\bopacity\s*:/.test(style));
        if (animates && styled) {
          const line = text.slice(0, match.index).split('\n').length;
          offenders.push(`${path.relative(SRC, file)}:${line}`);
        }
      }
    }

    expect(offenders, 'put the value in `animate`, not in `style`').toEqual([]);
  });
});
