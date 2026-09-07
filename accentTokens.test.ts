import { describe, it, expect } from 'vitest';
import fs from 'node:fs';
import path from 'node:path';
import { contrastRatio, compositeOver } from '@voxire/utils';

/**
 * `--accent` and `--accent-fg` are derived against the card surface. That
 * guarantee does not transfer to `--accent-muted`, which is a different
 * background, so `--accent-on-muted` exists and is derived against the tint
 * itself (rtylr#970, extended repo-wide in rtylr#1217).
 *
 * This lives at the root rather than once per app because the defect is not
 * app-specific and the earlier flow-only copy let the same 11 sites survive in
 * dash, insights and hr.
 *
 * The declaration test is here because of the way rtylr#1217 was first fixed:
 * every call site was repointed at `var(--accent-on-muted)` while only flow
 * declared it. A reference to an undeclared custom property is invalid at
 * computed-value time, so those elements silently inherited their parent
 * colour. Nothing caught it: the pairing test passed (the tint now had
 * `accent-on-muted` beside it), the hardcoded-hex guard passed (no hex), and
 * jsdom/happy-dom do not resolve custom properties, so no component test could
 * see it either.
 */
const ROOT = __dirname;

/**
 * Set on an element's own inline style rather than in a stylesheet, so a scan of
 * globals.css cannot see them. --pane-u is assigned in AppearancePreview itself,
 * a few lines above the rules that read it, and --tone is set per tinted element
 * so that one value feeds both the tint and its foreground (rtylr#1249).
 */
const RUNTIME_TOKENS = new Set(['--flow-r', '--flow-u', '--pane-u', '--tone']);
const APPS = fs
  .readdirSync(path.join(ROOT, 'apps'), { withFileTypes: true })
  .filter((e) => e.isDirectory() && fs.existsSync(path.join(ROOT, 'apps', e.name, 'src')))
  .map((e) => e.name)
  .sort();

function sourceFiles(dir: string, into: string[] = []): string[] {
  for (const entry of fs.readdirSync(dir, { withFileTypes: true })) {
    const full = path.join(dir, entry.name);
    if (entry.isDirectory()) {
      if (entry.name === 'node_modules' || entry.name === 'locales' || entry.name === 'dist') {
        continue;
      }
      sourceFiles(full, into);
      continue;
    }
    if (!/\.tsx?$/.test(entry.name)) continue;
    if (/\.test\.tsx?$/.test(entry.name)) continue;
    into.push(full);
  }
  return into;
}

const isComment = (line: string) => {
  const t = line.trim();
  return t.startsWith('*') || t.startsWith('//') || t.startsWith('/*');
};

