import fs from 'node:fs';
import path from 'node:path';
import { describe, expect, it } from 'vitest';

/**
 * A `var(--token)` that names nothing is a hardcoded colour wearing a token's
 * clothes.
 *
 * Flow had eleven of these. `var(--flow-danger, #ef4444)`,
 * `var(--flow-success, #10b981)`, `var(--flow-warning, #f59e0b)` and
 * `var(--flow-info, #3b82f6)` all reference tokens that are declared nowhere,
 * in this app or in the shared theme, so the literal fallback is what painted
 * every time. The reference read as a token, reviewed as a token, and behaved
 * as a hardcoded colour that could not follow the theme.
 *
 * `var(--flow-accent)` was worse: no fallback and no such token, so the link
 * in `PreviewRouteUnavailable` had no colour at all and inherited whatever was
 * around it.
 */
const SRC = path.resolve(__dirname, '../..');
const GLOBALS = path.join(SRC, 'globals.css');
const SHARED_THEME = path.resolve(SRC, '../../../packages/tailwind-config/theme.css');

/**
 * Written onto documentElement at runtime by AppearanceProvider rather than
 * declared in a stylesheet, so a static scan cannot see them. Every use in the
 * app supplies a fallback for the pre-hydration frame.
 */
const RUNTIME_TOKENS = new Set(['--flow-r', '--flow-u', '--pane-u']);

function declaredTokens(): Set<string> {
  const declared = new Set<string>();
  for (const file of [GLOBALS, SHARED_THEME]) {
    if (!fs.existsSync(file)) continue;
    for (const match of fs.readFileSync(file, 'utf8').matchAll(/^\s*(--[a-z0-9-]+)\s*:/gim)) {
      declared.add(match[1]);
    }
  }
  return declared;
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

describe('token references resolve', () => {
  const declared = declaredTokens();

  it('reads the stylesheets, so a passing run means something', () => {
    expect(declared.size).toBeGreaterThan(80);
    expect(declared.has('--flow-border')).toBe(true);
    expect(declared.has('--flow-danger-fg')).toBe(true);
  });

  it('every var(--token) in the app names a declared token', () => {
    const offenders: string[] = [];
    for (const file of sourceFiles(SRC)) {
      const lines = fs.readFileSync(file, 'utf8').split('\n');
      lines.forEach((line, index) => {
        for (const match of line.matchAll(/var\(\s*(--[a-z0-9-]+)/gi)) {
          const token = match[1];
          if (declared.has(token) || RUNTIME_TOKENS.has(token)) continue;
          offenders.push(`${path.relative(SRC, file)}:${index + 1}  ${token}`);
        }
      });
    }
    expect(offenders).toEqual([]);
  });

  it('the runtime tokens really are absent from the stylesheets', () => {
    // If one of these gains a declaration, it should leave the allowlist
    // rather than sit there hiding a future typo of the same name.
    for (const token of RUNTIME_TOKENS) {
      expect(declared.has(token), `${token} is declared now`).toBe(false);
    }
  });

  it('never uses the failing red as a foreground', () => {
    // #ef4444 measures 3.76:1 on white, under AA for normal text.
    // --flow-danger-fg is derived per mode and clears 4.5:1 on its ground.
    const offenders: string[] = [];
    for (const file of sourceFiles(SRC)) {
      const lines = fs.readFileSync(file, 'utf8').split('\n');
      lines.forEach((line, index) => {
        if (!/#ef4444/i.test(line)) return;
        if (!/\bcolor[:=]/.test(line) && !/\.style\.color/.test(line)) return;
        // A `background` on the same line is a fill, handled separately.
        if (/background/.test(line) && !/\bcolor:\s*'#ef4444'/.test(line)) return;
        offenders.push(`${path.relative(SRC, file)}:${index + 1}`);
      });
    }
    expect(offenders).toEqual([]);
  });
});
