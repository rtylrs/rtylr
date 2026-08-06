import { describe, expect, it } from 'vitest';
import { buildDiscountRequest } from './discountWorkflow';

describe('buildDiscountRequest', () => {
  it('builds a normalized percent discount request', () => {
    expect(buildDiscountRequest({ orderId: '42.9', mode: 'percent', valueInput: '150' })).toEqual({
      payload: { orderId: '42', discountType: 'percent', discountValue: 100, reason: '' },
    });
  });

  it('builds a normalized flat discount request', () => {
    expect(buildDiscountRequest({ orderId: '42', mode: 'flat', valueInput: '12.50' })).toEqual({
      payload: { orderId: '42', discountType: 'flat', discountValue: 12.5, reason: '' },
    });
  });

  it('rejects invalid order ids', () => {
    expect(() =>
      buildDiscountRequest({ orderId: 'current', mode: 'percent', valueInput: '10' })
    ).toThrow('Select a saved order before applying a discount.');
  });

  it('rejects zero or malformed discounts', () => {
    expect(() =>
      buildDiscountRequest({ orderId: '42', mode: 'flat', valueInput: 'abc' })
    ).toThrow('Enter a discount greater than zero.');
    expect(() =>
      buildDiscountRequest({ orderId: '42', mode: 'percent', valueInput: '0' })
    ).toThrow('Enter a discount greater than zero.');
  });
});
