import fs from 'node:fs';
import path from 'node:path';
import { describe, expect, it } from 'vitest';

const PLURAL_SUFFIX = /_(zero|one|two|few|many|other)$/;

function keys(value: unknown, prefix = ''): string[] {
  if (typeof value !== 'object' || value === null) return [prefix.replace(PLURAL_SUFFIX, '')];
  return Object.entries(value).flatMap(([key, child]) =>
    keys(child, prefix ? `${prefix}.${key}` : key)
  );
}

function load(lang: string): string[] {
  const file = path.resolve(__dirname, 'locales', lang, 'pager.json');
  return [...new Set(keys(JSON.parse(fs.readFileSync(file, 'utf8'))))].sort();
}

// Plural forms differ by language, so they are compared by their base key.
describe('pager locales', () => {
  it.each(['fr', 'ar'])('%s has exactly the English keys', (lang) => {
    expect(load(lang)).toEqual(load('en'));
  });
});
