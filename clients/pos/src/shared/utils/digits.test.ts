import { describe, expect, it } from 'vitest';
import { digitsOnly, normalizeDigits } from './digits';

const FULL_SET = '0123456789';

function block(start: number): string {
  return Array.from({ length: 10 }, (_, i) => String.fromCharCode(start + i)).join('');
}

describe('normalizeDigits', () => {
  it('maps Arabic-Indic digits to ASCII', () => {
    // U+0660-U+0669
    expect(normalizeDigits('٠١٢٣٤٥٦٧٨٩')).toBe(FULL_SET);
    expect(normalizeDigits('٤٥٦')).toBe('456');
  });

  it('maps Eastern Arabic-Indic digits to ASCII', () => {
    // U+06F0-U+06F9
    expect(normalizeDigits('۰۱۲۳۴۵۶۷۸۹')).toBe(FULL_SET);
    expect(normalizeDigits('۱۲۳۴')).toBe('1234');
  });

  it('maps the other common decimal digit blocks to ASCII', () => {
    const blocks: Array<[string, number]> = [
      ['Devanagari', 0x0966],
      ['Bengali', 0x09e6],
      ['Gurmukhi', 0x0a66],
      ['Gujarati', 0x0ae6],
      ['Tamil', 0x0be6],
      ['Telugu', 0x0c66],
      ['Kannada', 0x0ce6],
      ['Thai', 0x0e50],
      ['Lao', 0x0ed0],
      ['Fullwidth', 0xff10],
    ];
    for (const [name, start] of blocks) {
      expect(normalizeDigits(block(start)), name).toBe(FULL_SET);
    }
  });

  it('normalises digits mixed into other text', () => {
    expect(normalizeDigits('order ٤٥٦ total')).toBe('order 456 total');
    expect(normalizeDigits('١٢٣٤')).toBe('1234');
  });

  it('leaves ASCII, punctuation and non-digit scripts alone', () => {
    expect(normalizeDigits('abc 123-456.78')).toBe('abc 123-456.78');
    expect(normalizeDigits('مرحبا')).toBe('مرحبا');
    expect(normalizeDigits('')).toBe('');
  });

  it('does not treat non-decimal numerals as digits', () => {
    // Roman numeral one and superscript two are numeric to a human but are not
    // decimal digits, so they must not be silently translated.
    expect(normalizeDigits('Ⅰ²')).toBe('Ⅰ²');
  });
});

describe('digitsOnly', () => {
  it('keeps Arabic-Indic digits that a bare \\D strip would delete', () => {
    expect('٤٥٦'.replace(/\D/g, '')).toBe('');
    expect(digitsOnly('٤٥٦')).toBe('456');
    expect(digitsOnly('۱۲۳۴')).toBe('1234');
  });

  it('still drops characters that are not digits', () => {
    expect(digitsOnly('a1b2-c3')).toBe('123');
    expect(digitsOnly('1.5')).toBe('15');
    expect(digitsOnly('')).toBe('');
  });

  it('caps the result at maxLength so a PIN cannot overrun', () => {
    expect(digitsOnly('٤٥٦٧٨٩', 4)).toBe('4567');
    expect(digitsOnly('1234567890', 4)).toBe('1234');
  });
});
