import { describe, it, expect } from 'vitest';
import { cn } from './utils';

describe('cn', () => {
  it('merges class names', () => {
    expect(cn('foo', 'bar')).toBe('foo bar');
  });
  it('handles conditional classes', () => {
    expect(cn('base', true && 'active', false && 'hidden')).toBe('base active');
  });
  it('deduplicates tailwind classes with later value winning', () => {
    expect(cn('text-red-500', 'text-blue-500')).toBe('text-blue-500');
  });
  it('ignores falsy values', () => {
    expect(cn(undefined, null, false, 'valid')).toBe('valid');
  });
  it('handles object syntax', () => {
    expect(cn({ active: true, disabled: false })).toBe('active');
  });
  it('returns empty string for no classes', () => {
    expect(cn()).toBe('');
  });
});