describe('the accent tint is paired with its own foreground', () => {
  it('measures why: the accent always fails AA on its own tint in light mode', () => {
    // An assertion about the reason, not the code. If this ever passes the rule
    // below can be revisited rather than kept out of habit.
    const accents = ['#64c5b9', '#4583c4', '#8b5cf6', '#3b82f6', '#f59e0b', '#ef4444'];
    for (const accent of accents) {
      const tint = compositeOver(accent, 0.12, '#f4f4f5');
      expect(contrastRatio(accent, tint), `${accent} on its light tint`).toBeLessThan(4.5);
    }
  });

  it('measures why the same rule cannot be stated for dark mode', () => {
    // Dark is accent-dependent, so a blanket claim would be wrong. Flow's teal
    // passes on its own dark tint, which is why flow's token happens to equal
    // its accent in dark and had to be derived only for light. The two blues
    // and the purple do not, which is why dash, insights and hr needed a real
    // value in both modes.
    const tint = (accent: string) => compositeOver(accent, 0.12, '#161616');
    expect(contrastRatio('#64c5b9', tint('#64c5b9'))).toBeGreaterThanOrEqual(4.5);
    for (const accent of ['#4583c4', '#8b5cf6']) {
      expect(contrastRatio(accent, tint(accent)), `${accent} on its dark tint`).toBeLessThan(4.5);
    }
  });

  it('leaves no source pairing --accent with the muted tint', () => {
    const offenders: string[] = [];

    for (const app of APPS) {
      for (const file of sourceFiles(path.join(ROOT, 'apps', app, 'src'))) {
        const lines = fs.readFileSync(file, 'utf8').split('\n');
        lines.forEach((line, i) => {
          if (isComment(line)) return;
          if (!/color:\s*[^,\n]*var\(--accent(-fg)?\)/.test(line)) return;
          const window = lines.slice(Math.max(0, i - 3), i + 4).join('\n');
          if (!window.includes('accent-muted')) return;
          if (window.includes('accent-on-muted')) return;
          offenders.push(`${path.relative(ROOT, file)}:${i + 1}`);
        });
      }
    }

    expect(offenders, 'use var(--accent-on-muted) on an --accent-muted background').toEqual([]);
  });

  it('declares every accent token an app references, in both modes', () => {
    const missing: string[] = [];

    for (const app of APPS) {
      const src = path.join(ROOT, 'apps', app, 'src');
      const referenced = new Set<string>();

      for (const file of sourceFiles(src)) {
        for (const line of fs.readFileSync(file, 'utf8').split('\n')) {
          if (isComment(line)) continue;
          // A reference carrying a fallback -- var(--x, #fff) -- resolves to the
          // fallback rather than inheriting, so it is not this defect. It has its
          // own problem: the fallback cannot follow a user's chosen accent.
          for (const m of line.matchAll(/var\(\s*(--accent[a-z0-9-]*)\s*([,)])/gi)) {
            if (m[2] === ',') continue;
            referenced.add(m[1]);
          }
        }
      }
      if (referenced.size === 0) continue;

      const cssPath = path.join(src, 'globals.css');
      const css = fs.existsSync(cssPath) ? fs.readFileSync(cssPath, 'utf8') : '';

      // Walk the CSS tracking brace depth, so each declaration is attributed to
      // the selector stack it sits in. A flat /{...}/ regex cannot do this: the
      // files wrap rules in @media and @theme, and nested braces make the
      // selector capture meaningless.
      const contexts = new Map<string, string[]>();
      {
        const stack: string[] = [];
        let pending = '';
        for (const raw of css.split('\n')) {
          const line = raw.replace(/\/\*.*?\*\//g, '');
          for (const ch of line) {
            if (ch === '{') {
              stack.push(pending.trim());
              pending = '';
            } else if (ch === '}') {
              stack.pop();
              pending = '';
            } else {
              pending += ch;
            }
          }
          const decl = /^\s*(--[a-z0-9-]+)\s*:/i.exec(raw);
          if (decl) {
            const seen = contexts.get(decl[1]) ?? [];
            seen.push(stack.join(' | '));
            contexts.set(decl[1], seen);
            pending = '';
          }
        }
      }

      const where = (token: string) => contexts.get(token) ?? [];
      const declaresIn = (token: string, pred: (sel: string) => boolean) => where(token).some(pred);
      const isDark = (sel: string) => /\[data-theme='dark'\]/.test(sel);
      const isLight = (sel: string) => /\[data-theme='light'\]/.test(sel);
      const isBase = (sel: string) => !isDark(sel) && !isLight(sel);

      for (const token of [...referenced].sort()) {
        const dark = declaresIn(token, isDark);
        const light = declaresIn(token, isLight);
        const base = declaresIn(token, isBase);

        if (!dark && !light && !base) {
          missing.push(`${app}: ${token} referenced but never declared`);
          continue;
        }
        // A base :root declaration covers whichever mode does not override it,
        // which is how flow legitimately raises --accent-muted from 0.12 to 0.18
        // in light only. So a one-sided override is a defect only when there is
        // no base value behind it.
        if (base) continue;
        if (dark && !light) missing.push(`${app}: ${token} declared for dark but not light`);
        if (light && !dark) missing.push(`${app}: ${token} declared for light but not dark`);
      }
    }

    expect(missing, 'a var() on an undeclared property silently inherits').toEqual([]);
  });

  it('gives text on an accent fill a foreground that passes AA', () => {
    // The second defect the declaration test turned up. dash painted text on an
    // --accent fill with var(--accent-contrast, #fff): a property no stylesheet
    // declares, so it always took the white fallback, at 3.96:1 on #4583c4.
    // insights had the same white under the name --accent-fg, and HR's
    // --hr-on-accent was white too, at 4.23:1.
    //
    // Checked by value, not by name, because the name is not reliable: flow's
    // --accent-fg is the accent used AS a foreground, while insights had used
    // --accent-fg for text ON the fill. --<app>-on-accent is the name that says
    // which it means.
    const failures: string[] = [];

    for (const app of APPS) {
      const cssPath = path.join(ROOT, 'apps', app, 'src', 'globals.css');
      if (!fs.existsSync(cssPath)) continue;
      const css = fs.readFileSync(cssPath, 'utf8');

      const accent = /^\s*--accent:\s*(#[0-9a-f]{6})\s*;/im.exec(css)?.[1];
      if (!accent) continue;

      for (const m of css.matchAll(/^\s*(--[a-z-]*on-accent):\s*(#[0-9a-f]{6})\s*;/gim)) {
        const ratio = contrastRatio(m[2], accent);
        if (ratio < 4.5) {
          failures.push(`${app}: ${m[1]} ${m[2]} on ${accent} is ${ratio.toFixed(2)}:1`);
        }
      }
    }

    expect(failures, 'text on an accent fill must clear 4.5:1').toEqual([]);
  });

  it('leaves no accent fill painted with a literal foreground', () => {
    // Both shapes of the fill. The second one matters because the first sweep
    // for this defect only looked for background: var(--accent) and so missed
    // every fill painted with the accent's literal hex, which is how five more
    // sites in dash stayed hidden.
    const accents: Record<string, string> = {};
    for (const app of APPS) {
      const cssPath = path.join(ROOT, 'apps', app, 'src', 'globals.css');
      if (!fs.existsSync(cssPath)) continue;
      const hex = /^\s*--accent:\s*(#[0-9a-f]{6})\s*;/im.exec(
        fs.readFileSync(cssPath, 'utf8')
      )?.[1];
      if (hex) accents[app] = hex.toLowerCase();
    }

    const offenders: string[] = [];
    for (const [app, accent] of Object.entries(accents)) {
      for (const file of sourceFiles(path.join(ROOT, 'apps', app, 'src'))) {
        const lines = fs.readFileSync(file, 'utf8').split('\n');
        lines.forEach((line, i) => {
          if (isComment(line)) return;
          const onFill =
            /background:\s*'?var\(--accent\)/.test(line) ||
            new RegExp(`background:[^,\n]*'${accent}'`, 'i').test(line);
          if (!onFill) return;
          const window = lines.slice(i, i + 3).join('\n');
          if (/color:\s*'?#[0-9a-fA-F]{3,6}'?/.test(window)) {
            offenders.push(`${path.relative(ROOT, file)}:${i + 1} (hardcoded hex)`);
          } else if (/color:\s*[^,\n]*var\(\s*--[a-z0-9-]+\s*,/.test(window)) {
            offenders.push(`${path.relative(ROOT, file)}:${i + 1} (var with literal fallback)`);
          }
        });
      }
    }

    expect(offenders, 'use var(--<app>-on-accent) on an accent fill').toEqual([]);
  });

  it('declares every token an app references, whatever its prefix', () => {
    // The accent-family version of this check above missed three shapes, all
    // found once it was widened (rtylr#1240): dash reaching for flow's
    // --flow-on-accent, erp reaching for --finance-danger, and three HR sites
    // naming a --hr-card-bg that never existed. None carried a fallback, so all
    // six silently inherited.
    //
    // Prefix is not the signal. --hub-* is declared by six apps, so a --hub-
    // token in crm is perfectly legal; what made crm's --hub-accent a defect is
    // that no app declares it at all.
    const missing: string[] = [];

    for (const app of APPS) {
      const src = path.join(ROOT, 'apps', app, 'src');
      const cssPath = path.join(src, 'globals.css');
      const declared = new Set<string>();
      if (fs.existsSync(cssPath)) {
        for (const m of fs.readFileSync(cssPath, 'utf8').matchAll(/^\s*(--[a-z0-9-]+)\s*:/gim)) {
          declared.add(m[1]);
        }
      }

      for (const file of sourceFiles(src)) {
        const lines = fs.readFileSync(file, 'utf8').split('\n');
        lines.forEach((line, i) => {
          if (isComment(line)) return;
          for (const m of line.matchAll(/var\(\s*(--[a-z0-9-]+)\s*([,)])/gi)) {
            // A fallback means it resolves to the fallback rather than
            // inheriting, so it is not this defect.
            if (m[2] === ',') continue;
            if (declared.has(m[1]) || RUNTIME_TOKENS.has(m[1])) continue;
            missing.push(`${path.relative(ROOT, file)}:${i + 1} ${m[1]}`);
          }
        });
      }
    }

    expect(missing, 'a fallback-less var() on an undeclared property inherits').toEqual([]);
  });
});
