import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { copyToClipboard, readFromClipboard } from '../clipboard';
describe('copyToClipboard', () => {
  beforeEach(() => {
    Object.defineProperty(globalThis, 'navigator', {
      value: { clipboard: { writeText: vi.fn().mockResolvedValue(undefined) } },
      configurable: true,
    });
  });
  afterEach(() => {
    vi.restoreAllMocks();
  });
  it('returns true when available', async () => {
    expect(await copyToClipboard('hello')).toBe(true);
  });
});
describe('readFromClipboard', () => {
  beforeEach(() => {
    Object.defineProperty(globalThis, 'navigator', {
      value: { clipboard: { readText: vi.fn().mockResolvedValue('text') } },
      configurable: true,
    });
  });
  afterEach(() => {
    vi.restoreAllMocks();
  });
  it('returns clipboard text', async () => {
    expect(await readFromClipboard()).toBe('text');
  });
});
