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
