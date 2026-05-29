import { describe, it, expect } from 'vitest';
import {
  cartCountAriaLabel,
  ariaPressed,
  ariaExpanded,
  buttonRole,
  labelToId,
  isContrastSufficient,
} from './accessibilityUtils';

describe('cartCountAriaLabel', () => {
  it('returns empty cart label for 0', () => {
    expect(cartCountAriaLabel(0)).toBe('Cart is empty');
  });

  it('uses singular for 1 item', () => {
    expect(cartCountAriaLabel(1)).toBe('1 item in cart');
  });

  it('uses plural for more than 1', () => {
    expect(cartCountAriaLabel(5)).toBe('5 items in cart');
  });
});

describe('ariaPressed', () => {
  it('returns "true" when active', () => {
    expect(ariaPressed(true)).toBe('true');
  });

  it('returns "false" when inactive', () => {
    expect(ariaPressed(false)).toBe('false');
  });
});

describe('ariaExpanded', () => {
  it('returns "true" when open', () => {
    expect(ariaExpanded(true)).toBe('true');
  });

  it('returns "false" when closed', () => {
    expect(ariaExpanded(false)).toBe('false');
  });
});

describe('buttonRole', () => {
  it('returns "link" when isLink is true', () => {
    expect(buttonRole(true)).toBe('link');
  });

  it('returns "button" when isLink is false', () => {
    expect(buttonRole(false)).toBe('button');
  });
});

describe('labelToId', () => {
  it('lowercases and replaces spaces with hyphens', () => {
    expect(labelToId('Hot Dogs')).toBe('hot-dogs');
  });

  it('removes special characters', () => {
    expect(labelToId('Combo #1!')).toBe('combo-1');
  });

  it('handles empty string', () => {
    expect(labelToId('')).toBe('');
  });
});

describe('isContrastSufficient', () => {
  it('returns true for black on white (high contrast)', () => {
    expect(isContrastSufficient('#000000', '#ffffff')).toBe(true);
  });

  it('returns false for similar light colors (low contrast)', () => {
    expect(isContrastSufficient('#eeeeee', '#ffffff')).toBe(false);
  });
});
