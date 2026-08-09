import { describe, it, expect } from 'vitest';
import {
  DEFAULT_CURRENCY,
  DEFAULT_LOCALE,
  ORDER_POLL_INTERVAL_MS,
  ITEM_NOTE_MAX_LENGTH,
  ORDER_NOTE_MAX_LENGTH,
  CART_STORAGE_KEY,
  THEME_STORAGE_KEY,
  MIN_TAP_TARGET_PX,
  CSS_VARS,
} from './menuConstants';

describe('menu constants', () => {
  it('DEFAULT_CURRENCY is USD', () => {
    expect(DEFAULT_CURRENCY).toBe('USD');
  });

  it('DEFAULT_LOCALE is en-US', () => {
    expect(DEFAULT_LOCALE).toBe('en-US');
  });

  it('ORDER_POLL_INTERVAL_MS is 30000', () => {
    expect(ORDER_POLL_INTERVAL_MS).toBe(30_000);
  });

  it('ITEM_NOTE_MAX_LENGTH is a positive integer', () => {
    expect(ITEM_NOTE_MAX_LENGTH).toBeGreaterThan(0);
    expect(Number.isInteger(ITEM_NOTE_MAX_LENGTH)).toBe(true);
  });

  it('ORDER_NOTE_MAX_LENGTH is greater than ITEM_NOTE_MAX_LENGTH', () => {
    expect(ORDER_NOTE_MAX_LENGTH).toBeGreaterThan(ITEM_NOTE_MAX_LENGTH);
  });

  it('CART_STORAGE_KEY is a non-empty string', () => {
    expect(typeof CART_STORAGE_KEY).toBe('string');
    expect(CART_STORAGE_KEY.length).toBeGreaterThan(0);
  });

  it('THEME_STORAGE_KEY is a non-empty string', () => {
    expect(typeof THEME_STORAGE_KEY).toBe('string');
    expect(THEME_STORAGE_KEY.length).toBeGreaterThan(0);
  });

  it('MIN_TAP_TARGET_PX meets WCAG minimum of 44', () => {
    expect(MIN_TAP_TARGET_PX).toBeGreaterThanOrEqual(44);
  });

  it('CSS_VARS contains accent, accentText, and background', () => {
    expect(CSS_VARS.accent).toBeTruthy();
    expect(CSS_VARS.accentText).toBeTruthy();
    expect(CSS_VARS.background).toBeTruthy();
  });
});
