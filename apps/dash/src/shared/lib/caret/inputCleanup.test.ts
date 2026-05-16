import { describe, expect, it } from 'vitest';
import { cleanupInput } from './inputCleanup';

describe('cleanupInput', () => {
  it('normalizes CRLF to LF', () => {
    expect(cleanupInput('a\r\nb\r\n')).toBe('a\nb');
  });
  it('normalizes bare CR to LF', () => {
    expect(cleanupInput('a\rb')).toBe('a\nb');
  });
  it('strips trailing whitespace', () => {
    expect(cleanupInput('hello   \n   ')).toBe('hello');
  });
});
