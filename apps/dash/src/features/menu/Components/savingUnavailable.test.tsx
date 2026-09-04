import fs from 'node:fs';
import path from 'node:path';
import { describe, expect, it } from 'vitest';

/**
 * Dash's menu authoring must not report success it did not achieve.
 *
 * Both forms changed only the local zustand store and then toasted "Item
 * created" / "Banner created", while `useGetMenu` refetches over the top. A
 * user built a menu, watched it persist through `localStorage`, and it never
 * reached the point of sale or the public menu: absent on another device,
 * absent for every other consumer.
 *
 * Wiring the mutations needs product decisions the store's shape does not
 * answer (nested categories and banners have no home in the flat
 * `pos/menu_item/v1` contract, and `addMenuItem` mints client-side ids the
 * backend will not accept), so until then the forms say so. rtylr#1150.
 *
 * Source assertions: both screens are large forms whose render needs the
 * router, the upload client and the store, and what matters here is that the
 * write calls and the success toasts are gone rather than how the form looks.
 */
const FORMS = ['CreateItem.tsx', 'CreateBanner.tsx'] as const;
const DIR = path.resolve(__dirname);

function source(file: string): string {
  return fs.readFileSync(path.join(DIR, file), 'utf8');
}

describe('menu authoring does not claim success', () => {
  it.each(FORMS)('%s calls no store write on submit', (file) => {
    const code = source(file)
      .replace(/\/\*[\s\S]*?\*\//g, '')
      .split('\n')
      .filter((line) => !line.trim().startsWith('//'))
      .join('\n');

    expect(code).not.toMatch(/\baddMenuItem\(/);
    expect(code).not.toMatch(/\bupdateMenuItem\(/);
    expect(code).not.toMatch(/\baddBanner\(/);
  });

  it.each(FORMS)('%s raises no success toast', (file) => {
    expect(source(file)).not.toMatch(/toast\.success/);
  });

  it.each(FORMS)('%s does not navigate away as if it had saved', (file) => {
    // history.back() after a write read as "done, returning you to the list".
    // The Cancel button still uses it, so this checks the submit path only.
    const code = source(file);
    const submitBody = code.slice(code.indexOf('handleSubmit'), code.indexOf('return ('));
    expect(submitBody).not.toContain('history.back()');
  });

  it.each(FORMS)('%s disables its submit control', (file) => {
    const code = source(file);
    expect(code).toMatch(/type="submit"[\s\S]{0,120}disabled\b/);
  });

  it.each(FORMS)('%s points the disabled control at the reason', (file) => {
    expect(source(file)).toContain('aria-describedby={UNAVAILABLE_ID}');
  });

  it.each(FORMS)('%s explains the state on screen', (file) => {
    const code = source(file);
    expect(code).toContain('<UnavailableNotice>');
    expect(code).toContain("t('menu.savingUnavailable')");
  });

  it('the notice text says where the change does not reach', () => {
    // "Not available" alone would not tell anyone why their menu is missing
    // from the till.
    const en = JSON.parse(
      fs.readFileSync(path.resolve(DIR, '../../../locales/en/menu.json'), 'utf8')
    ) as Record<string, string>;
    expect(en.savingUnavailable).toMatch(/point of sale/i);
    expect(en.savingUnavailable).toMatch(/public menu/i);
  });

  it.each(['en', 'ar', 'fr'])('%s carries the copy', (lang) => {
    const dict = JSON.parse(
      fs.readFileSync(path.resolve(DIR, `../../../locales/${lang}/menu.json`), 'utf8')
    ) as Record<string, string>;
    for (const key of ['savingUnavailable', 'savingUnavailableToast', 'saveItem']) {
      expect(dict[key], `${lang}.${key}`).toBeTruthy();
    }
  });

  it('CreateBanner no longer casts its payload through any', () => {
    // Comments stripped: the comment recording the removal quotes the cast,
    // and a naive search matches that rather than any code. Same trap as the
    // "Project not found." assertion in #1164.
    const code = source('CreateBanner.tsx')
      .replace(/\/\*[\s\S]*?\*\//g, '')
      .split('\n')
      .filter((line) => !line.trim().startsWith('//'))
      .join('\n');
    expect(code).not.toContain('as any');
  });
});
