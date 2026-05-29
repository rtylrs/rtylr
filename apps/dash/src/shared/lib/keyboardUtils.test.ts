import { describe, it, expect } from 'vitest';
import { formatKeyCombo, isNavigationKey, isActionKey, isDismissKey } from './keyboardUtils';

describe('formatKeyCombo', () => {
  it('formats a key with no modifiers', () => {
    expect(formatKeyCombo({ key: 'k' })).toBe('K');
  });

  it('formats a key with ctrl modifier', () => {
    expect(formatKeyCombo({ key: 'k', modifiers: ['ctrl'] })).toBe('Ctrl+K');
  });

  it('formats a key with multiple modifiers', () => {
    expect(formatKeyCombo({ key: 's', modifiers: ['ctrl', 'shift'] })).toBe('Ctrl+Shift+S');
  });

  it('maps meta to Cmd', () => {
    expect(formatKeyCombo({ key: 'z', modifiers: ['meta'] })).toBe('Cmd+Z');
  });

  it('uppercases the key', () => {
    expect(formatKeyCombo({ key: 'enter' })).toBe('ENTER');
  });
});

describe('isNavigationKey', () => {
  it('returns true for arrow keys', () => {
    expect(isNavigationKey('ArrowUp')).toBe(true);
    expect(isNavigationKey('ArrowDown')).toBe(true);
  });

  it('returns true for Home and End', () => {
    expect(isNavigationKey('Home')).toBe(true);
    expect(isNavigationKey('End')).toBe(true);
  });

  it('returns false for non-navigation keys', () => {
    expect(isNavigationKey('a')).toBe(false);
    expect(isNavigationKey('Enter')).toBe(false);
  });
});

describe('isActionKey', () => {
  it('returns true for Enter', () => {
    expect(isActionKey('Enter')).toBe(true);
  });

  it('returns true for space', () => {
    expect(isActionKey(' ')).toBe(true);
  });

  it('returns false for other keys', () => {
    expect(isActionKey('a')).toBe(false);
  });
});

describe('isDismissKey', () => {
  it('returns true for Escape', () => {
    expect(isDismissKey('Escape')).toBe(true);
  });

  it('returns false for other keys', () => {
    expect(isDismissKey('Enter')).toBe(false);
    expect(isDismissKey('a')).toBe(false);
  });
});
