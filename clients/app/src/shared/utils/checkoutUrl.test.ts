import { describe, it, expect } from 'vitest';
import { isAllowedCheckoutUrl, CHECKOUT_HOSTS } from './checkoutUrl';

describe('isAllowedCheckoutUrl', () => {
  it('accepts the exact Paddle checkout hosts', () => {
    for (const host of CHECKOUT_HOSTS) {
      expect(isAllowedCheckoutUrl(`https://${host}/checkout/abc123`)).toBe(true);
    }
    expect(isAllowedCheckoutUrl('https://BUY.PADDLE.COM/checkout/x')).toBe(true);
  });

  it('rejects lookalike and non-checkout hosts', () => {
    expect(isAllowedCheckoutUrl('https://paddle.com.evil.com/checkout/x')).toBe(false);
    expect(isAllowedCheckoutUrl('https://buy.paddle.com.evil.com/checkout/x')).toBe(false);
    expect(isAllowedCheckoutUrl('https://evil.com/checkout/x')).toBe(false);
    expect(isAllowedCheckoutUrl('https://paddle.com/pricing')).toBe(false);
    expect(isAllowedCheckoutUrl('https://sub.buy.paddle.com/checkout/x')).toBe(false);
  });

  it('rejects non-https and malformed values', () => {
    expect(isAllowedCheckoutUrl('http://buy.paddle.com/checkout/x')).toBe(false);
    expect(isAllowedCheckoutUrl('javascript:alert(1)')).toBe(false);
    expect(isAllowedCheckoutUrl('buy.paddle.com/checkout/x')).toBe(false);
    expect(isAllowedCheckoutUrl('')).toBe(false);
    expect(isAllowedCheckoutUrl(null)).toBe(false);
    expect(isAllowedCheckoutUrl(undefined)).toBe(false);
    expect(isAllowedCheckoutUrl(42)).toBe(false);
    expect(isAllowedCheckoutUrl(`https://buy.paddle.com/${'x'.repeat(3000)}`)).toBe(false);
  });
});
