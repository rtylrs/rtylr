import { describe, expect, it, beforeEach } from 'vitest';
import { canSend, resetCooldown } from './preventDoubleSend';

beforeEach(() => resetCooldown());

describe('canSend', () => {
  it('allows the first call', () => {
    expect(canSend('hi')).toBe(true);
  });
  it('blocks an immediate repeat', () => {
    canSend('hi');
    expect(canSend('hi')).toBe(false);
  });
  it('treats different prompts independently', () => {
    canSend('hi');
    expect(canSend('hello')).toBe(true);
  });
});
