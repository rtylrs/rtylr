import { describe, it, expect } from 'vitest';
import type { DocContent } from '../../content';
import { buildNavItems, prevSlug, nextSlug, getBreadcrumbs } from '../nav';

const makeDoc = (title: string): DocContent => ({ title, body: '' });

const sampleMap: Record<string, DocContent> = {
  'quick-start': makeDoc('Quick Start'),
  'api-overview': makeDoc('API Overview'),
  introduction: makeDoc('Introduction'),
};

describe('buildNavItems', () => {
  it('returns items sorted alphabetically', () => {
    const items = buildNavItems(sampleMap, 'introduction');
    expect(items.map((i) => i.slug)).toEqual(['api-overview', 'introduction', 'quick-start']);
  });

  it('marks the current slug as active', () => {
    const items = buildNavItems(sampleMap, 'introduction');
    const active = items.filter((i) => i.active);
    expect(active).toHaveLength(1);
    expect(active[0].slug).toBe('introduction');
  });

  it('includes the title from the content map', () => {
    const items = buildNavItems(sampleMap, 'quick-start');
    const qs = items.find((i) => i.slug === 'quick-start')!;
    expect(qs.title).toBe('Quick Start');
  });
});

describe('prevSlug', () => {
  it('returns null for the first slug', () => {
    expect(prevSlug(sampleMap, 'api-overview')).toBeNull();
  });

  it('returns the preceding slug', () => {
    expect(prevSlug(sampleMap, 'introduction')).toBe('api-overview');
  });
});

describe('nextSlug', () => {
  it('returns null for the last slug', () => {
    expect(nextSlug(sampleMap, 'quick-start')).toBeNull();
  });

  it('returns the following slug', () => {
    expect(nextSlug(sampleMap, 'api-overview')).toBe('introduction');
  });

  it('returns null for an unknown slug', () => {
    expect(nextSlug(sampleMap, 'nonexistent')).toBeNull();
  });
});

describe('getBreadcrumbs', () => {
  it('starts with a root Docs crumb', () => {
    const crumbs = getBreadcrumbs(sampleMap, 'introduction');
    expect(crumbs[0]).toEqual({ label: 'Docs', href: '/docs' });
  });

  it('appends the current page crumb', () => {
    const crumbs = getBreadcrumbs(sampleMap, 'introduction');
    expect(crumbs[1]).toEqual({ label: 'Introduction', href: '/docs/introduction' });
  });

  it('returns only the root crumb for unknown slugs', () => {
    const crumbs = getBreadcrumbs(sampleMap, 'missing');
    expect(crumbs).toHaveLength(1);
  });
});
