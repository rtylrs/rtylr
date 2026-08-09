import { useState, useMemo } from 'react';
import { paginate, paginationResult, totalPages, clampPage } from '../utils/paginationUtils';
import type { PaginationResult } from '../utils/paginationUtils';

interface UsePaginationReturn<T> extends PaginationResult<T> {
  goTo: (page: number) => void;
  next: () => void;
  prev: () => void;
  setPageSize: (size: number) => void;
}

/**
 * Manage client-side pagination of an item list.
 * Returns the current page of items and navigation helpers.
 */
export function usePagination<T>(items: T[], initialPageSize = 10): UsePaginationReturn<T> {
  const [page, setPage] = useState(1);
  const [pageSize, setPageSizeState] = useState(initialPageSize);

  const pages = useMemo(() => totalPages(items.length, pageSize), [items.length, pageSize]);

  const currentItems = useMemo(() => paginate(items, page, pageSize), [items, page, pageSize]);

  const result = useMemo(
    () => paginationResult(currentItems, items, page, pageSize),
    [currentItems, items, page, pageSize]
  );

  const goTo = (target: number) => setPage(clampPage(target, pages));
  const next = () => goTo(page + 1);
  const prev = () => goTo(page - 1);

  const setPageSize = (size: number) => {
    setPageSizeState(size);
    setPage(1);
  };

  return { ...result, goTo, next, prev, setPageSize };
}
