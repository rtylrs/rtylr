import { describe, expect, it } from 'vitest';
import { DEFAULT_CURRENCY, formatMoney } from './money';

describe('formatMoney', () => {
  it('treats the amount as major units, not cents', () => {
    // The regression this exists for: 12.5 is twelve dollars fifty, because
    // pos_orders.subtotal is DECIMAL(10,2). Dividing by 100 rendered $0.13.
    expect(formatMoney(12.5)).toBe('$12.50');
    expect(formatMoney(1234.56)).toBe('$1,234.56');
    expect(formatMoney(0)).toBe('$0.00');
  });

  it('defaults to the currency the backend defaults to', () => {
    expect(DEFAULT_CURRENCY).toBe('USD');
    expect(formatMoney(1)).toBe(formatMoney(1, 'USD'));
  });

  it('takes the fraction digits from the currency, so zero-decimal currencies get none', () => {
    // JPY and KRW have no minor unit. Hardcoding two decimals would print
    // "¥1,234.00" for an amount that cannot have a fractional part.
    expect(formatMoney(1234, 'JPY')).toBe('¥1,234');
    expect(formatMoney(1234.56, 'JPY')).toBe('¥1,235');
    expect(formatMoney(1234, 'KRW')).toBe('₩1,234');
  });

  it('gives a three-decimal currency three decimals', () => {
    expect(formatMoney(1.5, 'KWD')).toContain('1.500');
  });

  it('formats a non-USD two-decimal currency with its own symbol', () => {
    expect(formatMoney(9.99, 'EUR')).toBe('€9.99');
  });

  it('falls back to the code and the amount for an unusable currency', () => {
    // Intl throws on a malformed code. A visible code beats a wrong symbol.
    expect(formatMoney(12.5, 'not-a-currency')).toBe('not-a-currency 12.50');
  });

  it('renders zero rather than NaN for a non-finite amount', () => {
    expect(formatMoney(Number.NaN)).toBe('$0.00');
    expect(formatMoney(Number.POSITIVE_INFINITY)).toBe('$0.00');
  });

  it('keeps negative amounts signed', () => {
    expect(formatMoney(-5.25)).toBe('-$5.25');
  });
});
