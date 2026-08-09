import { describe, it, expect } from 'vitest';
import { isBrowserEnv, isNodeEnv, getEnvVar, getAppVersion } from '../env';

describe('isBrowserEnv', () => {
  it('returns a boolean', () => {
    expect(typeof isBrowserEnv()).toBe('boolean');
  });
});

describe('isNodeEnv', () => {
  it('returns true in Node/Vitest', () => {
    expect(isNodeEnv()).toBe(true);
  });
});

describe('getEnvVar', () => {
  it('returns fallback for missing variable', () => {
    expect(getEnvVar('__NONEXISTENT_VAR__', 'fallback')).toBe('fallback');
  });

  it('returns empty string when no fallback given', () => {
    expect(getEnvVar('__NONEXISTENT_VAR__')).toBe('');
  });
});

describe('getAppVersion', () => {
  it('returns a string', () => {
    expect(typeof getAppVersion()).toBe('string');
  });

  it('returns a non-empty string', () => {
    expect(getAppVersion().length).toBeGreaterThan(0);
  });
});
