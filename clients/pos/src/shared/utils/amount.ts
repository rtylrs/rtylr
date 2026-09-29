// Money typed into a till is written the way the operator's locale writes it, so
// it cannot be read with parseFloat: parseFloat("12,50") is 12, and a hundred
// times out is written straight to the order. The locale's own decimal
// separator is therefore read from Intl rather than assumed.

import { normalizeDigits } from './digits';

interface LocaleSeparators {
  decimal: string;
  group: string;
}

/** The locale's own separators, read from Intl rather than guessed. */
function localeSeparators(locale: string): LocaleSeparators {
  try {
    const parts = new Intl.NumberFormat(locale).formatToParts(12345.6);
    return {
      decimal: parts.find((p) => p.type === 'decimal')?.value ?? '.',
      group: parts.find((p) => p.type === 'group')?.value ?? ',',
    };
  } catch {
    return { decimal: '.', group: ',' };
  }
}

/**
 * Keeps only what an amount can be written with: digits, a sign, whitespace and
 * the separators this locale can spell one with. A pasted amount carries a
 * currency symbol, and leaving that in would make the field hold text the
 * reader then has to interpret.
 */
function keepAmountCharacters(text: string, separators: LocaleSeparators): string {
  const allowed = new Set<string>(['-', '.', ',', separators.decimal, separators.group]);
  let out = '';
  for (const ch of normalizeDigits(text)) {
    if (allowed.has(ch) || /\d|\s/.test(ch)) out += ch;
  }
  return out;
}

/** Drops the grouping characters left over once the decimal point is located. */
function stripGrouping(digits: string, separators: LocaleSeparators): string {
  return digits.replace(/[.,]/g, '').split(separators.group).join('');
}

/**
 * Keeps what an amount can be typed with: digits, a leading sign, whitespace,
 * and at most one of each separator. A repeated separator is dropped rather
 * than stacked, so the text left in the field always reads one way.
 *
 * Non-ASCII digits are normalised, so a till on an Arabic keyboard reads what
 * the operator typed instead of discarding the keystroke. That only reaches the
 * field if the input can hold such a value at all: `type="number"` makes the
 * browser drop them before any sanitiser runs.
 */
export function sanitiseAmountInput(raw: string, locale: string): string {
  const separators = localeSeparators(locale);
  const allowed = new Set<string>(['.', ',', separators.decimal, separators.group]);
  const text = keepAmountCharacters(raw, separators).replace(/[\u2012\u2013\u2014\u2212]/g, '-');
  let out = '';
  const seen = new Set<string>();
  for (const ch of text) {
    if (ch === '-') {
      if (out === '') out += ch;
    } else if (/\d|\s/.test(ch)) {
      out += ch;
    } else if (allowed.has(ch) && !seen.has(ch)) {
      seen.add(ch);
      out += ch;
    }
  }
  return out;
}

/**
 * Reads typed text as a decimal number in the operator's locale.
 *
 * The locale's own decimal separator is never second-guessed. A foreign
 * separator is digit-grouping only when exactly three digits follow it, because
 * three digits after a separator is a thousands group, not a fraction. That is
 * what settles the two spellings both conventions share: "1.234" is one
 * thousand two hundred and thirty-four in de and 1.234 in en, and "12,5" is
 * twelve and a half rather than a hundred and twenty-five. Treating the foreign
 * separator as a group when it is not one is how "12,50" became 1250, a
 * hundred-fold overstatement of a real amount. Where a reading is genuinely
 * ambiguous the smaller one wins: 12.5 is visibly correctable in the field,
 * whereas 125 is a silent tenfold. Whitespace, including the non-breaking and
 * narrow ones a copy carries, groups everywhere.
 */
export function parseLocalizedAmount(text: string, locale: string): number {
  const separators = localeSeparators(locale);
  const sign = /^\s*-/.test(text) ? -1 : 1;
  // Whitespace only ever groups, so it goes before anything is classified: it
  // covers the plain, non-breaking and narrow spaces a copy can carry.
  const body = keepAmountCharacters(text, separators).replace(/[\s-]/g, '');
  if (body === '') return 0;

  // The locale's own decimal character joins '.' and ',' as a candidate, which
  // is what lets ar-EG's U+066B be read without special-casing a single script.
  const at = Math.max(
    body.lastIndexOf('.'),
    body.lastIndexOf(','),
    body.lastIndexOf(separators.decimal)
  );
  let whole = body;
  let fraction = '';
  const isGrouping = at !== -1 && body[at] !== separators.decimal && /^\d{3}$/.test(body.slice(at + 1));
  if (at !== -1 && !isGrouping) {
    whole = body.slice(0, at);
    fraction = body.slice(at + 1);
  }
  const value = Number(`${stripGrouping(whole, separators)}.${stripGrouping(fraction, separators)}`);
  return Number.isFinite(value) ? sign * value : 0;
}
