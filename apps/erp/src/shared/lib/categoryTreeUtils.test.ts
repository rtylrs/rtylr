import { describe, it, expect } from 'vitest';
import {
  buildCategoryTree,
  flattenCategoryTree,
  findCategoryNode,
  categoryBreadcrumb,
  subtreeProductCount,
  leafCategories,
} from './categoryTreeUtils';
import type { Category } from '../services/erpApi';

const cats: Category[] = [
  { id: 1, name: 'Electronics', parentId: null, productCount: 0, sortOrder: 1 },
  { id: 2, name: 'Phones', parentId: 1, productCount: 10, sortOrder: 1 },
  { id: 3, name: 'Laptops', parentId: 1, productCount: 5, sortOrder: 2 },
  { id: 4, name: 'Smartphones', parentId: 2, productCount: 8, sortOrder: 1 },
  { id: 5, name: 'Food', parentId: null, productCount: 0, sortOrder: 2 },
  { id: 6, name: 'Dairy', parentId: 5, productCount: 12, sortOrder: 1 },
];

describe('buildCategoryTree', () => {
  it('builds a tree with correct children', () => {
    const tree = buildCategoryTree(cats);
    const electronics = tree.find((n) => n.id === 1);
    expect(electronics).toBeDefined();
    expect(electronics!.children).toHaveLength(2);
  });

  it('sets depth correctly', () => {
    const tree = buildCategoryTree(cats);
    const flat = flattenCategoryTree(tree);
    const smartphones = flat.find((n) => n.id === 4);
    expect(smartphones!.depth).toBe(2);
  });

  it('sets path correctly', () => {
    const tree = buildCategoryTree(cats);
    const flat = flattenCategoryTree(tree);
    const smartphones = flat.find((n) => n.id === 4);
    expect(smartphones!.path).toEqual(['Electronics', 'Phones', 'Smartphones']);
  });
});

describe('flattenCategoryTree', () => {
  it('returns all nodes in depth-first order', () => {
    const tree = buildCategoryTree(cats);
    const flat = flattenCategoryTree(tree);
    expect(flat).toHaveLength(6);
  });
});

describe('findCategoryNode', () => {
  it('finds a node by id', () => {
    const tree = buildCategoryTree(cats);
    const node = findCategoryNode(tree, 4);
    expect(node).toBeDefined();
    expect(node!.name).toBe('Smartphones');
  });

  it('returns undefined for missing id', () => {
    const tree = buildCategoryTree(cats);
    expect(findCategoryNode(tree, 999)).toBeUndefined();
  });
});

describe('categoryBreadcrumb', () => {
  it('returns breadcrumb path with separator', () => {
    const breadcrumb = categoryBreadcrumb(cats, 4);
    expect(breadcrumb).toContain('Electronics');
    expect(breadcrumb).toContain('Phones');
    expect(breadcrumb).toContain('Smartphones');
  });

  it('returns empty string for missing category', () => {
    expect(categoryBreadcrumb(cats, 999)).toBe('');
  });
});

describe('subtreeProductCount', () => {
  it('sums products including all children', () => {
    const tree = buildCategoryTree(cats);
    const electronics = tree.find((n) => n.id === 1)!;
    // Electronics: 0 + Phones (10) + Laptops (5) + Smartphones (8) = 23
    expect(subtreeProductCount(electronics)).toBe(23);
  });
});

describe('leafCategories', () => {
  it('returns categories with no children', () => {
    const leaves = leafCategories(cats);
    const leafIds = leaves.map((l) => l.id);
    expect(leafIds).toContain(3); // Laptops
    expect(leafIds).toContain(4); // Smartphones
    expect(leafIds).toContain(6); // Dairy
    expect(leafIds).not.toContain(1); // Electronics has children
  });
});
