import { describe, expect, it } from 'vitest';
import { buildCsv, csvField, parseCsv } from './csv';

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

describe('parseCsv', () => {
  it('reads quoted cells, doubled quotes and line breaks inside quotes', () => {
    expect(parseCsv('﻿title,description\r\n"Login, page","Say ""hi""\nthen go"\nPlain,\n')).toEqual(
      [
        ['title', 'description'],
        ['Login, page', 'Say "hi"\nthen go'],
        ['Plain', ''],
      ]
    );
  });

  it('drops blank lines and reads what buildCsv writes', () => {
    expect(parseCsv('a,b\n\n,\n1,2')).toEqual([
      ['a', 'b'],
      ['1', '2'],
    ]);
    expect(parseCsv(buildCsv([['Title'], ['مهمة']]))).toEqual([['Title'], ['مهمة']]);
  });
});
