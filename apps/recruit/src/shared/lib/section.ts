// Section content helpers for the recruit public page.
// Company pages can have custom content sections (perks, FAQ, values,
// etc.) that the template renders. These helpers parse, sort, and
// derive display properties for sections.

import type { SectionKind } from './types';

export interface ContentSection {
  id: number;
  careersPageId: number;
  kind: SectionKind | string;
  title: string;
  itemsJson: string;
  position: number;
  isVisible: boolean;
}

/**
 * sectionKindLabel returns the human-readable label for a section kind.
 */
export function sectionKindLabel(kind: SectionKind | string): string {
  const map: Record<SectionKind, string> = {
    perks: 'Perks & Benefits',
    culture: 'Culture',
    testimonial: 'Testimonials',
    video: 'Video',
    faq: 'FAQ',
    values: 'Our Values',
  };
  return (map as Record<string, string>)[kind] ?? kind.charAt(0).toUpperCase() + kind.slice(1);
}

/**
 * visibleSections returns only sections where isVisible is true,
 * sorted by position ascending.
 */
export function visibleSections(sections: ContentSection[]): ContentSection[] {
  return [...sections].filter((s) => s.isVisible).sort((a, b) => a.position - b.position);
}

/**
 * sectionByKind returns the first section matching the given kind,
 * or undefined when none exists.
 */
export function sectionByKind(
  sections: ContentSection[],
  kind: SectionKind | string
): ContentSection | undefined {
  return sections.find((s) => s.kind === kind);
}

/**
 * parseItemsJson safely decodes the itemsJson field. Returns an empty
 * array on parse failure or non-array values.
 */
export function parseItemsJson<T = unknown>(itemsJson: string): T[] {
  if (!itemsJson) return [];
  try {
    const parsed = JSON.parse(itemsJson);
    if (Array.isArray(parsed)) return parsed as T[];
    return [];
  } catch {
    return [];
  }
}

/**
 * maxSectionPosition returns the highest position value in a section
 * list, or 0 for empty lists. Used to compute the next position when
 * appending a new section.
 */
export function maxSectionPosition(sections: ContentSection[]): number {
  if (sections.length === 0) return 0;
  return Math.max(...sections.map((s) => s.position));
}
