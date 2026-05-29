import { describe, it, expect } from 'vitest';
import { uniqueTags, tagsByColor, hasTag, sortTagsAlpha, tagLabels } from './candidateTag';
import type { CandidateTag } from './candidateTag';

function makeTag(id: number, label: string, color = 'blue'): CandidateTag {
  return { id, label, color };
}

describe('uniqueTags', () => {
  it('removes duplicate labels (case-insensitive)', () => {
    const tags = [makeTag(1, 'JavaScript'), makeTag(2, 'javascript'), makeTag(3, 'Python')];
    expect(uniqueTags(tags)).toHaveLength(2);
  });

  it('returns all tags when all are unique', () => {
    const tags = [makeTag(1, 'React'), makeTag(2, 'Vue')];
    expect(uniqueTags(tags)).toHaveLength(2);
  });
});

describe('tagsByColor', () => {
  it('groups tags by color', () => {
    const tags = [makeTag(1, 'A', 'blue'), makeTag(2, 'B', 'red'), makeTag(3, 'C', 'blue')];
    const groups = tagsByColor(tags);
    expect(groups['blue']).toHaveLength(2);
    expect(groups['red']).toHaveLength(1);
  });
});

describe('hasTag', () => {
  it('returns true when tag exists', () => {
    const tags = [makeTag(1, 'React')];
    expect(hasTag(tags, 'react')).toBe(true);
  });

  it('returns false when tag is missing', () => {
    expect(hasTag([], 'React')).toBe(false);
  });
});

describe('sortTagsAlpha', () => {
  it('sorts tags alphabetically', () => {
    const tags = [makeTag(1, 'Zebra'), makeTag(2, 'Apple'), makeTag(3, 'Mango')];
    const sorted = sortTagsAlpha(tags);
    expect(sorted[0].label).toBe('Apple');
    expect(sorted[2].label).toBe('Zebra');
  });

  it('does not mutate original array', () => {
    const tags = [makeTag(1, 'Zebra'), makeTag(2, 'Apple')];
    sortTagsAlpha(tags);
    expect(tags[0].label).toBe('Zebra');
  });
});

describe('tagLabels', () => {
  it('extracts labels from tags', () => {
    const tags = [makeTag(1, 'React'), makeTag(2, 'TypeScript')];
    expect(tagLabels(tags)).toEqual(['React', 'TypeScript']);
  });
});
