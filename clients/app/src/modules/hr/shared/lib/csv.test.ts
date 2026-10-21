import { describe, expect, it } from 'vitest';
import { escapeCsvValue } from './csv';

describe('escapeCsvValue', () => {
  it('passes plain values through', () => {
    expect(escapeCsvValue('Jane')).toBe('Jane');
    expect(escapeCsvValue(null)).toBe('');
    expect(escapeCsvValue(undefined)).toBe('');
  });

  it('quotes commas, quotes, and line breaks', () => {
    expect(escapeCsvValue('a,b')).toBe('"a,b"');
    expect(escapeCsvValue('say "hi"')).toBe('"say ""hi"""');
    expect(escapeCsvValue('a\nb')).toBe('"a\nb"');
    expect(escapeCsvValue('a\rb')).toBe('"a\rb"');
  });

  it('neutralizes text a spreadsheet would run as a formula', () => {
    expect(escapeCsvValue('=SUM(A1:A2)')).toBe("'=SUM(A1:A2)");
    expect(escapeCsvValue('+1 555')).toBe("'+1 555");
    expect(escapeCsvValue('@cmd')).toBe("'@cmd");
    expect(escapeCsvValue('-1+1')).toBe("'-1+1");
  });

  it('leaves real negative numbers alone', () => {
    expect(escapeCsvValue(-5)).toBe('-5');
  });
});
