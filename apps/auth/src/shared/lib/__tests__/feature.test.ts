import { describe, it, expect } from 'vitest';
import {
  isEnabled,
  enabledFlags,
  disabledFlags,
  withOverrides,
  allEnabled,
  anyEnabled,
} from '../feature';

const flags = { darkMode: true, beta: false, newDashboard: true };

describe('isEnabled', () => {
  it('returns true for enabled flag', () => {
    expect(isEnabled(flags, 'darkMode')).toBe(true);
  });

  it('returns false for disabled flag', () => {
    expect(isEnabled(flags, 'beta')).toBe(false);
  });

  it('returns false for missing flag', () => {
    expect(isEnabled(flags, 'nonexistent')).toBe(false);
  });
});

describe('enabledFlags', () => {
  it('returns only enabled flag keys', () => {
    expect(enabledFlags(flags).sort()).toEqual(['darkMode', 'newDashboard'].sort());
  });
});

describe('disabledFlags', () => {
  it('returns only disabled flag keys', () => {
    expect(disabledFlags(flags)).toEqual(['beta']);
  });
});

describe('withOverrides', () => {
  it('applies overrides', () => {
    const result = withOverrides(flags, { beta: true });
    expect(result.beta).toBe(true);
    expect(result.darkMode).toBe(true);
  });

  it('does not mutate original', () => {
    withOverrides(flags, { beta: true });
    expect(flags.beta).toBe(false);
  });
});

describe('allEnabled', () => {
  it('returns true when all keys are enabled', () => {
    expect(allEnabled(flags, ['darkMode', 'newDashboard'])).toBe(true);
  });

  it('returns false when any key is disabled', () => {
    expect(allEnabled(flags, ['darkMode', 'beta'])).toBe(false);
  });
});

describe('anyEnabled', () => {
  it('returns true if at least one key is enabled', () => {
    expect(anyEnabled(flags, ['beta', 'darkMode'])).toBe(true);
  });

  it('returns false if all keys are disabled', () => {
    expect(anyEnabled(flags, ['beta'])).toBe(false);
  });
});
