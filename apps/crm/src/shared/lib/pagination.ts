// Pagination helpers for CRM list views.

export interface PaginationState {
  page: number;
  pageSize: number;
  total: number;
}
export interface PaginationInfo extends PaginationState {
  totalPages: number;
  offset: number;
  hasPrev: boolean;
  hasNext: boolean;
  startIndex: number;
  endIndex: number;
}

export function getPaginationInfo(state: PaginationState): PaginationInfo {
  const { page, pageSize, total } = state;
  const totalPages = pageSize > 0 ? Math.ceil(total / pageSize) : 0;
  const clampedPage = Math.max(1, Math.min(page, totalPages || 1));
  const offset = (clampedPage - 1) * pageSize;
  return {
    page: clampedPage,
    pageSize,
    total,
    totalPages,
    offset,
    hasPrev: clampedPage > 1,
    hasNext: clampedPage < totalPages,
    startIndex: total === 0 ? 0 : offset + 1,
    endIndex: Math.min(offset + pageSize, total),
  };
}

export function clampPage(page: number, totalPages: number): number {
  return Math.max(1, Math.min(page, Math.max(1, totalPages)));
}

export function getPageNumbers(currentPage: number, totalPages: number, window = 2): number[] {
  if (totalPages <= 1) return totalPages === 1 ? [1] : [];
  const pages = new Set<number>([1, totalPages]);
  for (let i = currentPage - window; i <= currentPage + window; i++) {
    if (i >= 1 && i <= totalPages) pages.add(i);
  }
  const sorted = Array.from(pages).sort((a, b) => a - b);
  const result: number[] = [];
  for (let i = 0; i < sorted.length; i++) {
    result.push(sorted[i]);
    if (i < sorted.length - 1 && sorted[i + 1] - sorted[i] > 1) result.push(-1);
  }
  return result;
}
