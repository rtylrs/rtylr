export type SortDirection = 'asc' | 'desc';

export function sortByString<T>(items: T[], key: keyof T, direction: SortDirection = 'asc'): T[] {
  return [...items].sort((a, b) => {
    const aVal = String(a[key] ?? '');
    const bVal = String(b[key] ?? '');
    const cmp = aVal.localeCompare(bVal);
    return direction === 'asc' ? cmp : -cmp;
  });
}

export function sortByNumber<T>(items: T[], key: keyof T, direction: SortDirection = 'asc'): T[] {
  return [...items].sort((a, b) => {
    const aVal = Number(a[key] ?? 0);
    const bVal = Number(b[key] ?? 0);
    const cmp = aVal - bVal;
    return direction === 'asc' ? cmp : -cmp;
  });
}

export function sortByDate<T>(items: T[], key: keyof T, direction: SortDirection = 'asc'): T[] {
  return [...items].sort((a, b) => {
    const aTime = new Date(String(a[key] ?? '')).getTime();
    const bTime = new Date(String(b[key] ?? '')).getTime();
    const cmp = aTime - bTime;
    return direction === 'asc' ? cmp : -cmp;
  });
}

export function sortByBooleanDesc<T>(items: T[], key: keyof T): T[] {
  return [...items].sort((a, b) => {
    const aVal = a[key] ? 1 : 0;
    const bVal = b[key] ? 1 : 0;
    return bVal - aVal;
  });
}

export function stableSort<T>(items: T[], compareFn: (a: T, b: T) => number): T[] {
  return items
    .map((item, index) => ({ item, index }))
    .sort((a, b) => compareFn(a.item, b.item) || a.index - b.index)
    .map(({ item }) => item);
}
