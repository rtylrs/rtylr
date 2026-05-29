// Filter map helpers for the insights dashboard filters.

export type FilterMap = Record<string, string>;

/**
 * addFilter returns a new FilterMap with the given key-value pair added.
 */
export function addFilter(filters: FilterMap, key: string, value: string): FilterMap {
  return { ...filters, [key]: value };
}

/**
 * removeFilter returns a new FilterMap without the given key.
 */
export function removeFilter(filters: FilterMap, key: string): FilterMap {
  const result = { ...filters };
  delete result[key];
  return result;
}

/**
 * clearFilters returns an empty FilterMap.
 */
export function clearFilters(): FilterMap {
  return {};
}

/**
 * hasActiveFilters returns true when any filter key has a non-empty value.
 */
export function hasActiveFilters(filters: FilterMap): boolean {
  return Object.values(filters).some((v) => v !== '');
}

/**
 * activeFilterCount returns the number of filter keys with non-empty values.
 */
export function activeFilterCount(filters: FilterMap): number {
  return Object.values(filters).filter((v) => v !== '').length;
}

/**
 * applyFilters returns only rows whose string values include all active filter values.
 * Comparison is case-insensitive.
 */
export function applyFilters<T extends Record<string, unknown>>(
  rows: T[],
  filters: FilterMap
): T[] {
  const active = Object.entries(filters).filter(([, v]) => v !== '');
  if (active.length === 0) return rows;
  return rows.filter((row) =>
    active.every(([key, value]) => {
      const cell = String(row[key] ?? '').toLowerCase();
      return cell.includes(value.toLowerCase());
    })
  );
}
