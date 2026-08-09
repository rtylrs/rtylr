export function allOf<T>(...predicates: ((item: T) => boolean)[]): (item: T) => boolean {
  return (item) => predicates.every((p) => p(item));
}

export function anyOf<T>(...predicates: ((item: T) => boolean)[]): (item: T) => boolean {
  return (item) => predicates.some((p) => p(item));
}

export function not<T>(predicate: (item: T) => boolean): (item: T) => boolean {
  return (item) => !predicate(item);
}

export function filterBySubstring<T>(
  items: T[],
  getValue: (item: T) => string,
  query: string
): T[] {
  const q = query.trim().toLowerCase();
  if (!q) return items;
  return items.filter((item) => getValue(item).toLowerCase().includes(q));
}

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

export function filterByExact<T, V>(items: T[], getValue: (item: T) => V, value: V): T[] {
  return items.filter((item) => getValue(item) === value);
}

export function filterByOneOf<T, V>(items: T[], getValue: (item: T) => V, allowed: V[]): T[] {
  return items.filter((item) => allowed.includes(getValue(item)));
}

export function paginateItems<T>(items: T[], page: number, pageSize: number): T[] {
  const start = (page - 1) * pageSize;
  return items.slice(start, start + pageSize);
}

export function totalPages(totalItems: number, pageSize: number): number {
  if (pageSize <= 0) return 0;
  return Math.ceil(totalItems / pageSize);
}

export function filterByDateRange<T>(
  items: T[],
  getDate: (item: T) => string,
  from: string,
  to: string
): T[] {
  return items.filter((item) => {
    const d = getDate(item);
    return d >= from && d <= to;
  });
}

export function filterNonNull<T>(items: (T | null | undefined)[]): T[] {
  return items.filter((item): item is T => item != null);
}
