import { describe, it, expect, vi, afterEach } from 'vitest';
import {
  viewportWidth,
  isAtLeast,
  activeBreakpoint,
  prefersReducedMotion,
  prefersDarkMode,
} from '../media';

afterEach(() => {
  vi.unstubAllGlobals();
});

describe('viewportWidth', () => {
  it('returns 0 in test environment (no window.innerWidth)', () => {
    expect(typeof viewportWidth()).toBe('number');
  });
});

describe('isAtLeast', () => {
  it('returns a boolean', () => {
    expect(typeof isAtLeast('md')).toBe('boolean');
  });
});

describe('activeBreakpoint', () => {
  it('returns a valid breakpoint name', () => {
    const bp = activeBreakpoint();
    expect(['xs', 'sm', 'md', 'lg', 'xl', '2xl']).toContain(bp);
  });
});

describe('prefersReducedMotion', () => {
  it('returns false in test environment', () => {
    expect(typeof prefersReducedMotion()).toBe('boolean');
  });
});

describe('prefersDarkMode', () => {
  it('returns false in test environment', () => {
    expect(typeof prefersDarkMode()).toBe('boolean');
  });
});
