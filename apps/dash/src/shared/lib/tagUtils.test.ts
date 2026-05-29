import { describe, it, expect } from 'vitest';
import {
  normalizeTagName,
  isValidTagName,
  tagExists,
  addTag,
  removeTag,
  toggleTag,
  sortTags,
  filterTags,
  type Tag,
} from './tagUtils';

const t = (id: string, name: string): Tag => ({ id, name });

describe('normalizeTagName', () => {
  it('trims and lowercases', () => expect(normalizeTagName('  Hello  ')).toBe('hello'));
});

describe('isValidTagName', () => {
  it('valid', () => expect(isValidTagName('my-tag')).toBe(true));
  it('empty', () => expect(isValidTagName('')).toBe(false));
  it('special chars', () => expect(isValidTagName('bad@tag')).toBe(false));
  it('too long', () => expect(isValidTagName('a'.repeat(33))).toBe(false));
});

describe('tagExists', () => {
  it('case-insensitive match', () => expect(tagExists([t('1', 'Promo')], 'promo')).toBe(true));
  it('no match', () => expect(tagExists([t('1', 'Promo')], 'vip')).toBe(false));
});

describe('addTag', () => {
  it('adds new tag', () => expect(addTag([], t('1', 'vip'))).toHaveLength(1));
  it('no duplicate', () => expect(addTag([t('1', 'vip')], t('1', 'vip'))).toHaveLength(1));
});

describe('removeTag', () => {
  it('removes by id', () =>
    expect(removeTag([t('1', 'vip'), t('2', 'promo')], '1')).toHaveLength(1));
});

describe('toggleTag', () => {
  it('adds when absent', () => expect(toggleTag([], t('1', 'vip'))).toHaveLength(1));
  it('removes when present', () =>
    expect(toggleTag([t('1', 'vip')], t('1', 'vip'))).toHaveLength(0));
});

describe('sortTags', () => {
  it('alphabetical', () => {
    const tags = [t('1', 'zebra'), t('2', 'apple'), t('3', 'mango')];
    expect(sortTags(tags)[0].name).toBe('apple');
  });
});

describe('filterTags', () => {
  it('filters by query', () =>
    expect(filterTags([t('1', 'vip'), t('2', 'promo')], 'vi')).toHaveLength(1));
  it('returns all for empty query', () =>
    expect(filterTags([t('1', 'vip'), t('2', 'promo')], '')).toHaveLength(2));
});
