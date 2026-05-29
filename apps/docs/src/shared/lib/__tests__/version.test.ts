import { describe, it, expect } from 'vitest';
import {
  parseSemver,
  compareSemver,
  latestVersion,
  isMajorVersion,
  breakingVersions,
} from '../version';

describe('parseSemver', () => {
  it('parses a valid semver string', () => {
    expect(parseSemver('1.2.3')).toEqual({ major: 1, minor: 2, patch: 3 });
  });

  it('handles v-prefix', () => {
    expect(parseSemver('v2.0.0')).toEqual({ major: 2, minor: 0, patch: 0 });
  });

  it('returns null for invalid string', () => {
    expect(parseSemver('not-a-version')).toBeNull();
  });
});

describe('compareSemver', () => {
  it('returns 1 when a is greater', () => {
    expect(compareSemver('2.0.0', '1.9.9')).toBe(1);
  });

  it('returns -1 when a is less', () => {
    expect(compareSemver('1.0.0', '2.0.0')).toBe(-1);
  });

  it('returns 0 for equal versions', () => {
    expect(compareSemver('1.2.3', '1.2.3')).toBe(0);
  });

  it('compares patch versions', () => {
    expect(compareSemver('1.0.1', '1.0.0')).toBe(1);
  });
});

describe('latestVersion', () => {
  const entries = [{ version: '1.0.0' }, { version: '2.1.0' }, { version: '1.5.3' }];

  it('returns the latest version', () => {
    expect(latestVersion(entries)?.version).toBe('2.1.0');
  });

  it('returns undefined for empty list', () => {
    expect(latestVersion([])).toBeUndefined();
  });
});

describe('isMajorVersion', () => {
  it('returns true for x.0.0', () => {
    expect(isMajorVersion('2.0.0')).toBe(true);
  });

  it('returns false for minor versions', () => {
    expect(isMajorVersion('2.1.0')).toBe(false);
  });

  it('returns false for patch versions', () => {
    expect(isMajorVersion('2.0.1')).toBe(false);
  });
});

describe('breakingVersions', () => {
  it('returns only breaking versions', () => {
    const entries = [
      { version: '1.0.0', breaking: true },
      { version: '1.1.0' },
      { version: '2.0.0', breaking: true },
    ];
    expect(breakingVersions(entries).length).toBe(2);
  });

  it('returns empty for no breaking changes', () => {
    expect(breakingVersions([{ version: '1.0.0' }])).toEqual([]);
  });
});
