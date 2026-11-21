import { useCallback, useEffect, useMemo, useRef, useState } from 'react';
import { useNavigate, useSearch } from '@tanstack/react-router';
import {
  EMPTY_TASK_FILTERS,
  hasActiveTaskFilters,
  parseTaskFilters,
  serializeTaskFilters,
  type TaskFilters,
} from './taskFilters';

const storageKey = (scope: string): string => `flow:filters:${scope}`;
const DIM_KEY = 'flow:filters:dim';

/**
 * Filters last for the visit, not forever. They are kept in sessionStorage, so
 * moving between a project's board, list and calendar keeps them, but closing
 * the site and coming back opens every board unfiltered. Remembering them in
 * localStorage made a board reopen days later already narrowed to "Urgent",
 * which read as a default nobody set.
 */
function load(scope: string): TaskFilters {
  try {
    const raw = sessionStorage.getItem(storageKey(scope));
    return raw ? parseTaskFilters(JSON.parse(raw)) : EMPTY_TASK_FILTERS;
  } catch {
    // Private mode, quota, or a corrupt entry: start unfiltered.
    return EMPTY_TASK_FILTERS;
  }
}

function save(scope: string, filters: TaskFilters): void {
  try {
    const value = serializeTaskFilters(filters);
    if (value) sessionStorage.setItem(storageKey(scope), JSON.stringify(value));
    else sessionStorage.removeItem(storageKey(scope));
  } catch {
    // Not being remembered is fine; the URL still carries the filter.
  }
}

/**
 * Filters used to be saved in localStorage; drop those on mount so they never
 * resurface. The dim preference stays: it is a setting, not a filter.
 */
function clearLegacyFilters(): void {
  try {
    for (let i = localStorage.length - 1; i >= 0; i -= 1) {
      const key = localStorage.key(i);
      if (key && key.startsWith('flow:filters:') && key !== DIM_KEY) localStorage.removeItem(key);
    }
  } catch {
    // Storage unavailable: nothing to clear.
  }
}

function loadDim(): boolean {
  try {
    return localStorage.getItem(DIM_KEY) === '1';
  } catch {
    return false;
  }
}

export interface TaskFilterState {
  filters: TaskFilters;
  setFilters: (next: TaskFilters | ((current: TaskFilters) => TaskFilters)) => void;
  clear: () => void;
  active: boolean;
  /** Fade non-matching tasks instead of hiding them. One preference for every list. */
  dim: boolean;
  setDim: (dim: boolean) => void;
}

/**
 * Filters for one list, kept in two places:
 *
 * - the URL (`?filter=`), so a filtered board is a link someone else can open;
 * - sessionStorage per `scope`, so switching between a project's board, list
 *   and calendar keeps the filter for the rest of the visit.
 *
 * The URL wins. When it has no filter the stored one is used straight away
 * (no unfiltered flash) and then written into the URL, so the address bar is
 * always the truth and copying it shares exactly what is on screen.
 *
 * `scope` is e.g. `project:42` for every view of one project, or `my-tasks`.
 */
export function useTaskFilters(scope: string): TaskFilterState {
  const navigate = useNavigate();
  const search = useSearch({ strict: false }) as { filter?: unknown };
  const fromUrl = search.filter;

  // Bumped on every write, so clearing the URL filter falls back to the
  // just-saved (empty) entry instead of the one read when the page opened.
  const [version, setVersion] = useState(0);
  // eslint-disable-next-line react-hooks/exhaustive-deps
  const stored = useMemo(() => load(scope), [scope, version]);
  const filters = useMemo(
    () => (fromUrl === undefined ? stored : parseTaskFilters(fromUrl)),
    [fromUrl, stored]
  );

  const filtersRef = useRef(filters);
  filtersRef.current = filters;

  const writeUrl = useCallback(
    (next: TaskFilters) => {
      void navigate({
        to: '.',
        replace: true,
        search: (previous: Record<string, unknown>) => ({
          ...previous,
          filter: serializeTaskFilters(next),
        }),
      });
    },
    [navigate]
  );

  // Once per scope: a URL filter is remembered, a remembered filter goes into
  // the URL. Keyed on scope alone so typing in the search box does not re-run it.
  useEffect(() => {
    if (fromUrl !== undefined) save(scope, parseTaskFilters(fromUrl));
    else if (hasActiveTaskFilters(stored)) writeUrl(stored);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [scope]);

  const setFilters = useCallback(
    (next: TaskFilters | ((current: TaskFilters) => TaskFilters)) => {
      const value = typeof next === 'function' ? next(filtersRef.current) : next;
      filtersRef.current = value;
      save(scope, value);
      setVersion((v) => v + 1);
      writeUrl(value);
    },
    [scope, writeUrl]
  );

  const clear = useCallback(() => setFilters(EMPTY_TASK_FILTERS), [setFilters]);

  useEffect(clearLegacyFilters, []);

  const [dim, setDimState] = useState(loadDim);
  const setDim = useCallback((value: boolean) => {
    setDimState(value);
    try {
      localStorage.setItem(DIM_KEY, value ? '1' : '0');
    } catch {
      // A preference that does not stick is not worth an error.
    }
  }, []);

  return { filters, setFilters, clear, active: hasActiveTaskFilters(filters), dim, setDim };
}
