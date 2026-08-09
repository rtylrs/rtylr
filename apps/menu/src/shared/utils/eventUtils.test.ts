import { describe, it, expect, vi } from 'vitest';
import { isKey, isPrimaryClick, stopEvent, normaliseKey } from './eventUtils';

describe('isKey', () => {
  it('returns true when the key matches', () => {
    const event = new KeyboardEvent('keydown', { key: 'Enter' });
    expect(isKey(event, 'enter')).toBe(true);
  });

  it('is case-insensitive', () => {
    const event = new KeyboardEvent('keydown', { key: 'Escape' });
    expect(isKey(event, 'ESCAPE')).toBe(true);
  });

  it('returns false for a different key', () => {
    const event = new KeyboardEvent('keydown', { key: 'Tab' });
    expect(isKey(event, 'Enter')).toBe(false);
  });
});

describe('isPrimaryClick', () => {
  it('returns true for button 0 (left click)', () => {
    const event = new MouseEvent('click', { button: 0 });
    expect(isPrimaryClick(event)).toBe(true);
  });

  it('returns false for right click', () => {
    const event = new MouseEvent('click', { button: 2 });
    expect(isPrimaryClick(event)).toBe(false);
  });
});

describe('stopEvent', () => {
  it('calls stopPropagation and preventDefault', () => {
    const event = new Event('click');
    const stopPropagation = vi.spyOn(event, 'stopPropagation');
    const preventDefault = vi.spyOn(event, 'preventDefault');
    stopEvent(event);
    expect(stopPropagation).toHaveBeenCalled();
    expect(preventDefault).toHaveBeenCalled();
  });
});

describe('normaliseKey', () => {
  it('returns the key from the event', () => {
    const event = new KeyboardEvent('keydown', { key: 'ArrowDown' });
    expect(normaliseKey(event)).toBe('ArrowDown');
  });
});
