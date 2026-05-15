import { describe, expect, it } from 'vitest';
import { safeFilename } from './safeFilename';

describe('safeFilename', () => {
  it('appends the extension', () => {
    expect(safeFilename('hello', 'md')).toBe('hello.md');
  });
  it('strips path separators and unsafe characters', () => {
    expect(safeFilename('a/b\\c:d?', 'txt')).toBe('abcd.txt');
  });
  it('falls back to a default name when empty', () => {
    expect(safeFilename('', 'md')).toBe('caret-chat.md');
  });
});
