import fs from 'node:fs';
import path from 'node:path';
import { describe, expect, it } from 'vitest';

/**
 * Menu authoring forms must persist directly to the canonical backend APIs
 * (`pos/menu_item/v1` and `hub/menu/v1`) instead of mutating the local zustand
 * store (`addMenuItem` / `updateMenuItem` / `addBanner`).
 */
const FORMS = ['CreateItem.tsx', 'CreateBanner.tsx'] as const;
const DIR = path.resolve(__dirname);

function source(file: string): string {
  return fs.readFileSync(path.join(DIR, file), 'utf8');
}

describe('menu authoring persists through canonical APIs', () => {
  it.each(FORMS)('%s calls no browser-only store writes on submit', (file) => {
    const code = source(file)
      .replace(/\/\*[\s\S]*?\*\//g, '')
      .split('\n')
      .filter((line) => !line.trim().startsWith('//'))
      .join('\n');

    expect(code).not.toMatch(/\baddMenuItem\(/);
    expect(code).not.toMatch(/\baddBanner\(/);
  });

  it('CreateItem wires create and update mutations to pos/menu_item/v1 hooks', () => {
    const code = source('CreateItem.tsx');
    expect(code).toContain('useCreateMenuItem');
    expect(code).toContain('useUpdateMenuItem');
    expect(code).not.toContain('<UnavailableNotice>');
  });

  it('CreateBanner wires creation to hub/menu/v1 banner hook without any casts', () => {
    const code = source('CreateBanner.tsx');
    expect(code).toContain('useCreateMenuBanner');
    expect(code).not.toContain('<UnavailableNotice>');
    expect(code).not.toContain('as any');
  });
});
