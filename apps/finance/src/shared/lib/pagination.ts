/**
 * Pure pagination utility helpers for finance list views.
 */

export interface PaginationMeta {
  page: number;
  pageSize: number;
  totalCount: number;
  totalPages: number;
  hasNext: boolean;
  hasPrevious: boolean;
}

/**
 * Computes full pagination metadata from raw values.
 */
export function getPaginationMeta(
  page: number,
  pageSize: number,
  totalCount: number
): PaginationMeta {
  const totalPages = pageSize > 0 ? Math.ceil(totalCount / pageSize) : 0;
  return {
    page,
    pageSize,
    totalCount,
    totalPages,
    hasNext: page < totalPages,
    hasPrevious: page > 1,
  };
}

/**
 * Returns true if there is a previous page.
 */
export function hasPreviousPage(page: number): boolean {
  return page > 1;
}

/**
 * Returns true if there is a next page.
 */
export function hasNextPage(page: number, totalPages: number): boolean {
  return page < totalPages;
}

/**
 * Clamps a page number to the valid range [1, totalPages].
 */
export function clampPage(page: number, totalPages: number): number {
  if (totalPages <= 0) return 1;
  return Math.min(Math.max(1, page), totalPages);
}

/**
 * Returns the start index (0-based) of the current page in a list.
 */
export function getPageStartIndex(page: number, pageSize: number): number {
  return (page - 1) * pageSize;
}

/**
 * Returns the end index (0-based, exclusive) of the current page.
 */
export function getPageEndIndex(page: number, pageSize: number, totalCount: number): number {
  return Math.min(page * pageSize, totalCount);
}
