import { describe, it, expect } from 'vitest';
import {
  normalizeTag,
  deduplicateTags,
  stockStatusLabel,
  poStatusLabel,
  sortTags,
  parseTags,
  serializeTags,
} from './labelUtils';

describe('normalizeTag', () => {
  it('lowercases and trims', () => {
    expect(normalizeTag('  Electronics  ')).toBe('electronics');
  });

  it('replaces spaces with hyphens', () => {
    expect(normalizeTag('New Arrival')).toBe('new-arrival');
  });

  it('collapses multiple spaces', () => {
    expect(normalizeTag('a  b')).toBe('a-b');
  });
});

describe('deduplicateTags', () => {
  it('removes case-insensitive duplicates', () => {
    const result = deduplicateTags(['Apple', 'apple', 'APPLE', 'Banana']);
    expect(result).toHaveLength(2);
    expect(result[0]).toBe('Apple');
    expect(result[1]).toBe('Banana');
  });
});

describe('stockStatusLabel', () => {
  it('returns Out of Stock when stock is 0', () => {
    const label = stockStatusLabel(0, 5);
    expect(label.text).toBe('Out of Stock');
    expect(label.variant).toBe('danger');
  });

  it('returns Low Stock when stock <= minStock', () => {
    const label = stockStatusLabel(3, 5);
    expect(label.text).toBe('Low Stock');
    expect(label.variant).toBe('warning');
  });

  it('returns In Stock when stock > minStock', () => {
    const label = stockStatusLabel(10, 5);
    expect(label.text).toBe('In Stock');
    expect(label.variant).toBe('success');
  });
});

describe('poStatusLabel', () => {
  it('returns correct label for Draft', () => {
    expect(poStatusLabel('Draft').variant).toBe('default');
  });

  it('returns correct label for Received', () => {
    expect(poStatusLabel('Received').variant).toBe('success');
  });

  it('falls back for unknown status', () => {
    const label = poStatusLabel('Unknown');
    expect(label.text).toBe('Unknown');
    expect(label.variant).toBe('default');
  });
});

describe('sortTags', () => {
  it('sorts alphabetically', () => {
    expect(sortTags(['Banana', 'Apple', 'Cherry'])).toEqual(['Apple', 'Banana', 'Cherry']);
  });

  it('does not mutate original', () => {
    const tags = ['C', 'A', 'B'];
    sortTags(tags);
    expect(tags).toEqual(['C', 'A', 'B']);
  });
});

describe('parseTags', () => {
  it('splits by comma and trims', () => {
    expect(parseTags('apple, banana,  cherry ')).toEqual(['apple', 'banana', 'cherry']);
  });

  it('filters empty entries', () => {
    expect(parseTags('a,,b')).toEqual(['a', 'b']);
  });
});

describe('serializeTags', () => {
  it('joins with comma and space', () => {
    expect(serializeTags(['apple', 'banana'])).toBe('apple, banana');
  });
});
