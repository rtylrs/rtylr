import { describe, expect, it } from 'vitest';
import { initialsOf } from './initials';

describe('initialsOf', () => {
  it('returns ? for empty input', () => {
    expect(initialsOf('')).toBe('?');
    expect(initialsOf(null)).toBe('?');
    expect(initialsOf(undefined)).toBe('?');
  });
  it('uses the first two letters of a single name', () => {
    expect(initialsOf('Sara')).toBe('SA');
  });
  it('uses first + last of a full name', () => {
    expect(initialsOf('Mohamad Homsi')).toBe('MH');
  });
  it('skips internal whitespace', () => {
    expect(initialsOf('  Mary  Ann  Doe ')).toBe('MD');
  });
});
