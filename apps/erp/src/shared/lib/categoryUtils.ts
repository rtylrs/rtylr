/**
 * Category helper functions used across the product and category features.
 */
import type { Category } from '../services/erpApi';

/** Return true when a category is a root (top-level) category. */
export function isRootCategory(category: Category): boolean {
  return category.parentId == null;
}

/** Return all root (top-level) categories from a flat list. */
export function getRootCategories(categories: Category[]): Category[] {
  return categories.filter(isRootCategory);
}

/** Return all child categories of a given parent category. */
export function getChildCategories(categories: Category[], parentId: number): Category[] {
  return categories.filter((c) => c.parentId === parentId);
}

/** Return true when a category has child categories. */
export function hasChildren(category: Category, categories: Category[]): boolean {
  return categories.some((c) => c.parentId === category.id);
}

/** Build a flat path string for a category, e.g. "Electronics > Phones > Accessories". */
export function buildCategoryPath(category: Category, allCategories: Category[]): string {
  const parts: string[] = [category.name];
  let current = category;
  const visited = new Set<number>();
  while (current.parentId != null && !visited.has(current.id)) {
    visited.add(current.id);
    const parent = allCategories.find((c) => c.id === current.parentId);
    if (!parent) break;
    parts.unshift(parent.name);
    current = parent;
  }
  return parts.join(' > ');
}

/** Sort categories alphabetically by name. */
export function sortCategoriesByName(categories: Category[]): Category[] {
  return [...categories].sort((a, b) => a.name.localeCompare(b.name));
}

/** Sort categories by product count, highest first. */
export function sortCategoriesByProductCount(categories: Category[]): Category[] {
  return [...categories].sort((a, b) => b.productCount - a.productCount);
}

/** Return the total product count across all categories. */
export function totalProductCount(categories: Category[]): number {
  return categories.reduce((sum, c) => sum + c.productCount, 0);
}
