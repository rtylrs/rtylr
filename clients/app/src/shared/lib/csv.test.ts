import { describe, expect, it } from 'vitest';
import { buildCsv, csvField } from './csv';

const BOM = '﻿';

describe('csvField', () => {
  it('quotes every field and doubles quotes inside it', () => {
    expect(csvField('plain')).toBe('"plain"');
    expect(csvField('say "hi", then go')).toBe('"say ""hi"", then go"');
    expect(csvField('two\nlines')).toBe('"two\nlines"');
    expect(csvField(null)).toBe('""');
    expect(csvField(2.5)).toBe('"2.5"');
  });

  it.each(['=HYPERLINK("x")', '+1+1', '-2', '@SUM(A1)', '\tTAB', '\rCR'])(
    'keeps %j as text rather than a formula',
    (value) => {
      expect(csvField(value).startsWith(`"'`)).toBe(true);
    }
  );

  it('leaves a value that only contains those characters alone', () => {
    expect(csvField('a=b')).toBe('"a=b"');
  });
});

describe('buildCsv', () => {
  it('starts with a byte order mark and ends rows with CRLF', () => {
    expect(buildCsv([['a'], ['b']])).toBe(`${BOM}"a"\r\n"b"\r\n`);
  });
});
