import { describe, it, expect } from 'vitest';
import {
  isValidCashAmount,
  isValidReason,
  isValidDiscountPercent,
  isValidFlatDiscount,
  isValidTipPercent,
  isValidTipAmount,
  isValidGiftCardCode,
  isValidPin,
  isValidSplitSeats,
} from './paymentValidation';

describe('isValidCashAmount', () => {
  it('accepts a positive numeric string', () => {
    expect(isValidCashAmount('20.00')).toBe(true);
  });

  it('accepts an integer string', () => {
    expect(isValidCashAmount('10')).toBe(true);
  });

  it('rejects zero', () => {
    expect(isValidCashAmount('0')).toBe(false);
  });

  it('rejects a negative value', () => {
    expect(isValidCashAmount('-5')).toBe(false);
  });

  it('rejects a non-numeric string', () => {
    expect(isValidCashAmount('abc')).toBe(false);
  });

  it('rejects an empty string', () => {
    expect(isValidCashAmount('')).toBe(false);
  });
});

describe('isValidReason', () => {
  it('accepts a non-empty reason', () => {
    expect(isValidReason('Customer complaint')).toBe(true);
  });

  it('rejects an empty string', () => {
    expect(isValidReason('')).toBe(false);
  });

  it('rejects a whitespace-only string', () => {
    expect(isValidReason('   ')).toBe(false);
  });
});

describe('isValidDiscountPercent', () => {
  it('accepts a percentage within 0-100', () => {
    expect(isValidDiscountPercent(10)).toBe(true);
  });

  it('accepts exactly 100', () => {
    expect(isValidDiscountPercent(100)).toBe(true);
  });

  it('rejects 0', () => {
    expect(isValidDiscountPercent(0)).toBe(false);
  });

  it('rejects a negative value', () => {
    expect(isValidDiscountPercent(-5)).toBe(false);
  });

  it('rejects a value above 100', () => {
    expect(isValidDiscountPercent(101)).toBe(false);
  });

  it('rejects NaN', () => {
    expect(isValidDiscountPercent(NaN)).toBe(false);
  });
});

describe('isValidFlatDiscount', () => {
  it('accepts a positive flat amount', () => {
    expect(isValidFlatDiscount(5)).toBe(true);
  });

  it('rejects zero', () => {
    expect(isValidFlatDiscount(0)).toBe(false);
  });

  it('rejects a negative amount', () => {
    expect(isValidFlatDiscount(-10)).toBe(false);
  });

  it('rejects NaN', () => {
    expect(isValidFlatDiscount(NaN)).toBe(false);
  });
});

describe('isValidTipPercent', () => {
  it('accepts zero (no tip)', () => {
    expect(isValidTipPercent(0)).toBe(true);
  });

  it('accepts a positive percentage', () => {
    expect(isValidTipPercent(18)).toBe(true);
  });

  it('rejects a negative percentage', () => {
    expect(isValidTipPercent(-5)).toBe(false);
  });

  it('rejects a percentage above 100', () => {
    expect(isValidTipPercent(101)).toBe(false);
  });
});

describe('isValidTipAmount', () => {
  it('accepts zero', () => {
    expect(isValidTipAmount(0)).toBe(true);
  });

  it('accepts a positive amount', () => {
    expect(isValidTipAmount(3.5)).toBe(true);
  });

  it('rejects a negative amount', () => {
    expect(isValidTipAmount(-1)).toBe(false);
  });

  it('rejects NaN', () => {
    expect(isValidTipAmount(NaN)).toBe(false);
  });
});

describe('isValidGiftCardCode', () => {
  it('accepts a standard gift card code', () => {
    expect(isValidGiftCardCode('GIFT-ABCD-1234')).toBe(true);
  });

  it('accepts an alphanumeric code', () => {
    expect(isValidGiftCardCode('ABCD1234')).toBe(true);
  });

  it('rejects an empty string', () => {
    expect(isValidGiftCardCode('')).toBe(false);
  });

  it('rejects a code that is too short', () => {
    expect(isValidGiftCardCode('AB')).toBe(false);
  });

  it('rejects codes with spaces', () => {
    expect(isValidGiftCardCode('GIFT CARD')).toBe(false);
  });
});

describe('isValidPin', () => {
  it('accepts a 4-digit PIN', () => {
    expect(isValidPin('1234')).toBe(true);
  });

  it('accepts a 6-digit PIN', () => {
    expect(isValidPin('123456')).toBe(true);
  });

  it('rejects a 3-digit PIN', () => {
    expect(isValidPin('123')).toBe(false);
  });

  it('rejects a 7-digit PIN', () => {
    expect(isValidPin('1234567')).toBe(false);
  });

  it('rejects a PIN with letters', () => {
    expect(isValidPin('123a')).toBe(false);
  });

  it('rejects an empty string', () => {
    expect(isValidPin('')).toBe(false);
  });
});

describe('isValidSplitSeats', () => {
  it('accepts 2 seats', () => {
    expect(isValidSplitSeats(2)).toBe(true);
  });

  it('accepts a larger number of seats', () => {
    expect(isValidSplitSeats(8)).toBe(true);
  });

  it('rejects 1 seat', () => {
    expect(isValidSplitSeats(1)).toBe(false);
  });

  it('rejects 0 seats', () => {
    expect(isValidSplitSeats(0)).toBe(false);
  });

  it('rejects a fractional number', () => {
    expect(isValidSplitSeats(2.5)).toBe(false);
  });
});
