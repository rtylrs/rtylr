/**
 * Filter state utility helpers for the Dash app.
 * All functions are pure with no side-effects.
 */

export type FilterOperator = 'eq' | 'neq' | 'gt' | 'gte' | 'lt' | 'lte' | 'contains' | 'in';

export interface FilterRule<T = unknown> {
  field: string;
  operator: FilterOperator;
  value: T;
}

export interface ActiveFilters {
  [field: string]: string | number | boolean | string[] | null;
}

/** Return true when an ActiveFilters object has no active filter values. */
export function hasNoFilters(filters: ActiveFilters): boolean {
  return Object.values(filters).every((v) => v == null || (Array.isArray(v) && v.length === 0));
}

/** Return the number of active (non-null, non-empty) filter keys. */
export function activeFilterCount(filters: ActiveFilters): number {
  return Object.values(filters).filter((v) => v != null && !(Array.isArray(v) && v.length === 0))
    .length;
}

/** Clear a single filter key by setting it to null. */
export function clearFilter(filters: ActiveFilters, field: string): ActiveFilters {
  return { ...filters, [field]: null };
}

/** Clear all filters by returning an object with all keys set to null. */
export function clearAllFilters(filters: ActiveFilters): ActiveFilters {
  const cleared: ActiveFilters = {};
  for (const key of Object.keys(filters)) {
    cleared[key] = null;
  }
  return cleared;
}

/** Merge new filter values into the current filter state. */
export function applyFilters(
  current: ActiveFilters,
  updates: Partial<ActiveFilters>
): ActiveFilters {
  return { ...current, ...updates };
}

/** Apply a single FilterRule to a record, returning true when the record matches. */
export function matchesRule<T extends Record<string, unknown>>(
  record: T,
  rule: FilterRule
): boolean {
  const val = record[rule.field];
  switch (rule.operator) {
    case 'eq':
      return val === rule.value;
    case 'neq':
      return val !== rule.value;
    case 'gt':
      return typeof val === 'number' && val > (rule.value as number);
    case 'gte':
      return typeof val === 'number' && val >= (rule.value as number);
    case 'lt':
      return typeof val === 'number' && val < (rule.value as number);
    case 'lte':
      return typeof val === 'number' && val <= (rule.value as number);
    case 'contains':
      return (
        typeof val === 'string' && val.toLowerCase().includes(String(rule.value).toLowerCase())
      );
    case 'in':
      return Array.isArray(rule.value) && rule.value.includes(val);
    default:
      return false;
  }
}

/** Filter an array of records by a list of FilterRules (all rules must match). */
export function applyRules<T extends Record<string, unknown>>(
  records: T[],
  rules: FilterRule[]
): T[] {
  if (rules.length === 0) return records;
  return records.filter((r) => rules.every((rule) => matchesRule(r, rule)));
}
