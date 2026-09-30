/**
 * The tip and split-bill fields used to be read with parseFloat. parseFloat
 * stops at the first character it does not recognise, so a French cashier
 * typing a 12,50 tip produced 12: a hundred-fold under-charge written to the
 * order. The mirror image is just as wrong — stripping non-digits turned the
 * same input into 1250 — so these cases pin the exact value, not the shape of
 * the text, in every locale the app ships.
 */
import { describe, expect, it } from 'vitest';
import { parseLocalizedAmount, sanitiseAmountInput } from './amount';

const NBSP = '\u00a0';
const NNBSP = '\u202f';

const CASES: [locale: string, typed: string, expected: number][] = [
  // The locale's own convention.
  ['en', '12.50', 12.5],
  ['fr', '12,50', 12.5],
  ['de', '12,50', 12.5],
  ['es', '12,50', 12.5],
  ['it', '12,50', 12.5],
  ['ar', '12.50', 12.5],
  // The other convention, which the locale's own rule still reads correctly.
  ['fr', '12.50', 12.5],
  ['en', '12,50', 12.5],
  ['de', '12.50', 12.5],
  // Grouping, in each convention and each space a copy can carry.
  ['en', '1,234.56', 1234.56],
  ['en', '1 234.56', 1234.56],
  ['en', `1${NBSP}234.56`, 1234.56],
  ['en', `1${NNBSP}234.56`, 1234.56],
  ['fr', '1 234,56', 1234.56],
  ['fr', '1\u202F234,56', 1234.56],
  ['de', '1.234,56', 1234.56],
  ['en', '\u20AC1,234.56', 1234.56],
  // A currency symbol and stray text pasted around the amount.
  ['fr', '\u20AC1\u202F234,56', 1234.56],
  // Fractions, whole amounts and a separator with nothing after it yet.
  ['en', '0.5', 0.5],
  ['fr', '0,5', 0.5],
  ['en', '12', 12],
  ['fr', '12', 12],
  ['en', '12.', 12],
  ['fr', '12,', 12],
  ['en', '12.0', 12],
  // Negative.
  ['en', '-12.50', -12.5],
  ['fr', '-12,50', -12.5],
  // Text that is not an amount.
  ['en', '', 0],
  ['en', 'abc', 0],
  ['en', '.', 0],
  // The two spellings both conventions share. The locale's own wins.
  ['en', '1,234', 1234],
  ['en', '1.234', 1.234],
  ['fr', '1.234', 1234],
  ['fr', '1,234', 1.234],
  ['de', '1.234', 1234],
  ['de', '1,234', 1.234],
  // Exactly three digits after a foreign separator is a group; anything else is
  // a fraction. When that is genuinely ambiguous the smaller reading wins.
  ['en', '12,5', 12.5],
  ['fr', '12,5', 12.5],
  ['en', '12,50', 12.5],
  ['en', '12,500', 12500],
  ['fr', '12.500', 12500],
];

describe('parseLocalizedAmount', () => {
  it.each(CASES)('reads %s "%s" as %d', (locale, typed, expected) => {
    expect(parseLocalizedAmount(typed, locale)).toBe(expected);
  });

  it('never returns a hundred-fold error in either direction', () => {
    // The two failure modes this replaces: an under-charge and an over-charge
    // of the same real amount, one hundred times out.
    expect(Number.parseFloat('12,50')).toBe(12);
    expect('12,50'.replace(/\D/g, '')).toBe('1250');
    expect(parseLocalizedAmount('12,50', 'fr')).toBe(12.5);
  });

  it('reads a locale whose own separators are not . or ,', () => {
    // ar-EG writes the decimal separator as U+066B and groups with U+066C.
    expect(parseLocalizedAmount('\u0661\u0662\u066B\u0665\u0660', 'ar-EG')).toBe(12.5);
    expect(parseLocalizedAmount('\u0661\u066C\u0662\u0663\u0664\u066B\u0665\u0666', 'ar-EG')).toBe(1234.56);
  });

  it('is stable on a string that has already been sanitised', () => {
    for (const [locale, typed, expected] of CASES) {
      expect(parseLocalizedAmount(sanitiseAmountInput(typed, locale), locale)).toBe(expected);
    }
  });
});

describe('sanitiseAmountInput', () => {
  it('keeps the locale separator and the text the operator typed', () => {
    expect(sanitiseAmountInput('12,50', 'fr')).toBe('12,50');
    expect(sanitiseAmountInput('12.50', 'en')).toBe('12.50');
  });

  it('never stacks a repeated separator, which would invent digits', () => {
    expect(sanitiseAmountInput('1.2.3', 'en')).toBe('1.23');
    expect(sanitiseAmountInput('1,2,3', 'fr')).toBe('1,23');
  });

  it('drops characters that are not part of an amount', () => {
    expect(sanitiseAmountInput('abc', 'en')).toBe('');
    expect(sanitiseAmountInput('12abc.50', 'en')).toBe('12.50');
  });

  it('keeps a minus sign only in front', () => {
    expect(sanitiseAmountInput('1-2', 'en')).toBe('1-2'.replace('-', ''));
    expect(sanitiseAmountInput('-12.50', 'en')).toBe('-12.50');
  });

  it('normalises non-ASCII digits instead of deleting them', () => {
    expect('٤٥٦'.replace(/\D/g, '')).toBe('');
    expect(sanitiseAmountInput('١٢,٥٠', 'fr')).toBe('12,50');
  });
});
