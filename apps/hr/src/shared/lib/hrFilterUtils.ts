export type FilterOperator = 'eq' | 'neq' | 'gt' | 'lt' | 'gte' | 'lte' | 'contains';

export interface FilterRule<T> {
  key: keyof T;
  operator: FilterOperator;
  value: unknown;
}

export function applyFilter<T>(items: T[], rule: FilterRule<T>): T[] {
  return items.filter((item) => {
    const fieldVal = item[rule.key];
    switch (rule.operator) {
      case 'eq':
        return fieldVal === rule.value;
      case 'neq':
        return fieldVal !== rule.value;
      case 'gt':
        return Number(fieldVal) > Number(rule.value);
      case 'lt':
        return Number(fieldVal) < Number(rule.value);
      case 'gte':
        return Number(fieldVal) >= Number(rule.value);
      case 'lte':
        return Number(fieldVal) <= Number(rule.value);
      case 'contains':
        return String(fieldVal).toLowerCase().includes(String(rule.value).toLowerCase());
      default:
        return true;
    }
  });
}

export function applyFilters<T>(items: T[], rules: FilterRule<T>[]): T[] {
  return rules.reduce((acc, rule) => applyFilter(acc, rule), items);
}

export function searchByFields<T>(items: T[], query: string, fields: (keyof T)[]): T[] {
  if (!query.trim()) return items;
  const q = query.toLowerCase();
  return items.filter((item) =>
    fields.some((field) =>
      String(item[field] ?? '')
        .toLowerCase()
        .includes(q)
    )
  );
}

export function paginateItems<T>(items: T[], page: number, pageSize: number): T[] {
  const start = (page - 1) * pageSize;
  return items.slice(start, start + pageSize);
}

export function totalPages(totalItems: number, pageSize: number): number {
  if (pageSize <= 0) return 0;
  return Math.ceil(totalItems / pageSize);
}
