import { describe, it, expect } from 'vitest';
import {
  sectionKindLabel,
  visibleSections,
  sectionByKind,
  parseItemsJson,
  maxSectionPosition,
} from './section';
import type { ContentSection } from './section';

function makeSection(overrides: Partial<ContentSection> = {}): ContentSection {
  return {
    id: 1,
    careersPageId: 1,
    kind: 'perks',
    title: 'Perks',
    itemsJson: '[]',
    position: 1,
    isVisible: true,
    ...overrides,
  };
}

describe('sectionKindLabel', () => {
  it('returns Perks & Benefits for perks', () => {
    expect(sectionKindLabel('perks')).toBe('Perks & Benefits');
  });

  it('returns FAQ for faq', () => {
    expect(sectionKindLabel('faq')).toBe('FAQ');
  });

  it('returns Our Values for values', () => {
    expect(sectionKindLabel('values')).toBe('Our Values');
  });

  it('capitalizes unknown kinds', () => {
    expect(sectionKindLabel('gallery')).toBe('Gallery');
  });
});

describe('visibleSections', () => {
  it('returns only visible sections', () => {
    const sections = [
      makeSection({ id: 1, isVisible: true }),
      makeSection({ id: 2, isVisible: false }),
      makeSection({ id: 3, isVisible: true }),
    ];
    const result = visibleSections(sections);
    expect(result).toHaveLength(2);
    expect(result.every((s) => s.isVisible)).toBe(true);
  });

  it('sorts by position ascending', () => {
    const sections = [makeSection({ id: 1, position: 3 }), makeSection({ id: 2, position: 1 })];
    const result = visibleSections(sections);
    expect(result[0].id).toBe(2);
  });

  it('does not mutate the original array', () => {
    const sections = [makeSection({ id: 1, position: 2 }), makeSection({ id: 2, position: 1 })];
    visibleSections(sections);
    expect(sections[0].id).toBe(1);
  });
});

describe('sectionByKind', () => {
  it('returns the first matching section', () => {
    const sections = [makeSection({ kind: 'perks' }), makeSection({ kind: 'faq' })];
    const result = sectionByKind(sections, 'faq');
    expect(result?.kind).toBe('faq');
  });

  it('returns undefined when no match found', () => {
    const sections = [makeSection({ kind: 'perks' })];
    expect(sectionByKind(sections, 'values')).toBeUndefined();
  });
});

describe('parseItemsJson', () => {
  it('parses a valid JSON array', () => {
    const result = parseItemsJson<{ label: string }>('[{"label":"Remote"}]');
    expect(result).toHaveLength(1);
    expect(result[0].label).toBe('Remote');
  });

  it('returns empty array for empty string', () => {
    expect(parseItemsJson('')).toEqual([]);
  });

  it('returns empty array for invalid JSON', () => {
    expect(parseItemsJson('not json')).toEqual([]);
  });

  it('returns empty array when JSON is not an array', () => {
    expect(parseItemsJson('{"key":"value"}')).toEqual([]);
  });
});

describe('maxSectionPosition', () => {
  it('returns the highest position', () => {
    const sections = [
      makeSection({ position: 1 }),
      makeSection({ position: 5 }),
      makeSection({ position: 3 }),
    ];
    expect(maxSectionPosition(sections)).toBe(5);
  });

  it('returns 0 for an empty array', () => {
    expect(maxSectionPosition([])).toBe(0);
  });
});
