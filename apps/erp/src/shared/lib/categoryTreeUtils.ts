/**
 * Category tree building and traversal helpers for the ERP app.
 * All functions are pure with no side-effects.
 */

import type { Category } from '../services/erpApi';

export interface CategoryNode extends Category {
  children: CategoryNode[];
  depth: number;
  path: string[];
}

/** Build a tree from a flat list of categories. */
export function buildCategoryTree(categories: Category[]): CategoryNode[] {
  const map = new Map<number, CategoryNode>();

  // First pass: create nodes
  for (const cat of categories) {
    map.set(cat.id, { ...cat, children: [], depth: 0, path: [cat.name] });
  }

  const roots: CategoryNode[] = [];

  // Second pass: attach children
  for (const node of map.values()) {
    if (node.parentId == null) {
      roots.push(node);
    } else {
      const parent = map.get(node.parentId);
      if (parent) {
        node.depth = parent.depth + 1;
        node.path = [...parent.path, node.name];
        parent.children.push(node);
      } else {
        roots.push(node);
      }
    }
  }

  return roots;
}

/** Flatten a category tree back to an array (depth-first). */
export function flattenCategoryTree(nodes: CategoryNode[]): CategoryNode[] {
  const result: CategoryNode[] = [];
  for (const node of nodes) {
    result.push(node);
    result.push(...flattenCategoryTree(node.children));
  }
  return result;
}

/** Find a category node by id in a tree. */
export function findCategoryNode(nodes: CategoryNode[], id: number): CategoryNode | undefined {
  for (const node of nodes) {
    if (node.id === id) return node;
    const found = findCategoryNode(node.children, id);
    if (found) return found;
  }
  return undefined;
}

/** Return all ancestor nodes for a given category id. */
export function ancestorNodes(nodes: CategoryNode[], id: number): CategoryNode[] {
  const target = findCategoryNode(nodes, id);
  if (!target || target.depth === 0) return [];
  const flattened = flattenCategoryTree(nodes);
  return flattened.filter((n) => target.path.includes(n.name) && n.id !== id);
}

/** Return the breadcrumb path for a category as a string. */
export function categoryBreadcrumb(
  categories: Category[],
  categoryId: number,
  separator = ' > '
): string {
  const tree = buildCategoryTree(categories);
  const node = findCategoryNode(tree, categoryId);
  if (!node) return '';
  return node.path.join(separator);
}

/** Count the total products in a subtree (including children). */
export function subtreeProductCount(node: CategoryNode): number {
  return node.productCount + node.children.reduce((sum, c) => sum + subtreeProductCount(c), 0);
}

/** Return categories that have no children (leaf nodes). */
export function leafCategories(categories: Category[]): CategoryNode[] {
  const tree = buildCategoryTree(categories);
  return flattenCategoryTree(tree).filter((n) => n.children.length === 0);
}
