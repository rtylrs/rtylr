import { describe, expect, it } from 'vitest';
import { normaliseTicketNumber } from './ticketNumber';

describe('normaliseTicketNumber', () => {
  it.each([
    ['42', '42'],
    ['  42 ', '42'],
    ['a12', 'A12'],
    ['#A-12', 'A12'],
    ['٤٢', '42'],
    ['۴۲', '42'],
    ['٠١٢٣٤٥٦٧٨٩', '0123456789'],
    ['4 2', '42'],
    ['طلب', ''],
    ['', ''],
  ])('turns %j into %j', (typed, expected) => {
    expect(normaliseTicketNumber(typed)).toBe(expected);
  });

  it('keeps at most twelve characters', () => {
    expect(normaliseTicketNumber('12345678901234')).toBe('123456789012');
  });
});
