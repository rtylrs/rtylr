import { describe, expect, it } from 'vitest';
import { makeId } from './uuid';

describe('makeId', () => {
  it('prefixes the result', () => {
    expect(makeId('chat')).toMatch(/^chat-/);
  });

  it('produces unique values', () => {
    const ids = new Set<string>();
    for (let i = 0; i < 50; i++) ids.add(makeId('x'));
    expect(ids.size).toBe(50);
  });
});
