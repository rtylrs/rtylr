import { describe, it, expect } from 'vitest';
import {
  orderStatusLabel,
  isOrderEditable,
  isOrderFinalized,
  isRefundable,
  isVoidable,
  orderStatusColor,
} from './orderStatus';

describe('orderStatusLabel', () => {
  it('returns "Open" for open status', () => {
    expect(orderStatusLabel('open')).toBe('Open');
  });

  it('returns "Submitted" for sent status', () => {
    expect(orderStatusLabel('sent')).toBe('Submitted');
  });

  it('returns "Ready" for ready status', () => {
    expect(orderStatusLabel('ready')).toBe('Ready');
  });

  it('returns "Paid" for paid status', () => {
    expect(orderStatusLabel('paid')).toBe('Paid');
  });

  it('returns "Voided" for voided status', () => {
    expect(orderStatusLabel('voided')).toBe('Voided');
  });

  it('returns "Refunded" for refunded status', () => {
    expect(orderStatusLabel('refunded')).toBe('Refunded');
  });
});

describe('isOrderEditable', () => {
  it('returns true for open orders', () => {
    expect(isOrderEditable('open')).toBe(true);
  });

  it('returns true for sent orders', () => {
    expect(isOrderEditable('sent')).toBe(true);
  });

  it('returns false for paid orders', () => {
    expect(isOrderEditable('paid')).toBe(false);
  });

  it('returns false for voided orders', () => {
    expect(isOrderEditable('voided')).toBe(false);
  });

  it('returns false for refunded orders', () => {
    expect(isOrderEditable('refunded')).toBe(false);
  });
});

describe('isOrderFinalized', () => {
  it('returns true for paid orders', () => {
    expect(isOrderFinalized('paid')).toBe(true);
  });

  it('returns true for voided orders', () => {
    expect(isOrderFinalized('voided')).toBe(true);
  });

  it('returns true for refunded orders', () => {
    expect(isOrderFinalized('refunded')).toBe(true);
  });

  it('returns false for open orders', () => {
    expect(isOrderFinalized('open')).toBe(false);
  });

  it('returns false for sent orders', () => {
    expect(isOrderFinalized('sent')).toBe(false);
  });
});

describe('isRefundable', () => {
  it('returns true for paid orders', () => {
    expect(isRefundable('paid')).toBe(true);
  });

  it('returns false for open orders', () => {
    expect(isRefundable('open')).toBe(false);
  });

  it('returns false for voided orders', () => {
    expect(isRefundable('voided')).toBe(false);
  });

  it('returns false for refunded orders', () => {
    expect(isRefundable('refunded')).toBe(false);
  });
});

describe('isVoidable', () => {
  it('returns true for open orders', () => {
    expect(isVoidable('open')).toBe(true);
  });

  it('returns true for sent orders', () => {
    expect(isVoidable('sent')).toBe(true);
  });

  it('returns false for paid orders', () => {
    expect(isVoidable('paid')).toBe(false);
  });

  it('returns false for voided orders', () => {
    expect(isVoidable('voided')).toBe(false);
  });
});

describe('orderStatusColor', () => {
  it('returns green for paid', () => {
    expect(orderStatusColor('paid')).toBe('green');
  });

  it('returns blue for open', () => {
    expect(orderStatusColor('open')).toBe('blue');
  });

  it('returns yellow for sent', () => {
    expect(orderStatusColor('sent')).toBe('yellow');
  });

  it('returns yellow for ready', () => {
    expect(orderStatusColor('ready')).toBe('yellow');
  });

  it('returns red for voided', () => {
    expect(orderStatusColor('voided')).toBe('red');
  });

  it('returns red for refunded', () => {
    expect(orderStatusColor('refunded')).toBe('red');
  });
});
