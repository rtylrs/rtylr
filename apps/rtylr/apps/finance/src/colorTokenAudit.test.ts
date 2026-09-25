import { readFileSync } from 'node:fs';
import { dirname, join } from 'node:path';
import { fileURLToPath } from 'node:url';
import { describe, expect, it } from 'vitest';

const sourceFiles = import.meta.glob('./**/*.{ts,tsx}', {
  eager: true,
  import: 'default',
  query: '?raw',
}) as Record<string, string>;

const globalsCss = readFileSync(join(dirname(fileURLToPath(import.meta.url)), 'globals.css'), 'utf8');

const hash = String.fromCharCode(35);
const hexColor = new RegExp(`${hash}[0-9a-f]{3,8}\\b`, 'gi');

// A short hex colour and a GitHub issue reference are the same three characters
// after a hash, so scanning raw source reports every `(#894)` in a comment as a
// hardcoded colour. Comments are also the one place a hex literal is harmless:
// nothing in a comment reaches the rendered output. Strip them before scanning
// so this audit fails only on colours the browser would actually apply.
const stripComments = (source: string): string =>
  source.replace(/\/\*[\s\S]*?\*\//g, '').replace(/(^|[^:])\/\/.*$/gm, '$1');
const financeVariable = /var\((--finance-[a-z-]+)/g;
const financeDeclaration = /(--finance-[a-z-]+)\s*:/g;

const componentSources = Object.entries(sourceFiles).filter(
  ([path]) =>
    !path.endsWith('.test.ts') &&
    !path.endsWith('.test.tsx') &&
    path !== './shared/lib/settings.ts'
);

describe('Finance color tokens', () => {
  it('keeps component source free of hardcoded hex colors', () => {
    const violations = componentSources.flatMap(([path, source]) =>
      stripComments(source).match(hexColor) ? [path] : []
    );

    expect(violations).toEqual([]);
  });

  it('declares every Finance CSS variable used by component source', () => {
    const used = new Set(
      componentSources.flatMap(([, source]) =>
        [...source.matchAll(financeVariable)].map((match) => match[1])
      )
    );
    const declared = new Set([...globalsCss.matchAll(financeDeclaration)].map((match) => match[1]));

    expect([...used].filter((token) => !declared.has(token))).toEqual([]);
  });
});
