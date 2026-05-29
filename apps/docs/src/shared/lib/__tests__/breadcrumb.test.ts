import { describe, it, expect } from 'vitest';
import { pathToCrumbs, crumbLabels, currentCrumb, ancestorCrumbs } from '../breadcrumb';

describe('pathToCrumbs', () => {
  it('builds crumbs from a path', () => {
    const crumbs = pathToCrumbs('/guides/getting-started');
    expect(crumbs.length).toBe(2);
    expect(crumbs[0].label).toBe('Guides');
    expect(crumbs[1].label).toBe('Getting Started');
  });

  it('handles single-segment paths', () => {
    const crumbs = pathToCrumbs('/introduction');
    expect(crumbs.length).toBe(1);
    expect(crumbs[0].label).toBe('Introduction');
  });

  it('builds correct hrefs', () => {
    const crumbs = pathToCrumbs('/guides/api');
    expect(crumbs[0].href).toBe('/guides');
    expect(crumbs[1].href).toBe('/guides/api');
  });

  it('returns empty array for root path', () => {
    expect(pathToCrumbs('/')).toEqual([]);
  });
});

describe('crumbLabels', () => {
  it('returns labels in order', () => {
    const crumbs = pathToCrumbs('/guides/getting-started');
    expect(crumbLabels(crumbs)).toEqual(['Guides', 'Getting Started']);
  });
});

describe('currentCrumb', () => {
  it('returns the last crumb', () => {
    const crumbs = pathToCrumbs('/guides/api');
    expect(currentCrumb(crumbs)?.label).toBe('Api');
  });

  it('returns undefined for empty list', () => {
    expect(currentCrumb([])).toBeUndefined();
  });
});

describe('ancestorCrumbs', () => {
  it('returns all but last crumb', () => {
    const crumbs = pathToCrumbs('/a/b/c');
    const ancestors = ancestorCrumbs(crumbs);
    expect(ancestors.length).toBe(2);
    expect(ancestors[ancestors.length - 1].label).toBe('B');
  });

  it('returns empty for single crumb', () => {
    const crumbs = pathToCrumbs('/home');
    expect(ancestorCrumbs(crumbs)).toEqual([]);
  });
});
