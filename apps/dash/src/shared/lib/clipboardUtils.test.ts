import { describe, it, expect } from 'vitest';
import { isClipboardAvailable, clipboardPreview, clipboardButtonClass } from './clipboardUtils';

describe('isClipboardAvailable', () => {
  it('returns a boolean', () => {
    expect(typeof isClipboardAvailable()).toBe('boolean');
  });
});

describe('clipboardPreview', () => {
  it('returns original text when short enough', () => {
    expect(clipboardPreview('hello', 10)).toBe('hello');
  });

  it('truncates with ellipsis when too long', () => {
    const result = clipboardPreview('this-is-a-very-long-api-key', 15);
    expect(result).toHaveLength(15);
    expect(result.endsWith('...')).toBe(true);
  });

  it('defaults to 20 char max', () => {
    const result = clipboardPreview('a'.repeat(30));
    expect(result).toHaveLength(20);
  });
});

describe('clipboardButtonClass', () => {
  it('returns base class when supported', () => {
    expect(clipboardButtonClass(true)).toBe('btn-copy');
  });

  it('appends disabled class when not supported', () => {
    expect(clipboardButtonClass(false)).toBe('btn-copy btn-copy--disabled');
  });

  it('uses custom base class', () => {
    expect(clipboardButtonClass(true, 'copy-btn')).toBe('copy-btn');
    expect(clipboardButtonClass(false, 'copy-btn')).toBe('copy-btn copy-btn--disabled');
  });
});
