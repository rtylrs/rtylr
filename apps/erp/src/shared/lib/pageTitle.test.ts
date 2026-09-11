import fs from 'node:fs';
import path from 'node:path';
import { describe, expect, it } from 'vitest';
import { pageTitle } from './pageTitle';

const FEATURES = path.join(__dirname, '..', '..', 'features');

describe('pageTitle', () => {
  it('puts the product name after the screen name', () => {
    expect(pageTitle('Brands')).toBe('Brands · rtylr');
  });

  it('uses a middle dot rather than a hyphen', () => {
    expect(pageTitle('Brands')).toContain('·');
    expect(pageTitle('Brands')).not.toContain(' - ');
  });
});

/**
 * The point of the helper is that the product name lives in one place. A screen
 * that spells it out again defeats that, and nothing else would catch it: a
 * literal tab title is valid TypeScript and renders fine in English.
 */
describe('no screen spells out the product name', () => {
  function sourceFiles(dir: string, into: string[] = []): string[] {
    for (const entry of fs.readdirSync(dir, { withFileTypes: true })) {
      const full = path.join(dir, entry.name);
      if (entry.isDirectory()) sourceFiles(full, into);
      else if (/\.tsx?$/.test(entry.name) && !/\.test\.tsx?$/.test(entry.name)) into.push(full);
    }
    return into;
  }

  it('finds no "· rtylr" outside the helper', () => {
    const offenders = sourceFiles(FEATURES)
      .filter((file) => fs.readFileSync(file, 'utf8').includes('· rtylr'))
      .map((file) => path.relative(FEATURES, file));
    expect(offenders).toEqual([]);
  });

  it('checks a real number of files', () => {
    // A floor, so a walk that silently finds nothing cannot pass by default.
    expect(sourceFiles(FEATURES).length).toBeGreaterThanOrEqual(30);
  });
});
