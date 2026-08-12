import { readFileSync } from 'node:fs';
import { join } from 'node:path';
import { describe, expect, it } from 'vitest';

const sourceFiles = import.meta.glob('./**/*.{ts,tsx}', {
  eager: true,
  import: 'default',
  query: '?raw',
}) as Record<string, string>;

const globalsCss = readFileSync(join(process.cwd(), 'apps/finance/src/globals.css'), 'utf8');

const hash = String.fromCharCode(35);
const hexColor = new RegExp(`${hash}[0-9a-f]{3,8}\\b`, 'gi');
const financeVariable = /var\((--finance-[a-z-]+)/g;
const financeDeclaration = /(--finance-[a-z-]+)\s*:/g;

const componentSources = Object.entries(sourceFiles).filter(
  ([path]) =>
    !path.endsWith('.test.ts') &&
    !path.endsWith('.test.tsx') &&
    path !== './shared/lib/settings.ts' &&
    path !== './shared/schemas/expenses.ts'
);

describe('Finance color tokens', () => {
  it('keeps component source free of hardcoded hex colors', () => {
    const violations = componentSources.flatMap(([path, source]) =>
      source.match(hexColor) ? [path] : []
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
