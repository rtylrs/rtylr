import { describe, it, expect } from 'vitest';
import {
  resolveColorScheme,
  colorSchemeClass,
  isDarkMode,
  toggleScheme,
  colorSchemeOptions,
} from './themeUtils';

describe('resolveColorScheme', () => {
  it('returns "light" for light preference', () => {
    expect(resolveColorScheme('light', true)).toBe('light');
  });

  it('returns "dark" for dark preference', () => {
    expect(resolveColorScheme('dark', false)).toBe('dark');
  });

  it('returns system dark when preference is system and system prefers dark', () => {
    expect(resolveColorScheme('system', true)).toBe('dark');
  });

  it('returns system light when preference is system and system prefers light', () => {
    expect(resolveColorScheme('system', false)).toBe('light');
  });
});

describe('colorSchemeClass', () => {
  it('returns "dark" for dark scheme', () => {
    expect(colorSchemeClass('dark')).toBe('dark');
  });

  it('returns "light" for light scheme', () => {
    expect(colorSchemeClass('light')).toBe('light');
  });
});

describe('isDarkMode', () => {
  it('returns true for dark preference', () => {
    expect(isDarkMode('dark')).toBe(true);
  });

  it('returns false for light preference', () => {
    expect(isDarkMode('light')).toBe(false);
  });

  it('uses system preference for system mode', () => {
    expect(isDarkMode('system', true)).toBe(true);
    expect(isDarkMode('system', false)).toBe(false);
  });
});

describe('toggleScheme', () => {
  it('toggles dark to light', () => {
    expect(toggleScheme('dark')).toBe('light');
  });

  it('toggles light to dark', () => {
    expect(toggleScheme('light')).toBe('dark');
  });

  it('treats system as light and returns dark', () => {
    expect(toggleScheme('system')).toBe('dark');
  });
});

describe('colorSchemeOptions', () => {
  it('returns 3 options', () => {
    expect(colorSchemeOptions()).toHaveLength(3);
  });

  it('includes light, dark, and system', () => {
    const values = colorSchemeOptions().map((o) => o.value);
    expect(values).toContain('light');
    expect(values).toContain('dark');
    expect(values).toContain('system');
  });

  it('each option has a label', () => {
    colorSchemeOptions().forEach((o) => {
      expect(o.label).toBeTruthy();
    });
  });
});
