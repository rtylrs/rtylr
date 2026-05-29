/**
 * ERP filter state management and generic predicate helpers.
 * All functions are pure with no side-effects.
 */

/** ERP product filter state for list views. */
export interface FilterState {
  search: string;
  status: string | undefined;
  category: string | undefined;
  brand: string | undefined;
}

/** Return true if any filter field is set. */
export function hasActiveFilters(filters: FilterState): boolean {
  return (
    filters.search.length > 0 ||
    filters.status !== undefined ||
    filters.category !== undefined ||
    filters.brand !== undefined
  );
}

/** Count the number of active filter fields. */
export function activeFilterCount(filters: FilterState): number {
  let count = 0;
  if (filters.search.length > 0) count++;
  if (filters.status !== undefined) count++;
  if (filters.category !== undefined) count++;
  if (filters.brand !== undefined) count++;
  return count;
}

/** Return a new empty filter state. */
export function clearFilters(): FilterState {
  return { search: '', status: undefined, category: undefined, brand: undefined };
}

/** Return a new filter state with one field updated. */
export function updateFilter<K extends keyof FilterState>(
  filters: FilterState,
  key: K,
  value: FilterState[K]
): FilterState {
  return { ...filters, [key]: value };
}

/** Return true if two filter states are deeply equal. */
export function filtersEqual(a: FilterState, b: FilterState): boolean {
  return (
    a.search === b.search &&
    a.status === b.status &&
    a.category === b.category &&
    a.brand === b.brand
  );
}

/** Serialize a filter state to a query-param-compatible object, omitting empty fields. */
export function filtersToParams(filters: FilterState): Record<string, string> {
  const params: Record<string, string> = {};
  if (filters.search) params.search = filters.search;
  if (filters.status !== undefined) params.status = filters.status;
  if (filters.category !== undefined) params.category = filters.category;
  if (filters.brand !== undefined) params.brand = filters.brand;
  return params;
}

/** Generic predicate combinator - AND logic. */
export function allOf<T>(...predicates: Array<(item: T) => boolean>): (item: T) => boolean {
  return (item) => predicates.every((fn) => fn(item));
}

/** Generic predicate combinator - OR logic. */
export function anyOf<T>(...predicates: Array<(item: T) => boolean>): (item: T) => boolean {
  return (item) => predicates.some((fn) => fn(item));
}

/** Negates a predicate. */
export function not<T>(predicate: (item: T) => boolean): (item: T) => boolean {
  return (item) => !predicate(item);
}

/** Filters items matching a substring (case-insensitive) on a string field. */
export function filterBySubstring<T>(
  items: T[],
  getValue: (item: T) => string,
  query: string
): T[] {
  const q = query.trim().toLowerCase();
  if (!q) return items;
  return items.filter((item) => getValue(item).toLowerCase().includes(q));
}

/** Filters items where a numeric field falls within [min, max]. */
export function filterByRange<T>(
  items: T[],
  getValue: (item: T) => number,
  min: number,
  max: number
): T[] {
  return items.filter((item) => {
    const v = getValue(item);
    return v >= min && v <= max;
  });
}

/** Filters items by an exact field value. */
export function filterByExact<T, V>(items: T[], getValue: (item: T) => V, value: V): T[] {
  return items.filter((item) => getValue(item) === value);
}

/** Filters items where a field value is one of the allowed values. */
export function filterByOneOf<T, V>(items: T[], getValue: (item: T) => V, allowed: V[]): T[] {
  return items.filter((item) => allowed.includes(getValue(item)));
}

/** Paginates an array to the requested page (1-based) and page size. */
export function paginateItems<T>(items: T[], page: number, pageSize: number): T[] {
  const start = (page - 1) * pageSize;
  return items.slice(start, start + pageSize);
}
