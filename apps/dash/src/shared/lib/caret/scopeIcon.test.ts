import { describe, expect, it } from 'vitest';
import { scopeEmoji } from './scopeIcon';

describe('scopeEmoji', () => {
  it('returns 👀 for read', () => {
    expect(scopeEmoji('read')).toBe('👀');
  });
  it('returns ✏️ for write', () => {
    expect(scopeEmoji('write')).toBe('✏️');
  });
});
