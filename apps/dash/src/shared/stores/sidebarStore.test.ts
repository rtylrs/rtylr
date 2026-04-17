import { describe, it, expect, beforeEach } from 'vitest';
import { useSidebarStore } from './sidebarStore';

const { getState } = useSidebarStore;

beforeEach(() => {
  useSidebarStore.setState({ sectionsVisible: {} });
});

describe('sidebarStore', () => {
  it('starts with empty sectionsVisible', () => {
    expect(getState().sectionsVisible).toEqual({});
  });

  it('toggleSection adds section as true when not present', () => {
    getState().toggleSection('dashboard');
    // Section not initialized, so toggling !undefined => true
    expect(getState().sectionsVisible['dashboard']).toBe(true);
  });

  it('toggleSection toggles existing section', () => {
    useSidebarStore.setState({ sectionsVisible: { dashboard: true } });

    getState().toggleSection('dashboard');
    expect(getState().sectionsVisible['dashboard']).toBe(false);

    getState().toggleSection('dashboard');
    expect(getState().sectionsVisible['dashboard']).toBe(true);
  });

  it('initializeSections adds new sections as false', () => {
    getState().initializeSections(['menu', 'orders', 'settings']);

    expect(getState().sectionsVisible['menu']).toBe(false);
    expect(getState().sectionsVisible['orders']).toBe(false);
    expect(getState().sectionsVisible['settings']).toBe(false);
  });

  it('initializeSections preserves existing sections', () => {
    useSidebarStore.setState({ sectionsVisible: { menu: true } });

    getState().initializeSections(['menu', 'orders']);

    expect(getState().sectionsVisible['menu']).toBe(true);
    expect(getState().sectionsVisible['orders']).toBe(false);
  });

  it('initializeSections with empty array does nothing', () => {
    getState().initializeSections([]);
    expect(getState().sectionsVisible).toEqual({});
  });

  it('toggleSection does not affect other sections', () => {
    getState().initializeSections(['a', 'b', 'c']);
    getState().toggleSection('b');

    expect(getState().sectionsVisible['a']).toBe(false);
    expect(getState().sectionsVisible['b']).toBe(true);
    expect(getState().sectionsVisible['c']).toBe(false);
  });

  it('multiple toggles on different sections', () => {
    getState().initializeSections(['x', 'y']);
    getState().toggleSection('x');
    getState().toggleSection('y');

    expect(getState().sectionsVisible['x']).toBe(true);
    expect(getState().sectionsVisible['y']).toBe(true);
  });

  it('initializeSections called twice merges correctly', () => {
    getState().initializeSections(['a', 'b']);
    getState().toggleSection('a');
    getState().initializeSections(['b', 'c']);

    expect(getState().sectionsVisible['a']).toBe(true);
    expect(getState().sectionsVisible['b']).toBe(false);
    expect(getState().sectionsVisible['c']).toBe(false);
  });
});
