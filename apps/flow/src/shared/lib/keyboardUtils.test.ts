import { describe, it, expect } from 'vitest';

import {
  isModKey,
  isKey,
  isConfirmKey,
  isEscapeKey,
  isArrowKey,
  isSelectAll,
  shortcutDescription,
  KEYBOARD_SHORTCUTS,
} from './keyboardUtils';

function makeEvent(overrides: Partial<KeyboardEvent>): KeyboardEvent {
  return {
    key: '',
    ctrlKey: false,
    metaKey: false,
    shiftKey: false,
    altKey: false,
    target: document.createElement('div'),
    ...overrides,
  } as unknown as KeyboardEvent;
}

describe('isModKey', () => {
  it('returns true for Ctrl key', () => {
    expect(isModKey(makeEvent({ ctrlKey: true }))).toBe(true);
  });

  it('returns true for Meta key', () => {
    expect(isModKey(makeEvent({ metaKey: true }))).toBe(true);
  });

  it('returns false with no modifier', () => {
    expect(isModKey(makeEvent({}))).toBe(false);
  });
});

describe('isKey', () => {
  it('matches case-insensitively', () => {
    expect(isKey(makeEvent({ key: 'Enter' }), 'enter')).toBe(true);
    expect(isKey(makeEvent({ key: 'a' }), 'A')).toBe(true);
  });

  it('returns false for different key', () => {
    expect(isKey(makeEvent({ key: 'Escape' }), 'Enter')).toBe(false);
  });
});

describe('isConfirmKey', () => {
  it('returns true for Enter without shift', () => {
    expect(isConfirmKey(makeEvent({ key: 'Enter' }))).toBe(true);
  });

  it('returns false for Shift+Enter', () => {
    expect(isConfirmKey(makeEvent({ key: 'Enter', shiftKey: true }))).toBe(false);
  });
});

describe('isEscapeKey', () => {
  it('returns true for Escape', () => {
    expect(isEscapeKey(makeEvent({ key: 'Escape' }))).toBe(true);
  });

  it('returns false for other keys', () => {
    expect(isEscapeKey(makeEvent({ key: 'Enter' }))).toBe(false);
  });
});

describe('isArrowKey', () => {
  it('returns true for all arrow keys', () => {
    expect(isArrowKey(makeEvent({ key: 'ArrowUp' }))).toBe(true);
    expect(isArrowKey(makeEvent({ key: 'ArrowDown' }))).toBe(true);
    expect(isArrowKey(makeEvent({ key: 'ArrowLeft' }))).toBe(true);
    expect(isArrowKey(makeEvent({ key: 'ArrowRight' }))).toBe(true);
  });

  it('returns false for non-arrow keys', () => {
    expect(isArrowKey(makeEvent({ key: 'Tab' }))).toBe(false);
  });
});

describe('isSelectAll', () => {
  it('returns true for Ctrl+A', () => {
    expect(isSelectAll(makeEvent({ key: 'a', ctrlKey: true }))).toBe(true);
  });

  it('returns false without modifier', () => {
    expect(isSelectAll(makeEvent({ key: 'a' }))).toBe(false);
  });
});

describe('shortcutDescription', () => {
  it('returns description for known shortcut', () => {
    expect(shortcutDescription('/')).toBe('Search');
  });

  it('returns null for unknown shortcut', () => {
    expect(shortcutDescription('z')).toBeNull();
  });
});

describe('KEYBOARD_SHORTCUTS', () => {
  it('contains standard shortcuts', () => {
    expect('c' in KEYBOARD_SHORTCUTS).toBe(true);
    expect('/' in KEYBOARD_SHORTCUTS).toBe(true);
  });
});
