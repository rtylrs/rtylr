import { describe, it, expect } from 'vitest';
import {
  normaliseTag,
  uniqueTags,
  addTag,
  removeTag,
  hasTag,
  intersectTags,
  unionTags,
  tagFrequency,
} from './tagsUtils';

describe('normaliseTag', () => {
  it('trims and lowercases', () => {
    expect(normaliseTag('  VIP  ')).toBe('vip');
  });
});

describe('uniqueTags', () => {
  it('deduplicates', () => {
    expect(uniqueTags(['VIP', 'vip', 'Gold'])).toEqual(['vip', 'gold']);
  });
});

describe('addTag', () => {
  it('adds new tag', () => expect(addTag(['vip'], 'gold')).toEqual(['vip', 'gold']));
  it('does not duplicate', () => expect(addTag(['vip'], 'VIP')).toEqual(['vip']));
  it('ignores empty', () => expect(addTag(['vip'], '  ')).toEqual(['vip']));
});

describe('removeTag', () => {
  it('removes case-insensitively', () =>
    expect(removeTag(['vip', 'gold'], 'VIP')).toEqual(['gold']));
});

describe('hasTag', () => {
  it('true when present', () => expect(hasTag(['vip', 'gold'], 'Gold')).toBe(true));
  it('false when absent', () => expect(hasTag(['vip'], 'bronze')).toBe(false));
});

describe('intersectTags', () => {
  it('returns common tags', () => {
    expect(
      intersectTags([
        ['vip', 'gold'],
        ['vip', 'silver'],
      ])
    ).toEqual(['vip']);
  });
  it('returns empty for empty input', () => expect(intersectTags([])).toEqual([]));
});

describe('unionTags', () => {
  it('merges and deduplicates', () => {
    const r = unionTags([
      ['vip', 'gold'],
      ['gold', 'silver'],
    ]);
    expect(r).toHaveLength(3);
  });
});

describe('tagFrequency', () => {
  it('counts occurrences', () => {
    const freq = tagFrequency([['vip', 'gold'], ['vip'], ['vip']]);
    expect(freq['vip']).toBe(3);
    expect(freq['gold']).toBe(1);
  });
});
