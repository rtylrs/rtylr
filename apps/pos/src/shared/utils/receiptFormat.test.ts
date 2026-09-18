import { describe, it, expect } from 'vitest';
import {
  truncate,
  padEnd,
  padStart,
  receiptLine,
  formatQtyPrice,
  divider,
  centerText,
} from './receiptFormat';

describe('truncate', () => {
  it('returns the original string when it fits', () => {
    expect(truncate('Burger', 10)).toBe('Burger');
  });

  it('truncates and appends a period when the string is too long', () => {
    const result = truncate('Cheeseburger Deluxe', 10);
    expect(result.length).toBe(10);
    expect(result.endsWith('.')).toBe(true);
  });

  it('returns the original string when length equals maxLen', () => {
    expect(truncate('Hello', 5)).toBe('Hello');
  });

  it('handles an empty string', () => {
    expect(truncate('', 5)).toBe('');
  });
});

describe('padEnd', () => {
  it('pads a string with trailing spaces', () => {
    expect(padEnd('Hi', 5)).toBe('Hi   ');
  });

  it('returns the original string when it already meets the width', () => {
    expect(padEnd('Hello', 5)).toBe('Hello');
  });
});

describe('padStart', () => {
  it('pads a string with leading spaces', () => {
    expect(padStart('5', 4)).toBe('   5');
  });

  it('returns the original string when it already meets the width', () => {
    expect(padStart('$5.00', 5)).toBe('$5.00');
  });
});

describe('receiptLine', () => {
  it('produces a line of the specified width', () => {
    const line = receiptLine('Burger', '$12.50', 40);
    expect(line.length).toBe(40);
  });

  it('includes both the label and price', () => {
    const line = receiptLine('Fries', '$4.00', 30);
    expect(line).toContain('Fries');
    expect(line).toContain('$4.00');
  });

  it('truncates a label that is too long', () => {
    const longLabel = 'Extra Long Combo Burger with All Toppings';
    const line = receiptLine(longLabel, '$9.99', 20);
    expect(line.length).toBe(20);
  });
});

describe('formatQtyPrice', () => {
  it('formats quantity and unit price', () => {
    expect(formatQtyPrice(2, '$5.00')).toBe('2 x $5.00');
  });

  it('handles a quantity of 1', () => {
    expect(formatQtyPrice(1, '$10.00')).toBe('1 x $10.00');
  });
});

describe('divider', () => {
  it('returns a line of dashes by default', () => {
    const result = divider();
    expect(result).toBe('-'.repeat(40));
  });

  it('respects a custom character', () => {
    expect(divider('=', 10)).toBe('==========');
  });

  it('respects a custom width', () => {
    expect(divider('-', 20).length).toBe(20);
  });
});

describe('centerText', () => {
  it('centers a short string within the width', () => {
    const result = centerText('Hi', 10);
    expect(result.length).toBe(10);
    expect(result.trim()).toBe('Hi');
  });

  it('returns the original string when it equals the width', () => {
    expect(centerText('Hello', 5)).toBe('Hello');
  });

  it('returns the original string when it exceeds the width', () => {
    const long = 'This is quite a long string';
    expect(centerText(long, 5)).toBe(long);
  });

  it('handles an empty string', () => {
    const result = centerText('', 10);
    expect(result.length).toBe(10);
  });
});
